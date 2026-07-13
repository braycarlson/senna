mod api;
mod lcu;
mod settings;

use tauri::{AppHandle, Manager, State};

use crate::api::ApiState;
use crate::settings::{Settings, settings_load, settings_save};

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
fn settings_get(state: State<'_, ApiState>) -> Settings {
    state
        .settings
        .read()
        .expect("settings lock poisoned")
        .clone()
}

#[allow(clippy::needless_pass_by_value)]
#[tauri::command]
fn settings_set(
    app: AppHandle,
    state: State<'_, ApiState>,
    settings: Settings,
) -> Result<(), String> {
    if settings.base_url.trim().is_empty() {
        return Err("base url must not be empty".to_string());
    }

    settings_save(&app, &settings)?;

    {
        let mut current = state.settings.write().expect("settings lock poisoned");

        *current = settings;
    }

    lcu::lcu_settings_changed(&app);

    Ok(())
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .setup(|app| {
            let settings = settings_load(app.handle());

            app.manage(ApiState::new(settings));

            lcu::lcu_spawn(app.handle());

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            settings_get,
            settings_set,
            lcu::lcu_status,
            lcu::lcu_apply_stats_runes,
            lcu::lcu_bench_swap,
            lcu::lcu_current_champion,
            lcu::lcu_current_page,
            lcu::lcu_current_summoner,
            lcu::lcu_events,
            lcu::lcu_lobby,
            lcu::lcu_owned_skins,
            lcu::lcu_pickable_skins,
            lcu::lcu_reroll,
            lcu::lcu_selection,
            lcu::lcu_set_skin,
            lcu::lcu_set_spells,
            api::api_patches,
            api::api_stats,
            api::api_champions,
            api::api_tier,
            api::api_build,
            api::api_match,
            api::api_matchups,
            api::api_synergies,
            api::api_player,
            api::api_player_champions,
            api::api_player_refresh,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
