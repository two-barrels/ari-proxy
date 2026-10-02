<!-- Created by two-barrels in 2026 for ARI v6 modernization. SPDX-License-Identifier: Apache-2.0 -->

# ARI proxy v6.0.0-rc.2 release notes

Module: `github.com/two-barrels/ari-proxy/v6`. This is a prerelease Go module;
stable v6 and binary/container publication remain future work.

## Changes since rc.1

- Added prominent two-barrels modification/creation notices and retained upstream
  attribution; added NOTICE and checksum companion notices.
- CI checks per-file notices against a pinned upstream baseline.
- Standalone snapshot archives retain LICENSE, NOTICE and companion notices,
  with regression coverage.
- Updated the native dependency and contract source audit to ARI rc.2.
- Proxy APIs and runtime behavior are unchanged from rc.1.

## Changes

- Lossless event forwarding preserves unknown event types and fields, nested
  payloads, and keyless events. Dialog copies retain independent routing metadata;
  shared nested payloads must be treated as immutable.
- Explicit proxy methods reach all 108 pinned REST routes; ARI's event websocket
  is represented by message-bus subscriptions. All 175 non-path parameters have
  native/proxy wire evidence.
- Added newer operations, complete option forwarding, explicit route variants,
  and assigned-ID response handling.
- Stored recording transfer uses bounded 32 KiB chunks, node-bound sessions,
  retry handling, and cleanup instead of whole-file JSON buffering.
- Preserve HTTP status codes through proxy errors. Proxy Move uses `app_args`,
  translated to native ARI `appArgs`.
- Updated NATS to 1.53.0 and RabbitMQ amqp091-go to 1.15.0, plus networking,
  compression, and test dependencies. Go minimum is 1.25.0; recommended
  toolchain is patched Go 1.26.8.

## Migration and release dependency

Review [v6 migration](v6-migration.md) before updating imports. The public API
uses ARI v6 types, requiring a proxy major-version migration. The proxy requires
published `github.com/two-barrels/ari/v6 v6.0.0-rc.2`; no local replacement remains.
Normal Go CI resolves that dependency without a sibling checkout. Contract CI
checks out the same ARI tag solely to inspect source evidence.

Existing v5 tags remain available. Phone-apps has not been migrated. The user
reports no current consumers of the existing proxy fork.

## Validation and limits

Hosted Go and ARI-contract workflows passed on modernization commit `74764bd`:
Go 1.25/1.26.8 builds, examples, race tests, vet, module verification, security
scan, all 109 contract operations/175 parameters, and standalone snapshot
consumer builds without replacements. The security scan reported zero reachable
issues; three module-only advisories concern unimported x/crypto SSH/OpenPGP.

Selected proxy/native operations passed Asterisk 22.10.1 live checks, including
recording hash comparison and cleanup. The user accepts this for current scope
and defers further PBX validation. NATS/RabbitMQ live tests are deferred for now;
broker behavior after dependency updates is unverified. Remaining escaping,
context, timeout, throughput, and multi-node error audits remain open.

## Artifact publication

ARI v6.0.0-rc.2 was published first; proxy v6.0.0-rc.2 uses that exact tag. Legacy
image/GoReleaser jobs are restricted to CyCore upstream because their tooling
and registry targets require separate two-barrels preparation. Source CI does
not establish binary/container artifact readiness.
