# Rule: Google API Improvement Proposals (AIP) Specification

Apply these standards to all Protocol Buffer schema definitions, gRPC services, JSON/REST transcoding, and resource-oriented API lifecycles.

---

## AIP-1: AIP Purpose and Guidelines
- **Core Purpose**: API Improvement Proposals (AIPs) serve as the source of truth for API design consistency and best current practices.
- **Requirement Keywords**: Guidance follows RFC 2119 keyword requirements (`MUST`, `MUST NOT`, `SHOULD`, `SHOULD NOT`, `MAY`).
- **Lifecycle States**:
    - `Draft`: In progress and under initial team discussion.
    - `Reviewing`: Consensus reached; under formal review by designated AIP approvers.
    - `Approved`: Formally accepted as best current practice.
    - `Replaced`: Superseded by a newer AIP document with referenced rationale.

## AIP-2: AIP Numbering
- **Number Allocation**: AIPs use numeric identifiers reflecting their categorical scope.
- **Scope Allocation**:
    - `1–99`: Meta, process, and style.
    - `100–119`: High-level concepts and architecture.
    - `120–129`: Resource design fundamentals.
    - `130–139`: Standard and custom methods.
    - `140–149`: Field naming and types.
    - `150–179`: Operational design patterns.
    - `180–189`: Compatibility, versioning, and stability.
    - `190–199`: Style, structure, and documentation.
    - `200–236`: Advanced patterns, batch methods, and special types.
    - `1000+`: Platform- and language-specific client library guidance.

## AIP-3: AIP Versioning
- **Evolution**: AIP documents evolve in place while preserving backward clarity. Substantial design shifts that break established patterns must document transition paths and changelogs.

## AIP-8: AIP Style and Guidance
- **Document Structure**: Every AIP must cover a single discrete topic and define an imperative title (noun-focused).
- **Protobuf Formatting**: Protobuf files must use 2-space indentation, explicit scalar types, and avoid syntax anti-patterns.
- **Actionable Scope**: Guidance must be unambiguous and directly verifiable through linters or automated schema validation tools.

## AIP-9: Glossary
- **Resource Name**: The unique string identifying a resource (e.g., `publishers/123/books/456`).
- **Resource ID**: The final URI component identifying the specific resource within its parent collection (e.g., `456`).
- **Parent**: The URI path prefix preceding a collection (e.g., `publishers/123`).
- **Collection**: A homogenous set of addressable resources of the same type (e.g., `books`).
- **Standard Method**: One of the five primary CRUD operations (`Get`, `List`, `Create`, `Update`, `Delete`).
- **Custom Method**: An operation that cannot be cleanly modeled as a standard CRUD method.

## AIP-100: API Design Review FAQ
- **Review Readiness**: APIs must be reviewed prior to stabilizing interfaces (`v1beta1`, `v1`).
- **Design Review Scope**: Reviews inspect naming consistency, resource modeling, pagination semantics, backward compatibility risks, and error surface definitions.

## AIP-111: Planes
- **Control Plane**: APIs that manage configuration, provision resources, and govern infrastructure lifecycles. Prioritize declarative reconciliation, granular CRUD, and idempotent mutations.
- **Data Plane**: APIs that handle high-frequency payload ingestion, query execution, and high-throughput streaming. Prioritize low latency, minimal protocol overhead, and batched streaming.

## AIP-121: Resource-Oriented Design
- **Entities over RPCs**: Model APIs around distinct, addressable resources rather than procedural actions.
- **Resource Hierarchy**: Form logical parent-child relationships where appropriate (e.g., `users/{user}/orders/{order}`).
- **Stateless Operations**: Server implementations must remain stateless with respect to individual client connections; state lives exclusively inside persistent resources.

## AIP-122: Resource Names
- **Format**: Hierarchical pattern formatted as `collection/{id}/subCollection/{id}`.
- **Collection IDs**: Must be plural, `lowerCamelCase` in REST paths, and valid protobuf identifiers (e.g., `userProfiles/{user_profile}`).
- **Resource IDs**: Must be non-empty strings and must never contain forward slashes (`/`).
- **Full vs Relative Names**:
    - Full: `//library.googleapis.com/publishers/123/books/456`
    - Relative: `publishers/123/books/456`
