# Optional compact TCP headers

Paqet normally adds a TCP timestamp option to raw TCP packets. This remains the default for upstream compatibility and traffic-shape continuity.

This fork adds an opt-in setting:

```yaml
network:
  tcp:
    local_flag: ["PA"]
    remote_flag: ["PA"]
    timestamp: false
```

For ordinary non-SYN packets, disabling timestamps removes the two NOP bytes plus the 10-byte timestamp option, reducing the generated TCP header by 12 bytes per packet. Packet count is unchanged; this only reduces per-packet wire bytes and a small amount of header-generation work.

SYN packets still keep MSS, SACK-permitted, and window-scale options and are padded to a 32-bit boundary.

## Compatibility

- If `timestamp` is omitted, timestamps remain enabled exactly as before.
- Use `timestamp: false` only as an A/B-tested optimization. Raw TCP traffic shape changes slightly, so deployments that care about matching a particular traffic fingerprint should measure before enabling it broadly.
- This setting is independent from KCP `efficient` mode. You can test either optimization separately.
