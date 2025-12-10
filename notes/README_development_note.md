# Development Note — Previous README

This file contains the previous contents of the project's top-level README.md. It was moved here to keep a development/history record while a new user-focused README.md is created.

---

# IndicoDataFusion

Aggregating the Indico data into one desktop app

## About

This is a Wails desktop application with Svelte 5 as the frontend framework.

## Technology Stack

- **Backend**: Go 1.25.4
- **Frontend Framework**: Svelte 5.43.5
- **Build Tool**: Vite 7.2.2
- **Desktop Framework**: Wails v2.11.0
- **UI**: Svelte with Vite

## Prerequisites

Before you begin, ensure you have the following installed:

- Go 1.24 or later
- Node.js v24 or later
- npm 11 or later
- Wails CLI v2.11.0

### Linux Dependencies

On Linux systems, you'll also need:

```bash
sudo apt install libgtk-3-dev libwebkit2gtk-4.1-dev
```

## Installation

### Install Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Install Frontend Dependencies

```bash
cd frontend
npm install
```

## Development

To run in live development mode, run the following command in the project directory:

```bash
wails dev
```

This will run a Vite development server that will provide very fast hot reload of your frontend changes. 
If you want to develop in a browser and have access to your Go methods, there is also a dev server that 
runs on http://localhost:34115. Connect to this in your browser, and you can call your Go code from devtools.

## Building

To build a redistributable, production mode package, use:

```bash
wails build
```

The built application will be available in `build/bin/`.

## License

See LICENSE file for details.