- **Protobuf Annotation**:
  ```proto
  message Book {
    option (google.api.resource) = {
      type: "library.googleapis.com/Book"
      pattern: "publishers/{publisher}/books/{book}"
    };
    string name = 1 [(google.api.field_behavior) = IDENTIFIER];
  }
  ```

## AIP-123: Resource Types
- **Format**: Fully qualified string identifier formatted as `<service-name>/<Type>` (e.g., `pubsub.googleapis.com/Topic`).
- **Type Casing**: Must use `PascalCase` singular noun for the resource type suffix.

## AIP-124: Resource Association
- **References**: Reference other resources using string name fields annotated with `(google.api.resource_reference)`:
  ```proto
  string author = 2 [
    (google.api.field_behavior) = REQUIRED,
    (google.api.resource_reference) = { type: "library.googleapis.com/Author" }
  ];
  ```
- **Anti-Pattern**: Do not embed child resource bodies or foreign numeric database primary keys to represent associations.

## AIP-126: Enumerations
- **Zero-Value**: The `0` value must always be named `<ENUM_NAME>_UNSPECIFIED`.
- **Casing**: Enum type names use `PascalCase`; enum values use `UPPER_SNAKE_CASE` prefixed with the enum type name.
  ```proto
  enum BookFormat {
    BOOK_FORMAT_UNSPECIFIED = 0;
    BOOK_FORMAT_HARDCOVER = 1;
    BOOK_FORMAT_PAPERBACK = 2;
    BOOK_FORMAT_EBOOK = 3;
  }
  ```
- **Prohibited**: Do not define `UNKNOWN`, `DEFAULT`, or `NONE` values that collide with `UNSPECIFIED`.

## AIP-127: HTTP and gRPC Transcoding
- **HTTP Bindings**: All gRPC methods must declare `google.api.http` annotations.
- **Standard Mappings**:
    - `Get`: `GET /v1/{name=publishers/*/books/*}`
    - `List`: `GET /v1/{parent=publishers/*}/books`
    - `Create`: `POST /v1/{parent=publishers/*}/books` (`body: "book"`)
    - `Update`: `PATCH /v1/{book.name=publishers/*/books/*}` (`body: "book"`)
    - `Delete`: `DELETE /v1/{name=publishers/*/books/*}`
    - `Custom`: `POST /v1/{name=publishers/*/books/*}:archive` (`body: "*"`)

## AIP-128: Declarative-Friendly Interfaces
- **Tool Compatibility**: APIs must support declarative managers (Terraform, Kubernetes Operators).
- **Separation of Concerns**: Clearly distinguish declared user intent (`spec`) from server-computed status (`status` or `OUTPUT_ONLY` state).
- **Predictable Mutations**: Creation and updates must be idempotent when client-assigned IDs or `request_id`s are supplied.

## AIP-129: Server-Modified Values and Defaults
- **Transparency**: Explicitly document all default values applied when fields are omitted.
- **Output Annotations**: Fields modified asynchronously or normalized by the server must be documented clearly in proto field comments.

## AIP-130: Methods
- **Method Categories**: APIs must prioritize standard methods (`Get`, `List`, `Create`, `Update`, `Delete`) and standard batch methods before considering custom methods.

## AIP-131: Standard Methods: Get
- **Signature**: `rpc Get<Resource>(Get<Resource>Request) returns (<Resource>);`
- **Request Fields**:
    - `string name = 1 [(google.api.field_behavior) = REQUIRED, (google.api.resource_reference) = { ... }];`
    - Optional `View` enum or `google.protobuf.FieldMask read_mask`.
- **Behavior**: Returns `NOT_FOUND` if the resource does not exist or the caller lacks permission.

## AIP-132: Standard Methods: List
- **Signature**: `rpc List<Resources>(List<Resources>Request) returns (List<Resources>Response);`
- **Request Fields**:
    - `string parent = 1 [(google.api.field_behavior) = REQUIRED, (google.api.resource_reference) = { ... }];`
    - `int32 page_size = 2 [(google.api.field_behavior) = OPTIONAL];`
    - `string page_token = 3 [(google.api.field_behavior) = OPTIONAL];`
    - `string filter = 4 [(google.api.field_behavior) = OPTIONAL];`
    - `string order_by = 5 [(google.api.field_behavior) = OPTIONAL];`
