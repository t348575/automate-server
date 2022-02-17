use actix::{fut, ActorContext, WrapFuture, ContextFutureSpawner, ActorFuture, ResponseActFuture};
use actix_web::client::Client;
use jsonwebtoken::{DecodingKey, decode, Validation, Algorithm};
use log::error;
use serde::{Serialize, Deserialize};
use crate::messages::{Disconnect, Connect, WsMessage, WsData, Auth, CanCreateRoom, CreateRoom};
use crate::room::Session;
use crate::room_manager::RoomManager;
use actix::{Actor, Addr, Running, StreamHandler};
use actix::{AsyncContext, Handler};
use actix_web_actors::ws::{self, CloseReason};
use actix_web_actors::ws::Message::Text;
use std::time::{Duration, Instant, SystemTime};

static mut HEARTBEAT_INTERVAL: Duration = Duration::from_millis(4000);
static mut CLIENT_TIMEOUT: Duration = Duration::from_millis(30000);
static mut JWT_PRIVATE_KEY: DecodingKey = DecodingKey::from_secret("".as_bytes());
static mut GENERAL_SERVICES_URL: String = String::from("http://localhost:3002");

pub struct WsConn {
    user_id: i64,
    script_id: String,
    state: WsConnState,
    hb: Instant,
    hb_count: u8,
    manager: Addr<RoomManager>,
}

pub enum WsConnState {
    Connected,
    Authenticated,
    Ready,
    Close,
    Disconnected,
}

#[derive(Serialize, Deserialize)]
struct Claims {
    exp: usize,
    iat: usize,
    nbf: usize,
    sub: String,
    user: i64,
    scope: String,
}

pub unsafe fn init_statics(jpk: DecodingKey, hbi: Duration, ct: Duration) {
    JWT_PRIVATE_KEY = jpk;
    HEARTBEAT_INTERVAL = hbi;
    CLIENT_TIMEOUT = ct;
}

impl WsConn {
    pub fn new(script_id: String, manager: Addr<RoomManager>) -> WsConn {
        WsConn {
            user_id: 0,
            script_id,
            state: WsConnState::Connected,
            hb: Instant::now(),
            hb_count: 0,
            manager,
        }
    }
}

impl Actor for WsConn {
    type Context = ws::WebsocketContext<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        self.hb(ctx);

        let addr = ctx.address();
        self.lobby_addr
            .send(Connect {
                addr: addr.recipient(),
                lobby_id: self.room,
                self_id: self.id,
            })
            .into_actor(self)
            .then(|res, _, ctx| {
                match res {
                    Ok(_res) => (),
                    _ => ctx.stop(),
                }
                fut::ready(())
            })
            .wait(ctx);
    }

    fn stopping(&mut self, _: &mut Self::Context) -> Running {
        self.lobby_addr.do_send(Disconnect { id: self.id, room_id: self.room });
        Running::Stop
    }
}

impl WsConn {
    fn hb(&self, ctx: &mut ws::WebsocketContext<Self>) {
        ctx.run_later(Duration::from_millis(6000), |act, ctx| {
            match self.state {
                WsConnState::Connected => {
                    if self.hb_count > 0 {
                        ctx.stop();
                        return;
                    }
                },
                _ => {},
            }
        });

        let client_timeout = unsafe {
            CLIENT_TIMEOUT.clone()
        };
        ctx.run_interval(unsafe {
            HEARTBEAT_INTERVAL.clone()
        }, |act, ctx| {
            if Instant::now().duration_since(act.hb) > client_timeout {
                println!("Disconnecting failed heartbeat");
                act.lobby_addr.do_send(Disconnect { id: act.id, room_id: act.room });
                ctx.stop();
                return;
            }

            match self.state {
                WsConnState::Close | WsConnState::Disconnected => {
                    ctx.stop();
                    return;
                }
                _ => {},
            }

            ctx.ping(b"hi");
        });
    }

