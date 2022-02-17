use std::collections::HashMap;
use actix::{Addr, Actor, Context, Handler, WrapFuture, ActorFuture, ResponseFuture};
use crate::{room::Room, messages::{InsertRoom, CreateRoom}, redis::Redis};

pub struct RoomManager {
    rooms: HashMap<String, Addr<Room>>,
    redis: Addr<Redis>
}

impl RoomManager {
    pub fn new(redis: Addr<Redis>) -> RoomManager {
        RoomManager {
            rooms: HashMap::new(),
            redis
        }
    }
}

impl Actor for RoomManager {
    type Context = Context<Self>;
}

impl Handler<CreateRoom> for RoomManager {
    type Result = ResponseFuture<Result<bool, String>>;

    fn handle(&mut self, msg: CreateRoom, ctx: &mut Self::Context) -> Self::Result {Box::pin(
            async {
                let room = Room::new(
                    (msg.user_id, msg.session),
                    msg.script_id.clone()
                );
        
                if self.rooms.contains_key(&msg.script_id) {
                    return Err("room already exists".to_string());
                }
        
                self.rooms.insert(msg.script_id.clone(), room.start());
                
                let res = self.redis.send(InsertRoom {
                    user_id: msg.user_id,
                    script_id: msg.script_id,
                }).await;

                if res.is_err() {
                    return Err(format!("could not insert room into redis: {}", res.err().unwrap()));
                }
        
                let res = res.unwrap();
                if res.is_err() {
                    return Err(format!("could not insert room into redis: {}", res.err().unwrap()));
                }

                Ok(true)
            }
        )
    }
}