FROM node:22-bookworm-slim

# The slim base ships NO system CA bundle (and no git). Node carries its own
# CA store, so pnpm works regardless — but the dev container installs git via
# a feature, and git verifies TLS against the SYSTEM bundle: without this,
# every https git operation inside the container fails certificate
# verification (first victim: `git submodule update --init`).
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/*

ENV PNPM_HOME=/pnpm
ENV PATH=$PNPM_HOME:$PATH
ENV HOST=0.0.0.0
ENV CHOKIDAR_USEPOLLING=true

WORKDIR /app

RUN corepack enable

# pnpm-workspace.yaml is load-bearing here, not optional: pnpm 10+ reads the
# dependency overrides and allowBuilds from it, the lockfile records those
# overrides, and a frozen install refuses to proceed when the two disagree —
# so omitting the file fails the build outright (and even a non-frozen
# install would silently skip the allowed build scripts). .npmrc carries
# engine-strict, which should hold inside the image too.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./

RUN pnpm install --frozen-lockfile

COPY . .

EXPOSE 5173

# No "--" separator: npm needs it, but pnpm forwards it VERBATIM, so vite
# received `dev -- --host 0.0.0.0`, treated the flags as positional leftovers,
# and bound only the container's localhost — the published port then reset
# every connection. Vite's own log line "Network: use --host to expose" is the
# tell if this ever regresses.
CMD ["pnpm", "run", "dev", "--host", "0.0.0.0"]