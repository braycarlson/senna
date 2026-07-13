use std::collections::HashMap;
use std::sync::atomic::{AtomicBool, AtomicI32, Ordering};
use std::sync::{Arc, RwLock};
use std::time::{SystemTime, UNIX_EPOCH};

use asol::{Client, Config, Connector, Event, EventKind};
use serde_json::{Value, json};
use tauri::{AppHandle, Emitter, Manager};
use tokio::sync::mpsc;

use crate::api::{ApiState, get_json};
use crate::settings::{ItemBlock, Loadout, RunePage, Settings};

const BENCH_SWAP_URI: &str = "/lol-champ-select/v1/session/bench/swap";
const CHAMP_SELECT_URI: &str = "/lol-champ-select/v1/session";
const GAMEFLOW_URI: &str = "/lol-gameflow/v1/gameflow-phase";
const REROLL_URI: &str = "/lol-champ-select/v1/session/my-selection/reroll";
const GAMEFLOW_SESSION_URI: &str = "/lol-gameflow/v1/session";
const CURRENT_PAGE_URI: &str = "/lol-perks/v1/currentpage";
const ITEM_SETS_URI: &str = "/lol-item-sets/v1/item-sets";
const MY_SELECTION_URI: &str = "/lol-champ-select/v1/session/my-selection";
const PERK_INVENTORY_URI: &str = "/lol-perks/v1/inventory";
const PERK_PAGES_URI: &str = "/lol-perks/v1/pages";
const PICKABLE_SKINS_URI: &str = "/lol-champ-select/v1/pickable-skin-ids";
const READY_CHECK_URI: &str = "/lol-matchmaking/v1/ready-check";
const READY_CHECK_ACCEPT_URI: &str = "/lol-matchmaking/v1/ready-check/accept";
const SUMMONER_URI: &str = "/lol-summoner/v1/current-summoner";
const PAGE_NAME_PREFIX: &str = "senna:";
const PAGE_PERK_COUNT: usize = 9;

const PHASES_WITH_CHAMPION: [&str; 4] = ["ChampSelect", "GameStart", "InProgress", "Reconnect"];
const EVENTS_MAX: usize = 200;
const MODE_ARAM: &str = "ARAM";

enum SkinApply {
    Applied,
    Pending,
    Skipped,
}

#[derive(Clone, PartialEq)]
struct MemberState {
    champion: i64,
    intent: i64,
    spell1: i64,
    spell2: i64,
    name: String,
    this: bool,
}

