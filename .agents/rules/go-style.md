# Rule: Go Style Guide & Engineering Standards

Apply these conventions when writing, refactoring, or reviewing Go code,
tests, and protocol buffers in this repository.

---

## Source Code Representation

### Ordering

- **Declarations:** Strict declaration ordering within a Go file MUST be:
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
  individually on its own line using an explicit `const` or `var` keyword.
- **80-Character Limit:** Every `.go` file is strictly formatted to a
  maximum of **80 characters per line**. This applies to logic, long
  `fmt.Errorf` chains, and anonymous closures.
    - **String Literals:** Long string literals MUST NOT be concatenated
      with `+` across multiple lines. Keep strings on a single line even if
      they exceed 80 characters to preserve literal formatting.
- **Goimports:** **DO NOT** run `goimports`. The `goimports` engine does
  not enforce the 80-character limit or project formatting styles.
  Manage imports and formatting manually.

---

## Lexical Elements

### Identifiers

- **Import Aliases:** Avoid aliasing imports unless resolving package
  name collisions or standard `go.chromium.org` AIP packages. When collisions
  require aliasing, use descriptive naming that clearly defines boundaries
  (e.g., `internalfs` for the internal `fs` package, `internalaip160`).
    - **Generated Protos:** NEVER alias generated protobuf packages.
      Use default generated package names directly.
- **Variables / Functions:** Use `camelCase` for unexported and
  `PascalCase` for exported members.
- **Map Variable Naming (Key-Value Self-Documentation):** Map variables MUST be
  named explicitly using the format `map{KeyName}To{ValueName}` to denote the
  concepts represented by their key and value types (e.g.,
  `mapSchemaNameToSchema`, `mapTableNameToTable`, `mapUserIDToAccountID`).
- **Parameters:** Provide explicit types for every parameter
  (e.g., `(min int, max int)`, never `(min, max int)`).
- **Callee Validation:** Do NOT perform `nil` checks on pointer
  parameters (e.g., `if p == nil { return ... }`). Non-nil safety is the
  caller's responsibility.
- **Receivers:** Use short single-letter or two-letter abbreviations
  (e.g., `e *Emitter`, `p *PostgresAccumulator`, `g *Generator`).
- **Timestamps:** Use the `_at` suffix for timestamp fields and columns
  (e.g., `operations_ready_at`, `created_at`).
- **Contexts:** When deriving a context, prefix with `ctx` followed by a
  descriptive descriptor (e.g., `ctxLogger`, `ctxTarget`, `ctxInsert`).

---

## Types

### Struct Types (Models & DTOs)

- **Custom Unmarshaling:** Use the `type Alias T` pattern inside
  `UnmarshalJSON` to prevent infinite recursion during post-processing.

### Generics

- **Generics:** Use generic types where they improve type safety and reduce
  boilerplate (e.g., `Pipe[T any]` for fluent transformations).

---

## Declarations and Scope

### Documentation

- **Godocs:** All functions, methods, exported types, and package-private
  variables (e.g., `var myInternalMap = ...`) MUST have Godoc comments
  consisting of complete sentences starting with the element name.

---

## Statements

### Assignment Statements

- **Variable Assignments:** No inline `if` variables. Inline variable
  assignments inside `if` statements (`if err := fn(); err != nil`) are
  banned. Hoist variables above the `if` block for enhanced readability
  and debugging.

### If Statements

- **Control Flow & Guard Clauses:** Avoid deeply nested conditional
  logic. Use `if condition { return/continue/break }` guard clauses at
  the start of functions/loops to flatten logical flow.

### Defer Statements

- **Cleanup:** Use `WarnContext` if a logger is available to log resource
  closing errors. Otherwise, explicitly assign to the blank identifier
  (e.g., `_ = os.RemoveAll(tempDir)`).

---

## Errors and Logging

- **Explicit Checks:** Always check errors immediately after calls. All
  map lookups returning `ok` booleans must be checked.
