use std::{
    io::{self, BufRead,Write},
    process::{Command, Stdio},
    sync::mpsc,
    thread, fs::{self, create_dir_all, DirEntry}, path::{Path, PathBuf}, fmt::Display,
};

use derivative::Derivative;

use crate::{config::ServerConfig, fabric::init_server_jar, utils::{path::split_parent_and_file, regex::player_regex, fs::copy_dir, time::get_cur_time_str}, global_output_tx};

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
        let server = Server::new(ServerConfig {
            dir: "/Users/azurice/Game/MCServer/1.20.2".to_string(),
            jvm_options: String::new(),
            version: "1.20.2".to_string(),
        });

        println!("{:?}", server.get_snapshot_list());
        server.make_snapshot();
        println!("{:?}", server.get_snapshot_list());
        server.del_snapshot();
        println!("{:?}", server.get_snapshot_list());
    }
}

#[derive(Derivative, Clone)]
#[derivative(Default)]
pub struct Server {
    config: ServerConfig,
    input_tx: Option<mpsc::Sender<String>>,
    command_tx: Option<mpsc::Sender<String>>,
}

impl Server {
    pub fn command(&self) -> Command {
        let mut command = Command::new("java");
        let (dir, filename) = split_parent_and_file(
            init_server_jar(&self.config.dir, &self.config.version).expect("failed to init server jar"),
        );
        println!("[Server/new]: command's dir and jar_file is {dir} and {filename}");
        command.current_dir(dir);
        command.args(["-jar", &filename, "--nogui"]);
        command
    }

    pub fn new(config: ServerConfig) -> Self {
        Self {
            config,
            ..Default::default()
        }
    }

    // pub fn load_snapshot(&self, id: usize) {
        // TODO: load snapshot
        // self.writeln("stop")
    // }

    pub fn del_snapshot(&self) {
        println!("[del_snapshop]");
        let snapshot_dir = Path::new("./backups").join("snapshots");
        if let Err(err) = create_dir_all(&snapshot_dir) {
            println!("failed to create all dir: {err}");
            return
        }

        if let Ok(entries) = fs::read_dir(snapshot_dir) {
            let mut entries: Vec<DirEntry> = entries.into_iter().map(|entry|entry.unwrap()).collect();
            
            entries.sort_by_key(|entry| entry.metadata().unwrap().created().unwrap());
            let entries = entries.into_iter().map(|entry| entry.path()).collect::<Vec<PathBuf>>();
            if let Some(first) = entries.first() {
                println!("[del_snapshop]: Deleting {first:?}...");
                if let Err(err) = fs::remove_dir_all(first) {
                    println!("Failed to remove dir: {err}")
                }
                println!("[del_snapshop]: Snapshop deleted");
            }
        }
    }

    pub fn make_snapshot(&self) {
        let snapshot_dir = Path::new("./backups").join("snapshots");
        if let Err(err) = create_dir_all(&snapshot_dir) {
            println!("failed to create all dir: {err}");
            return
        }

        let src_path = Path::new(&self.config.dir).join("world");
        if !src_path.exists() {
            println!("skip world/, not exist");
            return
        }

        let time = get_cur_time_str();
        let backup_name = format!("{}", time);
        let dst_path = Path::new("./backups").join("snapshots").join(backup_name);
        if let Err(err) = copy_dir(&src_path, &dst_path) {
            println!("failed to copy: {err}")
        }
    }

    pub fn get_snapshot_list(&self) -> Vec<PathBuf>{
        let snapshot_dir = Path::new("./backups").join("snapshots");
        if let Err(err) = create_dir_all(&snapshot_dir) {
            println!("failed to create all dir: {err}");
            return Vec::new()
        }

        if let Ok(entries) = fs::read_dir(snapshot_dir) {
            let mut entries: Vec<DirEntry> = entries.into_iter().map(|entry|entry.unwrap()).collect();
            
            entries.sort_by_key(|entry| entry.metadata().unwrap().created().unwrap());
            let entries = entries.into_iter().map(|entry| entry.path()).collect::<Vec<PathBuf>>();
            // for entry in entries {
            //     if let Ok(entry) = entry {
            //         snapshot_list.push(Backup::new(entry))
            //     }
            // }
            entries
        } else {
            Vec::new()
        }
        // snapshot_list.sort_by_key(|snapshot|snapshot.metadata.created().unwrap());
        // snapshot_list
    }

    pub fn run(&mut self) {
        let mut child = self
            .command()
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .expect("failed to spawn");

        let child_in = child.stdin.take().expect("Failed to open child's stdin");
        let child_out = child.stdout.take().expect("Failed to open child's stdout");

        // 命令处理线程
        let (tx, rx) = mpsc::channel::<String>();
        self.command_tx = Some(tx);
        let mut server = self.clone();
        thread::spawn(move || {
            while let Ok(command_str) = rx.recv() {
                let command_str = command_str.strip_prefix('#').unwrap();
                let command_str = command_str.strip_suffix('\n').unwrap_or(command_str);

                let split: Vec<&str> = command_str.split(' ').collect();
                let command = split[0];
                let args = &split[1..];

                println!("command: {} {:?}", command, args);
                match command {
                    "bksnap" => {
                        if args.len() == 0 || args[0] == "list" {
                            let snapshot_list = server.get_snapshot_list();
                            server.say("snapshots: ");
                            for (i, snapshot) in snapshot_list.into_iter().enumerate() {
                                server.say(format!("{i}: {snapshot:?}"))
                            }
                        } else if args[0] == "make" {
                            while server.get_snapshot_list().len() >= 10 {
                                server.del_snapshot()
                            }
                            server.make_snapshot()
                        } else if args.len() == 2 && args[0] == "load" {
                            // TODO: load snap backup
                        }
                    }
                    "bkarch" => {
                        if args.len() < 1 || args[0] == "list" {
                            // TODO: show arch backup
                        } else if args[0] == "make" {
                            // let comment = args[1..].join(" ");
                            // TODO: make arch backup
                        } else if args.len() == 2 && args[0] == "load" {
                            // TODO: load arch backup
                        }
                    }
                    _ => {
                        println!("unknown command")
                    }
                }
            }
        });

        // 统一输入处理线程
        // 一切从终端、服务端输出识别到的玩家输入，都会通过 input_tx 输入到 channel 中
        // 然后在此统一处理，识别命令，作相关处理
        let (tx, rx) = mpsc::channel::<String>();
        self.input_tx = Some(tx);
        let command_tx = self.command_tx.clone().unwrap();
        thread::spawn(move || {
            let mut writer = io::BufWriter::new(child_in);
            while let Ok(input) = rx.recv() {
                if input.starts_with('#') {
                    command_tx.send(input).expect("failed to send to command_tx");
                } else {
                    writer.write_all(input.as_bytes()).expect("failed to write");
                    writer.flush().expect("failed to flush");
                }
            }
        });

        // 服务端 输出处理线程
        // 通过 global_tx 发送给主线程统一处理
        let global_output_tx = global_output_tx();
        let command_tx = self.command_tx.clone().unwrap();
        thread::spawn(move || {
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
                        command_tx.send(content.to_string()).expect("failed to send to command_tx");
                    }
                }
                global_output_tx.send(buf.clone()).expect("Failed to send to global_tx");
                // println!("{buf}");
            }
            println!("server end");
            child.wait().expect("failed to wait");
        });
    }

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
