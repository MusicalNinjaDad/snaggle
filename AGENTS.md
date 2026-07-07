# AGENTS.md

This repository is a small Go CLI and library for copying ELF binaries and their runtime dependencies into a minimal filesystem layout. Keep changes narrow, explicit, and easy to reason about.

## Primary goals

- Keep the tool correct for the supported Linux/ELF use case.
- Prefer explicit, conservative behavior over clever heuristics.
- Preserve a small, composable codebase with clear package boundaries.

## Code style and documentation: optimize for 3 audiences

- End users: behavior should be predictable, easy to use, and surfaced with clear errors. User-facing code is documented for end-users.
- Developers reading and maintaining the code: code should be easy to follow, with obvious structure and names. All non-user-facing code is documented for developers and designed to provide key insights in IDE pop-ups and full clarity when reading the code-base.
- Agents updating the codebase: future changes should be able to be clearly scoped, easy to review, and supported by enough context in-code to avoid guesswork.

## Errors

- Keep the error taxonomy explicit and typed. Use sentinel errors for stable categories and typed errors when callers need structured information.
- Wrap lower-level errors with context using `%w`, and preserve the original cause instead of discarding it.
- Use `errors.Is` for semantic checks and expected conditions such as sentinel errors, `errors.ErrUnsupported`, `fs.ErrNotExist`, and syscall-level conditions.
- Use `errors.As` when the caller needs to recover a specific error type or inspect structured data such as `*ErrElf`, `*fs.PathError`, `*SnaggleError`, or `*InvocationError`.
- Do not diagnose behavior by matching error strings; use the error chain and the typed error taxonomy.

## Documentation and APIs

- Document exported functions, types, and public behavior when they form part of the API.
- Prefer code that is self-explanatory over comments. Rename variables, split helpers, and make control flow obvious instead of adding a comment to explain the obvious.
- Use comments only for the why, non-obvious constraints, or behavior that cannot be made obvious from the code.
- Keep APIs clear and narrow so that the primary behavior can be understood without reading large amounts of documentation.

## Package boundaries and principles

- `internal/`: shared helpers that are not part of the public API. Use it for code reused across packages when it should stay implementation detail.
- `elf/`: ELF parsing, interpreter and dependency discovery, and ELF-specific errors. Keep this package focused on ELF semantics and isolated from CLI concerns.
- `snaggle.go` vs `cmd/snaggle/main.go`:
  - `snaggle.go` holds public orchestration and the main behavior of the library/CLI flow.
  - `cmd/snaggle/main.go` handles CLI entrypoints, argument parsing, exit codes, and user-facing error formatting.

## Functions and structure

- Keep functions focused and easy to read. A short function with a single responsibility is preferred where complex logic is involved.
- A longer function is preferred for orchestration paths to provide obvious structure, reading as a sequence of named steps.
- If a function is doing multiple complex tasks, split it into orchestration & helpers with clear names rather than a dense block of branching logic.

## Change process

- If a requirement is unclear, ask. Prefer a short clarification with options or ideas rather than guessing; always ask when uncertain.
- Update tests first. Make the tests describe the intended behavior, and confirm that they fail for the expected reason before implementing the change.
- Do not mix code and test updates in the same commit. Keep them separate.
- Changes to tests should directly reflect the intended behavior change.
- Prefer the most suitable level of test. API-boundary tests are often the clearest and most robust choice.
- Update documentation when behavior, CLI usage, or public semantics change.
- Ensure the tests, code, and docs are aligned with the original request.

## Repository map

- `cmd/snaggle/`: CLI entrypoint, help text, and exit-code handling.
- Root package: public orchestration and high-level behavior.
- `elf/`: ELF parsing and ELF-specific errors.
- `internal/`: shared helpers and path/file-system logic.
- `*_test.go`: keep tests close to the behavior they verify.
