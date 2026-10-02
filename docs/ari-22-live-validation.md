# Asterisk 22.10.1 live validation

Date: 2026-09-28. Target: user-provided, non-disposable test PBX running Asterisk 22.10.1. Credentials and host details are intentionally omitted from this report.

## What ran

- Native `ari` client: `Info`, `InfoWithOptions(only=system)`, `Ping`, and `List` for applications, bridges, channels, endpoints, sounds, stored recordings, device states, and modules. All 11 calls returned successfully. The `Info` response reported Asterisk 22.10.1.
- Proxy server: the same native client was used behind `dispatchRequest` with a test message-bus response sink. Ten read-only operations returned without errors and their responses survived a proxy JSON marshal/unmarshal roundtrip. See `server/live_readonly_test.go`.
- Lists observed 5 applications, 0 bridges, 0 channels, 6 stored recordings, 0 device states, and 294 modules. Endpoints and sounds returned thousands of items; counts may change between runs.

After authorization for state-changing tests, `ari-testing` was connected over the native ARI websocket and uniquely named test resources were created:

- Native bridge collection creation, assigned/requested ID, creation variable, single and bulk variable reads/writes, bridge data, deletion, and post-delete 404 all passed. A scoped bridge subscription delivered `BridgeDestroyed` after deletion.
- An isolated external-media channel targeting the PBX's own `127.0.0.1:9` passed channel data, single and bulk variables, bridge add/remove, a `ChannelVarset` event roundtrip, hangup, and post-hangup 404. RTP statistics returned 404 (`RTP info not found`) for this channel type; that is not evidence the route fails for ordinary RTP-bearing channels.
- The proxy server dispatch path passed live bridge create with variables, variable update/read, delete, and post-delete 404. Responses were JSON roundtripped through a test bus sink.
- A live `StasisStart` from a test external-media channel passed native websocket decoding, proxy server event forwarding, JSON serialization, and proxy client decoding with matching JSON after an event-model fix. This used an in-memory message-bus substitute, not NATS or RabbitMQ.
- The proxy stored-recording download path opened one existing file, transferred one bounded 32 KiB chunk through JSON, and closed the proxy session. It did not save or print recording content. The file advertised `audio/wav` and 49,644 bytes.

## Bugs found and fixed

1. Asterisk 22.10.1 rejects a nested `variables` object on collection bridge creation with HTTP 400, yet may create the bridge before returning that error. A flat variable map in the JSON body with bridge options in the query succeeded. The pinned [Asterisk 23 collection handler](https://github.com/asterisk/asterisk/blob/97d55b3306ca4aa26c0136c67b79470ca4b2b785/res/ari/resource_bridges.c) passes the body directly to its variable parser, while its path-ID handler extracts the nested object. `ari` now uses the verified flat shape only for collection creation with variables; the path-ID route and other JSON-body calls retain their existing encoding. The bridge left by the rejected request was found and deleted.
2. A wrapped native `Data` error hid its HTTP status because `errDataGet` lacked `Unwrap`. It now preserves the underlying 404 for callers and proxy responses.
3. An absent optional `replace_channel` in a live `StasisStart` was serialized as a fabricated empty channel with a generated timestamp. The event generator now uses `omitzero` for optional object payloads. A two-hop regression test and the live proxy event test pass.

Cleanup verification after all state-changing tests found **zero bridges and zero channels**. `ari-testing` was no longer listed after websocket disconnect. No test recording was created or retained.

## Limits and next checks

This run did not start a proxy NATS/RabbitMQ backend or test ordinary telephony calls, endpoint REFER, event claim, all option combinations, or every event type. It is partial progress on implementation sequence #2, not live certification for all 109 operations or Asterisk 23.

The opt-in proxy dispatch test reads its password from stdin. Compile it first with `go test -c -o /tmp/ari-proxy-server.test ./server`, then set `ARI_LIVE_READONLY=1`, `ARI_LIVE_URL`, and `ARI_LIVE_USERNAME` when running the binary with `-test.run '^TestLiveReadOnlyDispatch$'`. Supply the password through stdin; do not store it in this repository or include it in command-line arguments.

The state-changing opt-in tests use the same compiled binary and credential input. Set one of `ARI_LIVE_BRIDGE_TEST`, `ARI_LIVE_EVENT_TEST`, or `ARI_LIVE_RECORDING_TEST` to `1`, and select its corresponding test by name. Run them individually because each reads one password line from stdin. Only run bridge and event tests on a PBX where creating and cleaning up isolated ARI resources is authorized.

## Follow-up validation: 2026-10-01

The dev PBX was checked again using a race-instrumented server test binary
(`go test -race -c`). All ten read-only dispatch cases passed before and after
the state-changing tests. Bridge collection creation, initial variable,
bulk update/read, deletion, and post-delete 404 passed. The websocket
`StasisStart` native/proxy JSON forwarding test also passed under `ari-testing`.

The first bridge run exposed a native websocket shutdown race: `Client.Close`
and the listener both wrote `connected`, while `Connected` read it without
synchronization. The flag now uses `atomic.Bool`; the local websocket regression
`TestClientConcurrentShutdown` passed five race-instrumented runs. All live
checks were rerun successfully after the fix.

Stored-recording validation now compares a complete bounded file through
native and proxy streaming without saving or displaying its contents. The
49,644-byte WAV matched by SHA-256 and byte count. Retrying sequence zero
returned identical data, subsequent chunks reached EOF, and the proxy file
session was closed. Full comparison is limited to files with a known size of
at most 4 MiB; larger or unknown-size files receive the bounded first-chunk
and retry checks. Cleanup uses its own context rather than the test context.

Final lists showed zero bridges and zero channels, four applications (matching
the starting count), and six stored recordings (unchanged). These tests still
use an in-memory proxy transport; NATS/RabbitMQ and Asterisk 20/23 remain live
validation gaps.
