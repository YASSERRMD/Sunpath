# Sunpath — Master Build Specification
A solar exposure and shadow analysis application built on open-source maps.
Engine: OpenCode running DeepSeek V4 (deepseek-v4-pro).
Author: Mohamed Yasser | Solutions Architect
Version: v1.0

0. How to use this document
This file is the single source of truth for building Sunpath with OpenCode.

Place this file in the repository root as AGENTS.md (OpenCode reads it automatically), or keep it as SUNPATH_BUILD_SPEC.md and reference it explicitly in each prompt.
Section 4 (Git Discipline) is non-negotiable and applies to every phase. Read it before starting.
Each phase in Section 6 is a self-contained prompt block. Run them one at a time, in order. Do not paste multiple phases at once.
After each phase, run the verification block before moving on. If verification fails, fix on a correction branch (see Section 4).


1. Product definition
1.1 What Sunpath is
Sunpath answers a question that no clean consumer app currently owns well: how much direct sunlight does a specific point receive, and when.
A user drops a pin on a map (a terrace, a balcony, a future flat, a garden bed, a cafe seat) and Sunpath computes when that point is in direct sun versus in the shadow of surrounding buildings, across any day of the year. The primary deliverable is a year-round solar exposure profile for a fixed point.
1.2 Primary user outcome (build this first)
Year-round analysis for a fixed point. Given a latitude/longitude and a ground height (or floor level), produce:

A daily sun-hours figure for every day of the year.
A month-by-month heatmap (day of month on one axis, hour of day on the other) showing sun vs shadow.
Key dates: summer solstice, winter solstice, equinoxes, plus the worst and best days.
A plain-language summary, for example: "This balcony gets direct sun roughly 09:40 to 14:10 in midwinter and 06:30 to 19:50 at midsummer. It is fully shaded all day on 14 days of the year."

The two secondary outcomes (animated daily sweep, sunniest-spot finder) are built on the same engine in later phases. They are explicitly not part of the first shippable milestone.
1.3 Non-goals for v1.0

No user accounts, no authentication, no saved-project sync. A point can be shared via an encoded URL only.
No mobile native apps. Responsive web only.
No AI/ML features. All computation is deterministic astronomy and geometry.
No global coverage guarantee. Sunpath works wherever OpenStreetMap building data is good; it must degrade honestly where data is thin (see 3.4).


2. Architecture decisions (locked)
These were decided up front. Do not revisit them mid-build.
2.1 Shadow model: 2.5D core, DSM optional later

Core (Phases 1-9): Buildings are treated as 2.5D extruded prisms. Each OSM building footprint polygon is extruded vertically to its height. Shadow casting is geometric ray intersection between the sun vector and these prisms.
DSM layer (Phase 13, optional): A later, clearly isolated phase adds Digital Surface Model / terrain-and-vegetation shadows as an optional overlay. It must not be entangled with the core engine. If DSM is never built, the core app is still complete and correct.

2.2 Compute split: hybrid (backend precompute, client interpolation)

Backend (Go): Fetches and caches OSM building geometry, extrudes prisms, and for a requested point precomputes a horizon profile — the obstruction angle of surrounding buildings for each compass azimuth (typically 360 samples, one per degree). This is the expensive step and it is cached per point.
Client (browser): Given the cached horizon profile, the client computes sun position for any date/time (cheap astronomy math) and checks it against the horizon profile. Producing the full year heatmap is then thousands of cheap lookups, done instantly in the browser with no further backend calls.
Why this split: The horizon profile is the only geometrically expensive artifact, and it is independent of date and time. Compute it once on the backend, cache it, and the client can explore the entire year interactively with zero latency.

2.3 The horizon profile is the core abstraction
For a point P at ground/observer height h:

horizon[az] = the maximum elevation angle (degrees above the horizontal) at which a building edge obstructs the sky, looking in compass direction az.

A point is in direct sun at a given moment if and only if:

sun_elevation > horizon[round(sun_azimuth)] AND sun_elevation > 0.

