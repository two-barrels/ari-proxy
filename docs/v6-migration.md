# Migrating ari-proxy to v6

Status: release candidate, 2026-10-02. The current checkout declares
`github.com/two-barrels/ari-proxy/v6` and imports that path throughout source,
tests, and examples. It exposes `ari/v6` types and requires its published
`v6.0.0-rc.1` tag without a sibling replacement.
Do not publish the accumulated API changes as another v5 patch.

## Module transition

The proxy module declaration and internal imports have been converted to
`github.com/two-barrels/ari-proxy/v6`. Update application imports to match.
The client import becomes `github.com/two-barrels/ari-proxy/v6/client`.
Use `github.com/two-barrels/ari/v6` throughout the consumer. The README's
older `ari-proxy/client/v5` spelling is not the module layout.

The module requires Go 1.25.0. Custom ARI implementations and mocks must satisfy
the expanded interfaces described in the [ARI migration guide](../../ari/docs/v6-migration.md).
Build the server, application clients, and any shared adapters together. A
source-level import update does not establish mixed v5/v6 wire compatibility.

For candidate evaluation, use both modules at `v6.0.0-rc.1` in a separate
application migration branch. Existing v5 consumers and `main` remain intact.
The published-tag checker verifies real downloads without sibling repositories;
source snapshot checks alone do not certify a published tag.

## Events and routing

### JSON field naming

New proxy-owned envelope fields use snake_case, such as `channel_move`,
`app_args`, `playback_id`, and `media_uri`. Existing wire spellings remain stable;
do not rename them merely to normalize style. Embedded ARI option/request structs
retain their declared ARI-oriented spellings, including `appArgs`, `channelId`,
and `otherChannelId`. ARI itself uses a mixture of camelCase and snake_case.

For Move, the proxy payload is
`{"channel_move":{"app":"destination","app_args":"agent,ONCALL"}}`.
The server decodes it into Go fields, and the native client sends
`{"app":"destination","appArgs":"agent,ONCALL"}` to Asterisk. Empty optional
arguments are omitted when encoding. The v5 fork's `appArgs` proxy spelling is
not the canonical v6 spelling; the user confirmed that fork has no current
consumers, so v6 keeps `app_args` without adding a legacy alias.

`proxy.TestProxyJSONFieldNames` asserts exact field names, optional omission,
and embedded creation options. `server.TestMoveProxyJSONTranslatesToAsteriskJSON`
checks literal proxy JSON through server dispatch and a real native HTTP client,
asserting the exact Asterisk body. These tests prevent a symmetric encoder/decoder
rename from passing unnoticed.

The proxy now preserves raw JSON fields and unknown future event types across
server/client event delivery. Applications should handle `ari.UnknownEvent` and
exit when subscription channels close. Dialog clones have independent routing
metadata but share nested event values, which must be treated as immutable.
Application-wide subscriptions receive keyless events. NATS/RabbitMQ live
checks are deferred for now per the user; event transport checks so far use an
in-memory JSON transport substitute. On 2026-10-01 the user accepted existing
Asterisk 22.10.1 live validation for current release scope and deferred further
live PBX checks. Asterisk 20/23 remain unverified against real servers.

Requests for side-effecting explicit routes, endpoint REFER, and event claim
need a key containing the application and target Asterisk node. This prevents
broadcasting those actions to multiple PBXs. Inspect the returned handle key
when using Asterisk-assigned IDs. Text messages should use `SendWithKey` or
`SendByURIWithKey` in a multi-node application. Legacy unkeyed text methods
select a node only when exactly one live node serves the application; multiple
nodes produce an ambiguity error.

Deploy matching server/client versions for initial adoption. Validate rolling
upgrades explicitly before relying on cross-major transport compatibility.
Keep the message subject prefix and application/node routing configuration
consistent across each deployment.

## Recording and errors

Stored-recording file downloads stream through node-bound sessions in chunks of
at most 32 KiB. Always close the returned `RecordingFile.Body`. Retries preserve
the last chunk, and sessions expire after inactivity; clients should not leave
download handles open. Size can be unknown. Session capacity and backend
throughput still need operational validation.

Coded native HTTP errors now retain status through proxy responses, including
claim conflicts and not-found resources. Use error/status behavior instead of
matching old exact message strings. HTTPClient/timeouts on the native upstream
must be appropriate for media operations. Remaining audits include request
context propagation, escaping, and multi-node partial failures.

## Before switching an application

1. Update imports and custom interfaces/mocks, then compile the application.
2. Test prompt cancellation, subscription closure, and application teardown.
3. Test node-targeted side effects and returned IDs on the deployed PBX version.
4. Verify complete event payloads and unknown event handling with the actual bus.
5. Exercise a streamed recording, retry, EOF, cancellation, and cleanup.
6. Run the [release checklist](v6-release-checklist.md) before publishing tags.
