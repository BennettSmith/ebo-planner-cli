.DEFAULT_GOAL := help

.PHONY: help ci gen changelog-verify changelog-release release-help

SHELL := /bin/bash

# --- Local-dev helpers (CLI) ---
#
# These targets configure CLI profiles for local Docker Compose (API + Keycloak).
# They write to your normal ebo config location (or whatever EBO_CONFIG_DIR points at).
#
EBO_BIN ?= ./bin/ebo
EBO_API_URL ?= http://localhost:8081
EBO_OIDC_ISSUER_URL ?= http://localhost:8082/realms/ebo
EBO_OIDC_CLIENT_ID ?= ebo-client
EBO_OIDC_SCOPES_JSON ?= ["openid","profile","email"]

EBO_KEYCLOAK_BASE_URL ?= http://localhost:8082
EBO_KEYCLOAK_REALM ?= ebo
EBO_KEYCLOAK_LOGOUT_REDIRECT ?= http://localhost:8082/

# Backend repo location (used to read the seeded Keycloak realm import for local testing)
EBO_BACKEND_DIR ?= ../ebo-planner-backend
EBO_KEYCLOAK_REALM_IMPORT ?= $(EBO_BACKEND_DIR)/deploy/keycloak/import/ebo-realm.json

# Default profile used by local-dev helpers that call the API.
EBO_PROFILE ?= lois
EBO_TRIP_NAME ?= "Test Trip (make)"

help:
	@echo "CI / verification:"
	@echo "  make ci"
	@echo ""
	@echo "Codegen:"
	@echo "  make gen"
	@echo ""
	@echo "Local dev (CLI + Keycloak):"
	@echo "  make ebo                Build ./bin/ebo"
	@echo "  make profiles-keycloak  Create/update CLI profiles for local Keycloak test users"
	@echo "  make keycloak-users     Print local Keycloak test user credentials (reads backend realm import)"
	@echo "  make keycloak-logout    Open/print Keycloak logout URL (clear browser session)"
	@echo "  make keycloak-logout-url Print the logout URL only"
	@echo "  make trip-test          Create a draft trip and patch it with full details (requires auth + provisioned member)"
	@echo ""
	@echo "Changelog / releasing:"
	@echo "  make changelog-verify"
	@echo "  make changelog-release VERSION=0.8.0"
	@echo "Then commit, tag v0.8.0, and push the tag."

# Canonical "green" definition for this repo.
# Today, this repo is mostly release plumbing; when Go code lands, this will
# automatically start enforcing Go fmt/vet/test/build as well.
ci: changelog-verify
	@if [ -f go.mod ]; then \
		echo "Go checks (fmt-check + vet + test + build)..."; \
		unformatted="$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*' 2>/dev/null || true))"; \
		if [ -n "$$unformatted" ]; then \
			echo "ERROR: gofmt needed on:" >&2; \
			echo "$$unformatted" >&2; \
			echo "Fix: run 'gofmt -w' (or add a 'make fmt' target)." >&2; \
			exit 1; \
		fi; \
		go vet ./...; \
		go list ./... | grep -v '/e2e$$' | xargs go test -count=1; \
		covfile="$$(mktemp)"; \
		pkgs="$$(go list ./internal/... | grep -v '/internal/gen')"; \
		go test $$pkgs -count=1 -coverprofile="$$covfile" >/dev/null; \
		total="$$(go tool cover -func="$$covfile" | awk '/^total:/{gsub(/%/,"",$$3); print $$3}')"; \
		rm -f "$$covfile"; \
		echo "Internal coverage: $$total%"; \
		awk -v t="$$total" 'BEGIN{exit !(t+0 >= 85)}' \
			|| { echo "ERROR: internal (non-generated) coverage must be >= 85%" >&2; exit 1; }; \
		go build ./...; \
	else \
		echo "NOTE: No go.mod; skipping Go checks."; \
	fi

changelog-verify:
	@./scripts/verify_changelog.sh
	@./scripts/verify_spec_lock.sh

changelog-release:
	@if [ -z "$(VERSION)" ]; then \
		echo "ERROR: VERSION is required. Example: make changelog-release VERSION=0.8.0" >&2; \
		exit 2; \
	fi
	@./scripts/release_changelog.sh "$(VERSION)"

