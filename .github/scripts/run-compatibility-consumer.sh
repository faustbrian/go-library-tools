#!/usr/bin/env bash
set -euo pipefail
task="$(mktemp -d "${RUNNER_TEMP}/golib-compatibility-consumer.XXXXXX")"
# EXIT invokes cleanup on success, staging refusal, test failure and cancellation.
# shellcheck disable=SC2329
cleanup() {
  result=$?
  trap - EXIT
  chmod -R u+w "${task}" || result=1
  find "${task}" -depth -delete || result=1
  exit "${result}"
}
trap cleanup EXIT
# Give each long-running command its own process group so cancellation also
# reaches its compiler/test descendants, not only the waiting runner shell.
set -m
child=''
admitting=false
pending_signal=''
# shellcheck disable=SC2329
cancel() {
  if [[ "${admitting}" == true ]]; then
    pending_signal=$1
    return
  fi
  trap '' INT TERM
  if [[ -n "${child}" ]]; then
    kill -TERM -- "-${child}" 2>/dev/null || true
    wait "${child}" 2>/dev/null || true
    child=''
  fi
  exit "$1"
}
trap 'cancel 130' INT
trap 'cancel 143' TERM
run_command() {
  local result
  admitting=true
  "$@" &
  child=$!
  admitting=false
  if [[ -n "${pending_signal}" ]]; then cancel "${pending_signal}"; fi
  if wait "${child}"; then result=0; else result=$?; fi
  child=''
  return "${result}"
}
mkdir -p "${task}/cache" "${task}/consumer" "${task}/mod" "${task}/tmp"
python3 - "${task}/consumer" <<'PY'
import json,os,re,shutil,stat,sys
encoded = os.environ['SCENARIO_FILES']
if len(encoded) > 16384:
    raise ValueError('selection budget exceeded')
files = json.loads(encoded)
if not isinstance(files,list) or not files or len(files)>64 or len(set(files))!=len(files):
    raise ValueError('missing or ambiguous consumer selection')
source = 'release/compatibility-consumer'
for parent in ('release',source):
    if not stat.S_ISDIR(os.lstat(parent).st_mode):
        raise ValueError('unsafe consumer directory')
for name in files:
    if not isinstance(name,str) or not re.fullmatch(r'[a-z0-9_]+\.go',name):
        raise ValueError('invalid consumer source')
for name in ['go.mod','go.sum',*files]:
    path = os.path.join(source,name)
    info = os.lstat(path)
    if not stat.S_ISREG(info.st_mode) or info.st_size > 1024*1024:
        raise ValueError('missing or unsafe consumer input')
    shutil.copyfile(path,os.path.join(sys.argv[1],name),follow_symlinks=False)
PY
cd "${task}/consumer"
export GOCACHE="${task}/cache" GOMODCACHE="${task}/mod" GOTMPDIR="${task}/tmp"
export GOENV=off GOTOOLCHAIN=local GOWORK=off GOMAXPROCS=2 GOFLAGS=-p=1
export GOPROXY=https://proxy.golang.org GOSUMDB=sum.golang.org
export GOPRIVATE='' GONOPROXY='' GONOSUMDB=''
test "$(go env GOVERSION)" = go1.27.2
test "$(go env GOOS)/$(go env GOARCH)" = "${EXPECTED_GOOS}/${EXPECTED_GOARCH}"
go mod edit -json | python3 -c 'import json,sys; assert not json.load(sys.stdin).get("Replace"), "public consumer forbids replacements"'
# Authenticate public resolution; tidy prunes unused requirements in this disposable module.
run_command env GODEBUG=http2client=0 go mod tidy
run_command go test -mod=readonly -count=1 -timeout=20m ./...
