# Paqet v1.0.0-alpha.21 PV hardening audit

This document records the production-focused audit applied on top of upstream `hanselime/paqet` `v1.0.0-alpha.21`.

## Upstream baseline

- Upstream release: `v1.0.0-alpha.21`
- Upstream commit: `4b81ce0ab56b7cc9aec7501c87c4804676cdb1b4`
- The audit branch records that commit as an actual Git parent, so future upstream comparisons are ancestry-correct rather than content-only approximations.
- Go toolchain and release builders are aligned on Go 1.27.
- `golang.org/x/crypto` is aligned to v0.55.0 and `golang.org/x/net` to v0.57.0.

## Reliability and correctness fixes

- Added context-aware capped reconnect backoff to avoid CPU/log storms during outages.
- Close partially initialized KCP/smux connections when TCP-flag initialization fails.
- Clean up already-created sessions if multi-connection client startup fails partway through.
- Added initial protocol-handshake deadline on server streams.
- Added backoff on persistent accept/read errors in server, forwarder and SOCKS paths.
- Corrected protocol short-write handling and tightened malformed-frame validation.
- Added validation for manual KCP parameters, FEC shard combinations and SMUX buffer/keepalive relationships.
- Hardened SOCKS5 request/datagram validation and response writes.
- Synchronized pcap send/close operations and closed pcap handles on initialization failures.
- `PacketConn.LocalAddr()` now returns a stable cloned local address instead of `nil`, removing `%!s(<nil>)` log output and improving address reporting.
- Expected `context.Canceled` shutdown errors are suppressed by the normal error filter rather than being emitted as false ERROR logs.

## Static Linux builds

Two x86-64 Linux artifacts are produced from the same audited source:

- `amd64`: `GOAMD64=v1`, compatibility build for old and new x86-64 CPUs.
- `amd64-v3`: optimized modern build for CPUs that satisfy Go's complete amd64 v3 feature level.

Both are built with musl and a statically built libpcap. The pipeline rejects an artifact if ELF inspection finds a dynamic interpreter (`INTERP`) or shared-library dependency (`NEEDED`). Each archive also carries a SHA-256 file for its binary.

## Verification gates

The audit CI requires all of the following to pass:

- `gofmt`
- Linux installer shell syntax (`bash -n`)
- shuffled unit tests
- Go race detector
- `go vet`
- Staticcheck
- govulncheck
- normal Go build
- Linux `GOAMD64=v1` build
- Linux `GOAMD64=v3` build
- archive content and SHA-256 verification
- static ELF verification

A release must not be treated as verified unless the corresponding workflow runs are green.

## Upstream reports reviewed

The audit also reviewed recent and relevant upstream reports, including long-running reconnection/log symptoms, historical shared-libpcap Linux packaging problems, raw-packet provider filtering, and throughput tuning guidance.

Provider/network filtering is not a software defect: a network can still block or alter crafted raw TCP packets. Likewise, KCP performance is path-dependent; no single preset can guarantee the same throughput/latency on every network.

## Compatibility notes

The fork retains upstream alpha.21 protocol framing and KCP stream-mode behavior. The deprecated `SetStreamMode(true)` call is intentionally retained because kcp-go v5.6.72 exposes no non-deprecated public replacement and changing it would alter transport semantics. Staticcheck suppression is narrowly scoped to that call and documented in source.

No software can be proven to contain zero bugs. The release claim for this fork is therefore limited to the concrete checks and regression tests above plus the upstream issues examined at release time.
