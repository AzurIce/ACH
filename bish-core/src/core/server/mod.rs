mod backup;

use std::{
    fmt::Display,
    fs::File,
    io::{self, BufRead, Read, Write},
    path::PathBuf,
    process::{Command, Stdio},
    sync::{mpsc, Arc, Mutex},
    thread,
};

use derivative::Derivative;
use regex::Regex;

use crate::{
    config::ServerConfig,
    fabric::init_server_jar,
    global_output_tx,
    utils::{path::split_parent_and_file, regex::player_regex},
};

#[cfg(test)]
mod test {
    use crate::config::ServerConfig;

    use super::Server;

    #[test]
    fn t() {
        let s: [i32; 1] = [1];
        let b = s[0..].to_vec();
        let c = s[1..].to_vec();
        let d = s[2..].to_vec();
        println!("{:?}", b);
        println!("{:?}", c);
        println!("{:?}", d);
    }

    #[test]
    fn test_snapshop() {
        let server = Server::new(
            "1.20.2".to_string(),
            ServerConfig {
                dir: "/Users/azurice/Game/MCServer/1.20.2".to_string(),
                jvm_options: String::new(),
                version: "1.20.2".to_string(),
                ..Default::default()
            },
        );

        println!("{:?}", server.get_snapshot_list());
        server.make_snapshot();
        println!("{:?}", server.get_snapshot_list());
        server.del_snapshot();
        println!("{:?}", server.get_snapshot_list());
    }

    #[test]
    fn test_set_property() {
        let server = Server::new(
            "1.20.2".to_string(),
            ServerConfig {
                dir: "/Users/azurice/Game/MCServer/1.20.2".to_string(),
                jvm_options: String::new(),
                version: "1.20.2".to_string(),
                ..Default::default()
            },
        );
        server.set_property("difficulty", "hard");
        server.set_property("server-ip", "0.0.0.0");
    }
}

#[derive(Derivative, Clone)]
#[derivative(Default)]
pub struct Server {
    name: String,
    config: ServerConfig,
    input_tx: Option<mpsc::Sender<String>>,
    command_tx: Option<mpsc::Sender<String>>,
}

pub fn run(server: Arc<Mutex<Server>>) {
    let cloned_server = server.clone();
    // 命令处理线程
    let (command_tx, command_rx) = mpsc::channel::<String>();
    let _server = cloned_server.clone();
    thread::spawn(move || {
        let server = _server;
        while let Ok(command_str) = command_rx.recv() {
            let command_str = command_str.replace("\r\n", "\n");
            let command_str = command_str.strip_prefix('#').unwrap();
            let command_str = command_str.strip_suffix('\n').unwrap_or(command_str);

            let split: Vec<&str> = command_str.split(' ').collect();
            let command = split[0];
            let args = &split[1..];

            let mut server = server.lock().unwrap();
            println!("[server/{}]: command: {} {:?}", server.name, command, args);
            match command {
                "bksnap" => {
                    if args.is_empty() || args[0] == "list" {
                        let snapshot_list = server.get_snapshot_list();
                        server.say("snapshots: ");
                        for (i, snapshot) in snapshot_list.into_iter().enumerate() {
                            server.say(format!("{i}: {snapshot:?}"))
                        }
                    } else if args[0] == "make" {
                        while server.get_snapshot_list().len() >= 10 {
                            server.del_snapshot()
                        }
                        server.say("saving snapshot...");
                        server.make_snapshot();
                        server.say("saved snapshot")
                    } else if args.len() == 2 && args[0] == "load" {
                        // TODO: load snap backup
                    }
                }
                "bkarch" => {
                    if args.is_empty() || args[0] == "list" {
                        println!("bkarch list, not implemented yet")
                        // TODO: show arch backup
                    } else if args[0] == "make" {
                        println!("bkarch make, not implemented yet")
                        // let comment = args[1..].join(" ");
                        // TODO: make arch backup
                    } else if args.len() == 2 && args[0] == "load" {
                        println!("bkarch load, not implemented yet")
                        // TODO: load arch backup
                    }
                }
                _ => {
                    println!("unknown command")
                }
            }
        }
        println!("exit1")
    });

    let mut server = server.lock().unwrap();
    println!("[server/{}]: starting server...", server.name);
    let mut child = server
        .command()
        .stdin(Stdio::piped())
        .stdout(Stdio::piped())
        .spawn()
        .expect("failed to spawn");

    let child_in = child.stdin.take().expect("Failed to open child's stdin");
    let child_out = child.stdout.take().expect("Failed to open child's stdout");

    // 统一输入处理线程
    // 一切从终端、服务端输出识别到的玩家输入，都会通过 input_tx 输入到 channel 中
    // 然后在此统一处理，识别命令，作相关处理
    let (input_tx, input_rx) = mpsc::channel::<String>();
    server.input_tx = Some(input_tx);

    let _command_tx = command_tx.clone();
    let _server = cloned_server.clone();
    thread::spawn(move || {
        // let server = _server;
        let command_tx = _command_tx;
        let mut writer = io::BufWriter::new(child_in);
        while let Ok(input) = input_rx.recv() {
            if input.starts_with('#') {
                command_tx
                    .send(input)
                    .expect("failed to send to command_tx");
            } else {
                writer.write_all(input.as_bytes()).expect("failed to write");
                writer.flush().expect("failed to flush");
            }
        }
        // println!("exit2")
    });

    // 服务端 输出处理线程
    // 通过 global_tx 发送给主线程统一处理
    let global_output_tx = global_output_tx();
    let _command_tx = command_tx.clone();
    let _server = cloned_server.clone();
    thread::spawn(move || {
        let command_tx = _command_tx;
        let server = _server;

        let mut reader = io::BufReader::new(child_out);
        let mut buf = String::new();
        while let Ok(size) = reader.read_line(&mut buf) {
            if size == 0 {
                break;
            }
            if let Some(cap) = player_regex().captures(&buf) {
                let _player_name = cap.get(1).unwrap().as_str();
                let content = cap.get(2).unwrap().as_str();
                if content.starts_with('#') {
                    command_tx
                        .send(content.to_string())
                        .expect("failed to send to command_tx");
                }
            }
            global_output_tx
                .send(buf.clone())
                .expect("Failed to send to global_tx");
            // println!("{buf}");
        }
        println!("server end");
        child.wait().expect("failed to wait");
        // println!("exit3");

        // Drop 掉 command_tx 和 input_tx 使得上面的两个线程可以退出循环
        let mut server = server.lock().unwrap();
        server.command_tx = None;
        server.input_tx = None;
    });
}

