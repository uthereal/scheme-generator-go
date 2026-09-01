package parser_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uthereal/scheme-generator-go/internal/ast"
	"github.com/uthereal/scheme-generator-go/internal/parser"
)

func TestParseDDL_Basic(
	t *testing.T,
) {
	ddl := `
	CREATE SCHEMA security;
	CREATE TYPE security.user_role AS ENUM ('admin', 'member');

	CREATE TABLE security.accounts (
		id BIGSERIAL PRIMARY KEY,
		role security.user_role NOT NULL,
		email VARCHAR(255) UNIQUE,
		tags TEXT[],
		metadata JSONB NOT NULL DEFAULT '{}'
	);
	`

	state := parser.NewPostgresAccumulator()
	err := state.ParseDDL(ddl)
	require.NoError(t, err)

	require.Len(t, state.Schemas, 2)
	assert.Equal(t, "public", state.Schemas[0].Name)
	assert.Equal(t, "security", state.Schemas[1].Name)

	secSchema := state.Schemas[1]
	require.Contains(t, secSchema.Enums, "user_role")
	enum := secSchema.Enums["user_role"]
	assert.Equal(t, []string{"admin", "member"}, enum.Values)

	require.Contains(t, secSchema.Tables, "accounts")
	tbl := secSchema.Tables["accounts"]

	assert.Equal(t, "accounts", tbl.Name)
	assert.Equal(t, "security", tbl.Schema.Name)
	assert.Equal(t, secSchema, tbl.Schema)
	assert.Equal(t, []string{"id"}, ast.ColumnNames(tbl.PrimaryKey))

	assert.Len(t, tbl.Columns, 5)

	colID := tbl.Columns[0]
	assert.Equal(t, "id", colID.Name)
	assert.Equal(t, "bigserial", colID.Type)
	assert.True(t, colID.IsNullable)

	colRole := tbl.Columns[1]
	assert.Equal(t, "role", colRole.Name)
	assert.Equal(t, "user_role", colRole.Type)
	assert.Equal(t, ast.CustomEnum, colRole.CustomType)
	assert.Equal(t, "UserRole", colRole.ToModelType())
	assert.False(t, colRole.IsNullable)

	colEmail := tbl.Columns[2]
	assert.Equal(t, "email", colEmail.Name)
	assert.Equal(t, "varchar", colEmail.Type)
	assert.True(t, colEmail.IsNullable)

	colTags := tbl.Columns[3]
	assert.Equal(t, "tags", colTags.Name)
	assert.Equal(t, "text", colTags.Type)
	assert.True(t, colTags.IsArray)

	colMeta := tbl.Columns[4]
	assert.Equal(t, "metadata", colMeta.Name)
	assert.Equal(t, "jsonb", colMeta.Type)
	require.NotNil(t, colMeta.Default)
	assert.Equal(t, "default", *colMeta.Default)
}

func TestParseDDL_AlterTable(
	t *testing.T,
) {
	ddl := `
	CREATE TABLE items (
		id INT,
		tenant_id INT,
		extra TEXT,
		status TEXT DEFAULT 'pending'
	);
	ALTER TABLE items ADD PRIMARY KEY (id, tenant_id);
	ALTER TABLE items DROP COLUMN extra;
	ALTER TABLE items ADD COLUMN note TEXT;
	ALTER TABLE items ALTER COLUMN note SET NOT NULL;
	ALTER TABLE items ALTER COLUMN status DROP DEFAULT;
	ALTER TABLE items ALTER COLUMN note TYPE VARCHAR(255);
	`

	state := parser.NewPostgresAccumulator()
	err := state.ParseDDL(ddl)
	require.NoError(t, err)

	require.Len(t, state.Schemas, 1)
	sc := state.Schemas[0]
	require.Contains(t, sc.Tables, "items")
	tbl := sc.Tables["items"]

	assert.Equal(
		t,
		[]string{"id", "tenant_id"},
		ast.ColumnNames(tbl.PrimaryKey),
	)
	require.Len(t, tbl.Columns, 4)
	assert.Equal(t, "id", tbl.Columns[0].Name)
	assert.Equal(t, "tenant_id", tbl.Columns[1].Name)
	assert.Equal(t, "status", tbl.Columns[2].Name)
	assert.Nil(t, tbl.Columns[2].Default)
	assert.Equal(t, "note", tbl.Columns[3].Name)
	assert.Equal(t, "varchar", tbl.Columns[3].Type)
	assert.False(t, tbl.Columns[3].IsNullable)
}

