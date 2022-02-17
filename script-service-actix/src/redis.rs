use std::time::{SystemTime, UNIX_EPOCH};
use actix::{Context, Actor, Handler};
use redis::{Commands, RedisError};
use crate::messages::InsertRoom;

pub struct Redis {
    client: redis::Client,
    conn: redis::Connection
}

impl Redis {
    pub fn new(redis_conn: String) -> Self {
        let client = redis::Client::open(redis_conn).expect("Could not create redis client!");
        let conn = client.get_connection().expect("Could not connect to redis!");

        Redis {
            client,
            conn
        }
    }
}

impl Actor for Redis {
    type Context = Context<Self>;
}

impl Handler<InsertRoom> for Redis {
    type Result = Result<Option<String>, RedisError>;

    fn handle(&mut self, msg: InsertRoom, ctx: &mut Self::Context) -> Self::Result {
        self.conn.hset_multiple(msg.script_id, &[
            ("script_id".to_string(), msg.script_id),
            ("user:".to_string() + &msg.user_id.to_string(), SystemTime::now().duration_since(UNIX_EPOCH).unwrap().as_millis().to_string())
        ])
    }
}