Everything else in Sunpath is built on this one rule. The backend produces horizon[]; the client consumes it. Get this abstraction right in Phase 4 and the rest follows.
2.4 Technology stack
LayerChoiceReasonBackendGo (1.22+)Scalable backend, matches builder preferenceGeometry coreGo, orb library for geo types + custom ray castingNo heavyweight GIS dependencyBackend cacheSQLite (via modernc.org/sqlite, pure Go, no CGO)Single-file, zero-ops, cache of horizon profiles + OSM extractsOSM dataOverpass API for on-demand building fetch, with a local extract fallbackOpen data, no keyFrontendReact + Vite + TypeScriptReactive, low-latency frontendMap renderingMapLibre GL JSOpen-source, vector tiles, no Mapbox tokenTilesMapTiler open tiles or self-hosted style; no proprietary token committedOpen-source mapping requirementSun mathsuncalc (client) + a Go port for backend validationWell-tested astronomyChartsA lightweight canvas heatmap, hand-built (no charting dependency)Heatmap is custom; keeps bundle smallGeocodingNominatim (OSM)Open, no keyTestsGo testing + testify; Vitest + React Testing LibraryStandard
2.5 Repository shape
sunpath/
  AGENTS.md                 # this spec (or SUNPATH_BUILD_SPEC.md)
  README.md
  LICENSE                   # MIT
  .gitignore
  docker-compose.yml         # backend + tile proxy for local dev
  backend/
    cmd/sunpathd/main.go      # server entrypoint
    internal/
      geo/                    # geo types, polygon ops, extrusion
      osm/                    # Overpass client, building parsing, cache
      horizon/                # the horizon profile engine (core)
      sun/                    # solar position (backend validation copy)
      api/                    # HTTP handlers, routing
      store/                  # SQLite cache layer
    go.mod
  frontend/
    src/
      lib/sun.ts              # client sun position + sun/shade rule
      lib/horizon.ts          # horizon profile consumption, year compute
      components/             # Map, PinInspector, YearHeatmap, etc.
      App.tsx
    package.json
    vite.config.ts
  docs/
    ARCHITECTURE.md
    API.md

3. Domain detail the agent must get right
3.1 Solar position

Use standard solar position algorithms. suncalc on the client is acceptable for v1.0 accuracy targets (minutes-level).
Azimuth convention must be fixed and documented: 0° = North, 90° = East, 180° = South, 270° = West. The horizon profile array uses the same convention. A mismatch here silently inverts every shadow — call it out in code comments and test it explicitly.
Time handling: all internal computation in UTC. The point carries an IANA timezone (resolved from coordinates) and the UI displays local solar-relevant times. Store the timezone with the cached point.

3.2 Building heights from OSM
OSM building height data is inconsistent. Resolve height per building in this priority order:

height tag (metres) — use directly.
building:levels tag — multiply by an assumed 3.2 m per level, plus 1 m base.
No data — apply a configurable default (default 8 m for building, lower for building=garage, etc.) and flag the building as estimated.

The fraction of nearby buildings with estimated heights must be surfaced to the user as a confidence indicator (see 3.4).
3.3 The horizon computation
For point P:

Fetch all buildings within a radius (default 500 m; configurable, since a tall tower 800 m away can still cast a relevant winter shadow).
For each building, extrude its footprint polygon to its height.
For each compass azimuth az in 0..359:

Cast a vertical half-plane from P in direction az.
For every building edge it crosses, compute the elevation angle atan(building_top_height_above_P / horizontal_distance).
horizon[az] is the maximum such angle.


The observer height h matters: a 10th-floor balcony sees over many obstructions. h is subtracted from building tops when computing elevation angles. Negative results clamp to 0.
Cache horizon[] keyed by (rounded lat, rounded lng, h, building-data-hash).

Performance target: a horizon profile for a dense urban point computes in under 2 seconds on the backend. If a naive O(buildings x 360) loop is too slow, spatial-index the building edges (a simple grid bucket is enough — do not over-engineer).
3.4 Honest degradation
If fewer than a configurable fraction of nearby buildings have real (non-estimated) heights, the result is labelled low confidence in both the API response and the UI. Sunpath must never present an estimated-data result as if it were authoritative. This is a hard product requirement, not a nice-to-have.
3.5 Edge cases the tests must cover

