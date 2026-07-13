set windows-shell := ["cmd.exe", "/c"]

default:
    just --list

build:
    bun run tauri build

check:
    cargo clippy --manifest-path src-tauri/Cargo.toml --all-targets

dev:
    bun run tauri dev

pages:
    python -m http.server 8000 -d docs

web:
    bun run dev
