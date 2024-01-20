use std::{
    fs::{self, DirEntry},
    io::{self, stdin, BufRead, Write},
    path::{Path, PathBuf},
    process::{ChildStdin, ChildStdout, Command, Stdio},
    sync::{Arc, Mutex, mpsc},
    thread,
};

use log::{error, info, warn};

use bish_core::utils::{fs::copy_dir, regex::player_regex, time::get_cur_time_str};

use crate::{config::Config, core::loader::Loader};

pub struct Core {
    pub config: Config,

    pub output_tx: mpsc::Sender<String>,  // Sender for stdout_loop
    pub command_tx: mpsc::Sender<String>, // Sender for command_hanle_loop
    pub event_tx: mpsc::Sender<Event>,

    pub running_server: Arc<Mutex<Option<Server>>>,
}

pub enum Event {
    ServerDown,
    ServerLog(String),
    PlayerMessage { player: String, msg: String },
}

impl Core {
    pub fn run(config: Config) {
        let running_server = Arc::new(Mutex::new(None::<Server>));
        let (output_tx, output_rx) = mpsc::channel::<String>();
        thread::spawn(move || {
            while let Ok(buf) = output_rx.recv() {
                println!("{buf}")
            }
        });
        let (command_tx, command_rx) = mpsc::channel::<String>();

        let _running_server = running_server.clone();
        let _command_tx = command_tx.clone();
        thread::spawn(move || {
            let mut buf = String::new();
            while let Ok(_size) = stdin().read_line(&mut buf) {
                if buf.starts_with("#") {
                    _command_tx
                        .send(buf.clone())
                        .expect("failed to send to command_tx");
                } else {
                    let mut _running_server = _running_server.lock().unwrap();
                    if let Some(server) = _running_server.as_mut() {
                        server.writeln(&buf)
                    }
                }
                buf.clear();
            }
        });

        let (event_tx, event_rx) = mpsc::channel::<Event>();
        let _running_server = running_server.clone();
        let _command_tx = command_tx.clone();
        thread::spawn(move || {
            while let Ok(event) = event_rx.recv() {
                match event {
                    Event::ServerDown => {
                        *_running_server.lock().unwrap() = None;
                    }
                    Event::ServerLog(msg) => {
                        println!("{msg}")
                    }
                    Event::PlayerMessage { player, msg } => {
                        if msg.starts_with("#") {
                            _command_tx
                                .send(msg.clone())
                                .expect("failed to send to command_tx");
                        }
                    }
                }
            }
        });
        let mut core = Core {
            config,
            output_tx,
            command_tx,
            event_tx,
            running_server,
        };

        while let Ok(command) = command_rx.recv() {
            core.handle_command(command);
        }
    }

