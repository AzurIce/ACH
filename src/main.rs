mod config;
mod fabric;
mod server;
mod utils;
mod backup;

use std::{
    collections::HashMap,
    error::Error,
    io::stdin,
    sync::{
        mpsc::{self, Sender},
        Mutex, OnceLock,
    },
    thread,
};

use server::Server;
use utils::regex::forward_regex;

static GLOBAL_OUTPUT_TX: OnceLock<Sender<String>> = OnceLock::new();

pub fn servers() -> &'static Mutex<HashMap<String, Server>> {
    static SERVERS: OnceLock<Mutex<HashMap<String, Server>>> = OnceLock::new();
    SERVERS.get_or_init(|| Mutex::new(HashMap::new()))
}

pub fn global_output_tx() -> Sender<String> {
    GLOBAL_OUTPUT_TX.get().unwrap().clone()
}

fn main() -> Result<(), Box<dyn Error>> {
    let config = config::load_config()?;

    let mut servers = servers().lock().unwrap();
    for (server_name, server_config) in config.servers {
        let server = Server::new(server_config);
        servers.insert(server_name, server);
    }

    let (tx, rx) = mpsc::channel::<String>();
    GLOBAL_OUTPUT_TX.get_or_init(|| tx);
    // 输出处理线程
    // 统一处理来自所有 server 的输出
    thread::spawn(move || {
        while let Ok(buf) = rx.recv() {
            println!("{buf}")
        }
    });

    for (_server_name, server) in servers.iter_mut() {
        server.run()
    }

    // 主线程
    // 从终端接受输入，识别转发正则，转发到对应服务器的 input_tx
    // 或全部转发
    let mut buf = String::new();
    while let Ok(_size) = stdin().read_line(&mut buf) {
        // 正则捕获目标服务器，转发至对应服务器。
        // 或全部转发
        if let Some(cap) = forward_regex().captures(&buf) {
            let line = cap.get(1).unwrap().as_str();
            let server_name = cap.get(2).unwrap().as_str();
            if let Some(server) = servers.get_mut(server_name) {
                server.writeln(line)
            }
        } else {
            for (_server_name, server) in servers.iter_mut() {
                server.writeln(&buf)
            }
        }
        buf.clear();
    }

    Ok(())
}
