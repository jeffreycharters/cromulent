```markdown
# Cromulent — project memory document

## What is this

A reset document for continuing development of Cromulent with Claude. Paste this at the start of a new conversation to restore context.

---

## Github repository

https://github.com/jeffreycharters/jcc

---

## Communication style

- Concise answers preferred — no rambling
- Considered pushback welcome, don't just agree
- Technically capable — Go developer with reasonable experience
- Hobby project, wants to understand everything under the hood
- Likes "if you get it, you get it" humour (Simpsons-brained)
- Work one file at a time, one change at a time

---

## Project overview

**Cromulent** — a desktop SPC (statistical process control) charting and review application for an ISO 17025 accredited laboratory. Named after the Simpsons word. Tagline: *"perfectly cromulent data, every time."*

Replaces Northwest Analytical (NWA) Quality Analyst, which is slow, ugly, has no chart annotations, and doesn't enforce data locking.

### Core use case
- Technicians enter measurement data
- Reviewer (user) opens charts, checks for trends, adds traceable comments/rationale
- Data is locked (append-only) once entered
- Full audit trail for ISO 17025 compliance

---

## Tech stack

| Layer | Choice |
|---|---|
| Framework | Wails v2 (Go backend, web frontend) |
| Language | Go 1.26 |
| Frontend | Svelte 5 + TypeScript (plain Svelte, NOT SvelteKit — single view variable for routing) |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGO) |
| Auth | Username + password (bcrypt via `golang.org/x/crypto/bcrypt`) |
| Charts | Chart.js + chartjs-plugin-annotation (not yet implemented) |
| Font | Atkinson Hyperlegible Next + Mono (woff2, bundled in frontend assets) |
| Package manager | pnpm |
| Target OS | Windows only (dev on Linux, cross-compile with `-platform windows/amd64`) |
| Dev command | `wails dev -tags webkit2_41` |
| DnD | `svelte-dnd-action` |

### SQLite notes
- WAL mode enabled (`PRAGMA journal_mode=WAL`)
- Multi-user via network share (2-3 users per lab, low write contention)
- Network share path stored in local config file per workstation
- Append-only measurements table enforces data integrity architecturally
- Migrations via `PRAGMA user_version` — versioned migration slice in `db/migrate.go`
- Concurrent writes solved by using transactions rather than `Promise.all` parallel calls

### Go project structure
```
db/        — SQLite setup and migrations
models/    — data structs
handlers/  — functions exposed to Svelte via Wails bindings
frontend/  — Svelte app
```

---

## Authentication

- Username + password stored in SQLite `users` table
- Passwords bcrypt hashed
- Minimum 6 character password (threat model is audit trail integrity, not internet exposure)
- Session timeout: re-enter password after 30 minutes idle
- No magic links, no device-bound auth — credentials must work on any workstation
- Inactive users (active = 0) cannot log in — soft delete preserves audit trail

---

## Roles

| Role | Description |
|---|---|
| `technician` | Enters data, lands on data entry view |
| `reviewer` | Reviews charts, lands on chart review view |
| `supervisor` | Reviews charts, lands on chart review view |
| `admin` | Full access including user management |

All roles can create methods, materials, analytes, and MMA combos. Trust model is audit trail integrity, not access control — user_id on everything means bad actors are traceable.

---

## Database schema (migration v4)

### Core lookup tables

**`users`** — `id`, `username`, `password_hash`, `role`, `active` (default 1), `created_at`

**`materials`** — `id`, `name`, `description`

**`methods`** — `id`, `name`, `description`

**`analytes`** — `id`, `name`, `unit` — UNIQUE(name, unit) so molybdenum/ppm and molybdenum/ppb are distinct

### Relationships

**`material_method_analytes`** — `id`, `material_id FK`, `method_id FK`, `analyte_id FK`, `display_order`, `active` (default 1)
- UNIQUE(material_id, method_id, analyte_id)
- The unique combination that owns its own control limits and measurement history
- `active` allows soft-inactivation of retired combos — operates at combo level (all analytes), not per-analyte
- No deletes once referenced by measurements (FK enforcement)

**`control_limit_regions`** — `id`, `material_method_analyte_id FK`, `mean`, `ucl` (NOT NULL), `lcl` (NOT NULL), `uwl`, `lwl`, `uil`, `lil`, `effective_from_sequence`, `created_by FK`, `created_at`, `deleted_at`, `deleted_by FK`
- Soft delete only — `deleted_at`/`deleted_by` set on deletion, never hard deleted
- This preserves audit trail; audit log view can query `deleted_at IS NOT NULL` later

Three pairs of limits:
- `ucl`/`lcl` — control limits (required, typically ±3 SD)
- `uwl`/`lwl` — warning limits (optional, typically ±2 SD)
- `uil`/`lil` — inner limits (optional, typically ±1 SD)

Limits versioned by sequence number. To find limits for sequence N: query where `effective_from_sequence <= N` ORDER BY `effective_from_sequence DESC` LIMIT 1.

### Chart / run tables

**`control_charts`** — `id`, `material_id FK`, `method_id FK`, `batch_id`, `technician_id FK`, `created_at`, `locked_at`
- `locked_at` is set immediately on save — all charts are immutable from creation

**`chart_metadata_fields`** — `id`, `name`, `required`, `display_order`

**`chart_metadata_values`** — `id`, `control_chart_id FK`, `field_id FK`, `value`

**`measurements`** — `id`, `control_chart_id FK`, `material_method_analyte_id FK`, `value`, `sequence_order`, `sequence_number`, `entered_by FK`, `entered_at`
- `sequence_order` — column position within a chart (instrument order)
- `sequence_number` — immutable per-MMA run number, computed at insert time via `COALESCE(MAX(sequence_number), 0) + 1` inside the SaveChart transaction
- Append-only. No UPDATE statements ever issued against this table.

**`comments`** — `id`, `control_chart_id FK`, `measurement_id FK` (nullable), `comment_type`, `text`, `user_id FK`, `created_at`

**`spc_rule_sets`** — `id`, `beyond_sigma_enabled`, `beyond_sigma_n`, `run_trend_enabled`, `run_trend_n`, `one_side_enabled`, `one_side_n`, `effective_from_date`, `created_by FK`, `created_at`

---

## Key design decisions

- **No SvelteKit** — plain Svelte with top-level `let view = 'login'` for navigation
- **Append-only measurements** — data integrity enforced architecturally, not by policy
- **sequence_number** — per-MMA immutable run number on measurements, used as x-axis on charts and paperwork reference
- **sequence_order** — instrument column order within a chart, separate from sequence_number
- **Hybrid metadata** — `technician_id` is a proper FK, everything else flexible via metadata tables
- **Control limit regions** — versioned per `material_method_analyte` by sequence number, soft-deleted only
- **No pass/fail shown live during data entry** — deliberate, prevents pressure to fudge values before saving; shown after save only
- **SQLite over Postgres** — IT won't support a Postgres server; WAL mode acceptable at 2-3 users
- **Atkinson Hyperlegible Next** — bundled font for readability; Mono variant for data display
- **No Jet/sqlx** — plain `database/sql`; schema is simple enough, overhead not worth it
- **Wails v2 not v3** — v3 still alpha, no multi-window needed anyway
- **UserResponse not User** — Wails bindings can't handle `time.Time`; all frontend-bound returns use `UserResponse` with string timestamps
- **Light theme** — better for well-lit lab environment; backed by readability research
- **Go slices serialize as null** — empty slices from Go come back as `null` in JS, always use `?? []` on list results
- **Clipboard parsing** — LibreOffice Calc puts data as `text/html`, not `text/plain`. Parse HTML table cells via `DOMParser` first, fall back to tab-split plain text. Excel also uses HTML format so this handles both.
- **Decimal separator** — regex currently only accepts `.` as decimal separator. Locale issues deferred until if/when needed.
- **No LATERAL joins** — SQLite doesn't support them; use correlated subqueries instead
- **Combo-level deactivation** — DeactivateCombo/ActivateCombo operate on all analytes for a method+material at once
- **Chart data limit** — default 50 points, user-adjustable input; limit <= 0 means no limit in handler

---

## Models

### `models/user.go`
```go
type User struct {
    ID           int64
    Username     string
    PasswordHash string
    Role         string
    Active       bool
    CreatedAt    time.Time
}

