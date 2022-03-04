use std::{time::Duration, collections::HashMap};
use jsonwebtoken::{DecodingKey, decode, Validation, Algorithm};
use once_cell::sync::OnceCell;
use serde::{Serialize, Deserialize};

use crate::messages::StandardErrors;

static mut JWT_PRIVATE_KEY: OnceCell<DecodingKey> = OnceCell::new();
static mut INTERNAL_SERVICES_URL: OnceCell<String> = OnceCell::new();

#[derive(Serialize, Deserialize)]
#[serde(rename_all = "camelCase")]
pub struct Auth {
    pub message_id: i64,
    pub token: String,
    pub script_id: i64
}

#[derive(Serialize, Deserialize)]
struct Claims {
    exp: usize,
    iat: usize,
    nbf: usize,
    sub: String,
    user: i64,
    scope: String,
}

pub unsafe fn init_statics_auth(jpk: DecodingKey, isu: String) {
    let res = JWT_PRIVATE_KEY.set(jpk);
    if res.is_err() {
        panic!("Failed to set JWT private key");
    }

    let res = INTERNAL_SERVICES_URL.set(isu);
    if res.is_err() {
        panic!("Failed to set General services url");
    }
}

impl Auth {
    pub async fn authenticate(&self) -> Result<i64, StandardErrors> {
        let token = decode::<Claims>(&self.token, unsafe { JWT_PRIVATE_KEY.get().unwrap() }, &Validation::new(Algorithm::RS256));

        if token.is_err() {
            return Err(StandardErrors::Internal("JWT Auth error: ".to_string() + &token.err().unwrap().to_string()));
        }

        let claims = token.unwrap().claims;
        let data = HashMap::from([
            ("user_d", claims.user),
            ("script_id", self.script_id),
        ]);

        let client = reqwest::ClientBuilder::new().timeout(Duration::from_millis(5000)).user_agent("script-service").build().unwrap();
        let req = client.post(unsafe { INTERNAL_SERVICES_URL.get().unwrap() }.to_owned() + "/scripts/stream/").json(&data).send().await;

        if !req.is_ok() {
            return Err(StandardErrors::FatalError("internal_error".to_string()));
        }

        let res = req.unwrap();
        let status = res.status();
        let data = res.text().await;

        if !data.is_ok() {
            return  Err(StandardErrors::FatalError("internal_error".to_string()));
        }

        if status != 200 {
            return  Err(StandardErrors::FatalError(data.unwrap()));
        }

        Ok(claims.user)
    }
}