#!/usr/bin/env bash
set -Eeuo pipefail

REPO="${PAQET_REPO:-DashSaman/paqet}"
VERSION="${PAQET_VERSION:-v1.0.0-alpha.21-pv1}"
INSTALL_DIR="${PAQET_INSTALL_DIR:-/usr/local/bin}"
CPU_MODE="${PAQET_CPU:-auto}"

fail() {
  printf 'paqet installer: %s\n' "$*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

has_cpu_flag() {
  local flags="$1" flag="$2"
  case " $flags " in
    *" $flag "*) return 0 ;;
    *) return 1 ;;
  esac
}

amd64_supports_v3() {
  [[ -r /proc/cpuinfo ]] || return 1

  local flags
  flags="$(awk -F: '/^(flags|Features)[[:space:]]*:/ {print $2; exit}' /proc/cpuinfo | tr '[:upper:]' '[:lower:]')"
  [[ -n "$flags" ]] || return 1

  # Go's amd64 v3 level is v2 plus AVX/AVX2/BMI/FMA/F16C/LZCNT/MOVBE.
  # Linux names SSE3 as pni and commonly exposes LZCNT as abm. AVX being
  # exposed by Linux also means the OS has enabled the required XSAVE state.
  local flag
  for flag in cx16 lahf_lm popcnt pni ssse3 sse4_1 sse4_2 avx avx2 bmi1 bmi2 f16c fma movbe; do
    has_cpu_flag "$flags" "$flag" || return 1
  done
  if ! has_cpu_flag "$flags" lzcnt && ! has_cpu_flag "$flags" abm; then
    return 1
  fi
  return 0
}

select_asset() {
  local machine
  machine="$(uname -m)"

  case "$machine" in
    x86_64|amd64)
      case "$CPU_MODE" in
        auto)
          if amd64_supports_v3; then
            printf '%s\n' 'amd64-v3'
          else
            printf '%s\n' 'amd64'
          fi
          ;;
        baseline|old|v1|amd64)
          printf '%s\n' 'amd64'
          ;;
        modern|new|v3|amd64-v3)
          amd64_supports_v3 || fail "this machine does not satisfy the full GOAMD64=v3 feature set"
          printf '%s\n' 'amd64-v3'
          ;;
        *) fail "invalid PAQET_CPU=$CPU_MODE (use auto, baseline, or modern)" ;;
      esac
      ;;
    aarch64|arm64) printf '%s\n' 'arm64' ;;
    armv7l) printf '%s\n' 'arm32' ;;
    mips) printf '%s\n' 'mips' ;;
    mipsel) printf '%s\n' 'mipsle' ;;
    mips64) printf '%s\n' 'mips64' ;;
    mips64el) printf '%s\n' 'mips64le' ;;
    *) fail "unsupported Linux architecture: $machine" ;;
  esac
}

need curl
need tar
need install
need uname

asset="$(select_asset)"
archive="paqet-linux-${asset}-${VERSION}.tar.gz"
binary="paqet_linux_${asset}"
url="https://github.com/${REPO}/releases/download/${VERSION}/${archive}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

printf 'Installing Paqet %s (%s) from %s\n' "$VERSION" "$asset" "$REPO"
curl --fail --location --proto '=https' --tlsv1.2 --retry 3 --retry-delay 1 \
  "$url" -o "$tmp/$archive"

tar -xzf "$tmp/$archive" -C "$tmp"
[[ -f "$tmp/$binary" ]] || fail "archive does not contain $binary"

if [[ -f "$tmp/${binary}.sha256" ]]; then
  need sha256sum
  (
    cd "$tmp"
    sha256sum -c "${binary}.sha256"
  )
fi

mkdir -p "$INSTALL_DIR"
install -m 0755 "$tmp/$binary" "$INSTALL_DIR/paqet"

printf 'Installed: %s/paqet\n' "$INSTALL_DIR"
"$INSTALL_DIR/paqet" version || true