    fn handle_command(&mut self, command: String) {
        let server = self.running_server.clone();

        let command = command.replace("\r\n", "\n");
        let command = command.strip_prefix('#').unwrap();
        let command = command.strip_suffix('\n').unwrap_or(command);

        let split: Vec<&str> = command.split(' ').collect();
        let command = split[0];
        let args = &split[1..];

        info!("command: {} {:?}", command, args);
        match command {
            "start" => {
                info!("command start");
                let mut server = server.lock().unwrap();
                if server.is_some() {
                    error!("server is already running");
                } else {
                    *server = Some(Server::run(self.config.clone(), self.event_tx.clone()));
                }
            }
            "bksnap" => {
                if args.is_empty() || args[0] == "list" {
                    let snapshot_list = get_snapshot_list();

                    self.say("snapshots: ");
                    for (i, snapshot) in snapshot_list.into_iter().enumerate() {
                        self.say(format!("{i}: {snapshot:?}"))
                    }
                } else if args[0] == "make" {
                    while get_snapshot_list().len() >= 10 {
                        del_snapshot()
                    }
                    self.say("saving snapshot...");
                    make_snapshot();
                    self.say("saved snapshot")
                } else if args.len() == 2 && args[0] == "load" {
                    println!("bksnap load, not implemented yet")
                    // TODO: load snap backup
                }
            }
            "bkarch" => {
                if args.is_empty() || args[0] == "list" {
                    let archive_list = get_archive_list();

                    self.say("archives: ");
                    for (i, archive) in archive_list.into_iter().enumerate() {
                        self.say(format!("{i}: {archive:?}"))
                    }
                } else if args[0] == "make" {
                    let comment = args[1..].join(" ");
                    self.say("saving archive...");
                    make_archive(&comment);
                    self.say("saved archive")
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

    pub fn say<S: AsRef<str>>(&mut self, content: S) {
        let content = content.as_ref();
        println!("{content}");
        if let Some(server) = self.running_server.lock().unwrap().as_mut() {
            server.writeln(format!("say {}", content).as_str());
        }
    }
}

pub struct Server {
    pub child_in: ChildStdin,
}

impl Server {
    pub fn run(config: Config, event_tx: mpsc::Sender<Event>) -> Self {
        info!("Server::start");
        let jar_filename = match config.loader {
            Loader::Quilt => "quilt-server-launch.jar",
        };
        let mut command = Command::new("java");
        let mut args = vec!["-jar", jar_filename, "--nogui"];
        args.extend(config.jvm_options.split(' ').collect::<Vec<&str>>());

        command.current_dir("./server");
        command.args(args);

        let mut child = command
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .expect("failed to spawn");

        let child_in = child.stdin.take().expect("Failed to open child's stdin");

        let child_out = child.stdout.take().expect("Failed to open child's stdout");
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
                            let player = cap.get(1).unwrap().as_str().to_string();
                            let msg = cap.get(2).unwrap().as_str().to_string();
                            event_tx
                                .send(Event::PlayerMessage { player, msg })
                                .expect("failed to send to event_tx");
                            // if content.starts_with('#') {
                            //     command_tx
                            //         .send(content.to_string())
                            //         .expect("failed to send to command_tx");
                            // }
                        }
                        event_tx
                            .send(Event::ServerLog(buf.clone()))
                            .expect("Failed to send to event_tx");
                        // println!("{buf}");
                    }
                }
            }
            info!("server end");
            child.wait().expect("failed to wait");
            event_tx.send(Event::ServerDown).expect("failed to send to event_tx");
        });

        Self { child_in }
    }

    pub fn writeln(&mut self, line: &str) {
        let line = if line.ends_with('\n') {
            line.to_string()
        } else {
            format!("{line}\n")
        };

        self.child_in
            .write_all(line.as_bytes())
            .expect("failed to write to server");
    }

    pub fn say<S: AsRef<str>>(&mut self, content: S) {
        let content = content.as_ref();
        self.writeln(format!("say {}", content).as_str());
    }
}

// pub fn run(config: Config) {
//     let (output_tx, output_rx) = mpsc::channel::<String>();
//     info!("spawning output thread...");
//     thread::spawn(move || {
//         while let Ok(buf) = output_rx.recv() {
//             println!("{buf}")
//         }
//     });
//     let (command_tx, command_rx) = mpsc::channel::<String>();
//     info!("starting server...");
//     let server = Arc::new(Mutex::new(Server::init(config.clone())));

//     info!("entering main loop...");
//     let _server = server.clone();
//     let _command_tx = command_tx.clone();
//     let _output_tx = output_tx.clone();
//     thread::spawn(move || {
//         while let Ok(command_str) = command_rx.recv() {
//             let command_str = command_str.replace("\r\n", "\n");
//             let command_str = command_str.strip_prefix('#').unwrap();
//             let command_str = command_str.strip_suffix('\n').unwrap_or(command_str);

//             let split: Vec<&str> = command_str.split(' ').collect();
//             let command = split[0];
//             let args = &split[1..];

//             println!("[server/{}]: command: {} {:?}", config.name, command, args);
//             match command {
//                 "start" => {
//                     info!("command start");
//                     let mut server = _server.lock().unwrap();
//                     if *server.running.lock().unwrap() {
//                         error!("server is already running");
//                     } else {
//                         server.start(_command_tx.clone(), _output_tx.clone());
//                     }
//                 }
//                 "bksnap" => {
//                     // if let Some(server) = self.server {}
//                     let mut server = _server.lock().unwrap();