A point with no buildings nearby (open field): horizon is all zeros, sun-hours equals astronomical daylight.
A point inside a courtyard surrounded by tall buildings: heavy obstruction, possibly zero direct sun in winter.
High latitude in summer (sun never sets) and winter (sun never rises).
A point whose 500 m radius crosses a timezone or antimeridian — document the v1.0 limitation rather than silently breaking.
Observer height above all surrounding buildings: horizon collapses to near zero.


4. Git discipline (MANDATORY — applies to every phase)

OpenCode tends to batch many file edits and produce one large commit at the end, or stash intermediate work, so atomic commits never appear in the GitHub history. This section exists specifically to prevent that. The agent must treat these as hard rules, not suggestions.

4.1 Identity
All commits use exactly:

user.name = YASSERRMD
user.email = arafath.yasser@gmail.com

Never use claude, opencode, deepseek, or any AI/tool identifier as author or committer. Set this explicitly at repo init:
bashgit config user.name "YASSERRMD"
git config user.email "arafath.yasser@gmail.com"
Before the first commit of every phase, re-run git config user.name and git config user.email and confirm both values. If wrong, fix before committing.
4.2 Atomic commits — the core rule

An atomic task is one logically complete unit of work, roughly 3-15 minutes of effort (one file, one function group, one test file, one config).
Commit immediately after each atomic task is complete, before starting the next task. Do not accumulate multiple tasks into one commit. Do not defer commits to the end of the phase.
Each phase below lists its atomic tasks explicitly. There must be at least that many commits on the phase branch. More is fine; fewer means tasks were wrongly batched.
Commit messages: imperative mood, scoped, for example geo: add polygon extrusion to prism or horizon: cache profiles in sqlite. No "WIP", no "misc", no "various fixes".

4.3 Forbidden commands
The agent must NEVER run any of these, because they collapse or hide atomic history:

git stash (in any form) — do not stash intermediate work; commit it instead.
git commit --amend
git rebase (interactive or otherwise) on a branch that has commits being kept.
git merge --squash
git reset --soft / --mixed / --hard to fold commits together.
git cherry-pick to reassemble history.
Any squash option when merging a pull request.

If the agent believes history needs rewriting, it must STOP and ask the user instead.
4.4 Branching strategy

main is never committed to directly and never has work pushed to it directly.
Each phase gets its own branch: phase_1, phase_2, ... phase_N.
Start a phase: git checkout main && git pull && git checkout -b phase_N.
Do all the phase's atomic commits on phase_N.
Push the branch: git push -u origin phase_N.

4.5 Pull requests and merging — preserve atomic history
After pushing phase_N:

Open a pull request from phase_N into main.
Merge the PR using a merge commit (the "Create a merge commit" option) or a rebase merge. NEVER use squash merge — squash merge destroys exactly the atomic commit history this whole section is protecting.
After the PR is merged, delete the phase_N branch (remote and local).
git checkout main && git pull before starting the next phase.

If merging via the GitHub CLI:
bashgh pr create --base main --head phase_N --title "Phase N: <title>" --body "<summary>"
gh pr merge phase_N --merge --delete-branch
--merge is mandatory. Never --squash.
4.6 Corrections and fixes
If a defect from an already-merged phase must be fixed:

Do NOT amend or rebase.
Create a new branch fix_<short-description> off current main.
Commit the fix atomically on that branch, push it, open a PR, merge it with --merge, delete the branch.

4.7 Per-phase verification (run at the end of every phase, before the PR)
The agent must run and report the output of:
bashgit config user.name
git config user.email
git log --oneline main..HEAD
git log --oneline main..HEAD | wc -l
Then state explicitly: "Phase N produced X atomic commits on branch phase_N. Expected at least Y. [PASS/FAIL]." If FAIL because commits were batched, the agent must explain and the user decides whether to redo the phase. The agent must not silently proceed.
4.8 .gitignore (commit in Phase 0)
# Go
backend/sunpathd
*.test
*.out
# Node
frontend/node_modules/
frontend/dist/
# Local data / cache
*.sqlite
*.sqlite-*
data/
.env
.env.local
# OS / editor
.DS_Store
.idea/
.vscode/

