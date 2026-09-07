# Rule: Scheme Generator Go - Architecture & System Design

The Scheme Go schema generator (`scheme-generator-go`) parses PostgreSQL DDL migrations,
simulates schema evolution, resolves relational dependencies, and emits type-safe Go
code targeting PostgreSQL backed by `scheme-runtime-go`.

---

## Architectural Layering & Pipeline

The codebase is organized as a unidirectional data transformation pipeline:

```
SQL Migrations / DDL
         │
         ▼
internal/parser/     → PostgreSQL DDL parsing & schema evolution simulation
         │
         ▼
internal/ast/        → In-memory schema representation (AST)
         │
         ▼
internal/pipe/       → Generic fluent transformation pipeline (Pipe[T])
         │
         ▼
internal/emitter/    → Template-based Go code generation (text/template + go/format)
         │
         ▼
internal/fs/         → Atomic filesystem writes & directory management
         │
         ▼
Generated Go Code    → Type-safe models, queries, mutators, tables, schemas, relations
```

---

## Package Responsibilities

### 1. `internal/ast/` (Abstract Syntax Tree / Intermediate Representation)

Maintains the in-memory schema graph, tables, columns, constraints, and relationships:

- **`Schema`**: Top-level container representing a PostgreSQL schema holding tables and enums.
- **`Table`**: Table metadata, primary keys, columns, foreign keys, and unique constraints.
- **`Column`**: Column definitions, data types, default expressions, nullability, and array flags.
- **`ForeignKey`**: Foreign key constraints linking parent and child tables with action behaviors.
- **`Relation*`**: Relationship specifications detected between tables:
  - `RelationBelongsTo`: Many-to-one foreign key relationship.
  - `RelationHasOne`: One-to-one relationship.
  - `RelationHasMany`: One-to-many relationship.
  - `RelationBelongsToMany`: Many-to-many junction-table-backed relationship.
- **`Enum`**: PostgreSQL custom enum types and allowed label values.
- **`UniqueConstraint`**: Single-column or composite unique constraints.

### 2. `internal/parser/` (DDL Parsing & Schema Simulation)

Parses PostgreSQL DDL migrations and simulates schema state over time:

- **DDL Parser**: Uses `wasilibs/go-pgquery` (WebAssembly-compiled PostgreSQL query parser)
  to extract native PostgreSQL ASTs from DDL statements.
- **Schema Evolution**: Sequentially processes migration statements to simulate state changes:
  - `create.go`: Handles `CREATE TABLE`, `CREATE TYPE ... AS ENUM`, etc.
  - `alter.go`: Handles `ALTER TABLE` (add/drop/alter column, add/drop constraint).
  - `rename.go`: Handles `ALTER TABLE ... RENAME TO` and `RENAME COLUMN`.
  - `drop.go`: Handles `DROP TABLE`, `DROP TYPE`, etc.
- **Relation Detection (`relation.go`)**: Inspects foreign keys and junction tables across
  the evaluated schema to infer `BelongsTo`, `HasOne`, `HasMany`, and `BelongsToMany` relations.

### 3. `internal/emitter/` (Go Code Generation)

Compiles the AST into strongly-typed Go source code using embedded templates:

- Uses `text/template` with embedded templates via `//go:embed template/*.go.tmpl`.
- Formats emitted Go source code with `go/format` (`format.Source`).
- Strictly enforces the **80-character line width** constraint across all generated Go files
  (verified by `generator_test.go`).
- Emitted code imports `github.com/uthereal/scheme-runtime-go` runtime contracts
  (`contract`, `orm`, `column`, `relation`, `grammar`).

#### Embedded Templates (`internal/emitter/template/`)

- **`models.go.tmpl`**: Model structs with field tags, types, and column accessors.
- **`query.go.tmpl`**: Type-safe query builder methods and query execution helpers.
- **`mutator.go.tmpl`**: Insert, update, and upsert mutator structs.
- **`table.go.tmpl`**: Table descriptors, metadata, and strongly-typed column wrappers.
- **`schema.go.tmpl`**: Schema struct registering tables and providing builder entry points.
- **`relations.go.tmpl`**: Relationship descriptors for eager loading.
- **`hydrate.go.tmpl`**: Hydration functions mapping database rows to Go model structs.
- **`dehydrate.go.tmpl`**: Dehydration functions mapping model/mutator structs to column values.
- **`enum.go.tmpl`**: Go constants and types for PostgreSQL enums.

### 4. `internal/generator/` (Orchestrator)

Top-level coordinator wiring together parser, emitter, and filesystem:

- Reads SQL migrations from source paths.
- Drives parsing and schema evolution simulation.
- Executes code emission for tables, models, queries, mutators, and schemas.
- Coordinates writing output files to target destination directories.

### 5. `internal/pipe/` (Generic Fluent Pipeline)

- Generic fluent transformation utility (`Pipe[T any]`) for slice mapping, filtering,
  and transformation without imperative loops.

### 6. `internal/fs/` (Filesystem Utilities)

- Reads and sorts SQL migration files in deterministic order.
- Provides atomic directory creation, writing, and copying.

### 7. `internal/inflection/` (Inflection Rules)

- Handles English word pluralization and singularization for table, model,
  and relation naming.

### 8. `cmd/` (CLI Entry Points)

- Command-line interfaces for executing the generator tool.

### 9. `example/` (Reference Fixtures & Integration Tests)

- Migration SQL scripts (`01_init.sql`, `02_refactor_renames.sql`, etc.).
- Generated reference code (`example/generated/`).
- End-to-end integration tests validating generated code against live PostgreSQL databases.

---

## Core Dependencies

| Dependency | Purpose |
| --- | --- |
| `github.com/wasilibs/go-pgquery` | WASM-based PostgreSQL parser extracting native DDL ASTs |
| `github.com/uthereal/scheme-runtime-go` | Runtime ORM library backing generated models and queries |
| `github.com/ettle/strcase` | Case transformations (PascalCase, camelCase, snake_case) |
| `github.com/jinzhu/inflection` | Word pluralization and singularization |
| `github.com/jackc/pgx/v5` | PostgreSQL driver and identifier sanitization |
| `github.com/stretchr/testify` | Assertions and test suites (`assert`, `require`) |
| `github.com/testcontainers/testcontainers-go` | Ephemeral PostgreSQL containers for integration tests |

---

## Generator Engineering Conventions

### Project Structure

- **Singular Package Names**: Use singular names for packages (`ast`, `emitter`, `parser`,
  `generator`, `pipe`, `fs`, `inflection`).
- **Internal Visibility**: Keep core logic in `internal/` to avoid exposing private
  implementation details to external consumers.
- **CLI Entrypoints**: Reside in `cmd/`.
- **Fixtures & Tests**: Integration tests and migration fixtures reside in `example/`.

### Generated Code Standards

- **80-Character Line Limit**: Generated code MUST strictly respect the 80-character line
  limit (excluding long raw string literal lines like SQL statements).
- **Runtime Contract Alignment**: Generated code MUST cleanly target `scheme-runtime-go`
  contracts without circular dependencies.

### Agent Operations

- **Committing**: Never commit changes. Leave changes uncommitted for manual review.
- **Generated Files**: Never output generated files for test or debug purposes into the main
  source tree. All test generation commands MUST output to the `dist/` folder or temporary
  directories to keep the workspace clean.
