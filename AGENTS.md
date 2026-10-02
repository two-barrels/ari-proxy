# ARI proxy modernization handoff

## Goal and source of truth

Modernize this proxy together with `../ari`. The first priority is lossless ARI event forwarding: the proxy previously dropped event information needed by ARI applications. The wider goal is support for every endpoint and option in the pinned Asterisk 23 REST API, while fixing correctness and performance problems.

Start with `docs/ari-23-upgrade-plan.md`. `docs/ari-23-spec-manifest.json` is pinned to Asterisk 23 `rest-api/api-docs` commit `97d55b3306ca4aa26c0136c67b79470ca4b2b785`; `docs/ari-23-endpoint-inventory.csv` covers 109 operations and `docs/ari-23-parameter-coverage.csv` covers 175 non-path parameters. The checker in `tools/ari-contract` validates classification and optional drift against a local Asterisk specification. A CSV source reference is only a mapping; wire-test columns should name tests that assert the parameter on the relevant route.

## State as of 2026-09-28

- The native `ari` client explicitly reaches all 109 pinned routes. Proxy client and server explicitly reach 108 REST routes; the event websocket is deliberately exposed to proxy clients as a message-bus subscription. The proxy and native clients have explicit route methods for bridge creation, path-ID channel origination, bridge/channel playback, and snooping; collection routes can omit IDs for Asterisk to assign them or send optional IDs in JSON bodies.
- The contract checker last reported **109 operations classified, 175 parameters classified, 175 parameters with native and proxy wire-test evidence**. This does not establish live Asterisk behavior or exhaust response and error cases.
- Implementation sequence #1 in the upgrade plan is complete for the pinned local contract. Native and proxy response wire tests cover assigned/missing IDs, empty success, and representative 400/404/409 errors, including server-side status propagation. Shared `../ari/testfixtures/asterisk_versions.go` cases cover selected Asterisk 20/22/23 introduction boundaries; the checker validates them against the manifest and client tests exercise simulated available/unavailable responses.
- Live validation against a non-disposable Asterisk 22.10.1 test PBX passed 11 native read-only calls, ten proxy server dispatch/JSON roundtrip cases, isolated bridge and external-media channel operations, native and proxy event forwarding, and one bounded stored-recording chunk. See `docs/ari-22-live-validation.md` and opt-in `server/live_*_test.go`. NATS/RabbitMQ and Asterisk 23 remain untested live. All test bridges/channels were cleaned up; user authorized state-changing tests with cleanup under `ari-testing`. Never store the supplied credential in the repository.
- Proxy event transport preserves raw JSON and unknown fields/types in addition to generated Asterisk 23 typed events. Server tests cover nested payload details, keyless events, and dialog fanout. A top-level clone reduced a local 100-dialog decode/marshal benchmark from roughly 12.5 ms/2.08 MB to 3.7 ms/1.27 MB per source event; nested values remain shared and must be treated as immutable. Full message-bus contention has not been benchmarked.
- Newer operations and options have been carried through `proxy/types.go`, `client/`, and `server/`: application event filtering, ping, bridge/channel variables, channel redirect/progress/RTP statistics, endpoint REFER, event claim, text messages, and the option sets named in the upgrade plan. Side-effecting explicit-route, REFER, and claim calls require a target application and node to avoid broadcasting to multiple Asterisk servers.
- Stored recording file download uses bounded 32 KiB chunks and node-bound proxy sessions, with retry handling and cleanup tests. HTTP status codes for coded errors survive proxy responses. These features still need live Asterisk and both message-bus backend validation.
- Modernization is organized into local commits on `codex/v6-modernization`; inspect `git status` and preserve later work. It has not been pushed. `origin` is `git@github.com:two-barrels/ari-proxy.git`; `upstream` is CyCoreSystems/ari-proxy. Both modules/imports use their final two-barrels/v6 paths. Fork dependency updates were retained, including nats.go 1.49.0 and amqp091-go 1.10.0; the Go minimum is now 1.25.0. Both race suites passed on Go 1.25.7 and the host compiler, and standalone snapshots passed after reconciliation. The development-only `replace github.com/two-barrels/ari/v6 => ../ari` remains until a reviewed ARI tag is published; this state is not ready for a standalone release.

## Next work

User scope decision (2026-10-01): existing Asterisk 22.10.1 live evidence is
sufficient for current release scope. Further live PBX validation, including
Asterisk 20/23, is deferred and is not a current release gate. NATS/RabbitMQ
live tests were also deferred for now; the earlier required-test wording below
is superseded. Preserve local contract tests and unverified-version caveats.