type UserResponse struct {
    ID        int64  `json:"id"`
    Username  string `json:"username"`
    Role      string `json:"role"`
    Active    bool   `json:"active"`
    CreatedAt string `json:"created_at"`
}

type Role string

const (
    RoleTechnician Role = "technician"
    RoleReviewer   Role = "reviewer"
    RoleSupervisor Role = "supervisor"
    RoleAdmin      Role = "admin"
)
```

### `models/library.go`
```go
type Analyte struct { ID int64; Name string; Unit string }
type Method struct { ID int64; Name string; Description string }
type Material struct { ID int64; Name string; Description string }

type MMAEntry struct {
    ID int64; MaterialID int64; MaterialName string
    MethodID int64; MethodName string
    AnalyteID int64; AnalyteName string
    Unit string; DisplayOrder int; Active bool
}

type MethodWithMaterials struct { ID int64; Name string; Materials []MaterialSummary }
type MaterialSummary struct { ID int64; Name string }
type ComboAnalyte struct { MMAID int64; Name string; Unit string; DisplayOrder int }

type ControlLimitRegion struct {
    ID int64; MMAID int64
    Mean float64; UCL float64; LCL float64
    UWL *float64; LWL *float64; UIL *float64; LIL *float64
    EffectiveFromSequence int; CreatedBy int64; CreatedAt string
}

