GO_EXECUTABLE := /usr/local/go/bin/go

.PHONY: generate\:jwt-key
generate\:jwt-key:
	./scripts/generate-jwt-key.sh

.PHONY: generate\:app-key
generate\:app-key:
	./scripts/generate-app-key.sh

.PHONY: generate\:csrf-key
generate\:csrf-key:
	./scripts/generate-csrf-key.sh

.PHONY: generate\:store-key
generate\:store-key:
	./scripts/generate-store-key.sh

.PHONY: generate\:keys
generate\:keys:
	./scripts/generate-jwt-key.sh
	./scripts/generate-app-key.sh
	./scripts/generate-csrf-key.sh
	./scripts/generate-store-key.sh

.PHONY: detect\:docker-gateway
detect\:docker-gateway:
	./scripts/detect-docker-gateway.sh

.PHONY: api
# api/api.gen.go is NOT tracked; this target is how every environment gets it.
# The generator version is pinned HERE and only here — that pin is what makes
# each environment produce identical code. v2.2.0's go.mod still declares the
# pre-rename deepmap module path; "correcting" it to github.com/oapi-codegen/…
# fails with a module-path mismatch. When upgrading: bump the pin, regenerate,
# run the full test suite.
api:
	go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.2.0 --config=gen.config.yaml ./api/openapi.yaml

.PHONY: migrate
migrate:
	go run cmd/migrate/main.go

.PHONY: seed
seed:
	go run cmd/seed/main.go

.PHONY: up
up:
	docker compose up -d

.PHONY: up\:dev
# Hot-reload development (air + source mount) — never at a venue; see
# docker-compose.dev.yml's header for why.
up\:dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

.PHONY: up\:monitor
# Main stack + monitoring overlay. Use THIS (not plain `up`) for every
# restart/update of a monitored host: compose converges containers to
# whatever the -f list describes, so an `up` without the overlay quietly
# rebuilds app without SERVER_METRICSHOST and every Grafana panel goes
# flat with no error anywhere.
up\:monitor:
	docker compose -f docker-compose.yml -f docker-compose.monitor.yml up -d

.PHONY: dev
dev:
	air

.PHONY: build
build:
	go build -o bin/app ./cmd/main.go

.PHONY: mocks
mocks:
	rm -rf internal/mocks internal/app/shared/mocks; \
	mockery --config "" --all --outpkg sharedmocks --dir internal/app/shared --output internal/app/shared/mocks; \
	for dir in \
		internal/app/identity_domain/auth \
		internal/app/identity_domain/user \
		internal/app/evaluation_environment_domain/execution_environment \
		internal/app/project_domain/project \
		internal/app/contest_domain/contest \
		internal/app/contest_domain/activity \
		internal/app/judge_domain/judge \
		internal/app/judge_domain/worker; do \
		rm -rf $$dir/mocks; \
		mockery --config "" --all --recursive --dir $$dir --output $$dir/mocks; \
	done; \
	mockery
	
TEST_FLAGS := -v -count=1
TEST_UNIT_PACKAGES := internal
TEST_FEATURE_PACKAGES := ./tests/features/...
.PHONY: test\:unit
TEST_UNIT_ARG=$(filter-out $@,$(MAKECMDGOALS))
test\:unit:
	if [ -z "${TEST_UNIT_ARG}" ]; then \
		go test $(TEST_FLAGS) ./$(TEST_UNIT_PACKAGES)/... -failfast; \
	else \
		if [ -d "${TEST_UNIT_ARG}" ]; then \
			TARGET_DIR=$(shell echo "${TEST_UNIT_ARG}" | sed 's|^\./||; s|/$$||'); \
			go test $(TEST_FLAGS) ./$${TARGET_DIR}/... -failfast; \
		else \
			echo "Target directory not found"; \
			exit 1; \
		fi \
	fi

.PHONY: test\:feature
TEST_FEATURE_PACKAGES_ARG=$(filter-out $@,$(MAKECMDGOALS))
test\:feature:
	if [ -z "${TEST_FEATURE_PACKAGES_ARG}" ]; then \
		go run tests/test.go -count=1 $(TEST_FEATURE_PACKAGES); \
	else \
		if [ -d "${TEST_FEATURE_PACKAGES_ARG}" ]; then \
			TARGET_DIR=$(shell echo "${TEST_FEATURE_PACKAGES_ARG}" | sed 's|^\./||; s|/$$||'); \
			go run tests/test.go -count=1 ./$${TARGET_DIR}/...; \
		else \
			echo "Target directory not found"; \
			exit 1; \
		fi \
	fi

# Runs the short-TTL renewal feature on its own. The access lifetime is
# overridden for the whole process — viper's AutomaticEnv covers keys already
# present in the config file — so this cannot share a run with the rest of
# the suite: a two-second access token would make every other scenario flaky.
# TARGET_TEST keeps the other auth features from running alongside it.
.PHONY: test\:feature\:short-ttl
test\:feature\:short-ttl:
	AUTH_ACCESSTOKENTTLSECONDS=2 TARGET_TEST=short_ttl_renewal \
		go run tests/test.go -count=1 ./tests/features/identity_domain/auth/...

.PHONY: test\:unit\:watch
TEST_UNIT_WATCH_ARG=$(filter-out $@,$(MAKECMDGOALS))
test\:unit\:watch:
	if [ -z "${TEST_UNIT_WATCH_ARG}" ]; then \
		air --build.exclude_regex "" --build.include_dir="$(TEST_UNIT_PACKAGES)" --build.cmd "go test -count=1 ./$(TEST_UNIT_PACKAGES)/..." --build.bin "$(GO_EXECUTABLE) test -count=1 ./$(TEST_UNIT_PACKAGES)/..."; \
	else \
		if [ -d "${TEST_UNIT_WATCH_ARG}" ]; then \
			TARGET_DIR=$(shell echo "${TEST_UNIT_WATCH_ARG}" | sed 's|^\./||; s|/$$||'); \
			air --build.exclude_regex "" --build.include_dir="$${TARGET_DIR}" --build.cmd "go test -count=1 ./$${TARGET_DIR}/..." --build.bin "$(GO_EXECUTABLE) test -count=1 ./$${TARGET_DIR}/..."; \
		else \
			echo "Target directory not found"; \
			exit 1; \
		fi \
	fi
	

.PHONY: test\:unit\:coverage
test\:unit\:coverage:
	go test $(TEST_FLAGS) ./internal/app/.../application/usecase/... ./internal/app/.../domain/entity/... ./internal/app/.../domain/valueobject/... ./internal/app/.../infrastructure/repository/... ./internal/app/.../infrastructure/service/... -coverprofile=coverage.out
	go tool cover -func=coverage.out

# .PHONY: test
# test:
# 	go test $(TEST_FLAGS) $(TEST_UNIT_PACKAGES) $(TEST_FEATURE_PACKAGES)

# .PHONY: test\:feature\:watch
# test\:feature\:watch:
# 	air --build.cmd "go run tests/test.go $(TEST_FLAGS) $(TEST_FEATURE_PACKAGES)" --build.bin "${GO_EXECUTABLE} run tests/test.go $(TEST_FLAGS) $(TEST_FEATURE_PACKAGES)"
	
# .PHONY: test\:watch
# test\:watch:
# 	air --build.cmd "go test $(TEST_FLAGS) $(TEST_UNIT_PACKAGES) $(TEST_FEATURE_PACKAGES)" --build.bin "${GO_EXECUTABLE} test $(TEST_FLAGS) $(TEST_UNIT_PACKAGES) $(TEST_FEATURE_PACKAGES)"


.PHONY: down
down:
	docker compose down
