# Linux installation

This fork publishes two x86-64 Linux builds in addition to the upstream architecture set:

- `amd64` — `GOAMD64=v1`, the compatibility build for old and new 64-bit x86 CPUs.
- `amd64-v3` — `GOAMD64=v3`, the optimized build for systems that provide the complete v3 feature set (including AVX/AVX2, BMI1/BMI2, F16C, FMA, LZCNT, MOVBE and OSXSAVE).

Both binaries are built as static Linux executables with a statically built libpcap, following the upstream musl release approach. This avoids requiring a matching `libpcap.so` at runtime.

## Automatic installation

The installer defaults to `auto`. On x86-64 it selects `amd64-v3` only when the complete v3 CPU feature set can be confirmed; otherwise it safely falls back to `amd64`.

```bash
curl -fsSL https://raw.githubusercontent.com/DashSaman/paqet/master/scripts/install-linux.sh | sudo bash
```

Force the compatibility build:

```bash
curl -fsSL https://raw.githubusercontent.com/DashSaman/paqet/master/scripts/install-linux.sh | sudo PAQET_CPU=baseline bash
```

Force the modern build (the installer refuses to continue if the CPU does not satisfy v3):

```bash
curl -fsSL https://raw.githubusercontent.com/DashSaman/paqet/master/scripts/install-linux.sh | sudo PAQET_CPU=modern bash
```

The default custom release is `v1.0.0-alpha.21-pv1`. Override it when testing another published version:

```bash
curl -fsSL https://raw.githubusercontent.com/DashSaman/paqet/master/scripts/install-linux.sh | sudo PAQET_VERSION=v1.0.0-alpha.21-pv1 bash
```

## Manual installation

Compatibility / old CPU:

```bash
VERSION=v1.0.0-alpha.21-pv1
curl -fLO "https://github.com/DashSaman/paqet/releases/download/${VERSION}/paqet-linux-amd64-${VERSION}.tar.gz"
tar -xzf "paqet-linux-amd64-${VERSION}.tar.gz"
sudo install -m 0755 paqet_linux_amd64 /usr/local/bin/paqet
paqet version
```

Modern x86-64 / GOAMD64 v3:

```bash
VERSION=v1.0.0-alpha.21-pv1
curl -fLO "https://github.com/DashSaman/paqet/releases/download/${VERSION}/paqet-linux-amd64-v3-${VERSION}.tar.gz"
tar -xzf "paqet-linux-amd64-v3-${VERSION}.tar.gz"
sha256sum -c paqet_linux_amd64-v3.sha256
sudo install -m 0755 paqet_linux_amd64-v3 /usr/local/bin/paqet
paqet version
```

## Runtime requirements

Paqet still needs the same runtime privileges and network configuration as upstream because it captures and injects raw packets through pcap. A static binary removes the shared-libpcap dependency; it does not remove the need for the required packet-capture privileges.
