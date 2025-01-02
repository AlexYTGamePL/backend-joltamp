use std::sync::Arc;
use axum::extract::State;
use axum::http::StatusCode;
use axum::Json;
use scylla::{QueryResult, Session};
use scylla::transport::errors::QueryError;
use scylla::transport::query_result::IntoRowsResultError;
use serde::{Deserialize, Serialize};
use serde::de::{Error, StdError};
use uuid::Uuid;
use crate::security::passwords::hash_password;

#[derive(Deserialize)]
pub struct RequestUser {
    email: String,
    password: String,
    username: String,
}
#[derive(Serialize)]
pub struct ReturnUser {
    jwt: String,
    user_id: String,
}
#[derive(Serialize)]
pub struct RegisterError{
    message: String,
}

#[derive(Serialize)]
#[serde(untagged)]
pub enum ReturnType {
    ReturnUser(ReturnUser),
    Error(RegisterError),
}

pub async fn register(
    State(session): State<Arc<Session>>,
    Json(mut payload): Json<RequestUser>,
) -> (StatusCode, Json<ReturnType>) {
    if payload.email.is_empty() || payload.password.is_empty() || payload.username.is_empty(){
        return (StatusCode::BAD_REQUEST, Json(ReturnType::Error(RegisterError { message: String::from("Not every field satisfied") })));
    }
    if !payload.email.contains('@'){
        return (StatusCode::BAD_REQUEST, Json(ReturnType::Error(RegisterError { message: String::from("Invalid e-mail address")})));
    }
    if payload.password.len() < 3 || payload.username.len() < 3{
        return (StatusCode::BAD_REQUEST, Json(ReturnType::Error(RegisterError { message: String::from("Password or Username is too short (<4)")})));
    }

    if let Ok(used) = check_username_free(&session, &payload.username).await {
        if used {
            return (StatusCode::BAD_REQUEST, Json(ReturnType::Error(RegisterError { message: String::from("Username already used")})));
        }
    }else{
        return (StatusCode::INTERNAL_SERVER_ERROR, Json(ReturnType::Error(RegisterError { message: String::from("register#0x01 Internal server error")})));
    }
    if let Ok(user) = insert_user(&session, &mut payload).await{
        return (StatusCode::CREATED, Json(ReturnType::ReturnUser(ReturnUser{
            jwt: user.0.to_string(),
            user_id: user.1.to_string(),
        })));
    }else{
        return (StatusCode::INTERNAL_SERVER_ERROR, Json(ReturnType::Error(RegisterError { message: String::from("register#0x02 Internal server error")})));
    }
}

async fn check_username_free(session: &Arc<Session>, username: &String) -> Result<bool, Box<dyn StdError>> {
    let result = session.query_unpaged("SELECT user_id FROM joltamp.users WHERE username = ? ALLOW FILTERING", (username, )).await?.into_rows_result()?;

    return Ok(result.rows_num() != 0);
}

async fn insert_user(session: &Arc<Session>, mut payload: &mut RequestUser) -> Result<(Uuid, Uuid), Box<dyn StdError>> {
    let gen_jwt = Uuid::new_v4();
    let gen_user_id = Uuid::new_v4();
    hash_password(&mut payload.password).expect("Hashing error, disabling for safety.");
    session.query_unpaged("INSERT INTO joltamp.users (createdat, user_id, username, displayname, email, password, isadmin, jwt, status) VALUES (todate(now()), ?, ?, ?, ?, ?, false, ?, 0)",
                                     (gen_user_id, &payload.username, &payload.username, &payload.email, &payload.password, gen_jwt)
    ).await?;
    Ok((gen_jwt, gen_user_id))
}