impl Server {
    pub fn command(&self) -> Command {
        // init server.properties
        for (key, value) in &self.config.properties {
            self.set_property(key, value);
        }
        // init jar
        let (dir, filename) = split_parent_and_file(
            init_server_jar(&self.config.dir, &self.config.version)
                .expect("failed to init server jar"),
        );

        let mut command = Command::new("java");
        let mut args = vec!["-jar", &filename, "--nogui"];
        args.extend(self.config.jvm_options.split(' ').collect::<Vec<&str>>());

        println!("[Server/new]: command's dir and jar_file is {dir} and {filename}");
        command.current_dir(dir);
        command.args(args);
        command
    }

    pub fn new(name: String, config: ServerConfig) -> Self {
        Self {
            name,
            config,
            ..Default::default()
        }
    }

    pub fn set_property<S: AsRef<str>>(&self, key: S, value: S) {
        let key = key.as_ref();
        let value = value.as_ref();

        let mut buf = String::new();
        let property_file = PathBuf::from(&self.config.dir).join("server.properties");

        {
            let mut property_file =
                File::open(&property_file).expect("failed to open server.properties");

            property_file
                .read_to_string(&mut buf)
                .expect("failed to read properties");
        }

        let regex = Regex::new(format!(r"{}=([^#\n\r]+)", key).as_str()).unwrap();
        let res = regex.replace(&buf, format!("{}={}", key, value));
        // println!("{}", res);

        let mut property_file =
            File::create(&property_file).expect("failed to open server.properties");
        property_file
            .write_all(res.as_bytes())
            .expect("failed to write server.properties");
    }

    // pub fn load_snapshot(&self, id: usize) {
    // TODO: load snapshot
    // self.writeln("stop")
    // }

    pub fn writeln(&mut self, line: &str) {
        if let Some(tx) = &self.input_tx {
            let line = if line.ends_with('\n') {
                line.to_string()
            } else {
                format!("{line}\n")
            };

            tx.send(line).expect("failed to send to server's input_tx");
        }
    }

    pub fn say<S: AsRef<str> + Display>(&mut self, content: S) {
        self.writeln(&format!("say {content}"))
    }
}
