use std::{
    fs::{self, DirEntry},
    io::{self, stdin, BufRead, Write},
    path::{Path, PathBuf},
    process::{ChildStdin, ChildStdout, Command, Stdio},
    sync::{
        mpsc::{self, Sender},
        Arc, Mutex,
    },
    thread,
};

use log::{error, info};

use bish_core::utils::{fs::copy_dir, regex::player_regex, time::get_cur_time_str};

use crate::{config::Config, core::loader::Loader};

// pub struct Core {
//     pub config: Config,
//     command_tx: mpsc::Sender<String>,
//     output_tx: mpsc::Sender<String>, // To stdout
//     running_server: Arc<Mutex<Option<Server>>>,
// }

pub struct Server {
    // pub child_in: ChildStdin,
    pub config: Config,
    pub running: Arc<Mutex<bool>>,
    pub child_in: Option<ChildStdin>,
}

impl Server {
    pub fn init(config: Config) -> Self {
        Self {
            config,
            running: Arc::new(Mutex::new(false)),
            child_in: None,
        }
    }

    pub fn start(&mut self, command_tx: Sender<String>, output_tx: Sender<String>) {
        info!("Server::start");
        let jar_filename = match self.config.loader {
            Loader::Quilt => "quilt-server-launch.jar",
        };
        let mut command = Command::new("java");
        let mut args = vec!["-jar", jar_filename, "--nogui"];
        args.extend(self.config.jvm_options.split(' ').collect::<Vec<&str>>());

        command.current_dir("./server");
        command.args(args);

        let mut child = command
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .expect("failed to spawn");

        let child_in = child.stdin.take().expect("Failed to open child's stdin");
        let child_out = child.stdout.take().expect("Failed to open child's stdout");

        self.child_in = Some(child_in);

        let _running = self.running.clone();
        thread::spawn(move || {
            let mut reader = io::BufReader::new(child_out);
            loop {
                let mut buf = String::new();
                match reader.read_line(&mut buf) {
                    Err(err) => {
                        // TODO: 为何初次运行会有一段是 stream did not contain valid UTF-8？
                        error!("{}", err)
                    }
                    Ok(size) => {
                        if size == 0 {
                            info!("thread_read_output: readed Ok(0)");
                            break;
                        }
                        let buf = buf.replace("\r\n", "\n");
                        let buf = buf.strip_suffix("\n").unwrap_or(&buf).to_string();
                        if let Some(cap) = player_regex().captures(&buf) {
                            let _player_name = cap.get(1).unwrap().as_str();
                            let content = cap.get(2).unwrap().as_str();
                            if content.starts_with('#') {
                                command_tx
                                    .send(content.to_string())
                                    .expect("failed to send to command_tx");
                            }
                        }
                        output_tx
                            .send(buf.clone())
                            .expect("Failed to send to global_tx");
                        // println!("{buf}");
                    }
                }
            }
            info!("server end");
            child.wait().expect("failed to wait");
            *_running.lock().unwrap() = false;
        });

    }

    pub fn writeln(&mut self, line: &str) {
        let line = if line.ends_with('\n') {
            line.to_string()
        } else {
            format!("{line}\n")
        };

        if let Some(child_in) = &mut self.child_in {
            child_in
                .write_all(line.as_bytes())
                .expect("failed to write to server");
        }
    }

    pub fn say<S: AsRef<str>>(&mut self, content: S) {
        let content = content.as_ref();
        let content = content.as_ref();
        println!("{content}");
        if *self.running.lock().unwrap() {
            self.writeln(content);
        }
    }
}