- **Response Fields**:
    - `repeated <Resource> <resources> = 1;`
    - `string next_page_token = 2;`
    - `repeated string unreachable = 3 [(google.api.field_behavior) = OPTIONAL];`

## AIP-133: Standard Methods: Create
- **Signature**: `rpc Create<Resource>(Create<Resource>Request) returns (<Resource>);` (or `google.longrunning.Operation`).
- **Request Fields**:
    - `string parent = 1 [(google.api.field_behavior) = REQUIRED, (google.api.resource_reference) = { ... }];`
    - `string <resource>_id = 2 [(google.api.field_behavior) = OPTIONAL];`
    - `<Resource> <resource> = 3 [(google.api.field_behavior) = REQUIRED];`
    - `string request_id = 4 [(google.api.field_behavior) = OPTIONAL];`
    - `bool validate_only = 5 [(google.api.field_behavior) = OPTIONAL];`

## AIP-134: Standard Methods: Update
- **Signature**: `rpc Update<Resource>(Update<Resource>Request) returns (<Resource>);` (or `google.longrunning.Operation`).
- **Request Fields**:
    - `<Resource> <resource> = 1 [(google.api.field_behavior) = REQUIRED];`
    - `google.protobuf.FieldMask update_mask = 2 [(google.api.field_behavior) = OPTIONAL];`
    - `string request_id = 3 [(google.api.field_behavior) = OPTIONAL];`
    - `bool validate_only = 4 [(google.api.field_behavior) = OPTIONAL];`
    - `bool allow_missing = 5 [(google.api.field_behavior) = OPTIONAL];` (if upsert is supported).

## AIP-135: Standard Methods: Delete
- **Signature**: `rpc Delete<Resource>(Delete<Resource>Request) returns (google.protobuf.Empty);` (or `google.longrunning.Operation`).
- **Request Fields**:
    - `string name = 1 [(google.api.field_behavior) = REQUIRED, (google.api.resource_reference) = { ... }];`
    - `bool force = 2 [(google.api.field_behavior) = OPTIONAL];` (cascading delete).
    - `string etag = 3 [(google.api.field_behavior) = OPTIONAL];`
    - `string request_id = 4 [(google.api.field_behavior) = OPTIONAL];`
    - `bool validate_only = 5 [(google.api.field_behavior) = OPTIONAL];`

## AIP-136: Custom Methods
- **Naming Pattern**: `rpc <Verb><Noun>(<Verb><Noun>Request) returns (<Verb><Noun>Response);`
- **HTTP Verb**: Default to `POST`.
- **URI Structure**: Append `:<customVerb>` suffix to the resource or collection path:
  `POST /v1/{name=publishers/*/books/*}:archive`
- **Body**: Use `body: "*"` for custom methods that mutate resources.

## AIP-140: Field Names
- **Casing**: All field names must use `lower_snake_case`.
- **Prohibitions**: Never use uppercase letters, consecutive underscores (`__`), or leading/trailing underscores.
- **Numbers in Names**: Avoid ambiguous numerical suffixes (prefer `address_line1` over `address1`).

## AIP-141: Quantities
- **Unit Suffixes**: Fields representing physical or computational units must declare the unit suffix:
    - Time: `_seconds`, `_millis`, `_nanos`.
    - Data: `_bytes`, `_kibibytes`, `_mebibytes`.
    - Frequency: `_hz`, `_khz`.
- **Integrals vs Decimals**: Use `int64` for whole data counts; avoid floating point for exact quantity tracking.

## AIP-142: Time and Duration
- **Timestamps**: Use `google.protobuf.Timestamp`. Field names must end in `_time` (e.g., `create_time`, `start_time`, `expire_time`).
- **Durations**: Use `google.protobuf.Duration`. Field names must end in `_duration` (e.g., `timeout_duration`) or specify a recognized unit (e.g., `ttl`).
- **Prohibited**: Do not use numeric UNIX epoch integers or raw string representations for dates and times.

## AIP-143: Standardized Codes
- **Currencies**: Use ISO 4217 3-letter currency codes (e.g., `USD`, `EUR`).
- **Languages**: Use BCP 47 language tags (e.g., `en-US`, `es-ES`).
- **Countries/Regions**: Use ISO 3166-1 alpha-2 two-letter country codes (e.g., `US`, `GB`).
- **Time Zones**: Use IANA Time Zone Database names (e.g., `America/Los_Angeles`).

