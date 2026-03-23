# Candle Engine – Frontend

React-based frontend for the Candle Engine trading chart application.

## Tech Stack

| Package | Purpose |
|---|---|
| Vite 8 | Build tool & dev server |
| React 19 | UI framework |
| Chakra UI v2 | Component library |
| ECharts 6 + echarts-for-react | Charting |
| Zustand 5 | State management |
| react-icons | Icon library |
| Axios | HTTP client |
| ESLint 9 | Linting |

## Quick Start

```bash
# Install dependencies
npm install

# Start development server (proxies /api to localhost:8080)
npm run dev

# Lint
npm run lint

# Production build
npm run build
```

> **Prerequisites**: The Go backend must be running on `http://localhost:8080`.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `VITE_API_URL` | `http://localhost:8080` | Backend API base URL |

Copy `.env.example` to `.env.local` and adjust as needed.

## Project Structure

```
src/
├── api/
│   └── client.js          # Axios API client
├── components/
│   ├── chart/
│   │   ├── chartEngine.js  # ECharts option builder (pure functions)
│   │   └── ChartContainer.jsx
│   ├── layout/
│   │   ├── Navbar.jsx
│   │   ├── ThemeToggle.jsx
│   │   ├── PairView.jsx
│   │   └── ProjectView.jsx
│   └── logs/
│       ├── LogEntry.jsx
│       ├── LogLevelFilters.jsx
│       └── LogPanel.jsx
├── store/
│   ├── projectStore.js    # Projects + render data (Zustand)
│   └── uiStore.js         # UI state + localStorage persistence
├── theme/
│   └── theme.js           # Chakra UI theme
├── App.jsx
├── index.css
└── main.jsx
```

## Chart Architecture

The chart engine builds a single ECharts instance with multiple stacked grids:

- **Main chart** (top): candlestick + overlay indicators (`targetChart === ""`)
- **Sub-charts** (below): grouped by `targetChart`, sorted alphabetically

All charts share the same X-axis (timestamps) and are linked via `axisPointer`
so hovering highlights the same timestamp across every panel.
