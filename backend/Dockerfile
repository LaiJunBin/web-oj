FROM golang:1.27 as go_app_dev

WORKDIR /app/backend
COPY go.mod go.sum ./

RUN go mod download
RUN go install github.com/air-verse/air@v1.52.3
RUN go install github.com/laijunbin/go-migrate@latest
RUN go install github.com/roblaszczak/go-cleanarch@latest
RUN go install github.com/vektra/mockery/v2@v2.45.0

COPY . .
RUN go mod tidy

# api/api.gen.go is untracked. Generating at image build validates the spec
# early and pre-warms the generator's module cache. It does NOT make container
# start network-free: `go run pkg@version` performs a module lookup on every
# invocation, so the entrypoint's regeneration needs outbound network — an
# accepted premise (the judge host has internet; only contestants are limited
# to the venue LAN). The generated file is shadowed by the compose mount at
# runtime; the entrypoint regenerates it into the mounted tree on every start.
RUN make api

ENTRYPOINT ["./entrypoint.sh"]
CMD air



# Build the production binary with the SAME Go toolchain go.mod requires
# (1.23) — the previous 1.22 here was older than the module's `go 1.23.0`
# directive, an unreproducible-build footgun.
#
# Further hardening to do with a build to verify it (left as a follow-up rather
# than applied blind here): split a minimal non-root runtime stage that carries
# only the binary + CA certs instead of the full golang image with compiler and
# source, and pin every `go install ...@latest` to a tag. Note the gorm sqlite
# driver is cgo, so a fully static/scratch image needs the sqlite import dropped
# or swapped for a pure-Go driver first.
FROM golang:1.27 as go_app_prod

WORKDIR /app
COPY . .

RUN go mod download
RUN go install github.com/laijunbin/go-migrate@latest
# api/api.gen.go is untracked; nothing is mounted over this stage, so
# build-time generation is the real thing the binary compiles against.
RUN make api
RUN go build -o /app/bin/go_app cmd/main.go
ENTRYPOINT ["./entrypoint.sh"]
CMD /app/bin/go_app