## AIP-144: Repeated Fields
- **Plural Naming**: Repeated fields must use plural nouns (e.g., `repeated string display_names = 1;`).
- **No Arrays of Arrays**: Avoid nested repeated fields; wrap child lists in distinct protobuf messages.

## AIP-145: Ranges
- **Interval Pairing**: Continuous ranges must use explicit start and end pairs:
    - Time: `google.protobuf.Timestamp start_time` / `end_time`.
    - Date: `google.type.Date start_date` / `end_date`.
    - Values: `int32 min_value` / `max_value`.
- **Semantics**: Ranges default to half-open intervals `[start, end)` unless explicitly documented.

## AIP-146: Generic Fields
- **Avoid Untyped Payloads**: Avoid `google.protobuf.Any`, `google.protobuf.Struct`, or JSON-encoded strings in public APIs.
- **Type Safety**: Design explicit protobuf messages to preserve compile-time validation and backward-compatibility tooling.

## AIP-147: Sensitive Fields
- **PII & Secrets**: Fields containing passwords, private tokens, encryption keys, or PII must be documented as sensitive.
- **Logging Safety**: Mask sensitive fields during serialization, audit trail generation, and client logging.

## AIP-148: Standard Fields
Resource messages must use standard names and numbering conventions where applicable:
- `string name = 1 [(google.api.field_behavior) = IDENTIFIER];`
- `string display_name = ... [(google.api.field_behavior) = OPTIONAL];`
- `string description = ... [(google.api.field_behavior) = OPTIONAL];`
- `string uid = ... [(google.api.field_behavior) = OUTPUT_ONLY];`
- `string etag = ... [(google.api.field_behavior) = OPTIONAL];`
- `google.protobuf.Timestamp create_time = ... [(google.api.field_behavior) = OUTPUT_ONLY];`
- `google.protobuf.Timestamp update_time = ... [(google.api.field_behavior) = OUTPUT_ONLY];`
- `google.protobuf.Timestamp delete_time = ... [(google.api.field_behavior) = OUTPUT_ONLY];`
- `map<string, string> labels = ... [(google.api.field_behavior) = OPTIONAL];`

## AIP-149: Unset Field Values
- **Presence Semantics**: Rely on proto3 field presence.
- **Zero-Value Equivalence**: Do not assign semantic distinction between a field set to its default value (e.g., `0`, `""`, `false`) and an unset field unless `optional` presence is explicitly required.

## AIP-151: Long-Running Operations
- **Return Type**: Asynchronous tasks taking over a few seconds must return `google.longrunning.Operation`.
- **Annotation**: Annotate RPCs with `google.api.operation_response`:
  ```proto
  rpc CreateCluster(CreateClusterRequest) returns (google.longrunning.Operation) {
    option (google.api.http) = { ... };
    option (google.longrunning.operation_info) = {
      response_type: "Cluster"
      metadata_type: "OperationMetadata"
    };
  }
  ```

## AIP-152: Jobs
- **Pattern**: Asynchronous batch processes that run to completion should expose `Job` and `Execution` resources.
- **Lifecycle Control**: Provide standard CRUD plus custom methods for cancellation (`:cancel`) and retry (`:retry`).

## AIP-153: Import and Export
- **Pattern**: Mass data ingress and egress operations must be modeled as custom methods returning `google.longrunning.Operation`:
    - `rpc Import<Resources>(Import<Resources>Request) returns (google.longrunning.Operation);`
    - `rpc Export<Resources>(Export<Resources>Request) returns (google.longrunning.Operation);`
- **Source/Destination**: Use `inline_source` or external storage URIs (e.g., Cloud Storage / S3).

## AIP-154: Resource Freshness Validation
- **Concurrency Control**: Resources provide `string etag` for optimistic locking.
- **Conditional Writes**: Clients pass `etag` in `Update` or `Delete` requests. If etags mismatch, server returns `ABORTED` or `FAILED_PRECONDITION`.