5. Quality bar (applies throughout)

Every backend package ships with unit tests. The horizon and sun packages must have correctness tests against known reference values (a known solstice sun position; a hand-computed shadow case).
No secrets, API keys, or tile tokens committed. Tile style URL and any keys come from environment variables / .env (gitignored). A .env.example is committed.
No em dashes anywhere in code, comments, docs, or UI copy. Use commas, colons, or sentence breaks.
Use "explore" / "investigate", not "experience", in all prose and UI copy.
README must let a new developer run the whole app locally in under five commands.
Each phase leaves the build green: backend compiles and tests pass, frontend builds.


6. The phases
Thirteen phases. Phases 1-9 deliver the complete, shippable v1.0 (the year-round point analysis). Phases 10-12 add the secondary outcomes and polish. Phase 13 is the optional DSM layer.
Run each block below as a single OpenCode prompt. Each block already embeds the git rules by reference — keep AGENTS.md in the repo so the agent always has Section 4 in context.

PHASE 0 — Repository and tooling foundation
You are building "Sunpath", a solar exposure and shadow analysis web app.
Read AGENTS.md fully before doing anything. Section 4 (Git discipline) is
mandatory: atomic commits, one commit per atomic task committed immediately,
never stash, never amend, never squash, branch per phase, identity YASSERRMD /
arafath.yasser@gmail.com.

This is PHASE 0. Create branch phase_0 off main first.

Atomic tasks (one commit each, in order):
1. Initialise the repo: go.mod for module github.com/yasserrmd/sunpath/backend
   (Go 1.22), and frontend scaffold with Vite + React + TypeScript.
2. Add .gitignore exactly as specified in AGENTS.md section 4.8.
3. Add LICENSE (MIT, author Mohamed Yasser) and a README.md skeleton with
   project description, stack, and a "Local development" section placeholder.
4. Add docs/ARCHITECTURE.md containing the locked decisions from AGENTS.md
   sections 2 and 3 (summarised, in your own words).
5. Add docker-compose.yml with a placeholder backend service and a tileserver
   placeholder, plus a .env.example with TILE_STYLE_URL and OVERPASS_URL.
6. Set up the directory structure exactly as in AGENTS.md section 2.5, with
   empty package files where needed so the tree exists.

End of phase: run the section 4.7 verification block, report commit count
(expect at least 6), push phase_0, open a PR into main, merge with --merge
(never squash), delete the branch. Do not start Phase 1.

PHASE 1 — Geo core: types, polygons, extrusion
Read AGENTS.md. This is PHASE 1. Branch phase_1 off updated main.

Build the backend/internal/geo package: the geometric foundation.

Atomic tasks (one commit each):
1. Define core types: Point (lat, lng), a metric local projection helper
   (project lat/lng to local metres around an origin, and back).
2. Implement polygon types and basic operations: area, centroid,
   point-in-polygon, bounding box.
3. Implement the Building type: footprint polygon + height (metres) +
   heightEstimated bool + osmID.
4. Implement extrusion: turn a Building footprint into a Prism (the set of
   vertical wall quads + a flat roof at height h).
5. Implement a ray/segment helper: given an origin in local metres and a
   compass azimuth, intersect against a 2D segment, returning horizontal
   distance to the hit.
6. Write geo_test.go covering projection round-trip, point-in-polygon,
   extrusion vertex counts, and ray/segment intersection with known values.

End of phase: section 4.7 verification, report commit count (expect at
least 6), push, PR, merge with --merge, delete branch. Stop.

PHASE 2 — Sun position (backend validation copy)
Read AGENTS.md. This is PHASE 2. Branch phase_2 off updated main.

Build backend/internal/sun: a solar position library used to validate the
client and to support backend-side computation.

Atomic tasks (one commit each):
1. Implement Julian date and the core solar position algorithm: given UTC
   time and lat/lng, return sun azimuth and elevation. Azimuth convention:
   0=N, 90=E, 180=S, 270=W. Document this in comments.
2. Implement sunrise/sunset/solar-noon for a given date and location.
3. Implement a helper that returns sun azimuth/elevation sampled across a
   full day at a configurable interval (default 1 minute).
