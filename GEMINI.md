# GEMINI

## Introduction

This document outlines the architectural decisions, design patterns, and
rigorous formatting standards established for this code generator project.

## Source Code Representation

### Ordering

- **Declarations:** The strict ordering of declarations within a Go file MUST be:
  1. `package`
  2. `import`
  3. `interface` definitions
  4. `type` definitions (structs, aliases)
  5. `const` declarations
  6. `var` declarations
  7. Methods and functions

### Formatting

- **Declaration Grouping:** Do NOT use block wrappers for declarations
  (e.g., `const (...)`, `var (...)`). Declare each constant or variable
  individually on its own line using the explicit `const` or `var` keyword.
- **80-Character Limit:** Every `.go` file is strictly formatted to a
  maximum of **80 characters per line**. This rule applies to
  logic, long `fmt.Errorf` chains, and anonymous lambda functions.
  - **String Literals:** Long string literals MUST NOT be concatenated
    with the `+` operator across multiple lines to circumvent the
    80-character limit. Long strings should remain on a single line
    even if they violate the 80-character limit, preserving their raw
    literal formatting.
- **Goimports:** **DO NOT** run `goimports` on files. The formatting
  engine in `goimports` does not strictly enforce the 80-character
  line limits or other highly specific stylistic preferences of this
  project. Handle formatting and imports manually to preserve intent.

## Lexical Elements

### Identifiers

- **Import Aliases:** Avoid aliasing imports unless absolutely necessary to
  resolve package name collisions. When a collision requires aliasing, use
  a descriptive naming convention that clearly defines boundaries
  (e.g., `internalfs` for the internal `fs` package).
- **Variables/Functions:** Use `camelCase` for internal (unexported) and
  `PascalCase` for exported members.
- **Map Variable Naming (Key-Value Self-Documentation):** Map variables MUST be
  named explicitly using the format `map{KeyName}To{ValueName}` to denote the
  concepts represented by their key and value types. For example, a map of
  `map[string]*ast.Schema` mapping schema names to schema objects must be
  named `mapSchemaNameToSchema`. This eliminates implicit usage concerns and
  keeps mapping types self-evident.
- **Parameters:** Always provide an explicit type for every function parameter
  (e.g., use `(min int, max int)` instead of `(min, max int)`).
- **Callee Validation:** Do NOT perform `nil` checks on pointer parameters
  (e.g., `if p == nil { return ... }`). The responsibility for ensuring
  non-nil values lies with the caller. This reduces redundant checks and
  encourages clearer API contracts.
- **Receivers:** Use short, descriptive single-letter or two-letter
  abbreviations (e.g., `e *Emitter`, `p *PostgresAccumulator`).

## Types

### Struct Types

- **Generics:** Use generic types where they improve type safety and reduce
  boilerplate (e.g., `Pipe[T any]` for fluent transformations).

## Declarations and Scope

### Project Structure

- **Packages:** Use singular package names (`ast`, `emitter`, `parser`,
  `generator`, `pipe`, `fs`).
- **Internal:** Keep core logic in `internal/` to prevent external packages
  from importing private implementation details.
- **CMD:** Entry points for CLI tools reside in `cmd/`.
- **Example:** Integration tests and migration fixtures reside in `example/`.

### Package Responsibilities

- **`internal/ast`:** Intermediate representation types (`Schema`, `Table`,
  `Column`, `ForeignKey`, `Relation*`, `Enum`, `UniqueConstraint`).
- **`internal/parser`:** PostgreSQL DDL parsing via `wasilibs/go-pgquery`.
  Simulates schema evolution (CREATE, ALTER, DROP, RENAME) and detects
  relationships (BelongsTo, HasOne, HasMany, BelongsToMany).
- **`internal/emitter`:** Template-based Go code generation using `embed.FS`
  and `text/template`. Formats output with `go/format`.
- **`internal/fs`:** Filesystem utilities for reading SQL files and atomic
  directory copying.
- **`internal/pipe`:** Generic fluent transformation pipeline.
- **`internal/generator`:** Top-level orchestrator wiring parser → emitter → fs.

### Templates

