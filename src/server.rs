use std::{
    io::{self, BufRead,Write},
    process::{Command, Stdio},
    sync::mpsc,
    thread,
};

use derivative::Derivative;

use crate::{config::ServerConfig, fabric::init_server_jar, utils::path::split_parent_and_file};

#[derive(Derivative)]
#[derivative(Default)]
pub struct Server {
    #[derivative(Default(value = r#"Command::new("")"#))]
    command: Command,
    input_tx: Option<mpsc::Sender<String>>,
}

impl Server {
    pub fn new(config: ServerConfig) -> Self {
        let mut command = Command::new("java");
        let (dir, filename) = split_parent_and_file(
            init_server_jar(&config.dir, &config.version).expect("failed to init server jar"),
        );
        println!("[Server/new]: command's dir and jar_file is {dir} and {filename}");
        command.current_dir(dir);
        command.args(["-jar", &filename, "--nogui"]);

        Self {
            command,
            ..Default::default()
        }
    }
    pub fn run(&mut self) {
        let mut child = self
            .command
            .stdin(Stdio::piped())
            .stdout(Stdio::piped())
            .spawn()
            .expect("failed to spawn");

        let child_in = child.stdin.take().expect("Failed to open child's stdin");
        let child_out = child.stdout.take().expect("Failed to open child's stdout");

        let (tx, rx) = mpsc::channel::<String>();
        self.input_tx = Some(tx);

        // 统一输入处理线程
        // 一切从终端、服务端输出识别到的玩家输入，都会通过 input_tx 输入到 channel 中
        // 然后在此统一处理，识别命令，作相关处理
        thread::spawn(move || {
            let mut writer = io::BufWriter::new(child_in);
            while let Ok(input) = rx.recv() {
                writer.write_all(input.as_bytes()).expect("failed to write");
                writer.flush().expect("failed to flush");
            }
        });

        // 服务端 输出处理线程
        thread::spawn(move || {
            let mut reader = io::BufReader::new(child_out);
            let mut buf = String::new();
            while let Ok(size) = reader.read_line(&mut buf) {
                if size == 0 {
                    break;
                }
                println!("{buf}");
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
}