func TestParseDDL_Relations(
	t *testing.T,
) {
	t.Run("auto-detects 1:N BelongsTo and HasMany relations", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE posts (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id),
			title TEXT NOT NULL
		);
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		pubSchema := state.Schemas[0]
		usersTbl := pubSchema.Tables["users"]
		postsTbl := pubSchema.Tables["posts"]
		require.NotNil(t, usersTbl)
		require.NotNil(t, postsTbl)

		require.Len(t, usersTbl.HasMany, 1)
		assert.Equal(t, "Posts", usersTbl.HasMany[0].Name)
		assert.Equal(t, "PublicUser", usersTbl.HasMany[0].ParentModel)
		assert.Equal(t, "PublicPost", usersTbl.HasMany[0].ChildModel)
		assert.Equal(t, postsTbl, usersTbl.HasMany[0].ChildTable)

		require.Len(t, postsTbl.BelongsTo, 1)
		assert.Equal(t, "PublicUser", postsTbl.BelongsTo[0].Name)
		assert.Equal(t, "PublicPost", postsTbl.BelongsTo[0].ParentModel)
		assert.Equal(t, "PublicUser", postsTbl.BelongsTo[0].ChildModel)
		assert.Equal(t, usersTbl, postsTbl.BelongsTo[0].ChildTable)
	})

	t.Run("auto-detects 1:1 HasOne and BelongsTo relations", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE profiles (
			user_id BIGINT UNIQUE REFERENCES users(id),
			bio TEXT
		);
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		pubSchema := state.Schemas[0]
		usersTbl := pubSchema.Tables["users"]
		profilesTbl := pubSchema.Tables["profiles"]
		require.NotNil(t, usersTbl)
		require.NotNil(t, profilesTbl)

		require.Len(t, usersTbl.HasOne, 1)
		assert.Equal(t, "PublicProfile", usersTbl.HasOne[0].Name)
		assert.Equal(t, "PublicUser", usersTbl.HasOne[0].ParentModel)
		assert.Equal(t, "PublicProfile", usersTbl.HasOne[0].ChildModel)
		assert.Equal(t, profilesTbl, usersTbl.HasOne[0].ChildTable)

		require.Len(t, profilesTbl.BelongsTo, 1)
		assert.Equal(t, "PublicUser", profilesTbl.BelongsTo[0].Name)
	})

	t.Run("auto-detects M:N BelongsToMany on pivot table", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY
		);

		CREATE TABLE groups (
			id BIGSERIAL PRIMARY KEY
		);

		CREATE TABLE user_groups (
			user_id BIGINT REFERENCES users(id),
			group_id BIGINT REFERENCES groups(id),
			PRIMARY KEY (user_id, group_id)
		);
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		pubSchema := state.Schemas[0]
		usersTbl := pubSchema.Tables["users"]
		groupsTbl := pubSchema.Tables["groups"]
		require.NotNil(t, usersTbl)
		require.NotNil(t, groupsTbl)

		require.Len(t, usersTbl.BelongsToMany, 1)
		assert.Equal(t, "Groups", usersTbl.BelongsToMany[0].FieldName)
		assert.Equal(
			t,
			"PublicUser",
			usersTbl.BelongsToMany[0].ParentModel,
		)
		assert.Equal(
			t,
			"PublicGroup",
			usersTbl.BelongsToMany[0].ChildModel,
		)
		assert.Equal(t, groupsTbl, usersTbl.BelongsToMany[0].ChildTable)

		require.Len(t, groupsTbl.BelongsToMany, 1)
		assert.Equal(t, "Users", groupsTbl.BelongsToMany[0].FieldName)
		assert.Equal(
			t,
			"PublicGroup",
			groupsTbl.BelongsToMany[0].ParentModel,
		)
		assert.Equal(
			t,
			"PublicUser",
			groupsTbl.BelongsToMany[0].ChildModel,
		)
		assert.Equal(t, usersTbl, groupsTbl.BelongsToMany[0].ChildTable)
	})

	t.Run("handles composite foreign key relations", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TABLE parents (
			tenant_id INT,
			id INT,
			name VARCHAR(255),
			PRIMARY KEY (tenant_id, id)
		);

		CREATE TABLE children (
			id INT PRIMARY KEY,
			tenant_id INT,
			parent_id INT,
			FOREIGN KEY (tenant_id, parent_id)
				REFERENCES parents(tenant_id, id)
		);
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		pubSchema := state.Schemas[0]
		parentsTbl := pubSchema.Tables["parents"]
		childrenTbl := pubSchema.Tables["children"]
		require.NotNil(t, parentsTbl)
		require.NotNil(t, childrenTbl)

		require.Len(t, parentsTbl.HasMany, 1)
		assert.Equal(
			t,
			[]string{"tenant_id", "parent_id"},
			ast.ColumnNames(parentsTbl.HasMany[0].ForeignKeyCols),
		)
		assert.Equal(
			t,
			[]string{"tenant_id", "id"},
			ast.ColumnNames(parentsTbl.HasMany[0].LocalKeyCols),
		)
	})

	t.Run("handles cross-schema relations and name disambiguation", func(
		t *testing.T,
	) {
		ddl := `
		CREATE SCHEMA auth;
		CREATE SCHEMA content;

		CREATE TABLE auth.users (
			id BIGSERIAL PRIMARY KEY,
			email TEXT NOT NULL
		);

		CREATE TABLE auth.tags (
			id BIGSERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);

		CREATE TABLE content.tags (
			id BIGSERIAL PRIMARY KEY,
			label TEXT NOT NULL
		);

		CREATE TABLE content.posts (
			id BIGSERIAL PRIMARY KEY,
			author_id BIGINT NOT NULL REFERENCES auth.users(id),
			title TEXT NOT NULL
		);

		CREATE TABLE content.post_tags (
			post_id BIGINT NOT NULL REFERENCES content.posts(id),
			tag_id BIGINT NOT NULL REFERENCES content.tags(id),
			PRIMARY KEY (post_id, tag_id)
		);
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		var authSchema *ast.Schema
		var contentSchema *ast.Schema
		for _, sc := range state.Schemas {
			if sc.Name == "auth" {
				authSchema = sc
			}
			if sc.Name == "content" {
				contentSchema = sc
			}
		}
		require.NotNil(t, authSchema)
		require.NotNil(t, contentSchema)

		authUsers := authSchema.Tables["users"]
		contentPosts := contentSchema.Tables["posts"]
		contentTags := contentSchema.Tables["tags"]
		require.NotNil(t, authUsers)
		require.NotNil(t, contentPosts)
		require.NotNil(t, contentTags)

		require.Len(t, authUsers.HasMany, 1)
		assert.Equal(t, "Posts", authUsers.HasMany[0].Name)
		assert.Equal(t, contentPosts, authUsers.HasMany[0].ChildTable)

		require.Len(t, contentPosts.BelongsTo, 1)
		assert.Equal(t, "AuthUser", contentPosts.BelongsTo[0].Name)
		assert.Equal(t, authUsers, contentPosts.BelongsTo[0].ChildTable)

		require.Len(t, contentPosts.BelongsToMany, 1)
		assert.Equal(
			t,
			"Tags",
			contentPosts.BelongsToMany[0].FieldName,
		)
		assert.Equal(
			t,
			contentTags,
			contentPosts.BelongsToMany[0].ChildTable,
		)
	})
}

func TestParseDDL_Rename(
	t *testing.T,
) {
	ddl := `
	CREATE SCHEMA old_schema;
	CREATE TABLE old_schema.old_table (
		old_id BIGSERIAL PRIMARY KEY,
		val TEXT
	);
	CREATE TABLE old_schema.dependents (
		id BIGSERIAL PRIMARY KEY,
		target_id BIGINT REFERENCES old_schema.old_table(old_id)
	);

	ALTER SCHEMA old_schema RENAME TO new_schema;
	ALTER TABLE new_schema.old_table RENAME TO new_table;
	ALTER TABLE new_schema.new_table RENAME COLUMN old_id TO new_id;
	`

	state := parser.NewPostgresAccumulator()
	err := state.ParseDDL(ddl)
	require.NoError(t, err)

	var newSchema *ast.Schema
	for _, sc := range state.Schemas {
		if sc.Name == "new_schema" {
			newSchema = sc
			break
		}
	}
	require.NotNil(t, newSchema)
	require.Contains(t, newSchema.Tables, "new_table")
	require.Contains(t, newSchema.Tables, "dependents")

	newTbl := newSchema.Tables["new_table"]
	depTbl := newSchema.Tables["dependents"]

	assert.Equal(t, "new_id", newTbl.Columns[0].Name)
	assert.Equal(t, []string{"new_id"}, ast.ColumnNames(newTbl.PrimaryKey))

	require.Len(t, depTbl.ForeignKeys, 1)
	fk := depTbl.ForeignKeys[0]
	assert.Equal(t, newSchema, fk.RefSchema)
	assert.Equal(t, newTbl, fk.RefTable)
}

func TestParseDDL_Drop(
	t *testing.T,
) {
	t.Run("drops tables and schemas correctly", func(t *testing.T) {
		ddl := `
		CREATE SCHEMA temp_schema;
		CREATE TABLE temp_schema.t1 (id INT PRIMARY KEY);
		CREATE TABLE temp_schema.t2 (id INT PRIMARY KEY);
		DROP TABLE temp_schema.t1;
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		var tempSchema *ast.Schema
		for _, sc := range state.Schemas {
			if sc.Name == "temp_schema" {
				tempSchema = sc
				break
			}
		}
		require.NotNil(t, tempSchema)
		assert.NotContains(t, tempSchema.Tables, "t1")
		assert.Contains(t, tempSchema.Tables, "t2")
	})

	t.Run("drop schema removes schema", func(t *testing.T) {
		ddl := `
		CREATE SCHEMA to_delete;
		DROP SCHEMA to_delete;
		`

		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)

		for _, sc := range state.Schemas {
			assert.NotEqual(t, "to_delete", sc.Name)
		}
	})
}