//                     if args.is_empty() || args[0] == "list" {
//                         let snapshot_list = get_snapshot_list();

//                         server.say("snapshots: ");
//                         for (i, snapshot) in snapshot_list.into_iter().enumerate() {
//                             server.say(format!("{i}: {snapshot:?}"))
//                         }
//                     } else if args[0] == "make" {
//                         while get_snapshot_list().len() >= 10 {
//                             del_snapshot()
//                         }
//                         server.say("saving snapshot...");
//                         make_snapshot();
//                         server.say("saved snapshot")
//                     } else if args.len() == 2 && args[0] == "load" {
//                         println!("bksnap load, not implemented yet")
//                         // TODO: load snap backup
//                     }
//                 }
//                 "bkarch" => {
//                     let mut server = _server.lock().unwrap();

//                     if args.is_empty() || args[0] == "list" {
//                         let archive_list = get_archive_list();

//                         server.say("archives: ");
//                         for (i, archive) in archive_list.into_iter().enumerate() {
//                             server.say(format!("{i}: {archive:?}"))
//                         }
//                     } else if args[0] == "make" {
//                         let comment = args[1..].join(" ");
//                         server.say("saving archive...");
//                         make_archive(&comment);
//                         server.say("saved archive")
//                         // TODO: make arch backup
//                     } else if args.len() == 2 && args[0] == "load" {
//                         println!("bkarch load, not implemented yet")
//                         // TODO: load arch backup
//                     }
//                 }
//                 _ => {
//                     println!("unknown command")
//                 }
//             }
//         }
//     });

//     let mut buf = String::new();
//     while let Ok(_size) = stdin().read_line(&mut buf) {
//         let mut server = server.lock().unwrap();
//         if buf.starts_with("#") {
//             command_tx
//                 .send(buf.clone())
//                 .expect("failed to send to command_tx");
//         } else {
//             if *server.running.lock().unwrap() {
//                 server.writeln(&buf)
//             }
//         }
//         buf.clear()
//     }
// }

// Backup related
pub fn del_snapshot() {
    info!("deleting snapshot");
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
        error!("failed to create all dir: {err}");
        return;
    }

    let src_path = Path::new(&"./server/").join("world");
    if !src_path.exists() {
        warn!("skip world/, not exist");
        return;
    }

    let backup_name = get_cur_time_str();
    let dst_path = Path::new("./backups").join("snapshots").join(backup_name);
    info!("copying from {src_path:?} to {dst_path:?}...");
    if let Err(err) = copy_dir(&src_path, &dst_path) {
        error!("failed to copy: {err}")
    }
}

pub fn get_snapshot_list() -> Vec<PathBuf> {
    let snapshot_dir = Path::new("./backups").join("snapshots");
    if let Err(err) = fs::create_dir_all(&snapshot_dir) {
        error!("failed to create all dir: {err}");
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

pub fn get_archive_list() -> Vec<PathBuf> {
    let dir = Path::new("./backups").join("archives");
    if let Err(err) = fs::create_dir_all(&dir) {
        error!("failed to create all dir: {err}");
        return Vec::new();
    }

    if let Ok(entries) = fs::read_dir(dir) {
        let mut entries: Vec<DirEntry> = entries.into_iter().map(|entry| entry.unwrap()).collect();

        entries.sort_by_key(|entry| entry.metadata().unwrap().created().unwrap());
        entries
            .into_iter()
            .map(|entry| entry.path())
            .collect::<Vec<PathBuf>>()
    } else {
        Vec::new()
    }
}

pub fn make_archive(name: &str) {
    let dir = Path::new("./backups").join("archives");
    if let Err(err) = fs::create_dir_all(&dir) {
        error!("failed to create all dir: {err}");
        return;
    }

    let src_path = Path::new(&"./server/").join("world");
    if !src_path.exists() {
        warn!("skip world/, not exist");
        return;
    }

    let backup_name = format!("{} {}", get_cur_time_str(), name);
    let dst_path = dir.join(backup_name);
    info!("copying from {src_path:?} to {dst_path:?}...");
    if let Err(err) = copy_dir(&src_path, &dst_path) {
        error!("failed to copy: {err}")
    }
}
