#!/usr/bin/env bash
# Compile unit-tagged tests, then run with clean env, fresh cwd and OS network guard.
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
if [[ $(uname -s) != Darwin ]] || [[ ! -x /usr/bin/sandbox-exec ]]; then
  echo 'This runner requires macOS sandbox-exec; refusing an unisolated fallback.' >&2
  exit 2
fi
base=${TMPDIR:-/tmp}
work=$(mktemp -d "${base%/}/scopesentry-unit.XXXXXX")
echo "Isolated test evidence: $work" >&2
mkdir -p "$work/home"

# Only loopback is allowed for httptest. The build tag disables all store/config
# startup; environment cleanup additionally removes production credentials.
cat > "$work/network.sb" <<'EOF'
(version 1)
(allow default)
(deny network*)
(allow network-bind (local ip "localhost:*"))
(allow network-inbound (local ip "localhost:*"))
(allow network-outbound (remote ip "localhost:*"))
EOF
cat > "$work/check.py" <<'EOF'
import errno, socket
server = socket.socket()
server.bind(('127.0.0.1', 0))
server.listen(1)
client = socket.socket()
client.connect(server.getsockname())
peer, _ = server.accept()
peer.close(); client.close(); server.close()
for address in [('192.0.2.1', 80), ('198.51.100.1', 27017)]:
    s = socket.socket()
    s.settimeout(1)
    try:
        s.connect(address)
    except OSError as error:
        assert error.errno in (errno.EPERM, errno.EACCES), error
    else:
        raise AssertionError('network guard allowed external connection')
    finally:
        s.close()
print('PASS: OS network guard blocks external connections; loopback fixtures allowed')
EOF
/usr/bin/sandbox-exec -f "$work/network.sb" /usr/bin/python3 "$work/check.py"

cd "$root"
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
packages=("$@")
if [[ ${#packages[@]} == 0 ]]; then packages=(./...); fi
go list -tags scopesentry_unit -f '{{if .TestGoFiles}}{{.ImportPath}}{{else if .XTestGoFiles}}{{.ImportPath}}{{end}}' "${packages[@]}" > "$work/packages.txt"
if [[ -z $(sed '/^[[:space:]]*$/d' "$work/packages.txt") ]]; then
  echo 'No test packages selected.' >&2
  exit 2
fi
index=0
while IFS= read -r package; do
  [[ -n $package ]] || continue
  index=$((index + 1))
  dir="$work/package-$index"
  mkdir -p "$dir/home"
  binary="$dir/unit.test"
  flags=(-tags scopesentry_unit)
  if [[ ${RACE:-0} == 1 ]]; then flags+=(-race); fi
  go test "${flags[@]}" -c -o "$binary" "$package"
  echo "RUN $package" >&2
  (
    cd "$dir"
    /usr/bin/env -i PATH=/usr/bin:/bin HOME="$dir/home" TMPDIR="$dir/" TZ=UTC \
      MONGODB_IP=198.51.100.1 REDIS_IP=198.51.100.1 \
      /usr/bin/sandbox-exec -f "$work/network.sb" "$binary" -test.v -test.count=1 -test.timeout=90s
  ) > "$dir/result.log" 2>&1 || {
    cat "$dir/result.log" >&2
    exit 1
  }
  cat "$dir/result.log"
done < "$work/packages.txt"
[[ $index -gt 0 ]] || { echo 'No test packages selected.' >&2; exit 2; }
echo "PASS: $index isolated test packages; evidence: $work"
