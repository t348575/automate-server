use serde::Deserialize;

#[derive(Deserialize, Debug, Clone)]
pub struct Config {
    pub listen_addr: String,
    pub heartbeat_interval: u32,
    pub client_timeout: u32,
    pub jwt_private_key: String,
    pub redis_conn: String,
}