    fn Authenticate(&mut self, auth: &Auth, ctx: &mut ws::WebsocketContext<WsConn>) {
        let token = decode::<Claims>(&auth.token, &unsafe {
            JWT_PRIVATE_KEY.clone()
        }, &Validation::new(Algorithm::RS256));

        let authCopy = *auth;
        
        match token {
            Ok(token) => {
                if token.claims.sub != "access" {
                    ctx.stop();
                    return;
                }

                self.user_id = token.claims.user;
                self.state = WsConnState::Authenticated;

                ctx.address().send(CanCreateRoom {
                    user_id: self.user_id,
                    script_id: self.script_id,
                }).into_actor(self).then(|res, conn, ctx| {
                    if res.is_err() {
                        ctx.address().send(WsMessage::error(authCopy.message_id, format!("An unknown error occurred {}", res.err().unwrap().to_string())));
                    }

                    let res = res.unwrap();
                    if res.is_err() {
                        ctx.address().send(WsMessage::error(authCopy.message_id, res.err().unwrap().to_string()));
                    }
                    if res.unwrap() {
                        conn.manager.send(CreateRoom {
                            user_id: conn.user_id,
                            script_id: conn.script_id,
                            session: Session {
                                token: authCopy.token,
                                token_expiry: token.claims.exp,
                                session: ctx.address().recipient(),
                            }
                        }).into_actor(conn).map(|res, _, ctx| match res {
                            Ok(res) => {
                                println!("asd");
                            },
                            Err(err) => {
                                ctx.stop();
                            },
                        });
                    }

                    fut::ready(())
                }).wait(ctx);
            },
            Err(err) => {
                ctx.stop();
                return;
            }
        }
    }
}

impl StreamHandler<Result<ws::Message, ws::ProtocolError>> for WsConn {
    fn handle(&mut self, msg: Result<ws::Message, ws::ProtocolError>, ctx: &mut Self::Context) {
        match msg {
            Ok(ws::Message::Ping(msg)) => {
                self.hb = Instant::now();
                ctx.pong(&msg);
            }
            Ok(ws::Message::Pong(_)) => {
                self.hb = Instant::now();
            }
            Ok(ws::Message::Binary(bin)) => ctx.binary(bin),
            Ok(ws::Message::Close(reason)) => {
                ctx.close(reason);
                ctx.stop();
            }
            Ok(ws::Message::Continuation(_)) => {
                ctx.stop();
            }
            Ok(ws::Message::Nop) => (),
            Ok(Text(s)) => {
                match serde_json::from_str::<WsData>(&s) {
                    Ok(data) => {
                        match data {
                            WsData::Auth(auth) => self.Authenticate(&auth, ctx),
                        }
                    },
                    Err(err) => error!("{}", err),
                }
            },
            Err(e) => error!("{}", e),
        }
    }
}

impl Handler<WsMessage> for WsConn {
    type Result = ();

    fn handle(&mut self, msg: WsMessage, ctx: &mut Self::Context) {
        ctx.text(msg.message);
        
        if msg.close {
            self.state = WsConnState::Close;
            ctx.close(None);
            ctx.stop();
        }
    }
}

impl Handler<CanCreateRoom> for WsConn {
    type Result = ResponseActFuture<Self, Result<bool, String>>;

    fn handle(&mut self, msg: CanCreateRoom, ctx: &mut Self::Context) -> Self::Result {

        #[derive(Deserialize)]
        struct ScriptResponse {
            id: i64
        }

        #[derive(Deserialize)]
        struct ScriptError {
            error: String,
        }

        Box::pin(
            async {
                let client = Client::builder().timeout(Duration::from_millis(5000)).header("User-Agent", "script-service").finish();
                let res = client.get(unsafe { GENERAL_SERVICES_URL } + "/script/internal/" + &self.script_id.to_string()).send().await;
                if res.is_err() {
                    return Err(res.err().unwrap().to_string());
                }

                let res = res.unwrap();
                if res.status() != 200 {
                    let res = res.json::<ScriptError>().await;
                    if res.is_err() {
                        return Err(res.err().unwrap().to_string());
                    }

                    return Err(res.unwrap().error);
                }

                let res = res.json::<ScriptResponse>().await;
                if res.is_err() {
                    return Err(res.err().unwrap().to_string());
                }

                let res = res.unwrap();
                self.script_id = res.id.to_string();
                if res.id.to_string().len() > 0 {
                    return Ok(false);
                }

                return Ok(true);
            }.into_actor(self).map(|res, _, ctx| match res {
                Ok(res) => Ok(res),
                Err(e) => Err(e),
            })
        )
    }
}