release-help:
	@$(MAKE) help




gen:
	@spec_path="$$(go run ./tools/specpin)"; \
		echo "Generating OpenAPI client from $$spec_path..."; \
		go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.1 -response-type-suffix ClientResponse -config oapi-codegen.yaml "$$spec_path"

e2e:
	@go test ./e2e -count=1

# Build the CLI binary if missing/stale.
$(EBO_BIN): go.mod
	@mkdir -p "$$(dirname "$(EBO_BIN)")"
	@go build -o "$(EBO_BIN)" ./cmd/ebo

.PHONY: ebo
ebo: $(EBO_BIN)
	@echo "OK: built $(EBO_BIN)"

.PHONY: profiles-keycloak
profiles-keycloak: $(EBO_BIN)
	@set -euo pipefail; \
	echo "Configuring profiles for local Keycloak users..."; \
	echo "  apiUrl:      $(EBO_API_URL)"; \
	echo "  issuerUrl:   $(EBO_OIDC_ISSUER_URL)"; \
	echo "  clientId:    $(EBO_OIDC_CLIENT_ID)"; \
	echo "  scopes:      $(EBO_OIDC_SCOPES_JSON)"; \
	echo ""; \
	for p in alice lois aaron glen denise; do \
		"$(EBO_BIN)" profile create "$$p" --api-url "$(EBO_API_URL)" >/dev/null 2>&1 || true; \
		"$(EBO_BIN)" profile set "$$p" --api-url "$(EBO_API_URL)"; \
		"$(EBO_BIN)" config set "profiles.$$p.oidc.issuerUrl" "$(EBO_OIDC_ISSUER_URL)"; \
		"$(EBO_BIN)" config set "profiles.$$p.oidc.clientId" "$(EBO_OIDC_CLIENT_ID)"; \
		"$(EBO_BIN)" config set "profiles.$$p.oidc.scopes" '$(EBO_OIDC_SCOPES_JSON)'; \
	done; \
	echo ""; \
	echo "OK: profiles configured. Next steps:"; \
	echo "  $(EBO_BIN) --profile lois auth login"; \
	echo "  $(EBO_BIN) --profile lois member create --display-name \"Lois Caldwell\" --email lois.caldwell@example.com"

.PHONY: keycloak-logout-url
keycloak-logout-url:
	@python3 -c 'import os, urllib.parse; base=os.environ.get("EBO_KEYCLOAK_BASE_URL","http://localhost:8082"); realm=os.environ.get("EBO_KEYCLOAK_REALM","ebo"); client=os.environ.get("EBO_OIDC_CLIENT_ID","ebo-client"); redir=os.environ.get("EBO_KEYCLOAK_LOGOUT_REDIRECT","http://localhost:8082/"); print(f"{base}/realms/{realm}/protocol/openid-connect/logout?client_id={urllib.parse.quote(client)}&post_logout_redirect_uri={urllib.parse.quote(redir, safe="")}")'

.PHONY: keycloak-logout
keycloak-logout:
	@set -euo pipefail; \
	url="$$(EBO_KEYCLOAK_BASE_URL="$(EBO_KEYCLOAK_BASE_URL)" EBO_KEYCLOAK_REALM="$(EBO_KEYCLOAK_REALM)" EBO_OIDC_CLIENT_ID="$(EBO_OIDC_CLIENT_ID)" EBO_KEYCLOAK_LOGOUT_REDIRECT="$(EBO_KEYCLOAK_LOGOUT_REDIRECT)" $(MAKE) -s keycloak-logout-url)"; \
	echo "$$url"; \
	if command -v open >/dev/null 2>&1; then open "$$url" >/dev/null 2>&1 || true; \
	elif command -v xdg-open >/dev/null 2>&1; then xdg-open "$$url" >/dev/null 2>&1 || true; \
	fi