func TestParseDDL_StrictLookupErrors(
	t *testing.T,
) {
	t.Run("Alter table in nonexistent schema", func(t *testing.T) {
		ddl := `ALTER TABLE nonexistent_schema.items ADD COLUMN foo INT;`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"schema 'nonexistent_schema' does not exist",
		)
	})

	t.Run("Create table in nonexistent schema", func(t *testing.T) {
		ddl := `CREATE TABLE nonexistent_schema.items (id INT);`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"schema 'nonexistent_schema' does not exist",
		)
	})

	t.Run("Drop table in nonexistent schema", func(t *testing.T) {
		ddl := `DROP TABLE nonexistent_schema.items;`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"table 'nonexistent_schema.items' does not exist",
		)
	})

	t.Run("Drop schema does not exist", func(t *testing.T) {
		ddl := `DROP SCHEMA nonexistent_schema;`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"schema 'nonexistent_schema' does not exist",
		)
	})

	t.Run("Drop schema IF EXISTS does not error", func(t *testing.T) {
		ddl := `DROP SCHEMA IF EXISTS nonexistent_schema;`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.NoError(t, err)
	})
}

func TestParseDDL_CompoundSchemaStatements(
	t *testing.T,
) {
	ddl := `
	CREATE SCHEMA app_schema
		CREATE TABLE users (id BIGSERIAL PRIMARY KEY, name TEXT)
		CREATE TABLE orders (id BIGSERIAL PRIMARY KEY, user_id BIGINT);
	`

	state := parser.NewPostgresAccumulator()
	err := state.ParseDDL(ddl)
	require.NoError(t, err)

	var appSchema *ast.Schema
	for _, sc := range state.Schemas {
		if sc.Name == "app_schema" {
			appSchema = sc
			break
		}
	}
	require.NotNil(t, appSchema)
	assert.Contains(t, appSchema.Tables, "users")
	assert.Contains(t, appSchema.Tables, "orders")
}

