.PHONY: changelog-verify changelog-release release-help

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
	@echo "CLI releases:"
	@echo "  1) Update spec.lock to the spec tag targeted (e.g. v1.2.3)"
	@echo "  2) Add Unreleased notes in CHANGELOG.md (commands/flags/output/UX)"
	@echo "  3) make changelog-release VERSION=0.8.0"
	@echo "  4) Commit CHANGELOG.md (+ spec.lock if changed), tag v0.8.0, push tag"


