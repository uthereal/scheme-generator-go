# scheme-generator-go

A type-safe Go code generator for PostgreSQL. Parses DDL migration files and
emits strongly-typed models, query builders, hydrators, and relationship
bindings compatible with the
[scheme-runtime-go](https://github.com/uthereal/scheme-runtime-go) ORM runtime.

## Features

- **DDL-driven:** Reads standard PostgreSQL `CREATE`, `ALTER`, `DROP`, and
  `RENAME` statements — no custom schema DSL required.
- **Full migration replay:** Processes migrations in order, simulating the
  schema lifecycle to produce the final state.
- **Relation detection:** Automatically infers `BelongsTo`, `HasOne`,
  `HasMany`, and `BelongsToMany` relationships from foreign keys and
  pivot tables.
- **Cross-schema support:** Handles multi-schema databases with correct
  foreign key resolution across schema boundaries.
- **Composite keys:** Supports composite primary keys and composite foreign
  keys throughout the generation pipeline.
- **Atomic output:** Writes to a staging directory first, then copies to the
  target — broken DDL never leaves partial output.
- **Enum support:** Generates Go enum types from PostgreSQL `CREATE TYPE ... AS ENUM`.

## Generated Files

| File             | Contents                                          |
| ---------------- | ------------------------------------------------- |
| `models.go`      | Struct definitions for each table                 |
| `mutator.go`     | Mutation payload structs for insert/update         |
| `schema.go`      | Typed schema namespace with column references      |
| `table.go`       | Table metadata and column definitions              |
| `hydrate.go`     | Row scanner / model hydration methods              |
| `dehydrate.go`   | Mutator-to-SQL column/value extractors             |
| `relations.go`   | Runtime relationship bindings                      |
| `query.go`       | Query builder factory functions                    |
| `enums.go`       | Go enum types and constants (when enums exist)     |

## Installation

```bash
go install github.com/uthereal/scheme-generator-go/cmd@latest
```

## Usage

```bash
scheme-generator-go \
  -dir ./migrations \
  -package mydb \
  -output ./internal/mydb
```

| Flag       | Description                            |
| ---------- | -------------------------------------- |
| `-dir`     | Input directory containing SQL files   |
| `-package` | Go package name for generated output   |
| `-output`  | Output directory for generated files   |

SQL files are read in alphabetical order, allowing numbered migration files
(e.g., `01_init.sql`, `02_add_posts.sql`) to be processed sequentially.

## Example

The [`example/`](example/) directory contains a working integration test suite
with migrations that exercise schemas, renames, cross-schema references,
composite keys, and M:N pivot tables.

```bash
# Run the integration tests (requires Docker for testcontainers)
go test ./example/... -v
```

## Architecture

```
SQL DDL Files
     │  internal/fs.ReadSQLFiles
     ▼
PostgreSQL Parser (wasilibs/go-pgquery)
     │  AST Nodes
     ▼
PostgresAccumulator (internal/parser)
     │  Simulates schema evolution & detects relations
     ▼
Emitter (internal/emitter)
     │  Embedded Go templates + go/format
     ▼
Atomic FS Write → Output Directory
```

## Requirements

- Go 1.27+
- [scheme-runtime-go](https://github.com/uthereal/scheme-runtime-go) (imported
  by generated code at runtime)

## License

[MIT](LICENSE)
