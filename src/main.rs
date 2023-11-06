mod config;
mod fabric;
mod server;
mod utils;

use std::{
    collections::HashMap,
    error::Error,
    io::stdin,
    sync::{Mutex, OnceLock},
};

use server::Server;
use utils::regex::forward_regex;

pub fn servers() -> &'static Mutex<HashMap<String, Server>> {
    static SERVERS: OnceLock<Mutex<HashMap<String, Server>>> = OnceLock::new();
    SERVERS.get_or_init(|| Mutex::new(HashMap::new()))
}

fn main() -> Result<(), Box<dyn Error>> {
    let config = config::load_config()?;

    let mut servers = servers().lock().unwrap();
    for (server_name, server_config) in config.servers {
        let server = Server::new(server_config);
        servers.insert(server_name, server);
    }

    for (_server_name, server) in servers.iter_mut() {
        server.run()
    }

    let forward_regex = forward_regex();

    // 主线程
    // 从终端接受输入，识别转发正则，转发到对应服务器的 input_tx
    // 或全部转发
    let mut buf = String::new();
    while let Ok(_size) = stdin().read_line(&mut buf) {
        // 正则捕获目标服务器，转发至对应服务器。
        // 或全部转发
        if let Some(cap) = forward_regex.captures(&buf) {
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
