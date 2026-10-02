# Coordinated v6 release checklist

Status as of 2026-10-01: preparation in progress; release gates remain open.
No tags or publishing actions are performed by the standalone checker.
Both modernization checkouts now use local `codex/v6-modernization` branches,
with two-barrels forks as `origin` and CyCoreSystems as `upstream`. Existing fork
default branches and tags have not changed. The modernization work is still
organized into local commits and unpushed; both module/import paths use their final two-barrels owner.
The [two-barrels fork and phone-apps inspection](two-barrels-fork-inspection.md)
records current fork refs, pinned consumers, dependency differences, and migration risks.

## Evidence and open gates

| Gate | Evidence / remaining work |
| --- | --- |
| Pinned route/option contract | 109 operations, 175 parameters verified on native and proxy wires; checker passes. |
| Repository regression tests | Both full race suites passed after the 2026-10-01 live-discovered websocket fix. Rerun on the exact release commits. |
| Live Asterisk 22 | Selected read, bridge, external-media/event, and full bounded recording checks passed; not every operation. |
| Live Asterisk 20 and 23 | Open. Verify selected version boundaries, ARI 23 event payloads, returned IDs, option behavior, and version-specific failures. |
| Live NATS and RabbitMQ | Open. Validate event fidelity, discovery, concurrent shutdown, contention, and recording transfer/retry/cleanup. |
| Remaining correctness/performance audit | Open: escaping, request context propagation, timeout policy, binary throughput, multi-node partial failures. |
| Migration guides | Drafts exist in both repositories; confirm against reviewed release diffs and application upgrades. |
| Standalone source builds | Packaged snapshots for final two-barrels/ari/v6 and two-barrels/ari-proxy/v6 passed on 2026-10-01 without workspace or replacements; rerun on final commits. |
| Go minimum | Go 1.25.0 declared after retaining fork dependency upgrades. Both race suites and all examples passed on Go 1.25.7; race suites and standalone builds also passed with the host compiler. Rerun on final release commits. |
| Release CI | Open: align the old ARI Go 1.21 workflow with the module minimum, and supply the proxy's sibling dependency while the development override exists. Final CI must use the published dependency. |
| Proxy major path | Complete: module declaration, source/test/example imports use proxy/v6. Both race suites and contract checker passed after conversion. |
| Standalone published tags | Open: run tag mode after each real tag becomes available. |
| Review and commits | Local commits organized by module/dependencies, event/bus fidelity, native APIs, concurrency fixes, proxy implementation, examples, and release tooling. Maintainer review remains open. |

## Source packaging and consumer check

From `ari-proxy`:

```sh
GOWORK=off go run ./tools/release-check
```

The checker packages current source/build inputs, including uncommitted files,
into a temporary local Go module proxy, strips the sibling override from the
proxy archive, and updates that archive to the generated ARI snapshot version.
It builds an external consumer plus all library/proxy packages from downloaded
archives. The consumer imports native, mocks, playback, proxy client/server,
and representative new APIs. Its resolved graph must contain no replacements.
Temporary archives and its module cache are deleted on exit; repository module
files are unchanged. Snapshot mode can run from another directory with explicit
`--ari-dir` and `--proxy-dir` arguments.

Third-party dependencies use cached module-proxy files or the configured
upstream GOPROXY with normal checksum verification. Network access can be needed.
The archives include Go/module files and JSON/template/CSV fixtures; if embedded
assets are added later, extend the packaging allowlist. Snapshot mode requires
both modules to declare their v6 paths. Passing it does not certify future tags.

After actual approved v6 tags are available, run with their exact versions:

```text
go run ./tools/release-check --ari-version <ARI-v6-tag> --proxy-version <proxy-v6-tag>
```

Replace the placeholders with real tags. Tag mode resolves both v6 modules
through normal Go module resolution without generating source snapshots or
requiring sibling repositories. It rejects replacement dependencies. This is
the final consumer distribution check, not proof of PBX compatibility.

## Review and publication sequence

1. Review API/wire behavior and organize commits for contract/generation,
   native APIs/options/errors, proxy routing/events/recording, concurrency fixes,
   integration evidence, and release documentation. Do not drop existing work.
2. Finish the open integration/audit gates or explicitly agree on a narrower
   supported contract before release. Confirm the support matrix and changelog.
3. Prepare proxy/v6 module and imports together. Keep the local override only
   while developing; rerun both race suites, the contract checker, and snapshots.
4. Review and freeze the ARI release commit, then publish the approved ARI v6
   tag. Confirm an external consumer can resolve and build that exact tag.
5. Require that ARI tag in proxy, remove the local override, tidy, and rerun
   standalone builds/tests against the published dependency. Review and freeze
   the proxy release commit.
6. Publish the approved proxy v6 tag and run the two-tag checker. Verify server
   binaries/container artifacts against that commit and publish migration notes.
7. Keep any v5 maintenance patches on a separate maintenance line. Application
   rollback uses the previous modules/artifacts and needs a tested bus/PBX plan.

Tagging, pushing, and publishing are separate release actions. This checklist
and the checker prepare reviewable evidence without performing those actions.
