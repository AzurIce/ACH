
/// Inspired by https://github.com/Iru21/quick_fabric

async fn get_installer_url() -> Result<String> {
    Ok(reqwest::get("https://meta.fabricmc.net/v2/versions/installer").await?.json::<serde_json::Value>().await?[0]["url"].as_str().unwrap().to_string())
}