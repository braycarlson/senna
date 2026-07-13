use std::collections::BTreeMap;
use std::fs;
use std::path::PathBuf;

use serde::{Deserialize, Serialize};
use tauri::{AppHandle, Manager};

const SETTINGS_FILE_NAME: &str = "settings.json";
const BASE_URL_DEFAULT: &str = "http://127.0.0.1:8080";
const REGION_DEFAULT: &str = "NA1";
const SPELL_FLASH: i64 = 4;
const SPELL_GHOST: i64 = 6;

#[derive(Clone, Debug, Deserialize, Serialize)]
pub struct RunePage {
    pub primary_style: i64,
    pub sub_style: i64,
    pub perks: Vec<i64>,
}

#[derive(Clone, Debug, Deserialize, Serialize)]
pub struct ItemBlock {
    pub name: String,
    pub items: Vec<i64>,
}

fn item_blocks_compat<'de, D>(deserializer: D) -> Result<Option<Vec<ItemBlock>>, D::Error>
where
    D: serde::Deserializer<'de>,
{
    #[derive(Deserialize)]
    #[serde(untagged)]
    enum ItemsFormat {
        Blocks(Vec<ItemBlock>),
        Flat(Vec<i64>),
    }

    let parsed = Option::<ItemsFormat>::deserialize(deserializer)?;

    let blocks = parsed.map(|format| match format {
        ItemsFormat::Blocks(blocks) => blocks,
        ItemsFormat::Flat(items) => vec![ItemBlock {
            name: "Preferred build".to_string(),
            items,
        }],
    });

    Ok(blocks)
}

#[derive(Clone, Debug, Default, Deserialize, Serialize)]
pub struct Loadout {
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub skin_id: Option<i64>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub spells: Option<(i64, i64)>,
    #[serde(default, skip_serializing_if = "Option::is_none")]
    pub runes: Option<RunePage>,
    #[serde(
        default,
        skip_serializing_if = "Option::is_none",
        deserialize_with = "item_blocks_compat"
    )]
    pub items: Option<Vec<ItemBlock>>,
}

#[allow(clippy::struct_excessive_bools)]
#[derive(Clone, Debug, Deserialize, Serialize)]
pub struct Settings {
    pub base_url: String,
    pub token: String,
    pub region: String,
    #[serde(default = "enabled_default")]
    pub auto_accept: bool,
    #[serde(default = "enabled_default")]
    pub auto_runes: bool,
    #[serde(default = "enabled_default")]
    pub auto_spells: bool,
    #[serde(default = "enabled_default")]
    pub auto_items: bool,
    #[serde(default = "default_spells_default")]
    pub default_spells: Option<(i64, i64)>,
    #[serde(default)]
    pub mode_spells: BTreeMap<String, (i64, i64)>,
    #[serde(default)]
    pub random_skin: bool,
    #[serde(default)]
    pub loadouts: BTreeMap<i32, Loadout>,
    #[serde(default)]
    pub mode_loadouts: BTreeMap<String, BTreeMap<i32, Loadout>>,
}

fn enabled_default() -> bool {
    true
}

#[allow(clippy::unnecessary_wraps)]
fn default_spells_default() -> Option<(i64, i64)> {
    Some((SPELL_FLASH, SPELL_GHOST))
}

impl Default for Settings {
    fn default() -> Settings {
        Settings {
            base_url: BASE_URL_DEFAULT.to_string(),
            token: String::new(),
            region: REGION_DEFAULT.to_string(),
            auto_accept: true,
            auto_runes: true,
            auto_spells: true,
            auto_items: true,
            default_spells: default_spells_default(),
            mode_spells: BTreeMap::new(),
            random_skin: false,
            loadouts: BTreeMap::new(),
            mode_loadouts: BTreeMap::new(),
        }
    }
}

fn settings_path(app: &AppHandle) -> Result<PathBuf, String> {
    let dir = app
        .path()
        .app_config_dir()
        .map_err(|error| error.to_string())?;

    Ok(dir.join(SETTINGS_FILE_NAME))
}

