mod config;
mod fabric;

use std::error::Error;

fn main() -> Result<(), Box<dyn Error>>{

    let config = config::load_config()?;

    Ok(())
}
