# AGENTS.md

## Project
- Repository: `go-env`
- Module path: `github.com/rannday/go-env`
- Package name: `goenv`

## Purpose
This library loads environment configuration into Go structs. It is intended to be a small, dependency-free utility that can be reused across other projects.

## Conventions
- Keep the public API small and explicit.
- Preserve precedence:
  1. Process environment
  2. Optional `.env` fallback
  3. Struct tag `default`
- Prefer `allow_empty:"true"` over the legacy `allowempty:"true"` spelling.
- Do not mutate the process environment during load.
- Treat malformed `.env` lines as errors.
- Prefer standard library types and interfaces before introducing new abstractions.

## Supported values
- Scalars: strings, booleans, signed integers, unsigned integers, floats, `time.Duration`, `time.Time`
- Slices: comma-separated values with surrounding whitespace trimmed
- `[]byte`
- Types that implement `encoding.TextUnmarshaler`

## Workflow
- Run `gofmt` on edited Go files.
- Run `go test ./...` before finishing.
- If you change the loader behavior, add or update tests that cover the edge case.

## Cautions
- Do not rename tags or silently change precedence without updating the README and tests.
- Avoid broadening the parser in a way that makes invalid configuration harder to detect.
