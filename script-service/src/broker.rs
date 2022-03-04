use std::collections::HashSet;

use deadpool_redis::Pool;
use flume::{Receiver, Sender};
use tokio::time::{Instant, Duration};

pub struct Broker {
    pool: Pool
}

pub enum BrokerMessage {
    ToRedis,
    FromRedis,
    AddBroker(String),
    BrokerReady(String),
    Error(BrokerError),
    StreamReady(String),
}

pub enum BrokerError {
    CouldNotCreate
}

enum Internal {
    Error(InternalError),
}

enum InternalError {
    CouldNotCreatePool
}

impl Broker {
    async fn start(host: String, tx: Sender<Internal>, rx: Receiver<Internal>) {
        let cfg = deadpool_redis::Config::from_url(host);
        let pool = cfg.create_pool(Some(deadpool_redis::Runtime::Tokio1));
        if !pool.is_ok() {
            tx.send(Internal::Error(InternalError::CouldNotCreatePool));
        }

        loop {
            tokio::select! {
            }
        }
    }

    pub async fn broker_system(tx: Sender<BrokerMessage>, rx: Receiver<BrokerMessage>) {
        let mut brokers: HashSet<String> = HashSet::new();
        let mut interval = tokio::time::interval_at(Instant::now() + Duration::from_secs(1), Duration::from_secs(1));

        let (tx_int, rx_int) = flume::bounded::<Internal>(256);
        loop {
            tokio::select! {
                msg = rx.recv_async() => {
                    if !msg.is_ok() {
                        // TODO: handle error
                        continue;
                    }

                    let data = msg.unwrap();
                    match data {
                        BrokerMessage::ToRedis => todo!(),
                        BrokerMessage::FromRedis => todo!(),
                        BrokerMessage::BrokerReady(_) => todo!(),
                        BrokerMessage::StreamReady(_) => todo!(),
                        BrokerMessage::AddBroker(host) => {
                            if brokers.contains(&host) {
                                tx.send_async(BrokerMessage::BrokerReady(host)).await;
                                continue;
                            }

                            tokio::spawn(Self::start(host.clone(), tx_int.clone(), rx_int.clone()));
                            brokers.insert(host.clone());
                            tx.send_async(BrokerMessage::BrokerReady(host)).await;
                        },
                        BrokerMessage::Error(_) => {}
                    }
                }
            }
        }
    }
}


// let client = reqwest::ClientBuilder::new().timeout(Duration::from_millis(5000)).user_agent("script-service").build().unwrap();
// let req = client.get(service_discovery_url + "/redis-stream-server/" + &node_name).send().await;

// if !req.is_ok() {
//     panic!("Failed to connect to service discovery: {}", req.err().unwrap().to_string());
// }

// let res = req.unwrap();
// let status = res.status();
// let data = res.text().await;

// if !data.is_ok() {
//     panic!("Failed to get stream-server, status {}, data: {}", status, data.err().unwrap().to_string());
// }

// if status != 200 {
//     panic!("Failed to get stream-server, status {}, data: {}", status, data.unwrap());
// }

// let data: Result<RedisStreamServer, serde_json::Error> = serde_json::from_str(&data.unwrap().to_string());
// if !data.is_ok() {
//     panic!("Could not parse result: {}", data.err().unwrap().to_string());
// }

// let node = data.unwrap();
// let client = redis::Client::open(node.node.clone());
// if !client.is_ok() {
//     panic!("Could not connect to redis stream server: {}", client.err().unwrap().to_string());
// }