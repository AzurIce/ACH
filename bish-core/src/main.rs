mod config;
mod core;
mod fabric;
mod server;
mod utils;

use core::{server::run, Core};
use std::{
    error::Error,
    io::stdin,
    sync::{
        mpsc::{self, Sender},
        Arc, OnceLock,
    },
    thread,
};

use axum::{
    routing::{get, post},
    Router,
};
use server::services::servers::{get_servers, start_server, stop_server};
use utils::regex::forward_regex;

static GLOBAL_OUTPUT_TX: OnceLock<Sender<String>> = OnceLock::new();

// pub fn servers() -> &'static Mutex<HashMap<String, Server>> {
//     static SERVERS: OnceLock<Mutex<HashMap<String, Server>>> = OnceLock::new();
//     SERVERS.get_or_init(|| Mutex::new(HashMap::new()))
// }

pub fn global_output_tx() -> Sender<String> {
    GLOBAL_OUTPUT_TX.get().unwrap().clone()
}

// #[tokio::main]
async fn async_main(app_state: Arc<Core>) {
    let app = Router::new()
        .route("/servers", get(get_servers))
        .route("/servers/:name/start", post(start_server))
        .route("/servers/:name/stop", post(stop_server));
    let app = app.with_state(app_state.clone());

    axum::Server::bind(&"0.0.0.0:3000".parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}

fn main() -> Result<(), Box<dyn Error>> {
    let config = config::load_config().expect("Failed to load config");
    let app_state = Arc::new(Core::new(config));

    let (tx, rx) = mpsc::channel::<String>();
    GLOBAL_OUTPUT_TX.get_or_init(|| tx);
    // 输出处理线程
    // 统一处理来自所有 server 的输出
    thread::spawn(move || {
        while let Ok(buf) = rx.recv() {
            println!("{buf}")
        }
    });

    // for (_server_name, server) in app_state.servers.iter() {
    //     run(server.clone())
    // }

    // 主线程
    // 从终端接受输入，识别转发正则，转发到对应服务器的 input_tx
    // 或全部转发
    let _app_state = app_state.clone();
    thread::spawn(move || {
        let mut buf = String::new();
        while let Ok(_size) = stdin().read_line(&mut buf) {
            // 正则捕获目标服务器，转发至对应服务器。
            // 或全部转发
            if let Some(cap) = forward_regex().captures(&buf) {
                let line = cap.get(1).unwrap().as_str();
                let server_name = cap.get(2).unwrap().as_str();
                if let Some(server) = _app_state.running_servers.lock().unwrap().get(server_name) {
                    let mut server = server.lock().unwrap();
                    server.writeln(line)
                }
            } else {
                for (_server_name, server) in _app_state.running_servers.lock().unwrap().iter() {
                    let mut server = server.lock().unwrap();
                    server.writeln(&buf)
                }
            }
            buf.clear();
        }
    });

    tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .unwrap()
        .block_on(async_main(app_state));
    Ok(())
}