pub fn settings_load(app: &AppHandle) -> Settings {
    let Ok(path) = settings_path(app) else {
        return Settings::default();
    };

    let Ok(raw) = fs::read_to_string(path) else {
        return Settings::default();
    };

    serde_json::from_str(&raw).unwrap_or_default()
}

pub fn settings_save(app: &AppHandle, settings: &Settings) -> Result<(), String> {
    assert!(!settings.base_url.is_empty(), "base url must not be empty");

    let path = settings_path(app)?;

    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent).map_err(|error| error.to_string())?;
    }

    let raw = serde_json::to_string_pretty(settings).map_err(|error| error.to_string())?;

    fs::write(path, raw).map_err(|error| error.to_string())?;

    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn loadout_spells_roundtrip() {
        let raw = r#"{"base_url":"http://x","token":"","region":"NA1","auto_runes":true,"auto_spells":true,"auto_items":true,"default_spells":[4,6],"random_skin":true,"loadouts":{"12":{"skin_id":12003,"spells":[4,6]}}}"#;
        let settings: Settings = serde_json::from_str(raw).expect("settings parse");

        assert_eq!(
            settings.loadouts.get(&12).expect("loadout").spells,
            Some((4, 6))
        );

        let out = serde_json::to_string(&settings).expect("settings serialize");

        assert!(out.contains("\"spells\":[4,6]"), "spells lost: {out}");
    }

    #[test]
    fn mode_loadout_roundtrip() {
        let raw = r#"{"base_url":"http://x","token":"","region":"NA1","mode_loadouts":{"ARAM":{"22":{"spells":[4,32],"runes":{"primary_style":8100,"sub_style":8200,"perks":[8112,8139,8138,8135,8226,8236,5008,5008,5011]},"items":[3006,6672]}}}}"#;
        let settings: Settings = serde_json::from_str(raw).expect("settings parse");

        let loadout = settings
            .mode_loadouts
            .get("ARAM")
            .and_then(|by_champion| by_champion.get(&22))
            .expect("mode loadout");

        assert_eq!(loadout.spells, Some((4, 32)));

        let blocks = loadout.items.as_ref().expect("items migrate to blocks");

        assert_eq!(blocks.len(), 1);
        assert_eq!(blocks[0].name, "Preferred build");
        assert_eq!(blocks[0].items, vec![3006, 6672]);

        let runes = loadout.runes.as_ref().expect("runes");

        assert_eq!(runes.primary_style, 8100);
        assert_eq!(runes.perks.len(), 9);

        let out = serde_json::to_string(&settings).expect("settings serialize");

        assert!(
            out.contains("\"mode_loadouts\""),
            "mode loadouts lost: {out}"
        );
    }

    #[test]
    fn item_blocks_roundtrip() {
        let raw = r#"{"base_url":"http://x","token":"","region":"NA1","loadouts":{"22":{"items":[{"name":"Start","items":[1055]},{"name":"Core","items":[3031,3094]}]}}}"#;
        let settings: Settings = serde_json::from_str(raw).expect("settings parse");

        let blocks = settings
            .loadouts
            .get(&22)
            .and_then(|loadout| loadout.items.as_ref())
            .expect("item blocks");

        assert_eq!(blocks.len(), 2);
        assert_eq!(blocks[1].name, "Core");
        assert_eq!(blocks[1].items, vec![3031, 3094]);

        let out = serde_json::to_string(&settings).expect("settings serialize");

        assert!(out.contains("\"name\":\"Core\""), "blocks lost: {out}");
    }

    #[test]
    fn settings_without_mode_loadouts_parses() {
        let raw = r#"{"base_url":"http://x","token":"","region":"NA1"}"#;
        let settings: Settings = serde_json::from_str(raw).expect("settings parse");

        assert!(settings.mode_loadouts.is_empty());
        assert!(settings.loadouts.is_empty());
    }
}
