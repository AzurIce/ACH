/// Inspired by https://github.com/Iru21/quick_fabric
use std::{
    error::Error,
    fs::{self, File},
    io::Write,
    path::{Path, MAIN_SEPARATOR},
};

use curl::easy::Easy;

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_init_version() {
        let path = init_server_jar("/Users/azurice/Game/MCServer/1.20.2", "1.20.2")
            .expect("Failed to init version");
        println!("{path}")
    }
}

pub fn init_server_jar(folder: &str, version: &str) -> Result<String, Box<dyn Error>> {
    // 获取最新的 installer 版本
    let res = reqwest::blocking::get("https://meta.fabricmc.net/v2/versions/installer")?;
    let json = res.json::<serde_json::Value>()?;
    let installer_version = json[0]["version"].as_str().unwrap();

    // 获取最新的 loader 版本
    let res = reqwest::blocking::get("https://meta.fabricmc.net/v2/versions/loader")?;
    let json = res.json::<serde_json::Value>()?;
    let loader_version = json[0]["version"].as_str().unwrap();

    // 拼接出服务端 jar 文件的 url
    let url = format!("https://meta.fabricmc.net/v2/versions/loader/{version}/{loader_version}/{installer_version}/server/jar");

    // 下载服务端 jar
    println!("[fabric/init_version]: downloading server_jar to {folder}...");
    let path = format!("{folder}{MAIN_SEPARATOR}fabric-server-mc.{version}-loader.{loader_version}-launcher.{installer_version}.jar");
    download(&url, &path)?;
    println!("[fabric/init_version]: Downloaded to {path}");

    // 写入 eula=true 到 eula.txt
    let mut eula_file = File::create(format!("{folder}{MAIN_SEPARATOR}eula.txt"))
        .expect("failed to create eula file");
    eula_file
        .write_all("eula=true".as_bytes())
        .expect("failed to write into eula file");

    Ok(path)
}

fn download(url: &str, path: &str) -> Result<(), Box<dyn Error>> {
    let path = Path::new(&path);
    if !path.parent().unwrap().exists() {
        fs::create_dir_all(&path).unwrap();
    }
    if path.exists() {
        println!("File already exist, skipping download...\n");
    } else {
        println!("Downloading to {:?} from {}\n", path, url);
        let mut f = fs::File::create(&path)?;
        let mut easy = Easy::new();
        easy.url(&url).unwrap();
        easy.follow_location(true).unwrap();
        easy.write_function(move |data| {
            f.write_all(data).unwrap();
            Ok(data.len())
        })
        .unwrap();
        easy.perform().unwrap();
        println!("Downloaded!");
    }
    Ok(())
}