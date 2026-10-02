<!-- Created by two-barrels in 2026 for ARI v6 modernization. SPDX-License-Identifier: Apache-2.0 -->

# Coordinated v6 release checklist

## Current candidate status — 2026-10-02

This section supersedes the historical preparation notes below. Both PR #1s
are merged into permanent `v6` branches. Existing `main` and v5 tags are intact.
Both `v6.0.0-rc.1` tags are published; proxy requires ARI without replacements.
The final published-tag consumer check passed. Local proxy race, vet, contract,
and standalone snapshot gates pass with the published dependency. Normal proxy
Go CI no longer checks out the sibling; contract CI pins its source audit to
the ARI tag. Verify the final published pair with:

```sh
GOWORK=off go run ./tools/release-check --ari-version v6.0.0-rc.1 --proxy-version v6.0.0-rc.1
```

Release commits: ARI `05e769a972ff5b11bb6db96195135cfe52f6db75`, proxy
`8be33da28f494962efc18a95d05fc0d65deb7756`. Hosted [proxy Go CI](https://github.com/two-barrels/ari-proxy/actions/runs/36978623776)
and [contract CI](https://github.com/two-barrels/ari-proxy/actions/runs/36978623802)
passed for the proxy candidate; [ARI CI](https://github.com/two-barrels/ari/actions/runs/36978124329)
passed for the native candidate. Final published archives built the external
consumer and both modules without a sibling checkout or replacement.

Candidate publication covers Go modules only. Stable v6, phone-apps migration,
binary/container publication, and the remaining audits are future work.
Additional live PBX and broker checks remain deferred as agreed.

Status as of 2026-10-01: preparation in progress; release gates remain open.
No tags or publishing actions are performed by the standalone checker.
Both modernization checkouts now use local `codex/v6-modernization` branches,
with two-barrels forks as `origin` and CyCoreSystems as `upstream`. Existing fork
default branches and tags have not changed. The modernization work is still
organized into commits and pushed to origin/codex/v6-modernization; both module/import paths use their final two-barrels owner.
The [two-barrels fork and phone-apps inspection](two-barrels-fork-inspection.md)
records current fork refs, pinned consumers, dependency differences, and migration risks.

## Evidence and open gates

Hosted checks passed: [ARI Go CI](https://github.com/two-barrels/ari/actions/runs/36975590152)
at fe4592b, [proxy Go CI](https://github.com/two-barrels/ari-proxy/actions/runs/36975604921)
and [contract/standalone CI](https://github.com/two-barrels/ari-proxy/actions/runs/36975604943)
at 74764bd. Review PR submission is pending browser sign-in. Draft release notes
are in [proxy notes](v6-release-notes.md) and the sibling
[ARI notes](../../ari/docs/v6-release-notes.md). No tags have been created.

| Gate | Evidence / remaining work |
| --- | --- |
| Pinned route/option contract | 109 operations, 175 parameters verified on native and proxy wires; checker passes. |
| Repository regression tests | Both full race suites passed after the 2026-10-01 live-discovered websocket fix. Rerun on the exact release commits. |
| Live Asterisk 22 | Selected read, bridge, external-media/event, and full bounded recording checks passed; not every operation. User accepted 22.10.1 as sufficient live validation for current release scope on 2026-10-01. |
| Live Asterisk 20 and 23 | Deferred by the user on 2026-10-01; not a current release gate. Local version fixtures and pinned Asterisk 23 contract tests remain; these versions are not live-certified. |
| Live NATS and RabbitMQ | Deferred for now by the user on 2026-10-01. Broker behavior remains unverified after dependency updates. |
| Remaining correctness/performance audit | Open: escaping, request context propagation, timeout policy, binary throughput, multi-node partial failures. |
| Migration guides | Drafts exist in both repositories; confirm against reviewed release diffs and application upgrades. |
| Standalone source builds | Packaged snapshots for final two-barrels/ari/v6 and two-barrels/ari-proxy/v6 passed on 2026-10-01 without workspace or replacements; rerun on final commits. |
| Go minimum | Go 1.25.0 declared after retaining fork dependency upgrades. Both race suites and all examples passed on Go 1.25.7; race suites and standalone builds also passed with the host compiler. Rerun on final release commits. |
| Release CI | Source workflows prepared: main/master/codex branch pushes and PRs, Go 1.25.x/1.26.8 race tests, vet, module verification/tidy, explicit examples, vulnerability scans, contract and standalone snapshots. Proxy jobs check out two-barrels/ari at codex/v6-modernization beside the proxy; push that ARI branch first. Manual runs accept an ARI ref. Hosted execution remains unverified until branches are pushed. Final CI must use the published dependency and remove the development checkout. |
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

The user selected separate permanent `v6` branches on 2026-10-02. Both branches
start at their fork's current `main`; the modernization PRs now target `v6`.
Keep `main` and existing v5 tags intact while phone-apps migrates. After
validation, make `v6` the default branch; if the name `main` is desired, preserve
old `main` as `v5` before renaming `v6`. Branch/default changes do not replace
module-version tags. CI must include pushes to `v6`.

Legacy proxy image/release jobs are restricted to the CyCoreSystems upstream
repository because their Docker/GoReleaser configuration still uses CyCore
publication targets and old tooling. Prepare and review two-barrels artifact
publishing separately before enabling fork releases. The new source CI uses
`go vet` in place of the obsolete golangci-lint action; a separately configured
modern lint policy remains optional work.

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
