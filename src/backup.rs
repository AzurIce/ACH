use std::fs::{DirEntry, Metadata};

#[derive(Debug)]
pub struct Backup {
    pub filename: String,
    pub metadata: Metadata
}

impl Backup {
    pub fn new(entry: DirEntry) -> Self {
        Self {
            filename: entry.file_name().to_str().unwrap().to_string(),
            metadata: entry.metadata().unwrap()
        }
    }
}