4. Write sun_test.go: validate against known reference values for at least
   two locations and the 2025 solstices, within a 0.5 degree tolerance.
   Include a high-latitude polar-day and polar-night case.

End of phase: section 4.7 verification, commit count (expect at least 4),
push, PR, merge --merge, delete branch. Stop.

PHASE 3 — OSM building ingestion and cache
Read AGENTS.md. This is PHASE 3. Branch phase_3 off updated main.

Build backend/internal/osm and backend/internal/store.

Atomic tasks (one commit each):
1. store: SQLite layer (modernc.org/sqlite, no CGO). Schema for cached OSM
   building extracts keyed by bounding box, and a separate table for cached
   horizon profiles (used in Phase 4). Migrations run on startup.
2. osm: Overpass API client that, given a bounding box, fetches all
   building ways/relations with geometry.
3. osm: parse the Overpass response into geo.Building values. Resolve
   height per the priority rule in AGENTS.md 3.2 (height tag, then
   building:levels * 3.2 + 1, then default), setting heightEstimated.
4. osm: a fetch-with-cache function: check the store first, call Overpass
   on a miss, write through to the store.
5. Write osm_test.go using a recorded/fixture Overpass JSON response (do
   not hit the network in tests). Cover height resolution for all three
   cases and the estimated flag.

End of phase: section 4.7 verification, commit count (expect at least 5),
push, PR, merge --merge, delete branch. Stop.

PHASE 4 — The horizon profile engine (core abstraction)
Read AGENTS.md. This is PHASE 4, the most important phase. Branch phase_4
off updated main.

Build backend/internal/horizon: the core abstraction described in AGENTS.md
2.3 and 3.3.

Atomic tasks (one commit each):
1. Define HorizonProfile: 360 elevation-angle samples (one per integer
   azimuth degree), plus metadata (point, observer height h, confidence
   fraction, building count, estimated-building count, dataHash).
2. Implement the core compute: given a point P, observer height h, and a
   set of extruded buildings, fill horizon[az] for az in 0..359 using ray
   casting against building edges (use geo from Phase 1). Subtract h from
   building tops; clamp negative elevation angles to 0.
3. Add a simple grid spatial index over building edges so the computation
   meets the under-2-seconds target for a dense urban point. Do not
   over-engineer; a uniform grid is enough.
4. Implement confidence: compute the fraction of contributing buildings
   that have real (non-estimated) heights; expose it on the profile.
5. Integrate caching: compute keyed by (rounded lat, rounded lng, h,
   building-data-hash) via the store from Phase 3.
6. Write horizon_test.go: an open-field point (all-zero horizon), a point
   boxed in by four tall buildings (high obstruction on all sides), an
   observer raised above all buildings (near-zero horizon), and a
   hand-computed single-building shadow case with a known elevation angle.

End of phase: section 4.7 verification, commit count (expect at least 6),
push, PR, merge --merge, delete branch. Stop.

PHASE 5 — Backend HTTP API
Read AGENTS.md. This is PHASE 5. Branch phase_5 off updated main.

Build backend/internal/api and cmd/sunpathd: the HTTP server.

Atomic tasks (one commit each):
1. Server bootstrap in cmd/sunpathd/main.go: read config from env, open the
   store, graceful shutdown.
2. GET /api/horizon?lat=&lng=&h= : returns the HorizonProfile JSON
   (computing or serving from cache). Include confidence and building
   counts in the response.
3. GET /api/geocode?q= : proxy to Nominatim, return ranked matches. Respect
   Nominatim usage policy (a proper User-Agent, light use).
4. GET /api/healthz and structured request logging middleware.
5. CORS configuration for the frontend dev origin; JSON error envelope with
   a consistent shape.
6. Write api_test.go: handler tests with httptest covering the horizon
   endpoint (success, bad params, low-confidence response) using a fake
   building source.
7. Document every endpoint in docs/API.md (request, response, error shapes).

End of phase: section 4.7 verification, commit count (expect at least 7),
push, PR, merge --merge, delete branch. Stop.

PHASE 6 — Frontend shell and map
Read AGENTS.md. This is PHASE 6. Branch phase_6 off updated main.

