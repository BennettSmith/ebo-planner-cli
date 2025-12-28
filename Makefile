.PHONY: help ci gen changelog-verify changelog-release release-help

help:
	@echo "CI / verification:"
	@echo "  make ci"
	@echo ""
	@echo "Codegen:"
	@echo "  make gen"
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
