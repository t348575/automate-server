mod config;
mod ws;
mod room_manager;
mod room;
mod messages;
mod start_connection;
mod redis;
use std::time::Duration;

use jsonwebtoken::DecodingKey;
use room_manager::RoomManager;
use start_connection::start_connection as start_connection_route;
use actix::{Actor};
use actix_web::{App, HttpServer};
use ws::init_statics;

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    utils_rs::utils::load_env().expect("Failed to load env");

    let config = utils_rs::utils::parse_config::<config::Config>().expect("Failed to parse config");

    let redis = redis::Redis::new(config.redis_conn).start();
    let listen_addr = config.clone().listen_addr;
    let room_manager = RoomManager::new(redis).start();

    let rsa_decode_key = DecodingKey::from_rsa_pem(config.jwt_private_key.as_bytes()).expect("Failed to parse private key");

    unsafe {
        init_statics(rsa_decode_key, Duration::from_millis(config.heartbeat_interval.into()), Duration::from_millis(config.client_timeout.into()));
    }


    HttpServer::new(move || {
        App::new()
            .service(start_connection_route)
            .data(room_manager)
            .data(config.clone())
    })
    .bind(listen_addr)?
    .run()
    .await
}