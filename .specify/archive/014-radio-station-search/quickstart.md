# Quickstart: validating radio search and browse

## Automated

```sh
gofmt -l .
go vet ./...
go test ./...
```

Expected: no `gofmt` output, `go vet` exit 0, every suite passes, including the existing TTS diagnosis
tests unchanged (the shared inventory walk must not move a TTS message). The new tests:

- `tests/contract/radio_test.go`: each call's request (path, query parameters sent once and only when
  set, path-escaped UUID) and decoding against `api.Spec` key lists; every error row of
  [contracts/hub-package.md](contracts/hub-package.md).
- `tests/contract/extensions_test.go`: the `radio` diagnoses and `Extension.Version`.
- `tests/unit/render_radio_test.go`: YAML and JSON shapes, `null` fields, empty pages, and a fixture with
  ESC, C1 and U+202E characters that must come out escaped (SC-004).
- `tests/unit/radio_report_test.go`, `tests/unit/exitcode_test.go`: messages and exit 15.
- `tests/integration/radio_test.go`: each command end to end against a test server, every row of
  [contracts/cli.md](contracts/cli.md)'s exit-code table, and the request count (one radio request, plus
  one inventory request only after a "not offered" 404).

## Live hub (`multiroom.lan`, radio 0.2.0)

All of these are read-only: they touch no speaker, output, group, input or route, and are allowed without
asking. They do send requests on to radio-browser.info through the hub.

1. `sonora search radio-stations kiss --country UA` lists stations whose names contain "kiss", each with
   `uuid` and `uri`.
2. `sonora search radio-stations --limit 5 --json | jq .nextOffset` prints `5`; `--offset 5` gives the
   next five.
3. `sonora get radio-stations/<uuid from step 1>` shows that station;
   `sonora get radio-stations/00000000-0000-0000-0000-000000000000` exits 5;
   `sonora get radio-stations/not-a-uuid` exits 6 with the hub's message.
4. `sonora list radio-countries`, `sonora list radio-languages`, `sonora list radio-tags --name jazz`.
5. `sonora search radio-stations --limit 101` exits 2 without a request.
6. "Not available" reasons are covered by tests only: making them live would mean disabling the
   extension and restarting the hub.
7. SC-005: before and after, `sonora get routes` shows the same routes.

Playing a found station (SC-001) makes sound: ask the user first, then follow the production rules in
the workspace `AGENTS.md` (record state, `office` output, low volume, stop the route afterwards):
`sonora play radiobrowser://<uuid> outputs/office`.
