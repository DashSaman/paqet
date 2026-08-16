# Low-overhead KCP profile

This fork adds an opt-in `efficient` KCP mode for healthy, high-throughput paths where reducing redundant packets and protocol chatter matters more than aggressive loss recovery.

## Configuration

Use the same KCP mode and key on both ends:

```yaml
transport:
  protocol: "kcp"
  conn: 1
  kcp:
    mode: "efficient"
    block: "aes-128"
    key: "replace-with-a-long-random-key"
```

When omitted, the `efficient` profile supplies these defaults:

```yaml
mtu: 1420
sndwnd: 4096
rcvwnd: 4096
smuxbuf: 8388608
streambuf: 4194304
smuxkalive: 5
smuxktimeout: 20
```

The KCP behavior is equivalent to:

```yaml
nodelay: 0
interval: 10
resend: 0
nocongestion: 1
wdelay: true
acknodelay: false
```

Existing `normal`, `fast`, `fast2`, `fast3`, and `manual` modes retain their upstream defaults.

## Why

Aggressive fast-resend and immediate ACK behavior can be useful on throttled or lossy paths, but it can also create unnecessary retransmission/ACK traffic on a healthy path. The `efficient` profile is intentionally conservative about redundant traffic while keeping a short KCP update interval and large windows for high-bandwidth, higher-RTT links.

## When not to use it

If the path has sustained packet loss, active throttling, or rapidly changing latency, compare `efficient` against `fast3` or a tuned `manual` profile. Do not assume one profile is universally faster.

## Connections

Start with `conn: 1` for the lowest overhead. Increase it only when aggregate multi-stream throughput is measurably limited. Every extra connection creates another KCP/SMUX session and packet I/O state; it does not automatically stripe one application stream across all connections.

## Benchmarking

Compare profiles under the same route and load. Record at minimum:

- useful application throughput
- interface RX/TX bytes
- packets per second
- CPU usage
- packet loss/retransmission behavior
- latency and jitter

The goal is not the highest benchmark number at any cost; it is the best ratio of useful payload to wire traffic while preserving stable latency and throughput.
