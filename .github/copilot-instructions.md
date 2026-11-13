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
- Expense and market records are managed as separate Dexie tables; the income page (`/income`) combines market income with the manual `onlineIncome` setting and subtracts `expenses` for a display-only profit. Keep these independent from order profit calculations unless a task explicitly asks to combine them.
- Manual online income is stored in the `settings` table under the `onlineIncome` key with an `updatedAt` timestamp; update it through the store action, not by writing Dexie from a view.
- The planner (`/planner`) keeps saved templates in the `plannerTemplates` table and renders pages in millimetres. `src/components/planner/PlannerSheet.vue` and `PlannerBlock.vue` are shared by the on-screen preview and the print container; do not introduce a second print-only layout. Paper size is applied at print time via a dynamic `@page { size: Wmm Hmm; margin: 0 }` rule.
- The tape generator (`/tape`) is pure canvas with no database persistence; pattern drawing, millimetre-to-pixel conversion and PNG/PSD export live in `src/utils/tape.ts`. Import `ag-psd` dynamically only when exporting PSD so it stays out of the initial bundle.

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
- After each modification, please commit the local code. Do not push to the remote repository unless the user explicitly asks for it.

## References
- Project overview and quickstart: `README.md`
- Route structure example: `src/router/index.ts`
- Store and business logic example: `src/stores/useShopStore.ts`
- IndexedDB schema definition: `src/db/appDb.ts`