.PHONY: keycloak-users
keycloak-users:
	@set -euo pipefail; \
	f="$(EBO_KEYCLOAK_REALM_IMPORT)"; \
	if [[ ! -f "$$f" ]]; then \
		echo "ERROR: Keycloak realm import not found: $$f" >&2; \
		echo "Set EBO_BACKEND_DIR (default: ../ebo-planner-backend) or EBO_KEYCLOAK_REALM_IMPORT to point at ebo-realm.json" >&2; \
		exit 1; \
	fi; \
	if ! command -v jq >/dev/null 2>&1; then \
		echo "ERROR: jq is required for this target (brew install jq)" >&2; \
		exit 1; \
	fi; \
	out="$$(mktemp)"; \
	{ \
		echo -e "NAME\tUSERNAME\tPASSWORD\tEMAIL"; \
		jq -r '.users[] | [ \
			(((.firstName // "") + " " + (.lastName // "")) | gsub("^ +| +$$"; "")), \
			(.username // ""), \
			((.credentials // [] | map(select(.type == "password")) | .[0].value) // ""), \
			(.email // "") \
		] | @tsv' "$$f"; \
	} > "$$out"; \
	if command -v column >/dev/null 2>&1; then \
		column -t -s $$'\t' "$$out"; \
	else \
		cat "$$out"; \
	fi; \
	rm -f "$$out"

.PHONY: trip-test
trip-test: $(EBO_BIN)
	@set -euo pipefail; \
	if ! command -v jq >/dev/null 2>&1; then \
		echo "ERROR: jq is required for this target (brew install jq)" >&2; \
		exit 1; \
	fi; \
	echo "Creating a draft trip using profile: $(EBO_PROFILE)"; \
	# Ensure token + member are configured (best-effort checks; prints guidance on failure).
	if ! "$(EBO_BIN)" --profile "$(EBO_PROFILE)" auth status >/dev/null 2>&1; then \
		echo "ERROR: no valid token for profile '$(EBO_PROFILE)'. Run: $(EBO_BIN) --profile $(EBO_PROFILE) auth login" >&2; \
		exit 3; \
	fi; \
	if ! "$(EBO_BIN)" --profile "$(EBO_PROFILE)" member me >/dev/null 2>&1; then \
		echo "ERROR: member not provisioned for profile '$(EBO_PROFILE)'." >&2; \
		echo "Try: $(EBO_BIN) --profile $(EBO_PROFILE) member create --display-name \"<name>\" --email <email>" >&2; \
		exit 4; \
	fi; \
	create_json="$$(mktemp)"; \
	patch_json="$$(mktemp)"; \
	trap 'rm -f "$$create_json" "$$patch_json"' EXIT; \
	# Create request (JSON).
	printf '{\"name\": %s}\n' "$$(jq -Rn --arg v $(EBO_TRIP_NAME) '$$v')" > "$$create_json"; \
	out="$$( "$(EBO_BIN)" --profile "$(EBO_PROFILE)" --output json trip create --from-file "$$create_json" )"; \
	trip_id="$$(echo "$$out" | jq -r '.data.trip.tripId // empty')"; \
	if [[ -z "$$trip_id" || "$$trip_id" == "null" ]]; then \
		echo "ERROR: could not parse tripId from create output:" >&2; \
		echo "$$out" >&2; \
		exit 1; \
	fi; \
	echo "Created tripId=$$trip_id"; \
	# Patch request (JSON) - includes fields required for publish.
	cat > "$$patch_json" <<'JSON' \
{ \
  "description": "Created by make trip-test", \
  "startDate": "2026-01-10", \
  "endDate": "2026-01-10", \
  "capacityRigs": 5, \
  "difficultyText": "Moderate", \
  "commsRequirementsText": "GMRS recommended; HAM optional.", \
  "recommendedRequirementsText": "Full fuel, recovery points, basic recovery gear.", \
  "meetingLocation": { \
    "label": "Meetup spot", \
    "address": "123 Main St, Pleasanton, CA" \
  } \
} \
JSON
	"$(EBO_BIN)" --profile "$(EBO_PROFILE)" trip update "$$trip_id" --from-file "$$patch_json" >/dev/null; \
	echo "Patched trip $$trip_id"; \
	echo ""; \
	echo "Next steps:"; \
	echo "  $(EBO_BIN) --profile $(EBO_PROFILE) trip get $$trip_id"; \
	echo "  $(EBO_BIN) --profile $(EBO_PROFILE) trip visibility $$trip_id --public"; \
	echo "  $(EBO_BIN) --profile $(EBO_PROFILE) trip publish $$trip_id --print-announcement"
