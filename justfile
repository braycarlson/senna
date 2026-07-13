set windows-shell := ["cmd.exe", "/c"]

default:
    just --list

build:
    bun run tauri build

check:
    cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets

dev:
    bun run tauri dev

web:
    bun run dev