Build the frontend shell: React + Vite + TypeScript + MapLibre GL.

Atomic tasks (one commit each):
1. App layout: a full-screen map with a side panel; responsive so the panel
   becomes a bottom sheet on narrow screens.
2. Integrate MapLibre GL with an open vector style; tile style URL from an
   env var, never hardcoded or committed.
3. Pin interaction: click the map to drop a single Sunpath pin; show its
   coordinates; allow dragging the pin to move it.
4. A geocode search box wired to GET /api/geocode; selecting a result moves
   the map and drops the pin.
5. An observer-height control (ground level, or a floor number converted to
   metres) bound to the pin.
6. A typed API client module for the backend endpoints.

End of phase: section 4.7 verification, commit count (expect at least 6),
push, PR, merge --merge, delete branch. Stop.

PHASE 7 — Client sun engine and the sun/shade rule
Read AGENTS.md. This is PHASE 7. Branch phase_7 off updated main.

Build frontend/src/lib/sun.ts and horizon.ts: the client-side engine.

Atomic tasks (one commit each):
1. sun.ts: wrap suncalc to return sun azimuth/elevation for a date/time and
   location, using the same 0=N,90=E,180=S,270=W convention as the backend.
2. horizon.ts: fetch and hold a HorizonProfile; implement the core rule
   isInDirectSun(date) = sun_elevation > 0 AND sun_elevation >
   horizon[round(sun_azimuth)].
3. horizon.ts: computeDay(date) returning, at 1-minute resolution, the
   sun/shade state across the day, plus total direct-sun minutes.
4. horizon.ts: computeYear() returning sun-minutes for every day of the
   year and the full hour-by-day grid for the heatmap.
5. Resolve and apply the point's IANA timezone so displayed times are
   locally meaningful.
6. Vitest tests: validate isInDirectSun against the open-field profile
   (sun-hours equals astronomical daylight) and a fully obstructed profile
   (zero sun-hours), plus a solstice spot check.

End of phase: section 4.7 verification, commit count (expect at least 6),
push, PR, merge --merge, delete branch. Stop.

PHASE 8 — Year heatmap and point inspector (PRIMARY OUTCOME)
Read AGENTS.md. This is PHASE 8. Branch phase_8 off updated main. This phase
delivers the primary user outcome from AGENTS.md 1.2.

Atomic tasks (one commit each):
1. A hand-built canvas YearHeatmap component: x axis day-of-year, y axis
   hour-of-day, each cell coloured by sun (warm) vs shade (cool). No
   charting dependency.
2. A daily sun-hours line/bar strip beneath the heatmap.
3. The PinInspector panel: shows sun-hours for a selected day, the day's
   first and last direct-sun times, and a date scrubber.
4. Key dates summary: solstices, equinoxes, and the computed best and worst
   days, each clickable to load that day.
5. The plain-language summary generator (the example wording in AGENTS.md
   1.2), built from computed values, not hardcoded.
6. A confidence banner: when the backend reports low confidence, show a
   clear honest warning per AGENTS.md 3.4. Never hide it.
7. Wire the full flow: drop pin -> fetch horizon -> computeYear -> render
   heatmap, summary, and key dates. Show loading and error states.

End of phase: section 4.7 verification, commit count (expect at least 7),
push, PR, merge --merge, delete branch. Stop.

PHASE 9 — Shareable state, polish, and v1.0 release
Read AGENTS.md. This is PHASE 9. Branch phase_9 off updated main. This
completes shippable v1.0.

Atomic tasks (one commit each):
1. Encode the full point state (lat, lng, height) into the URL; loading
   such a URL restores the analysis. This is the only "save" mechanism.
2. Empty, loading, error, and offline states across the UI; a friendly
   message when OSM data is too thin to analyse.
3. An "about the method and its limits" panel: plainly explain 2.5D
   modelling, estimated heights, and the v1.0 non-goals from AGENTS.md 1.3.
4. Accessibility pass: keyboard navigation, focus order, colour-contrast
   check on the heatmap palette, sensible alt text and ARIA labels.
5. Finalise README: full local-development instructions in five commands
   or fewer, plus docker-compose usage.
