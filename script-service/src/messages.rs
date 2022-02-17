use pulsar::{SerializeMessage, Error as PulsarError, producer, DeserializeMessage, Payload};
use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Deserialize, Debug, Clone)]
pub struct Config {
    pub listen_addr: String,
    pub heartbeat_interval: u32,
    pub client_timeout: u32,
    pub jwt_private_key: String,
    pub redis_conn: String,
    pub general_services_url: String,
    pub pulsar_conn: String,
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

impl SerializeMessage for TestData {
    fn serialize_message(input: Self) -> Result<producer::Message, PulsarError> {
        let payload = serde_json::to_vec(&input).map_err(|e| PulsarError::Custom(e.to_string()))?;
        Ok(producer::Message {
            payload,
            ..Default::default()
        })
    }
}

impl DeserializeMessage for TestData {
    type Output = Result<TestData, serde_json::Error>;

    fn deserialize_message(payload: &Payload) -> Self::Output {
        serde_json::from_slice(&payload.data)
    }
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