func TestParseDDL_RenameCollisions(
	t *testing.T,
) {
	t.Run("Rename schema into existing schema returns error", func(
		t *testing.T,
	) {
		ddl := `
		CREATE SCHEMA s1;
		CREATE SCHEMA s2;
		ALTER SCHEMA s1 RENAME TO s2;
		`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "schema 's2' already exists")
	})

	t.Run("Rename table into existing table returns error", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TABLE t1 (id INT);
		CREATE TABLE t2 (id INT);
		ALTER TABLE t1 RENAME TO t2;
		`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "table 'public.t2' already exists")
	})

	t.Run("Rename column into existing column returns error", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TABLE c_table (c1 INT, c2 INT);
		ALTER TABLE c_table RENAME COLUMN c1 TO c2;
		`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"column 'c2' already exists in table 'public.c_table'",
		)
	})

	t.Run("Rename type into existing type returns error", func(
		t *testing.T,
	) {
		ddl := `
		CREATE TYPE e1 AS ENUM ('a', 'b');
		CREATE TYPE e2 AS ENUM ('c', 'd');
		ALTER TYPE e1 RENAME TO e2;
		`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "enum 'public.e2' already exists")
	})

	t.Run("Move table to schema with existing table returns error", func(
		t *testing.T,
	) {
		ddl := `
		CREATE SCHEMA custom_sc;
		CREATE TABLE custom_sc.same_name (id INT);
		CREATE TABLE same_name (id INT);
		ALTER TABLE same_name SET SCHEMA custom_sc;
		`
		state := parser.NewPostgresAccumulator()
		err := state.ParseDDL(ddl)
		require.Error(t, err)
		assert.Contains(
			t,
			err.Error(),
			"table 'custom_sc.same_name' already exists",
		)
	})
}