pub struct LcuHandle {
    pub connector: Arc<Connector>,
    pub current_champion: AtomicI32,
    pub applied_champion: AtomicI32,
    pub skinned_champion: AtomicI32,
    pub in_select: AtomicBool,
    pub lobby: RwLock<Value>,
    pub selection: RwLock<Value>,
    pub game_mode: RwLock<String>,
    pub events: RwLock<Vec<Value>>,
}

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
pub fn lcu_status(handle: tauri::State<'_, LcuHandle>) -> &'static str {
    if handle.connector.client().is_some() {
        "connected"
    } else {
        "disconnected"
    }
}

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
pub fn lcu_current_champion(handle: tauri::State<'_, LcuHandle>) -> i32 {
    handle.current_champion.load(Ordering::Relaxed)
}

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
pub fn lcu_lobby(handle: tauri::State<'_, LcuHandle>) -> Value {
    handle.lobby.read().expect("lobby lock poisoned").clone()
}

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
pub fn lcu_selection(handle: tauri::State<'_, LcuHandle>) -> Value {
    handle
        .selection
        .read()
        .expect("selection lock poisoned")
        .clone()
}

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
pub fn lcu_events(handle: tauri::State<'_, LcuHandle>) -> Vec<Value> {
    handle.events.read().expect("events lock poisoned").clone()
}

#[tauri::command]
pub async fn lcu_set_spells(
    handle: tauri::State<'_, LcuHandle>,
    spell_d: i64,
    spell_f: i64,
) -> Result<(), String> {
    if spell_d <= 0 || spell_f <= 0 || spell_d == spell_f {
        return Err("two distinct spells are required".to_string());
    }

    let client = client_of(&handle)?;

    apply_spell_pair(&client, spell_d, spell_f)
        .await
        .map(|_| ())
}

#[tauri::command]
pub async fn lcu_set_skin(handle: tauri::State<'_, LcuHandle>, skin_id: i64) -> Result<(), String> {
    if skin_id <= 0 {
        return Err("a skin id is required".to_string());
    }

    let client = client_of(&handle)?;
    let body = json!({ "selectedSkinId": skin_id });

    let _: Value = client
        .patch(MY_SELECTION_URI, &body)
        .await
        .map_err(|error| error.to_string())?;

    Ok(())
}

#[tauri::command]
pub async fn lcu_bench_swap(
    handle: tauri::State<'_, LcuHandle>,
    champion_id: i32,
) -> Result<(), String> {
    if champion_id <= 0 {
        return Err("a champion id is required".to_string());
    }

    let client = client_of(&handle)?;
    let uri = format!("{BENCH_SWAP_URI}/{champion_id}");

    let _: Value = client
        .post(&uri, &Value::Null)
        .await
        .map_err(|error| error.to_string())?;

    Ok(())
}

#[tauri::command]
pub async fn lcu_reroll(handle: tauri::State<'_, LcuHandle>) -> Result<(), String> {
    let client = client_of(&handle)?;

    let _: Value = client
        .post(REROLL_URI, &Value::Null)
        .await
        .map_err(|error| error.to_string())?;

    Ok(())
}

#[tauri::command]
pub async fn lcu_pickable_skins(handle: tauri::State<'_, LcuHandle>) -> Result<Vec<i64>, String> {
    let client = client_of(&handle)?;

    let pickable: Value = client
        .get(PICKABLE_SKINS_URI)
        .await
        .map_err(|error| error.to_string())?;

    let ids = pickable
        .as_array()
        .map(|entries| entries.iter().filter_map(Value::as_i64).collect())
        .unwrap_or_default();

    Ok(ids)
}

#[tauri::command]
pub async fn lcu_current_page(handle: tauri::State<'_, LcuHandle>) -> Result<Value, String> {
    let client = client_of(&handle)?;

    let page: Value = client
        .get(CURRENT_PAGE_URI)
        .await
        .map_err(|error| error.to_string())?;

    Ok(json!({
        "id": page.get("id").and_then(Value::as_i64).unwrap_or(0),
        "name": page.get("name").and_then(Value::as_str).unwrap_or(""),
    }))
}

#[tauri::command]
pub async fn lcu_apply_stats_runes(
    app: AppHandle,
    handle: tauri::State<'_, LcuHandle>,
    champion_id: i32,
) -> Result<bool, String> {
    if champion_id <= 0 {
        return Err("a champion id is required".to_string());
    }

    let client = client_of(&handle)?;
    let state = app.state::<ApiState>();
    let path = format!("/aram/champions/{champion_id}/build");
    let build = get_json(&state, &path, &[]).await?;

    apply_runes(&client, &build).await
}

fn client_of(handle: &tauri::State<'_, LcuHandle>) -> Result<Client, String> {
    handle
        .connector
        .client()
        .ok_or_else(|| "league client is not available".to_string())
}

#[tauri::command]
pub async fn lcu_current_summoner(handle: tauri::State<'_, LcuHandle>) -> Result<Value, String> {
    let Some(client) = handle.connector.client() else {
        return Err("league client is not available".to_string());
    };

    client
        .get(SUMMONER_URI)
        .await
        .map_err(|error| error.to_string())
}

#[tauri::command]
pub async fn lcu_owned_skins(
    handle: tauri::State<'_, LcuHandle>,
    champion_id: i32,
) -> Result<Vec<i64>, String> {
    let Some(client) = handle.connector.client() else {
        return Err("league client is not available".to_string());
    };

    let summoner: Value = client
        .get(SUMMONER_URI)
        .await
        .map_err(|error| error.to_string())?;

    let Some(summoner_id) = summoner.get("summonerId").and_then(Value::as_i64) else {
        return Err("summoner id is unavailable".to_string());
    };

    let uri = format!("/lol-champions/v1/inventories/{summoner_id}/champions/{champion_id}/skins");

    let skins: Value = client.get(&uri).await.map_err(|error| error.to_string())?;

    let Some(entries) = skins.as_array() else {
        return Ok(Vec::new());
    };

    let owned = entries
        .iter()
        .filter_map(|skin| {
            let base = skin.get("isBase").and_then(Value::as_bool).unwrap_or(false);

            let owned = skin
                .get("ownership")
                .and_then(|ownership| ownership.get("owned"))
                .and_then(Value::as_bool)
                .unwrap_or(false);

            if !base && !owned {
                return None;
            }

            skin.get("id").and_then(Value::as_i64)
        })
        .collect();

    Ok(owned)
}

pub fn lcu_spawn(app: &AppHandle) {
    let (connector, receiver) = Connector::new(Config::default());
    let connector = Arc::new(connector);

    connector.on_message(CHAMP_SELECT_URI, EventKind::Create);
    connector.on_message(CHAMP_SELECT_URI, EventKind::Update);
    connector.on_message(GAMEFLOW_URI, EventKind::Update);
    connector.on_message(CURRENT_PAGE_URI, EventKind::Create);
    connector.on_message(CURRENT_PAGE_URI, EventKind::Update);
    connector.on_message(READY_CHECK_URI, EventKind::Create);
    connector.on_message(READY_CHECK_URI, EventKind::Update);
    connector.on_message(PICKABLE_SKINS_URI, EventKind::Create);
    connector.on_message(PICKABLE_SKINS_URI, EventKind::Update);

    app.manage(LcuHandle {
        connector: Arc::clone(&connector),
        current_champion: AtomicI32::new(0),
        applied_champion: AtomicI32::new(0),
        skinned_champion: AtomicI32::new(0),
        in_select: AtomicBool::new(false),
        lobby: RwLock::new(Value::Array(Vec::new())),
        selection: RwLock::new(selection_cleared()),
        game_mode: RwLock::new(String::new()),
        events: RwLock::new(Vec::new()),
    });

    let runner = Arc::clone(&connector);

    tauri::async_runtime::spawn(async move {
        let _ = runner.run().await;
    });

    tauri::async_runtime::spawn(lcu_listen(app.clone(), connector, receiver));
}

struct ListenState {
    select_members: HashMap<i64, MemberState>,
    was_in_select: bool,
    ready_seen: bool,
}

impl ListenState {
    fn new() -> ListenState {
        ListenState {
            select_members: HashMap::new(),
            was_in_select: false,
            ready_seen: false,
        }
    }
}

async fn lcu_listen(
    app: AppHandle,
    connector: Arc<Connector>,
    mut receiver: mpsc::Receiver<Event>,
) {
    let lcu_state = app.state::<LcuHandle>();
    let mut state = ListenState::new();

    while let Some(event) = receiver.recv().await {
        match event {
            Event::Ready => on_ready(&app, &connector, &lcu_state).await,
            Event::Searching => {
                let _ = app.emit("lcu-status", "searching");
            }
            Event::Closed | Event::Failed(_) => on_disconnect(&app, &lcu_state, &mut state),
            Event::Message(message) => {
                on_message(
                    &app,
                    &connector,
                    &lcu_state,
                    &mut state,
                    message.uri.as_str(),
                    &message.data,
                )
                .await;
            }
            _ => {}
        }
    }
}

async fn on_ready(app: &AppHandle, connector: &Connector, lcu_state: &tauri::State<'_, LcuHandle>) {
    let _ = app.emit("lcu-status", "connected");

    let Some((champion_id, in_select)) = initial_sync(app, connector, lcu_state).await else {
        return;
    };

    debug_assert!(champion_id > 0, "initial sync yields a picked champion");

    lcu_state.in_select.store(in_select, Ordering::Relaxed);

    track_champion(app, lcu_state, champion_id);

    if in_select {
        champion_pipeline(app, connector, champion_id).await;
    }
}

fn on_disconnect(
    app: &AppHandle,
    lcu_state: &tauri::State<'_, LcuHandle>,
    state: &mut ListenState,
) {
    let _ = app.emit("lcu-status", "disconnected");

    lcu_state.applied_champion.store(0, Ordering::Relaxed);
    lcu_state.skinned_champion.store(0, Ordering::Relaxed);
    lcu_state.in_select.store(false, Ordering::Relaxed);

    state.select_members.clear();
    state.was_in_select = false;

    track_champion(app, lcu_state, 0);
    set_lobby(app, lcu_state, Value::Array(Vec::new()));
    set_selection(app, lcu_state, selection_cleared());
}

async fn on_message(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
    state: &mut ListenState,
    uri: &str,
    data: &Value,
) {
    match uri {
        GAMEFLOW_URI => on_gameflow(app, connector, lcu_state, state, data).await,
        CHAMP_SELECT_URI => on_champ_select(app, connector, lcu_state, state, data).await,
        PICKABLE_SKINS_URI => on_pickable_skins(app, connector, lcu_state).await,
        CURRENT_PAGE_URI => on_current_page(app, data),
        READY_CHECK_URI => on_ready_check(app, connector, lcu_state, state, data).await,
        _ => {}
    }
}

async fn on_gameflow(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
    state: &mut ListenState,
    data: &Value,
) {
    let phase = data.as_str().unwrap_or("");

    lcu_state
        .in_select
        .store(phase == "ChampSelect", Ordering::Relaxed);

    if phase != "ChampSelect" {
        lcu_state.applied_champion.store(0, Ordering::Relaxed);
        lcu_state.skinned_champion.store(0, Ordering::Relaxed);

        set_selection(app, lcu_state, selection_cleared());

        if state.was_in_select {
            state.was_in_select = false;

            state.select_members.clear();
            push_event(
                app,
                lcu_state,
                json!({
                    "kind": "phase",
                    "note": "Champ select ended",
                }),
            );
        }
    }

    if !PHASES_WITH_CHAMPION.contains(&phase) {
        track_champion(app, lcu_state, 0);
        set_lobby(app, lcu_state, Value::Array(Vec::new()));
    } else if phase != "ChampSelect" {
        sync_game_lobby(app, connector, lcu_state).await;
    }
}

async fn on_champ_select(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
    state: &mut ListenState,
    data: &Value,
) {
    lcu_state.in_select.store(true, Ordering::Relaxed);

    if !state.was_in_select {
        state.was_in_select = true;

        state.select_members.clear();
        sync_game_mode(connector, lcu_state).await;

        let mode = lcu_state
            .game_mode
            .read()
            .expect("game mode lock poisoned")
            .clone();

        push_event(
            app,
            lcu_state,
            json!({
                "kind": "phase",
                "note": format!("Champ select started ({})", mode_label(&mode)),
            }),
        );
    }

    for entry in diff_select_events(&mut state.select_members, data) {
        push_event(app, lcu_state, entry);
    }

    let mode = lcu_state
        .game_mode
        .read()
        .expect("game mode lock poisoned")
        .clone();

    set_lobby(app, lcu_state, select_roster(data));
    set_selection(app, lcu_state, select_selection(data, &mode));

    let Some(champion_id) = local_champion(data) else {
        return;
    };

    debug_assert!(
        champion_id > 0,
        "local champion id is positive when present"
    );

    track_champion(app, lcu_state, champion_id);

    champion_pipeline(app, connector, champion_id).await;
}

async fn on_pickable_skins(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
) {
    if !lcu_state.in_select.load(Ordering::Relaxed) {
        return;
    }

    let champion_id = lcu_state.current_champion.load(Ordering::Relaxed);

    if champion_id <= 0 {
        return;
    }

    lcu_state.skinned_champion.store(0, Ordering::Relaxed);

    champion_pipeline(app, connector, champion_id).await;
}

fn on_current_page(app: &AppHandle, data: &Value) {
    let page = json!({
        "id": data.get("id").and_then(Value::as_i64).unwrap_or(0),
        "name": data.get("name").and_then(Value::as_str).unwrap_or(""),
    });

    let _ = app.emit("lcu-runepage", page);
}

async fn on_ready_check(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
    state: &mut ListenState,
    data: &Value,
) {
    let response = data
        .get("playerResponse")
        .and_then(Value::as_str)
        .unwrap_or("");

    if response == "None" && !state.ready_seen {
        state.ready_seen = true;

        push_event(
            app,
            lcu_state,
            json!({
                "kind": "phase",
                "note": "Match found",
            }),
        );
    } else if response != "None" {
        state.ready_seen = false;
    }

    accept_ready_check(app, connector, data).await;
}

fn now_ms() -> i64 {
    let elapsed = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default();

    i64::try_from(elapsed.as_millis()).unwrap_or(i64::MAX)
}

fn mode_label(mode: &str) -> String {
    match mode {
        "" => "Unknown mode".to_string(),
        "CLASSIC" => "Summoner's Rift".to_string(),
        "ONEFORALL" => "One for All".to_string(),
        _ => title_case_mode(mode),
    }
}

fn title_case_mode(mode: &str) -> String {
    if mode.len() <= 1 || mode == "ARAM" || mode == "URF" {
        return mode.to_string();
    }

    format!("{}{}", &mode[..1], mode[1..].to_lowercase())
}

fn push_event(app: &AppHandle, lcu_state: &tauri::State<'_, LcuHandle>, mut entry: Value) {
    entry["ts"] = json!(now_ms());

    {
        let mut events = lcu_state.events.write().expect("events lock poisoned");

        events.push(entry.clone());

        if events.len() > EVENTS_MAX {
            let excess = events.len() - EVENTS_MAX;

            events.drain(..excess);
        }
    }

    let _ = app.emit("lcu-event", entry);
}

async fn sync_game_mode(connector: &Connector, lcu_state: &tauri::State<'_, LcuHandle>) {
    let Some(client) = connector.client() else {
        return;
    };

    let mode = match client.get::<Value>(GAMEFLOW_SESSION_URI).await {
        Ok(session) => session
            .get("gameData")
            .and_then(|data| data.get("queue"))
            .and_then(|queue| queue.get("gameMode"))
            .and_then(Value::as_str)
            .unwrap_or("")
            .to_string(),
        Err(_) => String::new(),
    };

    let mut current = lcu_state
        .game_mode
        .write()
        .expect("game mode lock poisoned");

    *current = mode;
}

fn member_states(session: &Value) -> HashMap<i64, MemberState> {
    let local_cell = session
        .get("localPlayerCellId")
        .and_then(Value::as_i64)
        .unwrap_or(-1);

    let Some(team) = session.get("myTeam").and_then(Value::as_array) else {
        return HashMap::new();
    };

    let mut states = HashMap::new();

    for member in team {
        let Some(cell) = member.get("cellId").and_then(Value::as_i64) else {
            continue;
        };

        let grab = |name: &str| member.get(name).and_then(Value::as_i64).unwrap_or(0);

        let mut name = member_name(member);

        if name.is_empty() {
            name = format!("Summoner {}", cell + 1);
        }

        states.insert(
            cell,
            MemberState {
                champion: grab("championId"),
                intent: grab("championPickIntent"),
                spell1: grab("spell1Id"),
                spell2: grab("spell2Id"),
                name,
                this: cell == local_cell,
            },
        );
    }

    states
}

fn diff_select_events(previous: &mut HashMap<i64, MemberState>, session: &Value) -> Vec<Value> {
    let current = member_states(session);
    let first_sync = previous.is_empty();

    let mut events = Vec::new();
    let mut champion_changes: Vec<(i64, i64, i64)> = Vec::new();

    collect_member_events(&current, previous, &mut events, &mut champion_changes);

    let traded = collect_trade_events(&current, &champion_changes, &mut events);

    collect_champion_events(
        &current,
        &champion_changes,
        &traded,
        first_sync,
        &mut events,
    );

    *previous = current;

    events
}

fn collect_member_events(
    current: &HashMap<i64, MemberState>,
    previous: &HashMap<i64, MemberState>,
    events: &mut Vec<Value>,
    champion_changes: &mut Vec<(i64, i64, i64)>,
) {
    for (cell, state) in current {
        let Some(old) = previous.get(cell) else {
            let appeared_with_champion = !previous.is_empty() && state.champion > 0;

            if appeared_with_champion {
                champion_changes.push((*cell, 0, state.champion));
            }

            continue;
        };

        if state.champion != old.champion {
            champion_changes.push((*cell, old.champion, state.champion));
        }

        let hovered =
            state.intent != old.intent && state.intent > 0 && state.intent != state.champion;

        if hovered {
            events.push(json!({
                "kind": "hover",
                "player": current[cell].name,
                "self": current[cell].this,
                "champion_id": state.intent,
            }));
        }

        let spells_updated = (state.spell1 != old.spell1 || state.spell2 != old.spell2)
            && state.spell1 > 0
            && state.spell2 > 0;

        if spells_updated {
            events.push(json!({
                "kind": "spells",
                "player": current[cell].name,
                "self": current[cell].this,
                "spells": [state.spell1, state.spell2],
            }));
        }
    }
}

fn collect_trade_events(
    current: &HashMap<i64, MemberState>,
    champion_changes: &[(i64, i64, i64)],
    events: &mut Vec<Value>,
) -> Vec<i64> {
    let mut traded: Vec<i64> = Vec::new();

    for index in 0..champion_changes.len() {
        for other in (index + 1)..champion_changes.len() {
            let (cell_a, old_a, new_a) = champion_changes[index];
            let (cell_b, old_b, new_b) = champion_changes[other];

            let swapped = old_a > 0 && old_b > 0 && new_a == old_b && new_b == old_a;
            let fresh_swap = swapped && !traded.contains(&cell_a) && !traded.contains(&cell_b);

            if !fresh_swap {
                continue;
            }

            debug_assert!(cell_a != cell_b, "a trade swaps two distinct cells");

            traded.push(cell_a);
            traded.push(cell_b);

            events.push(json!({
                "kind": "trade",
                "player": current[&cell_a].name,
                "self": current[&cell_a].this || current[&cell_b].this,
                "other": current[&cell_b].name,
                "champion_id": new_a,
                "other_champion_id": new_b,
            }));
        }
    }

    traded
}

fn collect_champion_events(
    current: &HashMap<i64, MemberState>,
    champion_changes: &[(i64, i64, i64)],
    traded: &[i64],
    first_sync: bool,
    events: &mut Vec<Value>,
) {
    for (cell, old, new) in champion_changes {
        if traded.contains(cell) || *new <= 0 {
            continue;
        }

        events.push(json!({
            "kind": if *old > 0 { "champion" } else { "assign" },
            "player": current[cell].name,
            "self": current[cell].this,
            "champion_id": new,
        }));
    }

    if !first_sync {
        return;
    }

    for state in current.values() {
        if state.champion > 0 {
            events.push(json!({
                "kind": "assign",
                "player": state.name,
                "self": state.this,
                "champion_id": state.champion,
            }));
        }
    }
}

fn track_champion(app: &AppHandle, lcu_state: &tauri::State<'_, LcuHandle>, champion_id: i32) {
    let previous = lcu_state
        .current_champion
        .swap(champion_id, Ordering::Relaxed);

    if champion_id == previous {
        return;
    }

    let _ = app.emit("lcu-champion", champion_id);
}

async fn champion_pipeline(app: &AppHandle, connector: &Connector, champion_id: i32) {
    debug_assert!(champion_id > 0, "the pipeline runs for a picked champion");

    let lcu_state = app.state::<LcuHandle>();

    if champion_id != lcu_state.applied_champion.load(Ordering::Relaxed) {
        match lcu_apply(app, connector, champion_id).await {
            Ok(()) => lcu_state
                .applied_champion
                .store(champion_id, Ordering::Relaxed),
            Err(error) => {
                let _ = app.emit("lcu-error", error);
            }
        }
    }

    if champion_id != lcu_state.skinned_champion.load(Ordering::Relaxed) {
        match lcu_apply_skin(app, connector, champion_id).await {
            Ok(SkinApply::Applied | SkinApply::Skipped) => {
                lcu_state
                    .skinned_champion
                    .store(champion_id, Ordering::Relaxed);
            }
            Ok(SkinApply::Pending) => {}
            Err(error) => {
                lcu_state
                    .skinned_champion
                    .store(champion_id, Ordering::Relaxed);

                let _ = app.emit("lcu-error", error);
            }
        }
    }
}

async fn accept_ready_check(app: &AppHandle, connector: &Connector, data: &Value) {
    let state = app.state::<ApiState>();

    if !state.settings_snapshot().auto_accept {
        return;
    }

    let response = data
        .get("playerResponse")
        .and_then(Value::as_str)
        .unwrap_or("");

    let check_state = data.get("state").and_then(Value::as_str).unwrap_or("");
    let timer = data.get("timer").and_then(Value::as_f64).unwrap_or(0.0);

    if response != "None" || check_state != "InProgress" || timer < 1.0 {
        return;
    }

    let Some(client) = connector.client() else {
        return;
    };

    let accepted: Result<Value, _> = client.post(READY_CHECK_ACCEPT_URI, &Value::Null).await;

    if accepted.is_ok() {
        let _ = app.emit("lcu-accepted", ());

        let lcu_state = app.state::<LcuHandle>();

        push_event(
            app,
            &lcu_state,
            json!({
                "kind": "phase",
                "note": "senna accepted the match",
            }),
        );
    }
}

pub fn lcu_settings_changed(app: &AppHandle) {
    let lcu_state = app.state::<LcuHandle>();

    lcu_state.applied_champion.store(0, Ordering::Relaxed);
    lcu_state.skinned_champion.store(0, Ordering::Relaxed);

    if !lcu_state.in_select.load(Ordering::Relaxed) {
        return;
    }

    let champion_id = lcu_state.current_champion.load(Ordering::Relaxed);

    if champion_id <= 0 {
        return;
    }

    let app = app.clone();

    tauri::async_runtime::spawn(async move {
        let connector = Arc::clone(&app.state::<LcuHandle>().connector);

        champion_pipeline(&app, &connector, champion_id).await;
    });
}

fn selection_cleared() -> Value {
    json!({
        "in_select": false,
        "champion_id": 0,
        "skin_id": 0,
        "spell_d": 0,
        "spell_f": 0,
        "bench_enabled": false,
        "bench": [],
        "rerolls": 0,
        "game_mode": "",
    })
}

fn select_selection(session: &Value, mode: &str) -> Value {
    let cell = session
        .get("localPlayerCellId")
        .and_then(Value::as_i64)
        .unwrap_or(-1);

    let member = session
        .get("myTeam")
        .and_then(Value::as_array)
        .and_then(|team| {
            team.iter()
                .find(|member| member.get("cellId").and_then(Value::as_i64) == Some(cell))
        });

    let field = |name: &str| {
        member
            .and_then(|member| member.get(name))
            .and_then(Value::as_i64)
            .unwrap_or(0)
    };

    let bench: Vec<i64> = session
        .get("benchChampions")
        .and_then(Value::as_array)
        .map(|entries| {
            entries
                .iter()
                .filter_map(|entry| entry.get("championId").and_then(Value::as_i64))
                .filter(|id| *id > 0)
                .collect()
        })
        .unwrap_or_default();

    json!({
        "in_select": true,
        "champion_id": field("championId"),
        "skin_id": field("selectedSkinId"),
        "spell_d": field("spell1Id"),
        "spell_f": field("spell2Id"),
        "bench_enabled": session
            .get("benchEnabled")
            .and_then(Value::as_bool)
            .unwrap_or(false),
        "bench": bench,
        "rerolls": session
            .get("rerollsRemaining")
            .and_then(Value::as_i64)
            .unwrap_or(0),
        "game_mode": mode,
    })
}

fn set_selection(app: &AppHandle, lcu_state: &tauri::State<'_, LcuHandle>, selection: Value) {
    {
        let mut current = lcu_state
            .selection
            .write()
            .expect("selection lock poisoned");

        if *current == selection {
            return;
        }

        *current = selection.clone();
    }

    let _ = app.emit("lcu-selection", selection);
}

fn set_lobby(app: &AppHandle, lcu_state: &tauri::State<'_, LcuHandle>, roster: Value) {
    {
        let mut lobby = lcu_state.lobby.write().expect("lobby lock poisoned");

        if *lobby == roster {
            return;
        }

        *lobby = roster.clone();
    }

    let _ = app.emit("lcu-lobby", roster);
}

async fn initial_sync(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
) -> Option<(i32, bool)> {
    let client = connector.client()?;
    let phase: Value = client.get(GAMEFLOW_URI).await.ok()?;
    let phase = phase.as_str()?;

    if !PHASES_WITH_CHAMPION.contains(&phase) {
        return None;
    }

    if phase == "ChampSelect" {
        let session: Value = client.get(CHAMP_SELECT_URI).await.ok()?;

        sync_game_mode(connector, lcu_state).await;

        let mode = lcu_state
            .game_mode
            .read()
            .expect("game mode lock poisoned")
            .clone();

        set_lobby(app, lcu_state, select_roster(&session));
        set_selection(app, lcu_state, select_selection(&session, &mode));

        let champion_id = local_champion(&session)?;

        return Some((champion_id, true));
    }

    let session: Value = client.get(GAMEFLOW_SESSION_URI).await.ok()?;
    let (puuid, internal) = local_identity(&client).await;

    set_lobby(app, lcu_state, game_roster(&session, &puuid, &internal));

    let champion_id = selection_champion(&session, &puuid, &internal)?;

    Some((champion_id, false))
}

async fn sync_game_lobby(
    app: &AppHandle,
    connector: &Connector,
    lcu_state: &tauri::State<'_, LcuHandle>,
) {
    let Some(client) = connector.client() else {
        return;
    };

    let session: Value = match client.get(GAMEFLOW_SESSION_URI).await {
        Ok(session) => session,
        Err(_) => return,
    };

    let (puuid, internal) = local_identity(&client).await;

    set_lobby(app, lcu_state, game_roster(&session, &puuid, &internal));

    if let Some(champion_id) = selection_champion(&session, &puuid, &internal) {
        track_champion(app, lcu_state, champion_id);
    }
}

async fn local_identity(client: &Client) -> (String, String) {
    let Ok(summoner) = client.get::<Value>(SUMMONER_URI).await else {
        return (String::new(), String::new());
    };

    let puuid = summoner
        .get("puuid")
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string();

    let internal = summoner
        .get("internalName")
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string();

    (puuid, internal)
}

fn member_name(member: &Value) -> String {
    let game_name = member.get("gameName").and_then(Value::as_str).unwrap_or("");
    let tag = member.get("tagLine").and_then(Value::as_str).unwrap_or("");

    if !game_name.is_empty() {
        if tag.is_empty() {
            return game_name.to_string();
        }

        return format!("{game_name}#{tag}");
    }

    member
        .get("summonerName")
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string()
}

fn select_roster(session: &Value) -> Value {
    let local_cell = session
        .get("localPlayerCellId")
        .and_then(Value::as_i64)
        .unwrap_or(-1);

    let Some(team) = session.get("myTeam").and_then(Value::as_array) else {
        return Value::Array(Vec::new());
    };

    let players = team
        .iter()
        .map(|member| {
            let cell = member.get("cellId").and_then(Value::as_i64).unwrap_or(-1);

            json!({
                "puuid": member.get("puuid").and_then(Value::as_str).unwrap_or(""),
                "name": member_name(member),
                "champion_id": member.get("championId").and_then(Value::as_i64).unwrap_or(0),
                "team": 1,
                "self": cell >= 0 && cell == local_cell,
            })
        })
        .collect();

    Value::Array(players)
}

fn selection_pairs(session: &Value) -> Vec<(String, String, i64)> {
    let Some(selections) = session
        .get("gameData")
        .and_then(|data| data.get("playerChampionSelections"))
        .and_then(Value::as_array)
    else {
        return Vec::new();
    };

    selections
        .iter()
        .map(|entry| {
            let puuid = entry.get("puuid").and_then(Value::as_str).unwrap_or("");

            let internal = entry
                .get("summonerInternalName")
                .and_then(Value::as_str)
                .unwrap_or("");

            let champion = entry.get("championId").and_then(Value::as_i64).unwrap_or(0);

            (puuid.to_string(), internal.to_string(), champion)
        })
        .collect()
}

fn selection_champion(session: &Value, puuid: &str, internal: &str) -> Option<i32> {
    let champion = selection_pairs(session)
        .iter()
        .find(|(entry_puuid, entry_internal, _)| {
            (!puuid.is_empty() && entry_puuid == puuid)
                || (!internal.is_empty() && entry_internal == internal)
        })
        .map(|(_, _, champion)| *champion)?;

    if champion <= 0 {
        return None;
    }

    i32::try_from(champion).ok()
}

fn game_roster(session: &Value, puuid: &str, internal: &str) -> Value {
    let selections = selection_pairs(session);
    let mut players = Vec::new();

    for (key, team) in [("teamOne", 1), ("teamTwo", 2)] {
        let Some(members) = session
            .get("gameData")
            .and_then(|data| data.get(key))
            .and_then(Value::as_array)
        else {
            continue;
        };

        for member in members {
            let member_puuid = member.get("puuid").and_then(Value::as_str).unwrap_or("");

            let member_internal = member
                .get("summonerInternalName")
                .and_then(Value::as_str)
                .unwrap_or("");

            let mut champion = member
                .get("championId")
                .and_then(Value::as_i64)
                .unwrap_or(0);

            if champion <= 0 {
                champion = selections
                    .iter()
                    .find(|(entry_puuid, entry_internal, _)| {
                        (!member_puuid.is_empty() && entry_puuid == member_puuid)
                            || (!member_internal.is_empty() && entry_internal == member_internal)
                    })
                    .map_or(0, |(_, _, champion)| *champion);
            }

            let this = (!puuid.is_empty() && member_puuid == puuid)
                || (!internal.is_empty() && member_internal == internal);

            players.push(json!({
                "puuid": member_puuid,
                "name": member_name(member),
                "champion_id": champion,
                "team": team,
                "self": this,
            }));
        }
    }

    Value::Array(players)
}

fn local_champion(session: &Value) -> Option<i32> {
    let cell = session.get("localPlayerCellId")?.as_i64()?;

    locked_champion(session, cell).or_else(|| hovered_champion(session, cell))
}

fn locked_champion(session: &Value, cell: i64) -> Option<i32> {
    let team = session.get("myTeam")?.as_array()?;

    let member = team
        .iter()
        .find(|member| member.get("cellId").and_then(Value::as_i64) == Some(cell))?;

    let champion = member.get("championId")?.as_i64()?;

    if champion <= 0 {
        return None;
    }

    i32::try_from(champion).ok()
}

fn hovered_champion(session: &Value, cell: i64) -> Option<i32> {
    let rounds = session.get("actions")?.as_array()?;

    for round in rounds {
        let Some(actions) = round.as_array() else {
            continue;
        };

        for action in actions {
            let actor = action.get("actorCellId").and_then(Value::as_i64);
            let kind = action.get("type").and_then(Value::as_str);

            if actor != Some(cell) || kind != Some("pick") {
                continue;
            }

            let champion = action
                .get("championId")
                .and_then(Value::as_i64)
                .unwrap_or(0);

            if champion > 0 {
                return i32::try_from(champion).ok();
            }
        }
    }

    None
}

fn loadout_resolve(settings: &Settings, mode: &str, champion_id: i32) -> Loadout {
    debug_assert!(champion_id > 0, "a loadout resolves for a picked champion");

    let mut resolved = settings
        .loadouts
        .get(&champion_id)
        .cloned()
        .unwrap_or_default();

    let by_mode = settings
        .mode_loadouts
        .get(mode)
        .and_then(|by_champion| by_champion.get(&champion_id));

    let Some(entry) = by_mode else {
        return resolved;
    };

    if entry.skin_id.is_some() {
        resolved.skin_id = entry.skin_id;
    }

    if entry.spells.is_some() {
        resolved.spells = entry.spells;
    }

    if entry.runes.is_some() {
        resolved.runes.clone_from(&entry.runes);
    }

    if entry.items.is_some() {
        resolved.items.clone_from(&entry.items);
    }

    resolved
}

fn rune_page_valid(page: &RunePage) -> bool {
    if page.primary_style <= 0 {
        return false;
    }

    if page.sub_style <= 0 {
        return false;
    }

    if page.sub_style == page.primary_style {
        return false;
    }

    if page.perks.len() != PAGE_PERK_COUNT {
        return false;
    }

    page.perks.iter().all(|perk| *perk > 0)
}

struct ApplyPlan {
    spell_pair: Option<(i64, i64)>,
    runes_custom: Option<RunePage>,
    items_custom: Option<Vec<ItemBlock>>,
    wants_spells: bool,
    wants_runes: bool,
    wants_items: bool,
}

impl ApplyPlan {
    fn wants_any(&self) -> bool {
        self.wants_spells || self.wants_runes || self.wants_items
    }

    fn needs_build(&self) -> bool {
        (self.wants_spells && self.spell_pair.is_none())
            || (self.wants_runes && self.runes_custom.is_none())
            || (self.wants_items && self.items_custom.is_none())
    }
}

struct Applied {
    spells: bool,
    runes: bool,
    items: bool,
}

impl Applied {
    fn any(&self) -> bool {
        self.spells || self.runes || self.items
    }
}

fn plan_apply(settings: &Settings, loadout: &Loadout, mode: &str) -> ApplyPlan {
    let is_aram = mode.is_empty() || mode == MODE_ARAM;

    let spell_pair = loadout
        .spells
        .or_else(|| settings.mode_spells.get(mode).copied())
        .or(settings.default_spells);

    let runes_custom = loadout.runes.clone().filter(rune_page_valid);

    let items_custom = loadout
        .items
        .clone()
        .filter(|blocks| blocks.iter().any(|block| !block.items.is_empty()));

    let wants_spells =
        (settings.auto_spells && (spell_pair.is_some() || is_aram)) || loadout.spells.is_some();

    let wants_runes = (settings.auto_runes && is_aram) || runes_custom.is_some();
    let wants_items = (settings.auto_items && is_aram) || items_custom.is_some();

    ApplyPlan {
        spell_pair,
        runes_custom,
        items_custom,
        wants_spells,
        wants_runes,
        wants_items,
    }
}

async fn lcu_apply(app: &AppHandle, connector: &Connector, champion_id: i32) -> Result<(), String> {
    debug_assert!(champion_id > 0, "apply runs for a picked champion");

    let state = app.state::<ApiState>();
    let lcu_state = app.state::<LcuHandle>();
    let settings = state.settings_snapshot();

    let mode = lcu_state
        .game_mode
        .read()
        .expect("game mode lock poisoned")
        .clone();

    let loadout = loadout_resolve(&settings, &mode, champion_id);
    let plan = plan_apply(&settings, &loadout, &mode);

    if !plan.wants_any() {
        return Ok(());
    }

    let Some(client) = connector.client() else {
        return Err("league client is not available".to_string());
    };

    let build = if plan.needs_build() {
        let path = format!("/aram/champions/{champion_id}/build");

        get_json(&state, &path, &[]).await?
    } else {
        Value::Null
    };

    let applied = apply_plan(&client, &build, champion_id, &plan).await?;

    emit_applied(app, &lcu_state, champion_id, &build, &applied);

    Ok(())
}

async fn apply_plan(
    client: &Client,
    build: &Value,
    champion_id: i32,
    plan: &ApplyPlan,
) -> Result<Applied, String> {
    debug_assert!(champion_id > 0, "apply runs for a picked champion");

    let mut applied = Applied {
        spells: false,
        runes: false,
        items: false,
    };

    if plan.wants_spells {
        applied.spells = match plan.spell_pair {
            Some((spell_a, spell_b)) => apply_spell_pair(client, spell_a, spell_b).await?,
            None => apply_spells(client, build).await?,
        };
    }

    if plan.wants_runes {
        applied.runes = match &plan.runes_custom {
            Some(page) => apply_runes_custom(client, build, page).await?,
            None => apply_runes(client, build).await?,
        };
    }

    if plan.wants_items {
        applied.items = match &plan.items_custom {
            Some(items) => apply_items_custom(client, build, champion_id, items).await?,
            None => apply_items(client, build, champion_id).await?,
        };
    }

    Ok(applied)
}

fn emit_applied(
    app: &AppHandle,
    lcu_state: &tauri::State<'_, LcuHandle>,
    champion_id: i32,
    build: &Value,
    applied: &Applied,
) {
    let payload = json!({
        "champion_id": champion_id,
        "champion_name": build.get("champion_name").cloned().unwrap_or(Value::Null),
        "runes": applied.runes,
        "spells": applied.spells,
        "items": applied.items,
        "skin": false,
    });

    let _ = app.emit("lcu-applied", payload);

    if !applied.any() {
        return;
    }

    push_event(
        app,
        lcu_state,
        json!({
            "kind": "apply",
            "champion_id": champion_id,
            "runes": applied.runes,
            "spells": applied.spells,
            "items": applied.items,
            "skin": false,
        }),
    );
}

async fn lcu_apply_skin(
    app: &AppHandle,
    connector: &Connector,
    champion_id: i32,
) -> Result<SkinApply, String> {
    debug_assert!(champion_id > 0, "skin apply runs for a picked champion");

    let state = app.state::<ApiState>();
    let settings = state.settings_snapshot();

    let mode = app
        .state::<LcuHandle>()
        .game_mode
        .read()
        .expect("game mode lock poisoned")
        .clone();

    let loadout_skin = loadout_resolve(&settings, &mode, champion_id).skin_id;

    if loadout_skin.is_none() && !settings.random_skin {
        return Ok(SkinApply::Skipped);
    }

    let Some(client) = connector.client() else {
        return Err("league client is not available".to_string());
    };

    let Some(ids) = pickable_skin_ids(&client, champion_id).await? else {
        return Ok(SkinApply::Pending);
    };

    let Some(skin_id) = choose_skin(loadout_skin, &ids, champion_id) else {
        return Ok(SkinApply::Skipped);
    };

    let body = json!({ "selectedSkinId": skin_id });
    let result: Result<Value, _> = client.patch(MY_SELECTION_URI, &body).await;

    if result.is_err() {
        return Ok(SkinApply::Pending);
    }

    emit_skin_applied(app, champion_id);

    Ok(SkinApply::Applied)
}

async fn pickable_skin_ids(client: &Client, champion_id: i32) -> Result<Option<Vec<i64>>, String> {
    debug_assert!(
        champion_id > 0,
        "skin ids are fetched for a picked champion"
    );

    let Ok(session) = client.get::<Value>(CHAMP_SELECT_URI).await else {
        return Ok(None);
    };

    let cell = session
        .get("localPlayerCellId")
        .and_then(Value::as_i64)
        .unwrap_or(-1);

    if locked_champion(&session, cell) != Some(champion_id) {
        return Ok(None);
    }

    let pickable: Value = client
        .get(PICKABLE_SKINS_URI)
        .await
        .map_err(|error| error.to_string())?;

    let Some(entries) = pickable.as_array() else {
        return Ok(None);
    };

    let ids: Vec<i64> = entries.iter().filter_map(Value::as_i64).collect();

    if ids.is_empty() {
        return Ok(None);
    }

    let list_current = ids.iter().any(|skin| skin / 1000 == i64::from(champion_id));

    if !list_current {
        return Ok(None);
    }

    Ok(Some(ids))
}

fn choose_skin(loadout_skin: Option<i64>, ids: &[i64], champion_id: i32) -> Option<i64> {
    debug_assert!(champion_id > 0, "a skin is chosen for a picked champion");
    debug_assert!(!ids.is_empty(), "a skin is chosen from a non-empty list");

    if let Some(id) = loadout_skin {
        if !ids.contains(&id) {
            return None;
        }

        return Some(id);
    }

    let base = i64::from(champion_id) * 1000;

    let pool: Vec<i64> = ids
        .iter()
        .copied()
        .filter(|id| *id != base && *id / 1000 == i64::from(champion_id))
        .collect();

    pool.get(random_index(pool.len())).copied()
}

fn emit_skin_applied(app: &AppHandle, champion_id: i32) {
    let applied = json!({
        "champion_id": champion_id,
        "champion_name": Value::Null,
        "runes": false,
        "spells": false,
        "items": false,
        "skin": true,
    });

    let _ = app.emit("lcu-applied", applied);

    let lcu_state = app.state::<LcuHandle>();

    push_event(
        app,
        &lcu_state,
        json!({
            "kind": "apply",
            "champion_id": champion_id,
            "runes": false,
            "spells": false,
            "items": false,
            "skin": true,
        }),
    );
}

fn random_index(length: usize) -> usize {
    let nanos = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .map_or(0, |duration| duration.subsec_nanos());

    let index = nanos as usize % length.max(1);

    debug_assert!(index < length.max(1), "a random index stays in bounds");

    index
}

async fn apply_spell_pair(client: &Client, spell_a: i64, spell_b: i64) -> Result<bool, String> {
    if spell_a <= 0 || spell_b <= 0 || spell_a == spell_b {
        return Ok(false);
    }

    let body = json!({ "spell1Id": spell_a, "spell2Id": spell_b });

    let _: Value = client
        .patch(MY_SELECTION_URI, &body)
        .await
        .map_err(|error| error.to_string())?;

    Ok(true)
}

async fn apply_spells(client: &Client, build: &Value) -> Result<bool, String> {
    let Some(pair) = build.get("summoner_spells").and_then(|value| value.get(0)) else {
        return Ok(false);
    };

    let spell_a = pair.get("spell_a").and_then(Value::as_i64).unwrap_or(0);
    let spell_b = pair.get("spell_b").and_then(Value::as_i64).unwrap_or(0);

    apply_spell_pair(client, spell_a, spell_b).await
}

fn page_name(build: &Value) -> String {
    let champion = build
        .get("champion_name")
        .and_then(Value::as_str)
        .unwrap_or("custom");

    format!("{PAGE_NAME_PREFIX} {champion}")
}

async fn apply_runes(client: &Client, build: &Value) -> Result<bool, String> {
    let Some(page) = build.get("rune_pages").and_then(|value| value.get(0)) else {
        return Ok(false);
    };

    let primary_style = page
        .get("primary_style")
        .and_then(Value::as_i64)
        .unwrap_or(0);
    let sub_style = page.get("sub_style").and_then(Value::as_i64).unwrap_or(0);
    let perks = page_perk_ids(page);

    if primary_style <= 0 || sub_style <= 0 || perks.len() != PAGE_PERK_COUNT {
        return Ok(false);
    }

    rune_page_write(client, &page_name(build), primary_style, sub_style, &perks).await
}

async fn apply_runes_custom(
    client: &Client,
    build: &Value,
    page: &RunePage,
) -> Result<bool, String> {
    assert!(
        rune_page_valid(page),
        "custom rune page must be validated at resolve"
    );

    rune_page_write(
        client,
        &page_name(build),
        page.primary_style,
        page.sub_style,
        &page.perks,
    )
    .await
}

async fn rune_page_write(
    client: &Client,
    name: &str,
    primary_style: i64,
    sub_style: i64,
    perks: &[i64],
) -> Result<bool, String> {
    assert!(primary_style > 0, "a rune page needs a primary style");
    assert!(sub_style > 0, "a rune page needs a sub style");
    assert!(primary_style != sub_style, "primary and sub styles differ");
    assert!(
        perks.len() == PAGE_PERK_COUNT,
        "a rune page needs nine perks"
    );

    let body = json!({
        "name": name,
        "primaryStyleId": primary_style,
        "subStyleId": sub_style,
        "selectedPerkIds": perks,
        "current": true,
    });

    let pages: Value = client
        .get(PERK_PAGES_URI)
        .await
        .map_err(|error| error.to_string())?;

    let pages = pages.as_array().cloned().unwrap_or_default();

    let senna_id = pages
        .iter()
        .find(|entry| {
            entry
                .get("name")
                .and_then(Value::as_str)
                .is_some_and(|name| name.starts_with(PAGE_NAME_PREFIX))
        })
        .and_then(|entry| entry.get("id").and_then(Value::as_i64));

    if let Some(id) = senna_id {
        let uri = format!("{PERK_PAGES_URI}/{id}");

        let _: Value = client
            .delete(&uri)
            .await
            .map_err(|error| error.to_string())?;
    } else {
        ensure_page_slot(client, &pages).await?;
    }

    let created: Value = client
        .post(PERK_PAGES_URI, &body)
        .await
        .map_err(|error| error.to_string())?;

    if let Some(id) = created.get("id").and_then(Value::as_i64) {
        let _: Result<Value, _> = client.put(CURRENT_PAGE_URI, &json!(id)).await;
    }

    Ok(true)
}

async fn ensure_page_slot(client: &Client, pages: &[Value]) -> Result<(), String> {
    let inventory: Value = client
        .get(PERK_INVENTORY_URI)
        .await
        .map_err(|error| error.to_string())?;

    let owned = inventory
        .get("ownedPageCount")
        .and_then(Value::as_i64)
        .unwrap_or(0);

    let deletable: Vec<&Value> = pages
        .iter()
        .filter(|page| {
            page.get("isDeletable")
                .and_then(Value::as_bool)
                .unwrap_or(false)
        })
        .collect();

    let deletable_count = i64::try_from(deletable.len()).unwrap_or(i64::MAX);

    if owned <= 0 || deletable_count < owned {
        return Ok(());
    }

    let current = deletable
        .iter()
        .find(|page| {
            page.get("current")
                .and_then(Value::as_bool)
                .unwrap_or(false)
        })
        .copied();

    let Some(victim) = current.or_else(|| deletable.first().copied()) else {
        return Ok(());
    };

    let Some(id) = victim.get("id").and_then(Value::as_i64) else {
        return Ok(());
    };

    let uri = format!("{PERK_PAGES_URI}/{id}");

    let _: Value = client
        .delete(&uri)
        .await
        .map_err(|error| error.to_string())?;

    Ok(())
}

fn page_perk_ids(page: &Value) -> Vec<i64> {
    let mut perks: Vec<i64> = Vec::with_capacity(PAGE_PERK_COUNT);

    for section in ["primary", "sub", "shards"] {
        let Some(entries) = page.get(section).and_then(Value::as_array) else {
            continue;
        };

        for entry in entries {
            if let Some(id) = entry.get("id").and_then(Value::as_i64) {
                perks.push(id);
            }
        }
    }

    perks
}

async fn apply_items(client: &Client, build: &Value, champion_id: i32) -> Result<bool, String> {
    let blocks = item_set_blocks(build);

    if blocks.is_empty() {
        return Ok(false);
    }

    item_set_write(client, build, champion_id, blocks).await
}

async fn apply_items_custom(
    client: &Client,
    build: &Value,
    champion_id: i32,
    blocks_pref: &[ItemBlock],
) -> Result<bool, String> {
    let blocks: Vec<Value> = blocks_pref
        .iter()
        .filter_map(|block| {
            let name = block.name.trim();
            let label = if name.is_empty() { "Items" } else { name };

            item_block(label, &block.items)
        })
        .collect();

    assert!(
        !blocks.is_empty(),
        "custom item blocks must be validated at resolve"
    );

    item_set_write(client, build, champion_id, blocks).await
}

async fn item_set_write(
    client: &Client,
    build: &Value,
    champion_id: i32,
    blocks: Vec<Value>,
) -> Result<bool, String> {
    assert!(!blocks.is_empty(), "an item set needs at least one block");

    let summoner: Value = client
        .get(SUMMONER_URI)
        .await
        .map_err(|error| error.to_string())?;

    let Some(summoner_id) = summoner.get("summonerId").and_then(Value::as_i64) else {
        return Ok(false);
    };

    let uri = format!("{ITEM_SETS_URI}/{summoner_id}/sets");

    let current: Value = client.get(&uri).await.map_err(|error| error.to_string())?;

    let mut sets: Vec<Value> = current
        .get("itemSets")
        .and_then(Value::as_array)
        .cloned()
        .unwrap_or_default();

    sets.retain(|set| {
        !set.get("title")
            .and_then(Value::as_str)
            .is_some_and(|title| title.starts_with(PAGE_NAME_PREFIX))
    });

    let champion = build
        .get("champion_name")
        .and_then(Value::as_str)
        .unwrap_or("build");

    sets.push(json!({
        "associatedChampions": [champion_id],
        "associatedMaps": [],
        "blocks": blocks,
        "map": "any",
        "mode": "any",
        "preferredItemSlots": [],
        "sortrank": 0,
        "startedFrom": "blank",
        "title": format!("{PAGE_NAME_PREFIX} {champion}"),
        "type": "custom",
        "uid": format!("senna-{champion_id}"),
    }));

    let body = json!({
        "accountId": current.get("accountId").cloned().unwrap_or(json!(summoner_id)),
        "itemSets": sets,
        "timestamp": current.get("timestamp").cloned().unwrap_or(json!(0)),
    });

    let _: Value = client
        .put(&uri, &body)
        .await
        .map_err(|error| error.to_string())?;

    Ok(true)
}

fn item_ids_of(entry: Option<&Value>) -> Vec<i64> {
    let Some(items) = entry
        .and_then(|value| value.get("items"))
        .and_then(Value::as_array)
    else {
        return Vec::new();
    };

    items
        .iter()
        .filter_map(|item| item.get("id").and_then(Value::as_i64))
        .collect()
}

fn flat_ids_of(list: Option<&Value>, limit: usize) -> Vec<i64> {
    let Some(entries) = list.and_then(Value::as_array) else {
        return Vec::new();
    };

    entries
        .iter()
        .take(limit)
        .filter_map(|entry| entry.get("id").and_then(Value::as_i64))
        .collect()
}

fn item_block(label: &str, ids: &[i64]) -> Option<Value> {
    debug_assert!(!label.is_empty(), "an item block needs a label");

    if ids.is_empty() {
        return None;
    }

    let items: Vec<Value> = ids
        .iter()
        .map(|id| json!({ "count": 1, "id": id.to_string() }))
        .collect();

    Some(json!({ "type": label, "items": items }))
}

fn item_set_blocks(build: &Value) -> Vec<Value> {
    let starting = item_ids_of(build.get("starting_items").and_then(|value| value.get(0)));
    let core = item_ids_of(build.get("core_builds").and_then(|value| value.get(0)));
    let boots = flat_ids_of(build.get("boots"), 3);
    let situational = flat_ids_of(build.get("items"), 6);

    [
        item_block("Starting items", &starting),
        item_block("Core build", &core),
        item_block("Boots", &boots),
        item_block("Situational", &situational),
    ]
    .into_iter()
    .flatten()
    .collect()
}
