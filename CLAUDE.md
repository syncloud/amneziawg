# Debugging CI failures

CI is Drone at `http://ci.syncloud.org:8080/syncloud/amneziawg` (JS SPA — use the API, not curl-scrape HTML).

List recent builds:
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/amneziawg/builds?limit=5" | python3 -c "
import json,sys
for b in json.load(sys.stdin):
    print(b['number'], b['status'], b.get('message','').split('\n')[0][:80])
"
```

Find the failing step (stage index + step index are what the log endpoint wants):
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/amneziawg/builds/{N}" | python3 -c "
import json,sys
b=json.load(sys.stdin)
for si,stage in enumerate(b.get('stages',[]), 1):
    for step in stage.get('steps',[]):
        if step.get('status') == 'failure':
            print(si, step['number'], stage['name'], step['name'])
"
```

Fetch the step log (use the stage/step numbers from above):
```
curl -s "http://ci.syncloud.org:8080/api/repos/syncloud/amneziawg/builds/{N}/logs/{stage}/{step}" | python3 -c "
import json,sys; [print(l.get('out',''), end='') for l in json.load(sys.stdin)]
" | tail -120
```

## CI Artifacts

Served at `http://ci.syncloud.org:8081` (JSON listings). Layout mirrors platform's.

Browse:
```
curl -s "http://ci.syncloud.org:8081/files/amneziawg/{build}-{arch}/"
curl -s "http://ci.syncloud.org:8081/files/amneziawg/{build}-{arch}/{distro}/"
curl -s "http://ci.syncloud.org:8081/files/amneziawg/{build}-{arch}/{distro}/app/"
curl -s "http://ci.syncloud.org:8081/files/amneziawg/{build}-{arch}/{distro}/desktop/"
```

Structure:
```
{build}-{arch}/
  {distro}/
    app/
      journalctl.log        # journal from integration test teardown
      ps.log, netstat.log   # process/network state at teardown
    platform/               # platform logs
  distro/
    desktop/                # UI test artifacts (amd64 only)
      journalctl.log
      screenshot/
        {test-name}-desktop.png
        {test-name}-desktop.html.log
      trace/                # Playwright traces
    video.mkv               # Playwright recording
```

Download directly:
```
curl -O "http://ci.syncloud.org:8081/files/amneziawg/{N}-amd64/bookworm/app/journalctl.log"
curl -O "http://ci.syncloud.org:8081/files/amneziawg/{N}-amd64/distro/desktop/screenshot/<name>.png"
```

# Build pipeline steps (per arch)

1. `version` — writes build number
2. `amneziawg.{distro}.com` — DNS service for UI test host aliasing
3. `nginx` / `nginx test {distro}` — build + test nginx (both distros)
4. `amneziawg-go` / `amneziawg-go test {distro}` — userspace WG replacement
5. `amneziawg-tools` / `amneziawg-tools test {distro}` — awg/awg-quick
6. `cli`, `backend`, `web` — Go CLI, Go backend, Vue SPA builds
7. `package` — create `.snap` file + test app
8. `test {distro}` — Python integration tests (syncloud-lib) per distro
9. (amd64 only) `test-ui-desktop`, `test-ui-mobile` — Playwright UI tests
10. `test-upgrade` — upgrade path (skipped on feature branches)
11. `upload` / `promote` — publish to release repo (stable/master only)
12. `artifact` — SCP test artifacts to ci.syncloud.org:8081

# UI test conventions

- Select by `data-testid` only — never `getByRole` / `getByText` / CSS selectors.
  Text and ARIA roles break on copy edits or Element Plus component internals.
- Navigate via clicks: one `page.goto('/')` per test, then `getByTestId('nav-peers').click()`.
  `page.goto('/peers')` bypasses nav and hides routing/auth bugs.

# DB schema migrations are mandatory

The peers DB lives in `$SNAP_DATA/db/data.db` and is preserved across `snap refresh` (upgrades). Any change to a table the running app reads — adding a column, renaming, changing type, dropping — must ship with a forward migration that runs on backend startup, otherwise an upgraded binary will hit a fresh-from-disk old schema and crash or corrupt rows. There is no migration framework yet; when the first schema change lands, add one (e.g. a `schema_version` table + ordered up-migrations applied at `db.Open`) before doing anything else.

# Running Drone builds locally

```
drone jsonnet --stdout --stream > .drone.yml
drone exec --pipeline amd64 --trusted \
  --include version \
  --include "build web" \
  --include build \
  --include package \
  --include "test bookworm" \
  .drone.yml
```
