#!/bin/bash
set -e

# The deployment image compiled the server at build time — regenerating
# api.gen.go at start would add a network round-trip (go run pkg@version
# contacts the module proxy on EVERY invocation, cache or not) for a file
# nothing would recompile. Only the dev overlay needs start-time generation:
# it mounts the host tree over the image, shadowing anything generated at
# build, so container start is the one moment generation lands where the
# code runs. A failure fails the start loudly (set -e); silently reusing a
# stale generated file would hide the real problem.
if [ ! -x /app/bin/go_app ]; then
	make api
	# The compose service runs as root, so on rootful Linux Docker the file
	# just written into the bind mount would be root-owned and block host-side
	# tools from rewriting it later. Align it with the tree's owner; harmless
	# elsewhere (Docker Desktop maps ownership itself; non-root runs skip).
	chown --reference=. api/api.gen.go 2>/dev/null || true
fi

exec "$@"
