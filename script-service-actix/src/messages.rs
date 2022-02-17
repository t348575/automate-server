use actix::prelude::{Message, Recipient};
use serde::{Serialize, Deserialize};
use redis::RedisError;
use uuid::Uuid;

use crate::room::Session;

#[derive(Message)]
#[rtype(result = "()")]
pub struct WsMessage {
    pub message: String,
    pub close: bool,
}

impl WsMessage {
    pub fn new_close<T>(msg: T) -> Option<WsMessage> where T: Serialize {
        Self::new_inner(msg, true)
    }

    pub fn error(message_id: i64, error: String) -> WsMessage {
        Self::new_inner(BasicError {
            message_id,
            error
        }, false).unwrap()
    }

    pub fn new<T>(msg: T) -> Option<WsMessage> where T: Serialize {
        Self::new_inner(msg, false)
    }

    fn new_inner<T>(msg: T, close: bool) -> Option<WsMessage> where T: Serialize {
        let res = Self::to_json(msg);
        if res.is_err() {
            return None;
        }

        Some(
            WsMessage {
                message: res.unwrap(),
                close,
            }
        )
    }

    fn to_json<T>(msg: T) -> Result<String, serde_json::Error> where T: Serialize {
        serde_json::to_string(&msg)
    }
}

#[derive(Message)]
#[rtype(result = "()")]
pub struct Connect {
    pub addr: Recipient<WsMessage>,
    pub lobby_id: Uuid,
    pub self_id: Uuid,
}

#[derive(Message)]
#[rtype(result = "()")]
pub struct Disconnect {
    pub id: Uuid,
    pub room_id: Uuid,
}

#[derive(Message)]
#[rtype(result = "()")]
pub struct ClientActorMessage {
    pub id: Uuid,
    pub msg: String,
    pub room_id: Uuid
}

#[derive(Serialize, Deserialize)]
#[serde(untagged)]
#[serde(rename_all = "camelCase")]
pub enum WsData {
    Auth(Auth),
}

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Auth {
    pub message_id: i64,
    pub token: String,
}

#[derive(Message)]
#[rtype(result = "Result<bool, String>")]
pub struct CanCreateRoom {
    pub user_id: i64,
    pub script_id: String,
}

#[derive(Message)]
#[rtype(result = "Result<bool, String>")]
pub struct CreateRoom {
    pub user_id: i64,
    pub script_id: String,
    pub session: Session,
}

#[derive(Message)]
#[rtype(result = "Result<Option<String>, RedisError>")]
pub struct InsertRoom {
    pub user_id: i64,
    pub script_id: String,
}

#[derive(Serialize)]
pub struct BasicError {
    pub message_id: i64,
    pub error: String,
}