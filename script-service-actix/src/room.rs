use std::{collections::HashMap, time::SystemTime};

use actix::{Recipient, Actor, Context};

use crate::messages::WsMessage;

pub struct Room {
    users: HashMap<i64, Session>,
    script_id: String,
}

pub struct Session {
    session: Recipient<WsMessage>,
    token: String,
    token_expiry: usize,
}

impl Default for Room {
    fn default() -> Self {
        Room {
            users: HashMap::new(),
            script_id: "".to_string(),
        }
    }
}

impl Room {
    pub fn new(user: (i64, Session), script_id: String) -> Self {
        let room = Self::default();

        room.users.insert(user.0, user.1);
        room.script_id = script_id;

        room
    }
}

impl Actor for Room {
    type Context = Context<Self>;
}

