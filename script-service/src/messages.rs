use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Deserialize, Debug, Clone)]
pub struct Config {
    pub listen_addr: String,
    pub heartbeat_interval: u32,
    pub client_timeout: u32,
    pub jwt_private_key: String,
    pub service_discovery_url: String,
    pub general_services_url: String,
    pub node_name: String
}

#[derive(Error, Debug)]
pub enum StandardErrors {
    #[error("Fatal error: {}", .0)]
    FatalError(String),

    #[error("Internal error: {}", .0)]
    Internal(String),
}

pub struct Action {
    pub close: bool,
    pub message: String
}


pub enum Internal {
    Error(StandardErrors),
    Ping,
    NewState(Option<Action>),
}

#[derive(Serialize, Deserialize)]
pub struct TestData {
    pub data: String,
}

impl Action {
    pub fn new_close(message: String) -> Option<Self> {
        Self::new(message, false)
    }

    pub fn new(message: String, close: bool) -> Option<Self> {
        Some(
            Action {
                close,
                message
            }
        )
    }

    pub fn empty() -> Option<Action> {
        None
    }
}