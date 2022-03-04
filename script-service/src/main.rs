mod messages;
mod broker;
mod auth;
mod ws;

use std::{thread, time::Duration, sync::Arc, task::{Context, Poll}, borrow::Cow, collections::HashMap};
use broker::BrokerMessage;
use flume::{Sender, Receiver};
use futures::{StreamExt, SinkExt};
use jsonwebtoken::DecodingKey;
use log::{info, error};
use messages::Internal;
use net2::{unix::UnixTcpBuilderExt, TcpBuilder};
use tokio::{net::{TcpListener, TcpStream}, time::Instant};
use tokio_tungstenite::{accept_async, tungstenite::{Message, protocol::CloseFrame, handshake::server::{Request, Response}}};
use ws::{WsConn, WsData, WsConnState};

use crate::{messages::{StandardErrors, Action}, broker::Broker};

static mut HEARTBEAT_INTERVAL: Duration = Duration::from_millis(4000);

fn main() {
    // console_subscriber::init();

    env_logger::init();
    utils_rs::utils::load_env().expect("Failed to load env");

    let config = utils_rs::utils::parse_config::<messages::Config>().expect("Failed to parse config");

    let listen_addr = config.clone().listen_addr;

    let decoded_jwt = base64::decode(config.jwt_private_key).expect("Failed to decode private key");
    let rsa_decode_key = DecodingKey::from_rsa_pem(&decoded_jwt).expect("Failed to parse private key");

    unsafe {
        auth::init_statics_auth(rsa_decode_key, config.internal_services_url);
        HEARTBEAT_INTERVAL = Duration::from_millis(config.heartbeat_interval.into());
    }

    let (tx, rx) = flume::unbounded::<BrokerMessage>();

    let mut threads = Vec::new();
    let tx_ = tx.clone();
    let rx_ = rx.clone();
    let addr = listen_addr.clone();

    threads.push(thread::spawn(move || {
        info!("Starting server!");
        let rt = tokio::runtime::Builder::new_multi_thread().enable_all().build().expect("Failed to build runtime");

        let server = async move {
            let listener = {
                let builder = TcpBuilder::new_v4().expect("Failed to create tcp builder");
                builder.reuse_address(true).expect("Failed to reuse address");
                builder.reuse_port(true).expect("Failed to reuse port");
                builder.bind(addr).expect("Failed to bind");
                builder.listen(10240).expect("Failed to listen")
            };

            let listener = TcpListener::from_std(listener).expect("Failed to convert to tcp listener");

            loop {
                let tx = tx_.clone();
                let rx = rx_.clone();
                let sock = listener.accept().await;
                if sock.is_ok() {
                    let sock = sock.unwrap();
                    info!("connection received from {}", sock.1);
                    tokio::spawn(process_request(sock.0, tx, rx));
                }
                else {
                    info!("Connection failed")
                }
            }
        };

        rt.block_on(server);
    }));

    threads.push(thread::spawn(move || {
        let rt = tokio::runtime::Builder::new_multi_thread().enable_all().build().expect("Failed to build runtime");
        rt.block_on(Broker::broker_system(tx, rx));
    }));

    info!("Listening on {}", listen_addr);
    for thread in threads {
        thread.join().unwrap();
    }
}

async fn process_request(stream: TcpStream, broker_tx: Sender<BrokerMessage>, broker_rx: Receiver<BrokerMessage>) {
    let mut ws_stream = accept_async(stream).await.expect("Failed to accept");

    let (tx, rx) = flume::bounded::<Internal>(32);
    let mut state = WsConn::new(tx, broker_tx);
    
    let mut exit = false;
    let mut interval = tokio::time::interval_at(Instant::now() + unsafe { HEARTBEAT_INTERVAL }, unsafe { HEARTBEAT_INTERVAL });
    while !exit {

        tokio::select! {
            result = ws_stream.next() => {
                if result.is_some() {
                    let result: Result<Message, tokio_tungstenite::tungstenite::Error> = result.unwrap();
                    if !result.is_ok() {
                        // TODO: process error
                        continue;
                    }

                    let res = state.process_message(result.unwrap()).await;
                }
                else {
                    exit = true;
                }
            },
            result = rx.recv_async() => {
                if !result.is_ok() {
                    // TODO: process error
                    continue;
                }

                match result.unwrap() {
                    Internal::Error(_) => todo!(),
                    Internal::Ping => {
                        info!("ping sent!");
                        ws_stream.send(Message::Ping(vec![])).await;
                    },
                    Internal::NewState(action) => {
                        match state.get_state() {
                            WsConnState::Authenticated => {},
                            WsConnState::Ready => todo!(),
                            WsConnState::Close => {
                                let get_close_frame = || -> Option<CloseFrame> {
                                    if action.is_some() {
                                        let action = action.unwrap();
                                        if action.message.len() > 0 {
                                            Some(CloseFrame {
                                                code: tokio_tungstenite::tungstenite::protocol::frame::coding::CloseCode::Error,
                                                reason: Cow::from(action.message)
                                            })
                                        } else {
                                            None
                                        }
                                    } else {
                                        None
                                    }
                                };
                                
                                info!("closing connection!");
                                ws_stream.close(get_close_frame()).await;
                            },
                            WsConnState::Disconnected => todo!(),
                            _ => {}
                        }
                    },
                }
            },
            _ = interval.tick() => {
                state.interval().await;
            },
        }
    }
}