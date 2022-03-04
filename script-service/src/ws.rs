use flume::Sender;
use log::{info};
use serde::{Serialize, Deserialize};
use tokio::{time::Instant, net::TcpStream};
use tokio_tungstenite::tungstenite::Message;

use crate::{auth::Auth, messages::{Internal, Action, StandardErrors}, broker::BrokerMessage};

type Stream = tokio_tungstenite::WebSocketStream<TcpStream>;

#[derive(Serialize, Deserialize)]
#[serde(untagged)]
#[serde(rename_all = "camelCase")]
pub enum WsData {
    Auth(Auth),
    Data(String),
}

pub struct WsConn {
    user_id: i64,
    script_id: String,
    state: WsConnState,
    hb: Instant,
    hb_count: u8,
    tx: Sender<Internal>,
    broker_tx: Sender<BrokerMessage>
}

#[derive(Clone)]
pub enum WsConnState {
    Connected,
    Authenticated,
    Ready,
    Close,
    Disconnected,
}

impl WsConn {
    pub fn new(tx: Sender<Internal>, broker_tx: Sender<BrokerMessage>) -> Self {
        WsConn {
            user_id: 0,
            script_id: String::new(),
            state: WsConnState::Connected,
            hb: Instant::now(),
            hb_count: 0,
            tx,
            broker_tx
        }
    }

    pub async fn process_message(&mut self, msg: Message) {
        match msg {
            Message::Text(msg) => {
                let res: Result<WsData, serde_json::Error> = serde_json::from_str(&msg);
                if !res.is_ok() {
                    return;
                }

                self.parse_text(res.unwrap()).await;

                self.tx.send_async(Internal::NewState(Action::empty())).await;
            },
            Message::Pong(buf) => {
                info!("pong received");
                self.hb_count += 1;
            },
            Message::Close(_) => {},
            _ => {},
        }
    }

    pub async fn parse_text(&mut self, res: WsData) {
        match res {
            WsData::Auth(auth) => {
                let res = auth.authenticate().await;
                if !res.is_ok() {
                    self.state = WsConnState::Close;
                    let reason = {
                        match res.err().unwrap() {
                            StandardErrors::FatalError(err) => err,
                            _ => "unknown_error".to_string()
                        }
                    };

                    self.tx.send_async(Internal::NewState(Action::new_close(reason))).await;
                    return;
                }

                self.script_id = auth.script_id.to_string();
                self.user_id = res.unwrap();
                self.state = WsConnState::Authenticated;
                self.tx.send_async(Internal::NewState(Action::empty())).await;
            },
            WsData::Data(data) => {
                info!("Sending data to producer");
                // self.broker_tx.send_async(data).await;
            },
        }
    }

    pub async fn interval(&mut self) {
        self.tx.send(Internal::Ping);
        match self.state {
            WsConnState::Connected => {
                if self.hb_count > 0 {
                    self.state = WsConnState::Close;
                    info!("requesting connection close!");
                    self.tx.send_async(Internal::NewState(Action::new_close("timeout_no_auth".to_string()))).await;
                }
            },
            _ => {}
        }
    }

    pub fn get_state(&self) -> WsConnState {
        self.state.clone()
    }
}