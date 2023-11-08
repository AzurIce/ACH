pub mod server;

use std::{collections::HashMap, sync::{Mutex, Arc}};

use crate::{config::BishConfig, core::server::Server};

#[derive(Default)]
pub struct Core {
    pub config: BishConfig,
    pub servers: HashMap<String, Arc<Mutex<Server>>>
}

impl Core {
    pub fn new(config: BishConfig) -> Self {
        let mut core = Self {
            config: config.clone(),
            ..Default::default()
        };

        for (server_name, server_config) in config.servers {
            let server = Server::new(server_name.clone(), server_config);
            core.servers.insert(server_name, Arc::new(Mutex::new(server)));
        }
        core
    }
}