type MeasurementResult struct {
    MMAID int64; AnalyteName string; Unit string
    Value float64; SequenceNumber int
    UCL *float64; LCL *float64
    Pass bool; NoLimits bool
}

type ChartPoint struct {
    SequenceNumber int; Value float64
    Mean *float64; UCL *float64; LCL *float64
    UWL *float64; LWL *float64; UIL *float64; LIL *float64
}
```

---

## Handlers (Wails-bound)

### `handlers/auth.go` — `app.Auth`
- `Login(username, password string) (*models.UserResponse, error)`
- `Logout()`
- `CurrentUser() *models.UserResponse`
- `IsAuthenticated() bool`
- `ListUsers() ([]models.UserResponse, error)`
- `DeactivateUser(id int64) error`
- `ActivateUser(id int64) error`
- `CreateUser(username, password, role string) error`

### `handlers/setup.go` — `app.Setup`
- `NeedsSetup() bool`
- `CreateAdminUser(username, password string) error`
- `UserExists(username string) bool`

### `handlers/config.go` — `app.Config`
- `GetDBPath() (string, error)`
- `InitDB() error`
- `OpenDBFilePicker() (string, error)`
- `OpenDBFolderPicker() (string, error)`
- `SetDBPath(path string) error`
- `SetContext(ctx context.Context)`

### `handlers/library.go` — `app.Library`
- `CreateAnalyte(name, unit string) error`
- `ListAnalytes() ([]models.Analyte, error)`
- `CreateMethod(name, description string) error`
- `ListMethods() ([]models.Method, error)`
- `CreateMaterial(name, description string) error`
- `ListMaterials() ([]models.Material, error)`

### `handlers/mma.go` — `app.MMA`
- `AddAnalyteToMMA(materialID, methodID, analyteID int64, displayOrder int) error`
- `ListMMAsForMethod(methodID int64) ([]models.MMAEntry, error)`
- `ListAllMMAs() ([]models.MMAEntry, error)`
- `RemoveAnalyteFromMMA(id int64) error`
- `UpdateDisplayOrders(ids []int64, orders []int) error`
- `ListUsedMMAIDs() ([]int64, error)`
- `DeactivateCombo(materialID, methodID int64) error`
- `ActivateCombo(materialID, methodID int64) error`

### `handlers/dataentry.go` — `app.DataEntry`
- `ListMethodsWithMaterials() ([]models.MethodWithMaterials, error)` — active combos only
- `GetAnalytesForCombo(methodID, materialID int64) ([]models.ComboAnalyte, error)`
- `SaveChart(methodID, materialID, technicianID int64, values map[string]float64) (int64, error)` — single transaction, sets locked_at immediately, computes sequence_number per MMA
- `GetChartResults(chartID int64) ([]models.MeasurementResult, error)` — pass/fail per analyte after save

### `handlers/limits.go` — `app.Limits`
- `GetCurrentSequencesForMMAs(ids []int64) (map[int64]int, error)`
- `SaveControlLimitRegions(regions []models.ControlLimitRegion) error`
- `ListControlLimitRegionsForCombo(materialID, methodID int64) ([]models.ControlLimitRegion, error)`
- `DeleteControlLimitRegionSet(materialID, methodID int64, effectiveFromSequence int, userID int64) error`

### `handlers/chartreview.go` — `app.ChartReview`
- `GetComboChartData(materialID, methodID int64, limit int) (map[int64][]models.ChartPoint, error)` — correlated subqueries for limit lookup per point, limit <= 0 means no limit

---

## Frontend structure

```
frontend/src/
├── App.svelte              — top-level view router, owns `let view` and `currentUser`
├── main.ts
├── style.css               — global styles, CSS vars, font-face declarations
├── assets/fonts/           — Atkinson Hyperlegible Next + Mono woff2 files
└── lib/
    ├── Login.svelte
    ├── Setup.svelte
    ├── Shell.svelte        — navbar + content slot, resets scroll on view change via bind:this
    ├── Admin.svelte
    ├── DBPicker.svelte
    ├── Settings.svelte
    ├── Library.svelte      — tabbed: Analytes, Methods, Materials, Combos, Limits
    ├── DataEntry.svelte    — sidebar + grid + paste + save + post-save pass/fail display
    └── ChartReview.svelte  — todo