pub fn run(config: Config) {
    let (output_tx, output_rx) = mpsc::channel::<String>();
    info!("spawning output thread...");
    thread::spawn(move || {
        while let Ok(buf) = output_rx.recv() {
            println!("{buf}")
        }
    });
    let (command_tx, command_rx) = mpsc::channel::<String>();
    info!("starting server...");
    let server = Arc::new(Mutex::new(Server::init(config.clone())));

    info!("entering main loop...");
    let _server = server.clone();
    let _command_tx = command_tx.clone();
    let _output_tx = output_tx.clone();
    thread::spawn(move || {
        while let Ok(command_str) = command_rx.recv() {
            let command_str = command_str.replace("\r\n", "\n");
            let command_str = command_str.strip_prefix('#').unwrap();
            let command_str = command_str.strip_suffix('\n').unwrap_or(command_str);

            let split: Vec<&str> = command_str.split(' ').collect();
            let command = split[0];
            let args = &split[1..];

            println!("[server/{}]: command: {} {:?}", config.name, command, args);
            match command {
                "start" => {
                    info!("command start");
                    let mut server = _server.lock().unwrap();
                    if *server.running.lock().unwrap() {
                        error!("server is already running");
                    } else {
                        server.start(_command_tx.clone(), _output_tx.clone());
                    }
                }
                "bksnap" => {
                    // if let Some(server) = self.server {}
                    let mut server = _server.lock().unwrap();

                    if args.is_empty() || args[0] == "list" {
                        let snapshot_list = get_snapshot_list();

                        server.say("snapshots: ");
                        for (i, snapshot) in snapshot_list.into_iter().enumerate() {
                            server.say(format!("{i}: {snapshot:?}"))
                        }
                    } else if args[0] == "make" {
                        while get_snapshot_list().len() >= 10 {
                            del_snapshot()
                        }
                        server.say("saving snapshot...");
                        make_snapshot();
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
    });

    let mut buf = String::new();
    while let Ok(_size) = stdin().read_line(&mut buf) {
        let mut server = server.lock().unwrap();
        if buf.starts_with("#") {
            command_tx.send(buf.clone()).expect("failed to send to command_tx");
        } else {
            server.writeln(&buf)
        }
        buf.clear()
    }
}

// Backup related
pub fn del_snapshot() {
    println!("[del_snapshop]");
    let snapshot_dir = Path::new("./backups").join("snapshots");
    if let Err(err) = fs::create_dir_all(&snapshot_dir) {
        println!("failed to create all dir: {err}");
        return;
    }

    if let Ok(entries) = fs::read_dir(snapshot_dir) {
        let mut entries: Vec<DirEntry> = entries.into_iter().map(|entry| entry.unwrap()).collect();

        entries.sort_by_key(|entry| entry.metadata().unwrap().created().unwrap());
        let entries = entries
            .into_iter()
            .map(|entry| entry.path())
            .collect::<Vec<PathBuf>>();
        if let Some(first) = entries.first() {
            println!("[del_snapshop]: Deleting {first:?}...");
            if let Err(err) = fs::remove_dir_all(first) {
                println!("Failed to remove dir: {err}")
            }
            println!("[del_snapshop]: Snapshop deleted");
        }
    }
}

pub fn make_snapshot() {
    let snapshot_dir = Path::new("./backups").join("snapshots");
    if let Err(err) = fs::create_dir_all(snapshot_dir) {
        println!("failed to create all dir: {err}");
        return;
    }

    let src_path = Path::new(&"./server/").join("world");
    if !src_path.exists() {
        println!("skip world/, not exist");
        return;
    }

    let backup_name = get_cur_time_str();
    let dst_path = Path::new("./backups").join("snapshots").join(backup_name);
    println!("copying from {src_path:?} to {dst_path:?}...");
    if let Err(err) = copy_dir(&src_path, &dst_path) {
        println!("failed to copy: {err}")
    }
}

pub fn get_snapshot_list() -> Vec<PathBuf> {
    let snapshot_dir = Path::new("./backups").join("snapshots");
    if let Err(err) = fs::create_dir_all(&snapshot_dir) {
        println!("failed to create all dir: {err}");
        return Vec::new();
    }

    if let Ok(entries) = fs::read_dir(snapshot_dir) {
        let mut entries: Vec<DirEntry> = entries.into_iter().map(|entry| entry.unwrap()).collect();

        entries.sort_by_key(|entry| entry.metadata().unwrap().created().unwrap());
        entries
            .into_iter()
            .map(|entry| entry.path())
            .collect::<Vec<PathBuf>>()
    } else {
        Vec::new()
    }
    // snapshot_list.sort_by_key(|snapshot|snapshot.metadata.created().unwrap());
    // snapshot_list
}