## AIP-155: Request Identification
- **Idempotency Token**: Mutation requests (`Create`, `Update`, `Delete`, and custom methods) should include `string request_id = ... [(google.api.field_behavior) = OPTIONAL];`.
- **Server Cache**: Servers deduplicate identical requests submitted with the same `request_id` within an operational time window (typically 48 hours).

## AIP-156: Singleton Resources
- **URI Structure**: Resources that exist exactly once within their parent scope omit the `{resource_id}` placeholder:
  `publishers/{publisher}/settings`
- **CRUD Adjustments**: Singleton methods use `GetSettings` and `UpdateSettings`; `Create` and `Delete` are omitted if the singleton is implicit.

## AIP-157: Partial Responses
- **Projection**: Read requests may accept `google.protobuf.FieldMask read_mask` to allow clients to request a subset of fields.
- **Server Behavior**: The server returns only the fields specified in the mask, reducing serialization and network overhead.

## AIP-158: Pagination
- **Parameters**: `List` requests must accept `int32 page_size` and `string page_token`.
- **Response**: Returns `string next_page_token`. An empty string indicates the final page.
- **Token Design**: Page tokens must be opaque, tamper-evident, and encode keyset pagination offsets rather than raw SQL offset integers.

## AIP-159: Reading Across Collections
- **Cross-Collection Query**: Allow querying resources across multiple parents using the `-` wildcard character in parent names:
  `GET /v1/publishers/-/books`
- **Result Identity**: Returned resources must always populate their full canonical `name` field so the caller knows the exact parent.

## AIP-160: Filtering
- **Syntax**: `List` requests support a `string filter` field using AIP-160 filter expression grammar.
- **Operators**:
    - Equality: `author = "Orwell"`
    - Comparison: `pages >= 100`, `create_time > "2026-01-01T00:00:00Z"`
    - Logic: `AND`, `OR`, `NOT`
    - Traversal: `attributes.color = "blue"`
    - Wildcard / Has: `title:1984`

## AIP-161: Field Masks
- **Masking Updates**: `Update` requests must take `google.protobuf.FieldMask update_mask`.
- **Semantics**:
    - Only paths listed in `update_mask` are mutated.
    - If a path is in `update_mask` but absent from the resource payload, that field is cleared to its default/zero state.
    - If `update_mask` is omitted, default to full replacement.

## AIP-162: Resource Revisions
- **Version Tracking**: Resources requiring history snapshots should expose `string revision_id` and `google.protobuf.Timestamp revision_create_time`.
- **Operations**: Provide `List<Resource>Revisions`, `Rollback<Resource>`, and `Tag<Resource>Revision` where version control is exposed to users.

## AIP-163: Change Validation
- **Dry-Run**: Mutation requests should include `bool validate_only = ... [(google.api.field_behavior) = OPTIONAL];`.
- **Execution**: When `validate_only = true`, the server validates all request schemas, authorization rules, and referential integrity without persisting changes.

## AIP-164: Soft Delete
- **Lifecycle Tracking**: Soft-deleted resources populate `google.protobuf.Timestamp delete_time = ... [(google.api.field_behavior) = OUTPUT_ONLY];`.
- **List Filtering**: By default, `List` methods exclude soft-deleted items unless `bool show_deleted = true;` is specified in the request.
- **Restoration**: Provide a custom `rpc Undelete<Resource>(Undelete<Resource>Request) returns (<Resource>);` method.

## AIP-165: Criteria-Based Delete
- **Bulk Purge**: Operations that delete resources based on query predicates must be modeled as custom purge methods returning an LRO:
  `rpc Purge<Resources>(Purge<Resources>Request) returns (google.longrunning.Operation);`
- **Request Filter**: Accepts `string filter` to define targets and `bool force` to confirm execution.

## AIP-180: Backwards Compatibility
- **Breaking Changes Prohibited**:
    - Changing field tag numbers or scalar field types.
    - Renaming existing fields or enum constants.
    - Removing fields or enum values (mark with `[deprecated = true]` instead).
    - Modifying HTTP paths or changing HTTP verbs on existing methods.
    - Adding required fields to existing request messages.

## AIP-181: Stability Levels
- **Alpha (`v1alpha`)**: Experimental. No backward compatibility guarantees. May be deprecated or changed without notice.
- **Beta (`v1beta`)**: Feature-complete. Backward compatibility maintained; breaking changes allowed only across minor version bumps with deprecation notices.
- **GA (`v1`)**: Production-ready. Strict backward compatibility enforced indefinitely.

