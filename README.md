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
   - Click "Add Indico Data Source" in the app.
   - Provide a Name, Base URL (or pick a suggested provider), and Event ID.
   - Provide an API token events and adjust the request timeout.

3. Use the sidebar to browse loaded events, contributions, and abstracts.

## Troubleshooting

- If a source fails to load, check the Base URL and Event ID first.
- Ensure your API token is valid and has the required permissions.