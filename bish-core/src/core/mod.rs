pub mod server;

use std::{
    collections::HashMap,
    sync::{Arc, Mutex},
};

use crate::{config::BishConfig, core::server::Server};

use self::server::run;

#[derive(Default)]
pub struct Core {
    pub config: BishConfig,
    pub running_servers: Mutex<HashMap<String, Arc<Mutex<Server>>>>,
}

impl Core {
    pub fn new(config: BishConfig) -> Self {
        Self {
            config: config.clone(),
            ..Default::default()
        }
    }

    pub fn run_server(&self, name: String) {
        if let Some(config) = self.config.servers.get(&name) {
            let server = Server::new(name.clone(), config.clone());
            let server = Arc::new(Mutex::new(server));

            self.running_servers
                .lock()
                .unwrap()
                .insert(name, server.clone());

            run(server);
        }
    }

    pub fn stop_server(&self, name: String) {
        let mut running_servers = self.running_servers.lock().unwrap();

        if let Some(server) = running_servers.get(&name) {
            server.lock().unwrap().writeln("stop");
            running_servers.remove(&name);
        }
    }
}
