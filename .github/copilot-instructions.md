# OneStarLab Project Guidelines

## Code Style
- Use Vue 3 SFC with `<script setup lang="ts">` and Composition API (`ref`, `reactive`, `computed`, `watch`).
- Keep business models centralized in `src/types/index.ts`; update type definitions first when adding fields.
- Keep pricing calculation logic in pure utility functions (see `src/utils/pricing.ts`) instead of embedding formulas in views.
- Prefer async/await and explicit error throws/messages in store actions.
- UI text is Chinese; keep new UI copy consistent with existing Chinese labels.

## Architecture
- This is a pure frontend SPA: Vue + Pinia + Vue Router + Dexie (IndexedDB). No backend service.
- Boundaries:
  - Views in `src/views/*` handle presentation and user interactions.
  - Store in `src/stores/useShopStore.ts` is the application orchestration layer for CRUD and business flows.
  - Database schema/indexes are defined in `src/db/appDb.ts`.
  - Route map and page-level lazy loading live in `src/router/index.ts`.
- Data flow should remain: View -> Pinia store actions -> Dexie tables.
- Expense records are managed as a separate Dexie table and page flow; keep them independent from order profit calculations unless a task explicitly asks to combine them.

## Build and Test
- Install deps: `npm install`
- Dev server: `npm run dev`
- Production build (includes type-check): `npm run build`
- Preview build: `npm run preview`
- There is currently no dedicated lint/test script in `package.json`; do not assume `npm test` exists.

## Conventions
- After mutating Dexie tables, refresh state via `loadAll()` pattern in the store to keep UI data consistent.
- For inventory/order consistency, use Dexie transactions when one action updates both orders and products.
- When changing IndexedDB schema, increment Dexie `version()` in `src/db/appDb.ts` and define indexes carefully.
- Router pages use lazy import and route `meta.title`; keep title behavior consistent with `router.afterEach`.
- For third-party instances (for example ECharts), ensure mount/unmount lifecycle cleanup to avoid leaks.
- After each modification, please commit the local code, but do not push it directly to the remote repository.

## References
- Project overview and quickstart: `README.md`
- Route structure example: `src/router/index.ts`
- Store and business logic example: `src/stores/useShopStore.ts`
- IndexedDB schema definition: `src/db/appDb.ts`
