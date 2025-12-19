This frontend uses `svelte` + `vite` and includes a small local SPA router compatible with the Wails desktop environment.

Router behavior

- The app uses a local router (`src/lib/local-router/Router.svelte`) so the app can build and run without external network dependencies.

Commands

- Development server: `npm run dev`
- Build for production: `npm run build`

Notes

- The app includes lightweight page components in `src/pages` which call the Wails client (`wailsjs/go/backend/IndicoClient.js`) to fetch data. The local router shim is intentionally kept to allow building and testing without an external router dependency.
