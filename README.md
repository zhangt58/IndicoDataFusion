# IndicoDataFusion

A desktop app to aggregate and browse Indico conference data (events, contributions, abstracts) from multiple Indico instances.

Designed for conference organizers and attendees who want a lightweight offline-capable viewer for Indico conference content.

## Key Features

- Aggregate data from multiple Indico servers (custom base URLs supported).
- Fetch and view event details, contributions, and abstracts in one place.
- Lightweight offline cache for faster repeat access.
- Simple UI for adding and managing Indico data sources.

## Quick Start

1. Installer packages are provided for Linux, Windows and macOS platforms.

2. Add an Indico data source:
   - Click "(+) Add" in the `Settings -> Configuration -> Data Sources`.
   - Provide a Name, Base URL (or pick a suggested provider), and Event ID.
   - Provide an API token events and adjust the request timeout.
   - See **API Token Management** below for details on managing API tokens.

3. Use the sidebar to browse loaded events, contributions, and abstracts.

## API Token Management

Overview

- Tokens are stored in the OS keyring (recommended) and referenced from the YAML config by a logical name.
- The config no longer stores raw API tokens on data-source entries. Instead, a top-level `api-tokens` list can include named entries and the `indico` data source must reference the token by `api_token_name`.
- At runtime the app resolves a named token in this order: 1) the `token` field on the top-level `api-tokens` entry (YAML, only used temporarily), 2) an OS keyring secret stored under the token entry `name`.
- If no token is available the app records a non-fatal init problem so the UI can prompt you to add the missing token.

Config example (YAML)

- Top-level api-tokens section (recommended to keep `token` empty and use keyring):
```yaml
  api-tokens:
    - name: bot
      base_url: https://example.indico.test
      username: bot
      token: ""  # leave empty and store the real secret in keyring
```

- Data source referencing the named token:

```yaml
  myconference:
    indico: true
    base_url: https://example.indico.test
    event_id: 42
    api_token_name: bot
    timeout: 7s
```

How the app resolves tokens

- The app looks up the `api_token_name` in the top-level `api-tokens` list by `name`.
- If the matched entry has a non-empty `token` field in YAML it will be used (this is kept for transition/migration only).
- If the YAML `token` is empty the app will try to read a secret from the OS keyring using the entry `name`.
- If neither is available the data source initialization will produce a non-fatal problem; the UI should allow adding/supplying the secret interactively.

Security recommendations

- Prefer the OS keyring for secrets; avoid committing raw tokens into `config/*.yaml`.
- Use distinct token names per base URL / username to keep names unique and self-explanatory.

## Troubleshooting

- If a source fails to load, check the Base URL and Event ID first.
- Ensure your API token is valid and has the required permissions.