6. End-to-end smoke test of the primary outcome and a manual QA checklist
   in docs/.
7. Tag v1.0 after the PR is merged (annotated tag on main).

End of phase: section 4.7 verification, commit count (expect at least 7),
push, PR, merge --merge, delete branch, then create the v1.0 tag on main.
Stop.

PHASE 10 — Animated daily sun and shadow sweep
Read AGENTS.md. This is PHASE 10 (secondary outcome). Branch phase_10 off
updated main.

Atomic tasks (one commit each):
1. A time-of-day slider and play/pause control for a chosen date.
2. Render the sun's current position and a sun/shade indicator for the pin
   that updates live as the slider moves.
3. Project and draw the actual shadow polygons of nearby buildings onto the
   map for the current sun position (2.5D shadow footprints).
4. Animate the sweep smoothly across the day; ensure it stays responsive.
5. Vitest/UI tests for the slider-to-state wiring.

End of phase: section 4.7 verification, commit count (expect at least 5),
push, PR, merge --merge, delete branch. Stop.

PHASE 11 — Sunniest-spot finder
Read AGENTS.md. This is PHASE 11 (secondary outcome). Branch phase_11 off
updated main.

Atomic tasks (one commit each):
1. Backend: an endpoint that, given a bounding box, a date or date range,
   and observer height, samples a grid of points and returns sun-hours per
   cell. Reuse the horizon engine; cache aggressively.
2. Frontend: let the user draw a rectangle on the map to define the area.
3. Render the returned grid as a sun-hours overlay (a heat surface) on the
   map.
4. Highlight the sunniest and shadiest cells; clicking a cell opens the
   full point inspector for it.
5. Tests for the grid endpoint and the overlay rendering.

End of phase: section 4.7 verification, commit count (expect at least 5),
push, PR, merge --merge, delete branch. Stop.

PHASE 12 — Performance, caching, and hardening
Read AGENTS.md. This is PHASE 12. Branch phase_12 off updated main.

Atomic tasks (one commit each):
1. Backend: rate-limit and politely batch Overpass and Nominatim calls;
   add retry-with-backoff.
2. Backend: a cache-warming and eviction policy for horizon profiles and
   OSM extracts; expose simple cache metrics.
3. Backend: load-test the horizon endpoint and tune the spatial index to
   confirm the under-2-seconds target for dense urban points.
4. Frontend: profile the year computation; move it to a Web Worker if it
   blocks the main thread, so the UI stays responsive.
5. Add structured error reporting and a basic backend metrics endpoint.

End of phase: section 4.7 verification, commit count (expect at least 5),
push, PR, merge --merge, delete branch. Stop.

PHASE 13 — Optional DSM / terrain shadow layer
Read AGENTS.md. This is PHASE 13, OPTIONAL and fully isolated. It must not
change the Phase 1-9 core behaviour. Branch phase_13 off updated main.

Atomic tasks (one commit each):
1. Backend: ingest an open DSM/elevation source for a configurable area;
   store it in the cache layer.
2. horizon: an optional pass that raises horizon[az] using terrain and
   vegetation surface heights, controlled by an explicit flag, leaving the
   2.5D-only path untouched and still the default.
3. API: expose the DSM-enhanced profile as an opt-in parameter.
4. Frontend: a toggle to include terrain shadows, clearly labelled as an
   enhanced layer with its own data caveats.
5. Tests proving the core 2.5D path is byte-for-byte unchanged when the
   DSM flag is off.

End of phase: section 4.7 verification, commit count (expect at least 5),
push, PR, merge --merge, delete branch. Stop.

7. Definition of done for v1.0
v1.0 is complete when Phases 0-9 are merged to main and:

A user can drop or search a pin, set observer height, and see a correct year-round solar exposure heatmap, daily sun-hours, key dates, and a plain-language summary.
Low-confidence (thin OSM data) results are clearly and honestly flagged.
The full point state round-trips through a shareable URL.
Backend and frontend test suites pass; the app runs locally in five commands or fewer.
git log on main shows the atomic commit history of every phase, with merge commits per phase and no squashed phases. Author is YASSERRMD throughout.
