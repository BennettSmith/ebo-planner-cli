# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

This repository contains a command-line interface for interacting with the Overland Trip Planning service.
Behavioral and API changes are defined in the spec repository; this changelog focuses on user-facing CLI behavior.

Each CLI release must declare which spec version it targets (see spec.lock).

Notes:

- Spec changelog = contract and behavior
- CLI changelog = commands, flags, UX, output, scripting impact

## [Unreleased]

### Added
- Added a stable JSON output envelope and standardized exit-code mapping for errors.
- Initialized the Go module and added a minimal `ebo` root command with global flags and environment variable equivalents.

### Changed
- CI now runs `go test` with `-count=1` to disable test result caching.

### Deprecated

### Removed

### Fixed

### Security