## AIP-182: External Software Dependencies
- **Abstraction**: Avoid exposing third-party vendor schemas directly in API definitions.
- **Decoupling**: Wrap external data formats in canonical protobuf messages under your service's domain.

## AIP-184: API Version Identifiers
- **Version Components**: Package names and URI paths must include the major API version:
    - Syntax: `package <org>.<service>.<version>;`
    - Examples: `package google.example.library.v1;`, `package google.example.library.v1beta1;`

## AIP-185: API Versioning
- **Major Version Transition**: Breaking changes require releasing a new major version package (e.g., `v1` to `v2`).
- **Coexistence**: Major versions must run concurrently without schema or route collisions.

## AIP-190: Naming Conventions
- **Services**: `PascalCase` ending in `Service` (e.g., `InventoryService`).
- **Messages**: `PascalCase` singular noun (e.g., `Warehouse`).
- **Fields**: `lower_snake_case` (e.g., `item_count`).
- **Enums**: Enum types use `PascalCase`; enum values use `UPPER_SNAKE_CASE` prefixed with the enum name.
- **RPC Methods**: `PascalCase` verb-first (e.g., `CreateOrder`, `CancelSubscription`).

## AIP-191: File and Directory Structure
- **Path Mapping**: Protobuf file locations must mirror their package declaration:
  `google/example/library/v1/library.proto` -> `package google.example.library.v1;`
- **Service Segregation**: Place services and resource definitions in distinct files if schemas grow large.

## AIP-192: Documentation
- **Completeness**: Every package, service, method, message, field, and enum value must include comments.
- **Format**: First sentence must be a concise, single-sentence summary ending in a period. Document constraints, valid ranges, and unit formats.

## AIP-193: Errors
- **Canonical Codes**: Use standard `google.rpc.Code` enum values (`OK`, `INVALID_ARGUMENT`, `NOT_FOUND`, `ALREADY_EXISTS`, `PERMISSION_DENIED`, `UNAUTHENTICATED`, `FAILED_PRECONDITION`, `ABORTED`, `UNAVAILABLE`, `DEADLINE_EXCEEDED`, `INTERNAL`).
- **Error Details**: Include rich typed error details:
    - `google.rpc.ErrorInfo`: Machine-readable reason, domain, and metadata.
    - `google.rpc.BadRequest`: Specific field-level violation descriptions.
    - `google.rpc.PreconditionFailure`: Violated system state requirements.
    - `google.rpc.ResourceInfo`: Type, name, and owner of the missing/conflicted resource.

## AIP-194: Automatic Retry Configuration
- **Idempotent Retries**: Client libraries should automatically retry idempotent methods (`Get`, `List`, and operations with `request_id`) on transient errors (`UNAVAILABLE`, `DEADLINE_EXCEEDED`).
- **Backoff Strategy**: Use truncated exponential backoff with jitter.

## AIP-200: Precedent
- **Consistency**: Follow established AIP design precedents. Avoid introducing bespoke patterns or non-standard naming when an existing AIP pattern addresses the scenario.

## AIP-202: Fields
- **Field Retention**: Never reuse or renumber retired field numbers.
- **Reservation**: When deleting deprecated fields, explicitly reserve their field numbers and names:
  `reserved 4, 8 to 12; reserved "legacy_sku", "temp_token";`

## AIP-203: Field Behavior Documentation
- **Annotations**: Annotate every field with `google.api.field_behavior`:
    - `REQUIRED`: Must be provided by client on creation/mutation.
    - `OUTPUT_ONLY`: Assigned and managed exclusively by the server.
    - `IMMUTABLE`: Set during creation; cannot be modified afterward.
    - `OPTIONAL`: May be omitted by the client.
    - `INPUT_ONLY`: Accepted on write; omitted from read responses.
    - `IDENTIFIER`: Identifies the resource's canonical name.

## AIP-205: Beta-Blocking Changes
- **Graduation Checklist**: Resolve all structural schema inconsistencies, naming mismatches, and deprecations before moving an API from `v1beta1` to `v1`.

## AIP-210: Unicode
- **Encoding**: All string fields must accept and store valid UTF-8 encoded text.
- **Normalization**: Servers should apply Unicode Normalization Form C (NFC) for string comparisons and indexing.

