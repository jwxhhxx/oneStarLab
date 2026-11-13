# OneStarLab Project Guidelines

## Code Style
- Use Vue 3 SFC with `<script setup lang="ts">` and Composition API (`ref`, `reactive`, `computed`, `watch`).
- Keep business models centralized in `src/types/index.ts` and mirrored in `backend/internal/model/model.go`; update both when adding fields.
- Keep pricing/formatting helpers in pure utility functions (see `src/utils/pricing.ts`) instead of embedding formulas in views; authoritative pricing runs on the server.
- Prefer async/await and explicit error messages; backend returns Chinese `{ "error": "..." }` messages that the UI shows directly.
- UI text is Chinese; keep new UI copy consistent with existing Chinese labels.
- Go code follows standard `gofmt` formatting; handlers stay thin and business rules live in `backend/internal/store`.

## Architecture
- Frontend: Vue 3 + Pinia + Vue Router + Element Plus. Backend: Go `net/http` + SQLite (`modernc.org/sqlite`, no CGO).
- Boundaries:
  - Views in `src/views/*` handle presentation and user interactions.
  - Store in `src/stores/useShopStore.ts` is the application orchestration layer; it calls `src/api/client.ts` and refreshes state with `loadAll()`.
  - `src/api/client.ts` is the only place that performs HTTP requests; views must not fetch directly.
  - Routes live in `src/router/index.ts`; pages are lazy-loaded with `meta.title`.
  - Backend routes/middleware live in `backend/internal/api/api.go`; business rules and SQL live in `backend/internal/store/*.go`; schema and seed data live in `backend/internal/db/db.go`.
- Data flow: View → Pinia store → API client → HTTP `/api/v1` → handler → store → SQLite.
- Expense and market records are separate tables; the income page (`/income`) combines market income with the manual `onlineIncome` setting and subtracts `expenses` for a display-only profit. Keep these independent from order profit calculations unless a task explicitly asks to combine them.
- Manual online income is stored in `settings` under the `onlineIncome` key with an `updatedAt` timestamp; update it through the API/store, not by writing SQL from a view.
- The planner (`/planner`) keeps saved templates in `planner_templates` and renders pages in millimetres. `src/components/planner/PlannerSheet.vue` and `PlannerBlock.vue` are shared by the on-screen preview and the print container; do not introduce a second print-only layout. Paper size is applied at print time via a dynamic `@page { size: Wmm Hmm; margin: 0 }` rule.
- The tape generator (`/tape`) is pure canvas with no database persistence; pattern drawing, millimetre-to-pixel conversion and PNG/PSD export live in `src/utils/tape.ts`. Import `ag-psd` dynamically only when exporting PSD so it stays out of the initial bundle.

## Build and Test
- Install deps: `npm ci`
- Dev server: `npm run dev` (proxies `/api` to `http://127.0.0.1:8080`)
- Backend dev server: `cd backend && go run ./cmd/server`
- Production build: `npm run build` (includes `vue-tsc`)
- Backend checks: `cd backend && go build ./... && go vet ./...`
- Docker: `docker compose up --build -d`
- There is currently no dedicated lint/test script in `package.json`; do not assume `npm test` exists.

## Conventions
- After mutating data, refresh state via `loadAll()` so the UI matches the server.
- Keep multi-table writes (orders + stock + inventory log, snapshot import) inside a single SQL transaction in `backend/internal/store`.
- When changing the SQLite schema, update `backend/internal/db/db.go` and keep migrations idempotent (`CREATE TABLE IF NOT EXISTS`, guarded `ALTER TABLE`).
- Router pages use lazy import and route `meta.title`; keep title behavior consistent with `router.afterEach`.
- For third-party instances (for example ECharts), ensure mount/unmount lifecycle cleanup to avoid leaks.
- After each modification, commit the local code. Do not push to the remote repository unless the user explicitly asks for it.

## References
- Project overview and quickstart: `README.md`
- Architecture and data flow: `docs/ARCHITECTURE.md`
- REST API reference: `docs/API.md`
- Local acceptance checklist: `docs/ACCEPTANCE.md`
- Store and business logic example: `src/stores/useShopStore.ts`
- Backend business layer: `backend/internal/store`