```

### View routing in `App.svelte`
- `loading` → checks NeedsSetup on mount
- `db-pick` → first launch or no config
- `setup` → first run only
- `login` → standard login
- `data-entry` → technician/all roles
- `chart-review` → reviewer/supervisor/admin
- `library` → all roles
- `admin` → admin only
- `settings` → all roles

### CSS variables (in `style.css`)
```css
--font-sans, --font-mono
--colour-bg, --colour-surface, --colour-border
--colour-text, --colour-text-muted
--colour-primary, --colour-primary-hover
--colour-danger, --colour-success
--radius, --radius-lg
--shadow-sm, --shadow-md
```

---

## Library.svelte notes

- Tabs: Analytes, Methods, Materials, Combos, Limits
- All Go list calls return null for empty slices — always `?? []`
- Combos tab: method selector → material selector (all materials shown, filtered by showHidden for inactive)
- `comboActiveForMat(matID)` — returns true if no combo exists yet (unlinked), or if combo is active
- `showHidden` checkbox shows inactive combos in material dropdown with `(inactive)` label
- Deactivate/Activate operates at combo level via `DeactivateCombo`/`ActivateCombo`
- Analyte list is drag-to-reorder via `svelte-dnd-action`
- Reorder fires `UpdateDisplayOrders` in a single transaction
- Remove button hidden if MMA has any measurements (`usedMMAIDs` set)
- Limits tab: combo cards → select combo → shows existing region sets + new paste grid
- Existing regions grouped by `effective_from_sequence`, each deletable (soft delete)
- New region: 7-row × N-analyte paste grid, 2D anchor (fills right and down from paste cell)
- `newGrid` initialized reactively when `limitAnalytes.length > 0`
- `Object.entries(groupedRegions)` types values as `unknown` — cast inside `getRegionValue(regions: unknown)`

## DataEntry.svelte notes

- Sidebar: methods as labels, materials as buttons, active combos only (from ListMethodsWithMaterials)
- Grid: wrapping flex of analyte cards
- Paste: 1D, fills right from anchor cell using HTML-first clipboard parsing
- After save: cards show result value + Pass/Fail/No limits badge, colored border
- Results cleared on combo change or new save
- `results` typed as `Record<string, MeasurementResult>` (Object.fromEntries always gives string keys)

## Shell.svelte notes
- `bind:this={contentEl}` on `.content` div
- `$: if (view && contentEl) contentEl.scrollTop = 0` resets scroll on navigate

---

## Migrations

- v1 — full initial schema
- v2 — `ALTER TABLE material_method_analytes ADD COLUMN active INTEGER NOT NULL DEFAULT 1`
- v3 — `ALTER TABLE measurements ADD COLUMN sequence_number INTEGER`
- v4 — `ALTER TABLE control_limit_regions ADD COLUMN deleted_at DATETIME; ALTER TABLE control_limit_regions ADD COLUMN deleted_by INTEGER REFERENCES users(id)`

Plan: merge all into single v1 migration before release — easy cleanup task.

---

## What's done

- [x] Wails + Svelte + TypeScript project scaffold
- [x] pnpm configured in wails.json
- [x] Go 1.26, modernc.org/sqlite, golang.org/x/crypto/bcrypt
- [x] DB init with WAL mode + foreign keys + busy timeout
- [x] Versioned migrations via PRAGMA user_version (v4)
- [x] Full schema
- [x] Auth handler with session timeout (30 min)
- [x] Role-based post-login routing
- [x] First-run setup screen
- [x] Login screen
- [x] App shell with role-filtered navbar, scroll reset on navigate
- [x] Admin user management screen
- [x] config package
- [x] ConfigHandler
- [x] DBPicker.svelte
- [x] Settings.svelte
- [x] Library.svelte — analytes, methods, materials CRUD + MMA combo wiring + drag-to-reorder + deactivate/activate combo + show hidden toggle + control limits tab with paste grid + soft delete
- [x] DataEntry.svelte — sidebar, analyte card grid, clipboard paste, save, post-save pass/fail
- [x] ChartReviewHandler.GetComboChartData — Go handler written, not yet wired to frontend

## What's next

1. ChartReview.svelte — XmR individuals control charts with Chart.js
   - Combo card picker (same pattern as limits tab)
   - Per-analyte X chart + mR chart stacked
   - Limit lines via chartjs-plugin-annotation
   - Point limit input (default 50, user-adjustable)
2. Trend detection against spc_rule_sets
3. Comment/annotation system
4. Audit log view

---

## Planned build order (revised)

1. ~~Wails init~~ ✅
2. ~~SQLite setup + migrations~~ ✅
3. ~~Auth~~ ✅
4. ~~Library setup~~ ✅
5. ~~Data entry~~ ✅
6. ~~Control limits setup~~ ✅
7. XmR control charts ← current
8. Comment/annotation system
9. Audit log view
```