# Dynamic Project Creation — Refactor Plan

Goal: replace hardcoded project setup (currently in `cmd/webserver/sample/`) with full CRUD through the REST API, backed by a persistent-friendly data model.

---

## 1. `pkg/sources/models.go`

Add `json` struct tags to every field that must survive serialisation:
- `SourcePairId` → `SourceName`, `PairId`
- `TimeRange` → `Start`, `End`
- `CandleColorConfig` → `BullishColor`, `BearishColor`

---

## 2. `pkg/registry`

The registry becomes the **single source of truth for available implementations** (sources and indicators registered at startup). It also absorbs the data-fetching responsibilities currently in `ChartSourceManager`.

Changes:
- Move `sourcemanager.go` from `pkg/chartproject/` into this package and rename it `sourceresolver.go` (or similar). Expose its fetch/align logic as methods on `Registry` (or as a standalone `SourceResolver` struct held by the registry).
- `Registry` now has two distinct concerns:
  1. **Implementation registry** — `RegisterDataSources`, `RegisterIndicators`, `GetDataSource`, `GetIndicator`, `ListDataSources`, `ListIndicators` (already exists).
  2. **Fetch/resolve** — given a list of `SourcePairId`s and a time range/timeframe, resolve the actual `Source` objects from the registry and fetch+align data. This is `ChartSourceManager.FetchSources` + alignment logic moved here.
- `GetCommonTimeFrames` and `GetTimeFrameSeconds` move here too (operating on a given subset of source names rather than a pre-loaded map).
- The registry is initialised once at startup (in `main.go`) and injected where needed.

---

## 3. `pkg/chartproject`

`ChartProject` becomes a **lightweight, serialisable config object**. It no longer owns source objects or indicator implementations.

### `ChartProject` struct (persistent fields only)
```
id             string                   `json:"id"`           // UUID
title          string                   `json:"title"`
selectedPairs  []sources.SourcePairId   `json:"selectedPairs"`
indicatorNames []string                 `json:"indicatorNames"`
timeFrame      string                   `json:"timeFrame"`
timeRange      sources.TimeRange        `json:"timeRange"`
colors         sources.CandleColorConfig `json:"colors"`
```

Remove the `sourceManager *ChartSourceManager` field entirely.
Remove the `indicators []indicators.BaseIndicator` field; replace with `indicatorNames []string`.

### Mutation methods (all thread-safe)
- `AddPair(pair sources.SourcePairId)`
- `RemovePair(pair sources.SourcePairId)`
- `AddIndicatorByName(name string)`
- `RemoveIndicatorByName(name string)`
- `UpdateConfig(cfg ChartProjectConfig)` — update timeframe/timerange/colors

### Registry field (non-persistent)
`ChartProject` holds a `registry *registry.Registry` field that is **not** serialised to JSON (no `json` tag, or explicitly `json:"-"`). It is injected after load (e.g. `project.SetRegistry(reg)`) so the project can call `c.registry.GetDataSource(...)` / `c.registry.GetIndicator(...)` internally without needing it passed on every call. Any name not found in the registry is surfaced as a `ProjectError`.

### `GetRender` behaviour change
When `RenderConfig` contains non-zero override values (timeFrame, rangeFrom/rangeTo, candleColors), `GetRender` **updates the corresponding fields on the struct** rather than using them only for the current render. This means a PATCH to the project settings and a render with overrides share the same write path, keeping the project state consistent.

### `ChartProjectConfig` / `models.go`
Add `json` tags to `ChartProjectConfig` and `RenderConfig` fields that are user-facing.

### Delete `sourcemanager.go`
All logic moves to `pkg/registry`.

---

## 4. `pkg/chartmanager`

The manager stores projects by **string ID** instead of slice index so that IDs are stable across additions/deletions.

Changes:
- Change internal storage from `[]ChartProject` to `map[string]*ChartProject` (pointer so mutations are reflected without re-inserting).
- Add `CreateProject(title string, cfg chartproject.ChartProjectConfig) *ChartProject` — generates a UUID, creates the project, stores it.
- Add `DeleteProject(id string) bool`.
- Add `GetProjectByID(id string) (*ChartProject, bool)`.
- Hold a reference to `*registry.Registry` so it can pass it into `GetRender`.
- Add `SaveToFile(path string) error` / `LoadFromFile(path string) error` for JSON persistence of all projects (serialize `map[string]*ChartProject`).
- `GetChartProjectsList` returns `[]ProjectSummary{ID, Title}` using the string IDs.

---

## 5. `pkg/webserver/api`

### New routes
```
POST   /api/v1/projects                        — create a project
DELETE /api/v1/projects/:id                    — delete a project
PATCH  /api/v1/projects/:id                    — partially update title/timeframe/timerange/colors
POST   /api/v1/projects/:id/pairs              — add a trading pair
DELETE /api/v1/projects/:id/pairs              — remove a trading pair
POST   /api/v1/projects/:id/indicators         — add an indicator by name
DELETE /api/v1/projects/:id/indicators/:name   — remove an indicator
GET    /api/v1/sources                         — list registered data sources
GET    /api/v1/sources/:source/pairs           — list valid trading pairs for a source
GET    /api/v1/indicators                      — list registered indicators
```

Existing routes keep their shape but switch from integer index to string UUID in the `:id` path parameter.

### New request models (in `models/`)
- `CreateProjectRequest` — `title`, optional `timeFrame`, `timeRange`, `candleColors`
- `UpdateProjectRequest` — same optional fields, all nullable/zero-omittable
- `AddPairRequest` — `sourceName`, `pairId`
- `AddIndicatorRequest` — `name`

### `handlers/projects.go`
Replace index-based lookups with ID-based lookups. Wire `CreateProject`, `DeleteProject`, `UpdateProject`, `AddPair`, `RemovePair`, `AddIndicator`, `RemoveIndicator` handlers.

### `handlers/render.go`
No registry plumbing needed here — the project already holds its registry reference. `GetRender(cfg)` signature stays the same; any overrides in `cfg` are written back to the project struct by `GetRender` itself.

---

## 6. `cmd/webserver/main.go`

- Instantiate `registry.Registry`, register all sources and indicators at startup.
- Instantiate `chartmanager.ChartManager` with the registry reference.
- Attempt to load persisted projects from a JSON file on startup; if absent, start empty.
- Remove all calls to `sample.Get*` — the sample package can be kept for integration tests or deleted.

