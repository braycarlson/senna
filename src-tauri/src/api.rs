use std::sync::RwLock;

use serde_json::Value;
use serde_json::value::RawValue;
use tauri::State;

use crate::settings::Settings;

const REQUEST_TIMEOUT_S: u64 = 15;

pub struct ApiState {
    pub http: reqwest::Client,
    pub settings: RwLock<Settings>,
}

impl ApiState {
    pub fn new(settings: Settings) -> ApiState {
        let http = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(REQUEST_TIMEOUT_S))
            .build()
            .expect("http client build");

        ApiState {
            http,
            settings: RwLock::new(settings),
        }
    }

    pub(crate) fn settings_snapshot(&self) -> Settings {
        self.settings
            .read()
            .expect("settings lock poisoned")
            .clone()
    }
}

async fn request_body(
    state: &ApiState,
    method: reqwest::Method,
    path: &str,
    query: &[(&str, Option<String>)],
) -> Result<String, String> {
    assert!(path.starts_with('/'), "path must start with a slash");

    let settings = state.settings_snapshot();
    let url = format!("{}{path}", settings.base_url.trim_end_matches('/'));

    let pairs: Vec<(&str, &str)> = query
        .iter()
        .filter_map(|(name, value)| value.as_deref().map(|value| (*name, value)))
        .collect();

    let mut request = state.http.request(method, url).query(&pairs);

    if !settings.token.is_empty() {
        request = request.bearer_auth(&settings.token);
    }

    let response = request.send().await.map_err(|error| error.to_string())?;
    let status = response.status();
    let body = response.text().await.map_err(|error| error.to_string())?;

    if !status.is_success() {
        return Err(format!("{status}: {body}"));
    }

    Ok(body)
}

async fn request_raw(
    state: &ApiState,
    method: reqwest::Method,
    path: &str,
    query: &[(&str, Option<String>)],
) -> Result<Box<RawValue>, String> {
    let body = request_body(state, method, path, query).await?;

    RawValue::from_string(body).map_err(|error| error.to_string())
}

async fn get_raw(
    state: &ApiState,
    path: &str,
    query: &[(&str, Option<String>)],
) -> Result<Box<RawValue>, String> {
    request_raw(state, reqwest::Method::GET, path, query).await
}

pub(crate) async fn get_json(
    state: &ApiState,
    path: &str,
    query: &[(&str, Option<String>)],
) -> Result<Value, String> {
    let body = request_body(state, reqwest::Method::GET, path, query).await?;

    serde_json::from_str(&body).map_err(|error| error.to_string())
}

#[tauri::command]
pub async fn api_patches(state: State<'_, ApiState>) -> Result<Box<RawValue>, String> {
    get_raw(&state, "/patches", &[]).await
}

#[tauri::command]
pub async fn api_stats(state: State<'_, ApiState>) -> Result<Box<RawValue>, String> {
    get_raw(&state, "/stats", &[]).await
}

#[tauri::command]
pub async fn api_champions(
    state: State<'_, ApiState>,
    patch: Option<String>,
    queue: Option<i32>,
) -> Result<Box<RawValue>, String> {
    let query = [
        ("patch", patch),
        ("queue", queue.map(|value| value.to_string())),
    ];

    get_raw(&state, "/aram/champions", &query).await
}

#[tauri::command]
pub async fn api_tier(
    state: State<'_, ApiState>,
    patch: Option<String>,
    queue: Option<i32>,
    games_min: Option<i64>,
) -> Result<Box<RawValue>, String> {
    let query = [
        ("patch", patch),
        ("queue", queue.map(|value| value.to_string())),
        ("games_min", games_min.map(|value| value.to_string())),
    ];

    get_raw(&state, "/aram/tier", &query).await
}

#[allow(clippy::similar_names)]
#[tauri::command]
pub async fn api_build(
    state: State<'_, ApiState>,
    champion_id: i32,
    patch: Option<String>,
    queue: Option<i32>,
) -> Result<Box<RawValue>, String> {
    let path = format!("/aram/champions/{champion_id}/build");
    let query = [
        ("patch", patch),
        ("queue", queue.map(|value| value.to_string())),
    ];

    get_raw(&state, &path, &query).await
}

#[allow(clippy::similar_names)]
#[tauri::command]
pub async fn api_matchups(
    state: State<'_, ApiState>,
    champion_id: i32,
    patch: Option<String>,
    queue: Option<i32>,
    games_min: Option<i64>,
) -> Result<Box<RawValue>, String> {
    let path = format!("/aram/champions/{champion_id}/matchups");

    let query = [
        ("patch", patch),
        ("queue", queue.map(|value| value.to_string())),
        ("games_min", games_min.map(|value| value.to_string())),
    ];

    get_raw(&state, &path, &query).await
}

#[allow(clippy::similar_names)]
#[tauri::command]
pub async fn api_synergies(
    state: State<'_, ApiState>,
    champion_id: i32,
    patch: Option<String>,
    queue: Option<i32>,
    games_min: Option<i64>,
) -> Result<Box<RawValue>, String> {
    let path = format!("/aram/champions/{champion_id}/synergies");

    let query = [
        ("patch", patch),
        ("queue", queue.map(|value| value.to_string())),
        ("games_min", games_min.map(|value| value.to_string())),
    ];

    get_raw(&state, &path, &query).await
}

#[tauri::command]
pub async fn api_match(
    state: State<'_, ApiState>,
    match_id: String,
) -> Result<Box<RawValue>, String> {
    assert!(!match_id.is_empty(), "match id must not be empty");

    let path = format!("/matches/{match_id}");

    get_raw(&state, &path, &[]).await
}

#[tauri::command]
pub async fn api_player_champions(
    state: State<'_, ApiState>,
    puuid: String,
    limit: Option<i64>,
) -> Result<Box<RawValue>, String> {
    assert!(!puuid.is_empty(), "puuid must not be empty");

    let path = format!("/players/{puuid}/champions");
    let query = [("limit", limit.map(|value| value.to_string()))];

    get_raw(&state, &path, &query).await
}

#[tauri::command]
pub async fn api_player(
    state: State<'_, ApiState>,
    name: String,
    tag: String,
    region: Option<String>,
    limit: Option<i64>,
) -> Result<Box<RawValue>, String> {
    assert!(!name.is_empty(), "name must not be empty");
    assert!(!tag.is_empty(), "tag must not be empty");

    let path = format!("/players/by-riot-id/{name}/{tag}");

    let query = [
        ("region", region),
        ("limit", limit.map(|value| value.to_string())),
    ];

    get_raw(&state, &path, &query).await
}

#[tauri::command]
pub async fn api_player_refresh(
    state: State<'_, ApiState>,
    name: String,
    tag: String,
    region: Option<String>,
) -> Result<Box<RawValue>, String> {
    assert!(!name.is_empty(), "name must not be empty");
    assert!(!tag.is_empty(), "tag must not be empty");

    let path = format!("/players/by-riot-id/{name}/{tag}/refresh");
    let query = [("region", region)];

    request_raw(&state, reqwest::Method::POST, &path, &query).await
}
