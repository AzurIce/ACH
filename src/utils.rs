pub mod path {
    use std::path::Path;

    pub fn split_parent_and_file(path: String) -> (String, String) {
        let path = Path::new(&path);
        let parent_path = path.parent().unwrap().to_str().unwrap();
        let file_path = path.file_name().unwrap().to_str().unwrap();
        (parent_path.to_string(), file_path.to_string())
    }
}

#[allow(unused)]
pub mod fs {
    use std::{fs, path::Path};

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
}

pub mod regex {
    use regex::Regex;

    const FORWARD: &str = r"^(.+) *\| *(\S+?)\n";
    pub fn forward_regex() -> Regex {
        Regex::new(FORWARD).expect("regex err")
    }
}
