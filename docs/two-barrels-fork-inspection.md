# two-barrels fork and phone-apps inspection

Inspection date: 2026-10-01. This was read-only for the forks and phone-apps;
no remote branches/tags or dependency pins were changed. Remote GitHub refs
were checked with `git ls-remote`, and matching local fork checkouts supplied
source/history comparisons.

## Current repositories and versions

| Repository | Remote default / HEAD | Relevant released tags | Module declaration |
| --- | --- | --- | --- |
| two-barrels/ari | main, `e8c2ec2adcb9fe3c4d758b260f6ca01870f6d944` | v5.3.2=`9c5b716`; v5.3.3=`e8c2ec2` | github.com/CyCoreSystems/ari/v5 |
| two-barrels/ari-proxy | main, `a3df0d066b2259f7a7aa791f4f28095a6e2f154f` | v5.3.2=`a3df0d0` | github.com/CyCoreSystems/ari-proxy/v5 |

Neither remote advertises v6 tags or a v6 modernization branch. The forks retain
upstream module declarations; existing consumers reach the fork through module
replacements. Preserve all existing tags, default branches, and commit IDs.
The modernization checkouts still have CyCoreSystems as their origin and large
uncommitted changes; do not push those changes to the upstream default branch.

## Fork-specific work to retain

ARI fork history after shared commit `681c62b` contains the `/v6` to `/v5`
reversion (`9c5b716`) and dependency updates (`e8c2ec2`). The inspected changes
are module/import/README/mock configuration and dependency changes, not bespoke
ARI resource behavior. Its Go minimum is 1.25.0, with newer x/exp, x/net, x/text,
and testify than the modernization tree. Compare and reconcile these versions
before releasing; do not silently regress dependency maintenance. Raising the
modernization minimum, if chosen, also requires updated documentation/checks.

Proxy fork commits after shared `38a40c1` include channel Move support
(`0f48f13`, corrected in `6c492d5`) and dependency updates. The modernization
proxy already implements Move through client, request, dispatcher, and handler;
client/server simple-command wire tests cover the operation and app arguments.
There is a transport difference: the fork uses `ChannelMove.appArgs`, whereas
the modernization proxy emits `ChannelMove.app_args`. Do not claim mixed v5/v6
wire compatibility without explicit tests or compatibility decoding. Initial
rollout should use matching proxy client/server majors.

Follow-up: the user confirmed the existing proxy fork has no current consumers.
The v6 convention intentionally retains `app_args` for proxy-owned fields and
ARI spellings for embedded options. Exact naming and native translation now have
regression tests; see `docs/v6-migration.md`.

The proxy fork's go.mod replaces CyCore ARI/v5 v5.3.2 with two-barrels ARI/v5
v5.3.3. Dependency-module replacements do not configure a downstream consumer's
main module, so application replacements must be audited independently. Newer
fork backend dependencies include nats.go 1.49.0 and amqp091-go 1.10.0 versus
1.28.0 and 1.8.1 in the modernization checkout.

## phone-apps dependency evidence

Checkout: `/home/ronlockard/Projects/phone-apps`, branch `dev`, HEAD
`9ff2f687160c2e3b31c75d7aa654a0b3090445a5`. Its origin is the corporate GitLab
phone-apps repository. Existing untracked analysis notes were left untouched.

Active go.mod selection:

```go
require github.com/CyCoreSystems/ari/v5 v5.3.2
replace github.com/CyCoreSystems/ari/v5 v5.3.2 => github.com/two-barrels/ari/v5 v5.3.3
```

`go list -mod=readonly -m -json` confirms that exact replacement, with the
v5.3.3 module checksum and cached source. No active proxy dependency appears in
its resolved module graph or Go imports; the proxy replacement is commented
out. A local ARI replacement is also commented out. `go env GOWORK` is empty.
These pins do not follow moving GitHub branches. The replacement is scoped to
upstream v5.3.2, so blindly changing only the require version would bypass it.

The application imports CyCore ARI/v5 types, native client, stdbus, play, and
record across 60 Go files. Migration must update imports and the requirement together to the final
`github.com/two-barrels/ari/v6` path and remove the old v5 replacement. Adding
proxy is a separate architecture decision; it is not currently needed merely
to upgrade this application's native ARI use.

The agent monitor also compares exact old 404 strings (`Non-2XX response: 404
Variable Not Found` and `Non-2XX response: 404 Not Found`). New native errors
preserve status and add server detail, so convert these checks to status-based
handling (`native.CodeFromError`) and search for similar comparisons throughout
phone-apps during migration.

GitLab CI builds seven application binaries from go.mod/go.sum using Go 1.25.7.
No active dependency-upgrade commands or `@main`/`@latest` selections were found
in the inspected CI/build scripts. Test stages are commented out. Workflow rules
skip branches other than main, staging, and dev, so a v6 migration branch would
not automatically receive normal pipeline validation. Add an explicit migration
branch/MR test path before relying on CI for the upgrade.

`TestOnCallDialRequiresConfirmationAndKeepsCallerForRetry` still defers context
cancellation and bus closure without joining agent monitors. Cancellation runs
first due to defer order, but asynchronous workers can still be active when the
bus closes. Fix that test's teardown on the migration branch by cancelling and
waiting for monitor completion before bus closure; retain the targeted race
check and run the broader calls suite. No phone-apps files were changed here.

## Recommended next sequence

Follow-up reconciliation: both v6 modules now retain the inspected fork dependency
versions and declare Go 1.25.0. Proxy nats.go is 1.49.0 and amqp091-go is 1.10.0.
The source-only nats-server/v2 test dependency remains present where required by
the modernization tests. Both race suites passed on Go 1.25.7 and the host
compiler; all examples and standalone packaged snapshots passed. The initial
inspection recommendations below describe the sequence, with local branches,
owner paths, dependency reconciliation, and commit organization now completed.

1. Preserve the current main branches and release tags in both existing forks.
   Work on `codex/v6-modernization`; account for existing local uncommitted files.
2. Review/reconcile the fork-specific dependency updates and Move behavior.
3. Rename the modernization module/import paths to two-barrels ARI/v6 and
   proxy/v6, including standalone checker constants, examples, mocks, and docs.
4. Commit/review/push only the new development branches. Update their CI for
   sibling development builds, then remove that arrangement after ARI publication.
5. Publish an unused ARI v6 prerelease after candidate checks, then require it
   from the proxy without a local replacement and publish its own prerelease.
6. Create a phone-apps migration branch; update native ARI imports and exact
   dependency pin, fix teardown and CI branch/test rules, and run application
   race/integration/build checks before changing deployed branches.
7. Publish stable tags and roll out when all agreed gates pass. Existing v5
   builds retain their exact tags and repository URLs throughout.

This inspection establishes source/dependency state, not deployed artifact
provenance. Separately inspect deployed binary build metadata when scheduling
rollout. Other local projects (queue-ari, dispatcher, user-status, visual-ivr,
and audio-diagnostic) also contain exact fork replacement pins; keep those
versions reachable and plan their migrations independently.
