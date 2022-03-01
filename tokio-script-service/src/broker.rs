use redis::Client;
use serde::{Serialize, Deserialize};
use tokio::time::Duration;

pub struct Broker {
    redis_node_conn: String
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
struct RedisStreamServer {
    conn: String,
    client: Client
}

impl Broker {

    pub async fn new(service_discovery_url: String, node_name: String) -> Self {
        let client = reqwest::ClientBuilder::new().timeout(Duration::from_millis(5000)).user_agent("script-service").build().unwrap();
        let req = client.get(service_discovery_url + "/redis-stream-server/" + &node_name).send().await;

        if !req.is_ok() {
            panic!("Failed to connect to service discovery: {}", req.err().unwrap().to_string());
        }

        let res = req.unwrap();
        let status = res.status();
        let data = res.text().await;

        if !data.is_ok() {
            panic!("Failed to get stream-server, status {}, data: {}", status, data.err().unwrap().to_string());
        }

        if status != 200 {
            panic!("Failed to get stream-server, status {}, data: {}", status, data.unwrap());
        }

        let data: Result<RedisStreamServer, serde_json::Error> = serde_json::from_str(&data.unwrap().to_string());
        if !data.is_ok() {
            panic!("Could not parse result: {}", data.err().unwrap().to_string());
        }

        let client = redis::Client::open(data.unwrap().conn);
        if !client.is_ok() {
            panic!("Could not connect to redis stream server: {}", client.err().unwrap().to_string());
        }

        Broker {
            node_name
        }
    }
}