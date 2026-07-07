# AGENTS.md

This repository is a small Go CLI and library for copying ELF binaries and their runtime dependencies into a minimal filesystem layout. Future changes should preserve that narrow scope, keep behavior predictable, and fit the existing package structure.

## 1. Primary goals

- Keep the tool correct for the supported Linux/ELF use case.
- Prefer explicit, conservative behavior over clever heuristics.
- Preserve a small, composable codebase with clear package boundaries.
- Make changes easy to reason about and easy to test.

## 2. Repository map

Use the existing structure as the default home for new code:

- `cmd/snaggle/`: CLI entrypoint, Cobra wiring, help text, exit-code handling.
- Root package (`snaggle.go`): public orchestration and high-level behavior.
- `elf/`: ELF parsing, interpreter/dependency discovery, and ELF-specific error types.
- `internal/`: shared helpers that are not part of the public API.
- `*_test.go` files: keep tests close to the behavior they verify.

If a change is primarily about CLI behavior, implement it in `cmd/snaggle/` or the root package. If it is primarily about ELF parsing or dependency detection, implement it in `elf/`.

## 3. Engineering principles to follow

### 3.1 Prefer correctness and explicitness over convenience

- Do not guess about unsupported ELF cases; represent them with typed errors.
- If the repository currently only supports a subset of binaries, keep that boundary explicit rather than silently broadening behavior.
- Prefer deterministic behavior and easy-to-follow control flow over magic.

### 3.2 Keep functions small and focused

- Each function should have one clear responsibility.
- Use helper functions for branching, path resolution, error handling, and file-system actions.
- Avoid long nested conditionals when they can be split into named helpers.

### 3.3 Keep package boundaries clean

- Do not let CLI concerns leak into parsing logic.
- Do not let low-level helpers depend on Cobra, logging, or command-line UX.
- Keep public API changes intentional and documented.

### 3.4 Preserve the existing error model

- Use typed errors where the repository already uses them.
- Wrap lower-level errors with context instead of discarding them.
- Prefer `errors.Is` and `errors.As` for error handling.
- Avoid panics for ordinary control flow; use explicit errors and return paths.

### 3.5 Be conservative with dependencies and abstractions

- Prefer the Go standard library and the dependencies already used by the repo.
- Avoid introducing new frameworks or large abstractions unless the change clearly needs them.
- If a feature can be implemented with a small helper, prefer that over a new package.

## 4. Code conventions already present in the repo

### 4.1 Go style

- Write idiomatic Go.
- Use `gofmt` on edited Go files.
- Keep names descriptive and specific.
- Favor small, local variables over overly clever one-liners.

### 4.2 Comments and docs

- Add package comments where the repo already uses them.
- Add comments for exported functions, types, and non-obvious behavior.
- Document behavior that affects users, especially CLI flags and public API semantics.
- Keep comments factual and tied to code behavior.

### 4.3 Public API

- Preserve existing exported names and semantics unless a change explicitly requires otherwise.
- If adding a new option or public entrypoint, follow the existing options pattern and keep the API small.
- Avoid adding hidden behavior that is not discoverable from the code or docs.

### 4.4 File-system and path handling

- Resolve paths carefully and preserve the original source path in errors when relevant.
- Handle symlinks explicitly.
- Prefer path handling helpers in `internal/` or existing path-resolution code rather than ad hoc logic.

### 4.5 Concurrency and performance

- The repo already uses concurrency for independent work where it is safe.
- Keep parallelism bounded and predictable.
- Preserve the distinction between silent/default mode and verbose mode.
- Avoid introducing unnecessary concurrency unless it improves clarity or throughput in a controlled way.

## 5. Decision rules for new work

When implementing a change, use these rules in order:

1. Place the logic in the most specific package that owns it.
2. Keep the change minimal and consistent with the surrounding code.
3. Preserve or improve error clarity.
4. Add or update tests for the new behavior.
5. Update user-facing docs or help text if behavior or CLI usage changes.

## 6. Testing expectations

- Add tests for new behavior rather than relying on manual reasoning alone.
- Prefer targeted tests that exercise the real behavior of the code path.
- Keep tests deterministic and avoid unnecessary mocks.
- If a change affects CLI output, usage text, or public behavior, update the relevant tests.

## 7. Documentation and CLI changes

- If a change alters command-line behavior, update the relevant help text and doc comments.
- If the change affects documented examples or user-visible semantics, update README content where appropriate.
- Keep the generated/help text and source comments aligned.

## 8. Default approach when unsure

When a task is ambiguous, choose the smallest change that:

- fits the existing package structure,
- preserves current behavior for supported cases,
- leaves the code easier to understand,
- and adds tests for the new requirement.