## AIP-211: Authorization Checks
- **Granular Checks**: Verify caller permissions for every RPC method before evaluating complex business logic or accessing backend data.
- **Missing Resources**: Return `NOT_FOUND` instead of `PERMISSION_DENIED` when the caller lacks access to discover that a resource exists.

## AIP-213: Common Components
- **Reusable Types**: Reuse standard Google Common Types rather than creating duplicates:
    - `google.type.Date`: Calendar dates.
    - `google.type.TimeOfDay`: Time without timezone.
    - `google.type.Money`: Currency amounts with currency codes.
    - `google.type.PostalAddress`: Physical mailing addresses.
    - `google.type.LatLng`: Geographic coordinates.

## AIP-214: Resource Expiration
- **Timestamp vs Duration**:
    - Use `google.protobuf.Timestamp expire_time` when the exact termination point is known.
    - Use `google.protobuf.Duration ttl` when expiration is set relative to creation or update time.
- **Server Lifecycle**: Mark expired resources as unavailable or purge them according to documented retention policies.

## AIP-215: API-Specific Protos
- **Decoupling**: Protos defining an API must not import internal protos or API-specific definitions from unrelated services.
- **Cross-Service References**: Reference external resources strictly via string resource name references (`google.api.resource_reference`).

## AIP-216: States
- **Enum Modeling**: Model resource lifecycles with an enum named `State`.
- **Field Definition**:
  ```proto
  enum State {
    STATE_UNSPECIFIED = 0;
    PENDING = 1;
    ACTIVE = 2;
    SUSPENDED = 3;
    DELETED = 4;
  }
  State state = 10 [(google.api.field_behavior) = OUTPUT_ONLY];
  ```

## AIP-217: Unreachable Resources
- **Partial Failure in List**: When listing resources across distributed partitions or multi-region federations, include `repeated string unreachable` in the response message:
  ```proto
  message ListClustersResponse {
    repeated Cluster clusters = 1;
    string next_page_token = 2;
    repeated string unreachable = 3 [(google.api.field_behavior) = OUTPUT_ONLY];
  }
  ```
- **Semantics**: Lists regional URI endpoints that failed to respond without failing the overall request.

## AIP-231: Batch Methods: Get
- **Signature**: `rpc BatchGet<Resources>(BatchGet<Resources>Request) returns (BatchGet<Resources>Response);`
- **HTTP Mapping**: `POST /v1/{parent=publishers/*}/books:batchGet`
- **Request Fields**: `parent` (required), `repeated string names` (required).
- **Response Fields**: `repeated <Resource> <resources> = 1;`.

## AIP-233: Batch Methods: Create
- **Signature**: `rpc BatchCreate<Resources>(BatchCreate<Resources>Request) returns (BatchCreate<Resources>Response);` (or `google.longrunning.Operation`).
- **HTTP Mapping**: `POST /v1/{parent=publishers/*}/books:batchCreate`
- **Request Fields**: `parent` (required), `repeated Create<Resource>Request requests` (required).
- **Response Fields**: `repeated <Resource> <resources> = 1;`.

## AIP-234: Batch Methods: Update
- **Signature**: `rpc BatchUpdate<Resources>(BatchUpdate<Resources>Request) returns (BatchUpdate<Resources>Response);` (or `google.longrunning.Operation`).
- **HTTP Mapping**: `POST /v1/{parent=publishers/*}/books:batchUpdate`
- **Request Fields**: `parent` (required), `repeated Update<Resource>Request requests` (required).
- **Response Fields**: `repeated <Resource> <resources> = 1;`.

## AIP-235: Batch Methods: Delete
- **Signature**: `rpc BatchDelete<Resources>(BatchDelete<Resources>Request) returns (google.protobuf.Empty);` (or `google.longrunning.Operation`).
- **HTTP Mapping**: `POST /v1/{parent=publishers/*}/books:batchDelete`
- **Request Fields**: `parent` (required), `repeated string names` (required).

## AIP-236: Policy Preview
- **Pre-Execution Evaluation**: Provide custom preview methods (e.g., `PreviewPolicyEvaluation`) allowing callers to simulate policy changes (such as IAM or validation constraints) and review the resulting access graph before committing updates.