func TestParseDDL_EnumResolution(
	t *testing.T,
) {
	ddl := `
	CREATE SCHEMA nexus;
	CREATE TYPE nexus.franchise_inquiry_status AS ENUM (
		'pending',
		'contacted',
		'archived'
	);

	CREATE TABLE nexus.franchise_inquiries (
		id UUID PRIMARY KEY NOT NULL,
		name VARCHAR(255) NOT NULL,
		status nexus.franchise_inquiry_status NOT NULL DEFAULT 'pending',
		backup_status nexus.franchise_inquiry_status,
		all_statuses nexus.franchise_inquiry_status[] NOT NULL DEFAULT '{}'
	);
	`

	state := parser.NewPostgresAccumulator()
	err := state.ParseDDL(ddl)
	require.NoError(t, err)

	require.Len(t, state.Schemas, 2)
	nexusSc := state.Schemas[0]
	if nexusSc.Name != "nexus" {
		nexusSc = state.Schemas[1]
	}
	require.Equal(t, "nexus", nexusSc.Name)
	require.Contains(t, nexusSc.Tables, "franchise_inquiries")

	tbl := nexusSc.Tables["franchise_inquiries"]
	require.Len(t, tbl.Columns, 5)

	colStatus := tbl.Columns[2]
	assert.Equal(t, "status", colStatus.Name)
	assert.Equal(t, "franchise_inquiry_status", colStatus.Type)
	assert.Equal(t, ast.CustomEnum, colStatus.CustomType)
	assert.Equal(t, "FranchiseInquiryStatus", colStatus.ToModelType())
	assert.Equal(
		t,
		"column.StringColumn[NexusFranchiseInquiry, FranchiseInquiryStatus]",
		colStatus.ToQueryColumnType("NexusFranchiseInquiry"),
	)

	colBackup := tbl.Columns[3]
	assert.Equal(t, "backup_status", colBackup.Name)
	assert.Equal(t, ast.CustomEnum, colBackup.CustomType)
	assert.Equal(t, "*FranchiseInquiryStatus", colBackup.ToModelType())
	assert.Equal(
		t,
		"column.NullableStringColumn[NexusFranchiseInquiry, FranchiseInquiryStatus]",
		colBackup.ToQueryColumnType("NexusFranchiseInquiry"),
	)

	colAll := tbl.Columns[4]
	assert.Equal(t, "all_statuses", colAll.Name)
	assert.Equal(t, ast.CustomEnum, colAll.CustomType)
	assert.Equal(t, "[]FranchiseInquiryStatus", colAll.ToModelType())
	assert.Equal(
		t,
		"column.ArrayColumn[NexusFranchiseInquiry, FranchiseInquiryStatus]",
		colAll.ToQueryColumnType("NexusFranchiseInquiry"),
	)
}
