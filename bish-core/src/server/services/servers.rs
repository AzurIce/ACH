use std::sync::Arc;

use axum::{
    extract::{Path, State},
    Json,
};
use reqwest::StatusCode;
use serde::Serialize;
use serde_json::json;

use crate::{config::ServerConfig, core::Core};

#[derive(Serialize)]
struct ServerInfo {
    pub name: String,
    pub config: ServerConfig,
    pub running: bool,
}

pub async fn get_servers(State(state): State<Arc<Core>>) -> (StatusCode, Json<serde_json::Value>) {
    let mut servers = vec![];
    let running_servers = state.running_servers.lock().unwrap();

    for (name, config) in &state.config.servers {
        servers.push(ServerInfo {
            name: name.clone(),
            config: config.clone(),
            running: running_servers.get(name).is_some(),
        })
    }

    (
        StatusCode::OK,
        Json(json!({
            "servers": servers
        })),
    )
}

pub async fn start_server(
    State(state): State<Arc<Core>>,
    Path(name): Path<String>,
) -> (StatusCode, Json<serde_json::Value>) {
    if state.running_servers.lock().unwrap().get(&name).is_none() {
        tokio::task::block_in_place(|| {
            state.run_server(name);
        });
        (
            StatusCode::OK,
            Json(json!({
                "msg": "success"
            })),
        )
    } else {
        (
            StatusCode::BAD_REQUEST,
            Json(json!({
                "msg": "The server is already running"
            })),
        )
    }
}

pub async fn stop_server(
    State(state): State<Arc<Core>>,
    Path(name): Path<String>,
) -> (StatusCode, Json<serde_json::Value>) {
    if state.running_servers.lock().unwrap().get(&name).is_some() {
        state.stop_server(name);
        // let mut server = server.lock().unwrap();
        // server.writeln("stop");
        (
            StatusCode::OK,
            Json(json!({
                "msg": "success"
            })),
        )
    } else {
        (
            StatusCode::BAD_REQUEST,
            Json(json!({
                "msg": "The server is not running"
            })),
        )
    }
}