- Templates live in `internal/emitter/template/*.go.tmpl` and are embedded
  via `//go:embed`.
- Generated code imports `github.com/uthereal/scheme-runtime-go` for runtime
  ORM contracts (`contract`, `orm`, `column`, `relation`, `grammar`).
- Generated output MUST respect the 80-character line limit (verified by
  `generator_test.go`).

### Documentation

- **Godocs:** All functions, methods, exported types, and package-private
  variables (e.g., `var myInternalMap = ...`) MUST have Godoc comments. Godoc
  comments should consist of complete sentences starting with the element name.

## Statements

### Assignment Statements

- **Variable Assignments:** No inline `if` variables. Inline variable
  assignments inside `if` blocks (e.g., `if err := action(); err != nil {`)
  are explicitly banned. Variables must be hoisted above the `if` block for
  enhanced readability and debugging.

### If Statements

- **Control Flow & Guard Clauses:** Avoid deeply nested conditional logic.
  Prefer using `if condition { return/continue/break }` at the beginning of
  functions or loops to validate state or handle errors immediately. This
  flattens the logical flow and drastically improves code readability.

### Defer Statements

- **Cleanup:** Explicitly assign to the blank identifier (e.g.,
  `_ = os.RemoveAll(tempDir)`) to signify the error is intentionally unhandled.

## Errors and Logging

- **Explicit Checks:** Always check errors immediately after they occur. All
  map accesses that return an `ok` boolean must be checked and handled.
- **Error Wrapping:** Native Go errors must NOT be complete sentences. They must
  start with a lowercase letter, must not contain ending punctuation, and must
  use the `->` delimiter for wrapping (e.g.,
  `fmt.Errorf("failed to parse DDL -> %w", err)`). Do not wrap with generic
  context strings. Use `errors.New("...")` for static error strings instead of
  `fmt.Errorf("...")` without formatting verbs.
- **Logging:** Use `slog` for structured logging. Logger messages **must be
  complete sentences**. They must begin with a capital letter and end with a
  period `.` (e.g., `logger.ErrorContext(ctx, "Failed to generate code.")`).
- **Log OR Return:** Never both log and return the same error.
- **Structured Attributes:** All contextual logging variables must be provided
  using strongly typed `slog` attributes (e.g., `slog.String("package", pkg)`,
  `slog.Any("error", err)`) rather than loose key-value pairs. When an error
  is being logged, `slog.Any("error", err)` MUST be the first argument
  provided after the message.

## Testing Standards

- **Table-Driven & Closures:** Use table-driven tests (`[]struct{...}`) for all
  unit and logic testing. Where testing independent state mutations makes
  table-driven arrays unwieldy, use independent
  `t.Run("name", func(t *testing.T) { ... })` closures.
- **Assertions:** Use `github.com/stretchr/testify/assert` and
  `github.com/stretchr/testify/require` for all test assertions. Prefer
  `require` for setup steps and critical failures, and `assert` for
  individual property checks.
- **Testing Types:**
  - **White Box (Internal):** Use `package mypkg` in `{{file}}_test.go` to test
    unexported logic and internal helpers.
  - **Black Box (External):** Use `package mypkg_test` in `{{file}}_test.go` to
    test the public API as a consumer, preventing import cycles.
- **Organization:** Use `main_test.go` (in the same package) to contain shared
  test mocks, helper functions, and setup/teardown logic (e.g.,
  `TestMain(m *testing.M)`).
- **Integration:** Use `testcontainers-go` for database integration tests.
  Integration tests MUST use the isolated template database strategy (e.g.,
  `db, cleanup, err := pgContainer.CreateDB(ctx, t)` in each test) to ensure
  complete test isolation.
- **Generated Output Verification:** Generator integration tests MUST verify
  the 80-character line width constraint on all generated files, excluding
  long string literal lines (e.g., `sqlStr :=`).

## Appendix A: Agent Operations

- **Committing:** Never commit your changes. Leave changes uncommitted for
  manual review.
- **Generated Files:** Never output generated files for test or debug purposes
  into the main source tree. All test generation commands MUST output to the
  `dist/` folder to keep the workspace clean.
