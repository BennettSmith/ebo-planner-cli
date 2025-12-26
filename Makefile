.PHONY: help ci changelog-verify changelog-release release-help

help:
	@echo "CI / verification:"
	@echo "  make ci"
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
		go test ./...; \
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