- **Error Wrapping:** Native Go errors must NOT be complete sentences.
  Start with a lowercase letter, omit ending punctuation, and use `->` as
  the wrapping delimiter:
  `fmt.Errorf("failed to parse DDL -> %w", err)`
  Use `errors.New("...")` for static error strings instead of
  `fmt.Errorf("...")` without formatting verbs.
- **Logging:** Use `slog` for structured logging (`ErrorContext`,
  `WarnContext`). Log messages MUST be complete sentences starting with a
  capital letter and ending with a period (e.g.,
  `logger.Error("Failed to generate code.")`).
- **Log OR Return:** Never both log and return the same error.
    - **Exceptions:**
        1. Critical local state snapshots too large/sensitive for errors.
        2. Asynchronous error handoffs to channels or goroutines.
- **Error Handling Guide:**
  | Action | When to use it |
  | --- | --- |
  | Return `err` | Default behavior. Let caller decide. |
  | Wrap & Return | Adding context (e.g., "processing record X"). |
  | Log & Stop | Top level of application. |
  | Log & Continue | Non-critical errors (e.g., cache miss). |
- **Structured Attributes:** Pass contextual variables using strongly
  typed `slog` attributes (e.g., `slog.String("package", pkg)`). When an
  error is logged, `slog.Any("error", err)` MUST be the first attribute.

---

## Concurrency and Context Management

- **Parallelism:** Use `golang.org/x/sync/errgroup` for concurrent tasks
  that return errors.
- **Context Propagation:** Always propagate `context.Context`. Use
  `context.WithoutCancel` for cleanup/logging that must finish post-cancel.
- **Strict Timeouts:** All database queries, external API calls, and
  blocking I/O operations MUST use `context.WithTimeout`. Never reuse raw
  parent/global request contexts directly for execution calls.

---

## System Considerations (Database & API)

### API Design (AIP Compliance)

- **Filtering (AIP-160):** Use `internal/aip160` (or project AIP parser).
  Configure via `filter_fields` in validation context.
- **Ordering (AIP-132):** Use `internal/aip132` (or project AIP parser).
  Configure via `order_fields` in validation context.
- **Pagination (AIP-158):** Use standard `PageSize` and `PageToken`
  fields. Configure via `request` in validation context.
- **Query Builder:** Always use the generated query builder, typed
  columns, and methods (`db.Or(...)`, `column.Lt(...)`, `column.IsNull()`)
  rather than raw `db.WhereCondition{Sql: ...}` strings.

---

## Testing Standards

- **Table-Driven & Closures:** Use table-driven tests (`[]struct{...}`)
  or `t.Run("name", func(t *testing.T) { ... })` subtest closures.
- **Assertions:** Use `github.com/stretchr/testify/assert` and `require`.
  Use `require` for setup/fatal checks, `assert` for properties.
- **Testing Packages:**
    - **White Box:** Use `package mypkg` in `{{file}}_test.go` for internals.
    - **Black Box:** Use `package mypkg_test` to test public surface and
      prevent import cycles.
- **Organization:** Use `main_test.go` for shared mocks and `TestMain`.
- **Integration:** Use `testcontainers-go` for database testing. Use
  isolated template databases (`db, cleanup, err := pg.CreateDB(ctx, t)`)
  for total test isolation rather than table truncation.
- **Timestamp Fidelity:** Use `.Truncate(time.Microsecond)` on timestamps
  to ensure stable equality comparisons across DB/JSON round-trips.
- **Generated Output Verification:** Generator integration tests MUST verify
  the 80-character line width constraint on all generated files, excluding
  long string literal lines (e.g., `sqlStr :=`).

---

## Agent Operations

- **Committing:** Never commit your changes. Leave changes uncommitted for
  manual review.
- **Generated Files:** Never output generated files for test or debug purposes
  into the main source tree. All test generation commands MUST output to the
  `dist/` folder or temporary directories to keep the workspace clean.
