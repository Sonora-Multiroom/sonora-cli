# Quickstart / Validation: Public Hub Client Package

**Feature**: 010-public-hub-package

Run from the repo root in Git Bash (Windows) or a Linux shell. `make` is optional — the
commands below are what `make check` runs.

## 1. Build, vet, test (US2, SC-002)

```sh
gofmt -l .                 # expect no output
go vet ./...
go build ./...
go test ./...              # includes the build-and-run version test
go test -short ./...       # same minus the slow build-and-run test
```

Repeat on Linux (CI `test.yml` runs on `ubuntu-latest`).

## 2. No stray references (SC-004)

```sh
go test ./tests/unit -run StrayRefs                                  # automated guard
git grep -n '"sonora-cli/' -- ':!specs' ':!docs/reviews'            # expect nothing
git grep -n 'internal/hub' -- ':!specs' ':!docs/reviews' ':!.specify'  # expect nothing
git grep -n 'X sonora-cli/' -- ':!specs'                            # expect nothing
```

## 3. Version injection (US2 scenario 2, SC-007)

```sh
v=$(git describe --tags --always --dirty)
sh build.sh && ./sonora.exe --version     # expect "$v", not "dev"
```

`build.sh` is the maintainer's local, untracked helper; without it, run
`go build -ldflags "-X github.com/Sonora-Multiroom/sonora-cli/internal/version.Version=$v" -o sonora.exe ./cmd/sonora`.

The automated guards are the tests described in [research.md §2](research.md).

## 4. Public API reads self-contained (US1 scenario 3, SC-005)

```sh
go doc -all ./hub | grep -nE 'research\.md|data-model\.md|FR-[0-9]|§|Principle'  # expect nothing
go doc ./hub ClassUsage     # expect "no symbol"
go doc ./hub ErrorClass.ExitCode   # expect "no method"
go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' ./hub  # expect only .../sonora-cli/hub
```

## 5. External consumer (US1, SC-001)

Before release, from a scratch dir **outside** the repo, using a local workspace:

```sh
mkdir -p "$TMP/hubconsumer" && cd "$TMP/hubconsumer"
go mod init example.com/hubconsumer
go work init . /d/projects-sonora/sonora-cli
```

`main.go` imports `github.com/Sonora-Multiroom/sonora-cli/hub` and `…/api`, starts an
`httptest.Server` returning a fixed `GET /api/v2/outputs` body, calls `hub.ListOutputs`,
calls `hub.SetMasterMute` against the same server, then points a call at a closed port and
prints `hub.ClassifyError(err)` (expect `ClassNetwork` and "could not reach the hub"), and
prints `len(api.Spec) > 0`. Also confirm that importing
`github.com/Sonora-Multiroom/sonora-cli/internal/config` fails with "use of internal
package … not allowed".

After tagging `v0.1.0` (step 6), repeat with no workspace:

```sh
rm go.work
GOWORK=off go get github.com/Sonora-Multiroom/sonora-cli@v0.1.0
GOWORK=off go run .
```

## 6. Release (US4)

```sh
./release.sh --minor        # v0.0.17 → v0.1.0; refuses existing tags
```

Release notes: new module path; public `hub` and `api` packages; `ClassUsage` and
`ErrorClass.ExitCode()` not in the public API; Go ≥ 1.27. Never move or re-create a pushed
tag.

## 7. CLI behavior unchanged (US2 scenario 1, SC-003)

Covered by the unchanged-in-intent unit/integration/contract suites. Optional manual spot
check against a real hub with the previous release binary and the new build:

```sh
for c in "--help" "--version" "get" "get outputs" "get outputs --json" "get routes" "get outputs/no-such-id" "set outputs/no-such-id volume 10"; do
  old/sonora $c; echo "exit=$?"; ./sonora.exe $c; echo "exit=$?"
done
```

Expect identical output and exit codes. `--version` differs only in the version string
itself.