CI preparation (2026-10-01): Go workflows run module checks, vet,
package/example builds, race suites on Go 1.25.x/1.26.8, and govulncheck on
1.26.8. Contract CI also builds standalone snapshots. Both proxy workflows
check out two-barrels/ari at codex/v6-modernization beside ari-proxy; push ARI
first. Manual runs accept an ari_ref override. GOTOOLCHAIN=local prevents matrix
compiler switching. Hosted runs remain unverified until branches are pushed.
Legacy CyCore image/release jobs are restricted to the upstream repository;
two-barrels publishing configuration is still a release gate. Remove sibling
checkouts once the proxy requires an approved published ARI tag.

Dependency refresh (2026-10-01): amqp091-go is 1.15.0, nats.go 1.53.0,
compress 1.20.1, x/crypto 0.55.0, x/net 0.58.0, x/text 0.41.0, and
testify 1.12.1. These supersede the fork dependency versions above. The Go
minimum remains 1.25.0; the module recommends patched toolchain go1.26.8.
Explicit GOTOOLCHAIN overrides bypass that recommendation. Live RabbitMQ
and NATS validation is still required for the upgraded clients.
Both repositories passed Go 1.26.8 race suites, Go 1.25.7 compatibility
tests, the 109-operation/175-parameter checker, and standalone snapshot builds.
govulncheck on Go 1.26.8 reported no ARI vulnerabilities and no reachable
proxy vulnerabilities (three module-only proxy advisories remain).
Those module-only reports concern unused x/crypto SSH (GO-2026-6354/6355)
and OpenPGP (GO-2026-5932) packages; no affected package is imported by the
scanned proxy roots. Reassess this if SSH or OpenPGP is introduced.

Release preparation drafts are in `docs/v6-migration.md` and
`docs/v6-release-checklist.md`. `go run ./tools/release-check` packages both
working trees into a temporary module proxy and builds all packages plus an
external consumer with GOWORK=off and no replacements. It strips the sibling
override only in its archive. Published-tag mode requires both actual v6 tags.
Both module paths are now /v6; the source checker requires that state.

The 2026-10-01 dev-PBX rerun passed race-instrumented read-only, bridge, event,
and full bounded recording tests. Native and proxy hashes matched for a
49,644-byte WAV; chunk retry, EOF, and session cleanup passed. It also exposed
a native websocket connected-state race now fixed in `../ari`. Cleanup verified
zero bridges and channels. See `docs/ari-22-live-validation.md`.

1. Deferred: live Asterisk 20/23 validation. Asterisk 22.10.1 is accepted for current scope; no live Asterisk 23 certification has been completed. Check Asterisk handler compatibility before changing JSON encoding; generated ARI handlers parse JSON-body fields for many POST/PUT operations.
2. Extend response/error cases where live tests expose a gap; the pinned local wire contract is complete, but representative responses do not prove every server response shape.
3. Benchmark and fix full event delivery under realistic NATS/RabbitMQ subscriber contention. Audit cancellation, binary download throughput, multi-node list partial failures, request context propagation, path/query escaping, and the 500 ms default request timeout.
4. Review coordinated v6 migration notes and standalone builds. Both module paths are now /v6. Publish `ari/v6` first after gates pass, remove the local replace, and test against the tagged dependency before the proxy release. Do not publish these v6-facing API changes as another proxy v5 release.

## Commands and workflow

- New proxy-owned JSON fields use snake_case; embedded ARI structs retain their
  declared spellings. Preserve established wire names. Move intentionally uses
  proxy `app_args`, translated to native `appArgs`; the existing proxy fork has
  no current consumers per the user. Exact naming and native translation tests
  are in `proxy/json_naming_test.go` and `server/channel_move_json_wire_test.go`.

- From this repo: `GOWORK=off GOCACHE=/tmp/ari-go-cache go test ./...`, `GOWORK=off GOCACHE=/tmp/ari-go-cache go run ./tools/ari-contract`, and `git diff --check`.
- From `../ari`: `GOWORK=off GOCACHE=/tmp/ari-go-cache go test ./...` and `git diff --check`. Some native tests use loopback listeners and may need permission outside a restricted sandbox.
- After adding an endpoint or option, update both libraries, the proxy request model/dispatch, focused wire tests, inventory CSVs, and the plan. The full Go suites and contract checker passed after all 175 parameter evidence rows were filled; rerun after code changes.
- Application subscription and global variable writes now have explicit JSON-body wire tests, including an empty variable value; see `../ari/client/native/body_write_wire_test.go`, `client/ari23_application_asterisk_test.go`, and `server/query_options_wire_test.go`.
