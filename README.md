# go-automater (ExternalSSD Process Manager)

A small macOS utility with a local web UI that finds processes locking a mounted volume (default: `/Volumes/ExternalSSD`), safely terminates known development tools, and attempts to unmount/eject the drive. Built in Go with embedded static assets for easy distribution.

## Overview
- **Purpose:** Help cleanly eject an external SSD by terminating only safe, whitelisted dev tools used by the current user.
- **How it works:**
  - Scans running processes referencing files under the target volume using `lsof`.
  - Filters to the current user and matches a whitelist (e.g., Node, npm, pnpm, yarn, VS Code, IntelliJ, Java).
  - On run, sends `SIGTERM` to matched processes, then tries `diskutil unmount`; if needed, falls back to `diskutil unmount force`.
- **UI:** Simple Tailwind-based page served locally; see [public/index.html](public/index.html).

## Features
- **Dry-run mode:** Preview what would be terminated without making changes.
- **Safe termination:** Sends `SIGTERM` to whitelisted processes only for the current user.
- **Automated unmount:** Attempts normal unmount, then force-eject if required.
- **Prod/Dev serving:** Embeds static assets for production; serves from filesystem in development.

## Requirements
- **OS:** macOS (uses `lsof`, `ps`, `whoami`, `diskutil`).
- **Go:** Go 1.25+ (see [go.mod](go.mod)).
- **Network:** Local only; the UI runs on `localhost`.

## Quick Start
Using the provided Makefile targets:

```bash
# Development mode (serves ./public from filesystem)
make dev

# Production mode (builds, embeds assets, runs binary)
make prod

# Build + run (production defaults)
make run

# Clean build artifacts
make clean
```

Environment variables:
- **`ENV`:** `development` or `production` (defaults to `production`).
- **`PORT`:** HTTP port (defaults to `3000`).

Examples:
```bash
# Dev on a custom port
PORT=4000 ENV=development make dev

# Prod binary on default port
make prod
```

Then open: http://localhost:3000 (or your chosen `PORT`).

## Configuration
- **Target volume:** Set in [process_handler.go](process_handler.go) as `targetVolume = "/Volumes/ExternalSSD"`.
  - Change this constant to point at a different mount path if needed.
- **Whitelist:** Controlled by `whitelist` regex in [process_handler.go](process_handler.go) to match safe dev tools.

## API
The server exposes two endpoints used by the UI:

- **GET `/api/dry-run`**
  - Runs discovery only and returns a JSON `Result` without terminating processes or attempting unmount.
- **POST `/api/run`**
  - Terminates matched processes for the current user and attempts to unmount/eject the target volume.

`Result` shape (see [process_handler.go](process_handler.go)):
- **`dryRun`:** boolean — whether this was a dry run
- **`processes`:** array of `{ pid, command, user, action }`
- **`unmountAttempted`:** boolean — unmount logic was triggered
- **`unmountSuccess`:** boolean — unmount/eject succeeded
- **`forceEjectUsed`:** boolean — force-eject was required
- **`message`:** string — status message

Notes:
- The handlers do not enforce HTTP method; the UI uses `GET` for dry-run and `POST` for run.

## Web UI
- Located in [public/index.html](public/index.html); styled with Tailwind via CDN.
- Buttons:
  - **Dry Run:** Calls `/api/dry-run`, renders the process table, and shows an informational banner.
  - **Run Cleanup:** Calls `/api/run`, renders results, and shows success/warning/error banners based on unmount outcome.

## Project Structure
- [main.go](main.go): Server bootstrap, static file serving, endpoint wiring.
- [config.go](config.go): Environment and port resolution with defaults.
- [process_handler.go](process_handler.go): Core logic for process discovery, termination, and unmount.
- [public/index.html](public/index.html): Lightweight UI.
- [Makefile](Makefile): Common dev/build/test tasks.

## Development
Useful Make targets:
```bash
make fmt          # Format code
make vet          # Static analysis
make test         # Run tests (if present)
make deps tidy    # Manage modules
make upgrade      # Update dependencies
make watch        # Auto-reload dev mode (requires `entr`)
```

Info and helpers:
- **Dev serving:** In `development`, static files are served from `./public` for rapid iteration.
- **Prod serving:** In `production`, assets are embedded via `go:embed` (see [main.go](main.go)).
- **Binary output:** Built to `./bin/automater`.

## Safety & Caveats
- **User scope:** Only processes owned by the current user are considered.
- **Signals:** Uses `SIGTERM` (graceful); you can adjust behavior if needed.
- **Force eject:** Safe but not clean; prefer normal unmount when possible.
- **Permissions:** Some operations may require sufficient privileges depending on system configuration.
- **Docker:** Makefile has Docker targets, but a Dockerfile is not included; add one if you plan to containerize.

## License
No explicit license file is present. If you plan to distribute, consider adding one.
