use crate::{ws::WsConn, room_manager::RoomManager};
use actix::Addr;
use actix_web::{get, web::Data, web::Path, web::Payload, Error, HttpResponse, HttpRequest};
use actix_web_actors::ws;

#[get("/{script_id}")]
pub async fn start_connection(
    req: HttpRequest,
    stream: Payload,
    Path(script_id): Path<String>,
    srv: Data<Addr<RoomManager>>
) -> Result<HttpResponse, Error> {
    let ws = WsConn::new(
        script_id,
        srv.get_ref().clone()
    );


    let resp = ws::start(ws, &req, stream)?;
    Ok(resp)
}