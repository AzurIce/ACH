use std::sync::Arc;

use axum::{
    extract::{Path, State},
    Json,
};
use reqwest::StatusCode;
use serde_json::json;

use crate::core::Core;

pub async fn get_servers(State(state): State<Arc<Core>>) -> (StatusCode, Json<serde_json::Value>) {
    let server_list = state.servers.keys().cloned().collect::<Vec<String>>();

    (
        StatusCode::OK,
        Json(json!({
            "servers": server_list
        })),
    )
}

pub async fn start_server(State(state): State<Arc<Core>>, Path(name): Path<String>) -> StatusCode {
    if let Some(server) = state.servers.get(&name) {
        let _server = server.lock().unwrap();
        // TODO: Start server
        StatusCode::OK
    } else {
        StatusCode::BAD_REQUEST
    }
}

pub async fn stop_server(State(state): State<Arc<Core>>, Path(name): Path<String>) -> StatusCode {
    if let Some(server) = state.servers.get(&name) {
        let mut server = server.lock().unwrap();
        server.writeln("stop");
        // TODO: Stop server
        StatusCode::OK
    } else {
        StatusCode::BAD_REQUEST
    }
}
