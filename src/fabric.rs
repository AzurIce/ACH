/// Inspired by https://github.com/Iru21/quick_fabric
use std::{
    error::Error,
    fs,
    io::Write,
    path::{Path, MAIN_SEPARATOR},
};

use curl::easy::Easy;

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn test_init_version() {
        init_version("1.20.2", "/Users/azurice/Game/MCServer/1.20.2")
            .expect("Failed to init version");
    }
}

fn init_version(version: &str, folder: &str) -> Result<(), Box<dyn Error>> {
    let res = reqwest::blocking::get("https://meta.fabricmc.net/v2/versions/installer")?;
    let json = res.json::<serde_json::Value>()?;
    let installer_version = json[0]["version"].as_str().unwrap();

    let res = reqwest::blocking::get("https://meta.fabricmc.net/v2/versions/loader")?;
    let json = res.json::<serde_json::Value>()?;
    let loader_version = json[0]["version"].as_str().unwrap();

    let url = format!("https://meta.fabricmc.net/v2/versions/loader/{version}/{loader_version}/{installer_version}/server/jar");

    println!("[fabric/init_version]: downloading server_jar to {folder}...");
    let path = format!("{folder}{MAIN_SEPARATOR}fabric-server-mc.{version}-loader.{loader_version}-launcher.{installer_version}.jar");
    download(&url, &path)?;
    println!("[fabric/init_version]: Downloaded to {path}");

    Ok(())
}

fn download(url: &str, path: &str) -> Result<(), Box<dyn Error>> {
    let path = Path::new(&path);
    if !path.parent().unwrap().exists() {
        fs::create_dir_all(&path).unwrap();
    }
    if path.exists() {
        println!("File already exist, skipping download...\n");
    } else {
        // if !is_empty(&Path::new(&folder)) {
        //     println!("Found an old installer, removing...\n");
        //     println!("File already exist, skipping download...\n");
        //     clear_dir(&Path::new(&folder));
        // }

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

fn is_empty(path: &Path) -> bool {
    match fs::read_dir(path) {
        Ok(entries) => entries.count() == 0,
        Err(_) => true,
    }
}

fn clear_dir(path: &Path) {
    if !is_empty(path) {
        for entry in fs::read_dir(path).unwrap() {
            let entry = entry.unwrap();
            if entry.file_type().unwrap().is_file() {
                fs::remove_file(entry.path()).unwrap();
            }
        }
    }
}
