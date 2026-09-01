package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDDL = `
CREATE DATABASE test_db;

CREATE TYPE user_status AS ENUM ('active', 'inactive');

CREATE TABLE users (
	id BIGSERIAL PRIMARY KEY,
	status user_status NOT NULL,
	email VARCHAR(255) NOT NULL UNIQUE,
	preferences JSONB
);

CREATE TABLE profiles (
	user_id BIGINT UNIQUE REFERENCES users(id),
	bio TEXT,
	location POINT
);

CREATE TABLE posts (
	id SERIAL PRIMARY KEY,
	author_id BIGINT NOT NULL REFERENCES users(id),
	title VARCHAR(255) NOT NULL
);

CREATE TABLE groups (
	id UUID PRIMARY KEY,
	name VARCHAR(255) NOT NULL
);

CREATE TABLE user_groups (
	user_id BIGINT REFERENCES users(id),
	group_id UUID REFERENCES groups(id),
	PRIMARY KEY (user_id, group_id)
);
`

const compositeDDL = `
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
	FOREIGN KEY (tenant_id, parent_id) REFERENCES parents(tenant_id, id)
);
`

func TestGeneratorIntegration(
	t *testing.T,
) {
	tempDir, err := os.MkdirTemp("", "generator-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fSys := fstest.MapFS{
		"schema.sql": &fstest.MapFile{
			Data: []byte(testDDL),
		},
	}

	err = Run(fSys, "testdb", tempDir)
	require.NoError(t, err)

	// Check generated files directly in tempDir
	files := []string{
		"models.go",
		"mutator.go",
		"schema.go",
		"table.go",
		"hydrate.go",
		"dehydrate.go",
		"relations.go",
		"query.go",
		"enums.go",
	}

	for _, f := range files {
		p := filepath.Join(tempDir, f)
		assert.FileExists(t, p)

		// Ensure 80-character limit is respected in generated files
		var bytes []byte
		bytes, err = os.ReadFile(p)
		require.NoError(t, err)
		lines := strings.Split(string(bytes), "\n")
		for idx, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "sqlStr :=") {
				continue
			}
			assert.LessOrEqual(
				t,
				len(line),
				80,
				"Line %d in file %s exceeds 80 characters: %q",
				idx+1,
				f,
				line,
			)
		}

	}
}


func TestCompositeKeyIntegration(
	t *testing.T,
) {
	tempDir, err := os.MkdirTemp("", "generator-composite-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fSys := fstest.MapFS{
		"schema.sql": &fstest.MapFile{
			Data: []byte(compositeDDL),
		},
	}

	err = Run(fSys, "testcomp", tempDir)

	require.NoError(t, err)

	p := filepath.Join(tempDir, "relations.go")
	assert.FileExists(t, p)

	bytes, err := os.ReadFile(p)
	require.NoError(t, err)
	content := string(bytes)

	// Verify that composite columns are written to ForeignKeyColumns
	assert.Contains(t, content, "Schema.Public.PublicChild.TenantID")
	assert.Contains(t, content, "Schema.Public.PublicChild.ParentID")

	// Verify that the composite key extractors return slice slices of any
	assert.Contains(t, content, "[]any{")
}

const nexusDDL = `
CREATE SCHEMA IF NOT EXISTS nexus;

CREATE TABLE IF NOT EXISTS nexus.users
(
    id                        UUID PRIMARY KEY NOT NULL DEFAULT uuidv7(),
    name                      VARCHAR(255)     NOT NULL,
    email                     VARCHAR(255)     NOT NULL,
    password                  VARCHAR(255)     NOT NULL,
    two_factor_secret         VARCHAR(512),
    two_factor_recovery_codes VARCHAR(128)[],
    created_at                TIMESTAMPTZ(6) NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ(6) NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_b36ae3d17c33fcb518047247480b2e31 ON nexus.users (email);
`

func TestNexusSchemaIntegration(
	t *testing.T,
) {
	tempDir, err := os.MkdirTemp("", "generator-nexus-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fSys := fstest.MapFS{
		"schema.sql": &fstest.MapFile{
			Data: []byte(nexusDDL),
		},
	}

	err = Run(fSys, "nexusdb", tempDir)
	require.NoError(t, err)

	// Verify models.go
	p := filepath.Join(tempDir, "models.go")
	assert.FileExists(t, p)
	bytes, err := os.ReadFile(p)
	require.NoError(t, err)
	content := string(bytes)
	assert.Contains(t, content, "// NexusUser represents a model for nexus.users.")
	assert.Contains(t, content, "type NexusUser struct {")
	assert.NotContains(t, content, "NexuUser")

	// Verify schema.go
	pSchema := filepath.Join(tempDir, "schema.go")
	assert.FileExists(t, pSchema)
	bytesSchema, err := os.ReadFile(pSchema)
	require.NoError(t, err)
	contentSchema := string(bytesSchema)
	assert.Contains(t, contentSchema, "Nexus struct {\n\t\tNexusUser struct {")
	assert.Contains(t, contentSchema, "column.UUIDColumn[NexusUser]")
	assert.NotContains(t, contentSchema, "NexuUser")

	// Verify mutator.go
	pMutator := filepath.Join(tempDir, "mutator.go")
	assert.FileExists(t, pMutator)
	bytesMutator, err := os.ReadFile(pMutator)
	require.NoError(t, err)
	contentMutator := string(bytesMutator)
	assert.Contains(t, contentMutator, "type NexusUserMutator struct")
	assert.NotContains(t, contentMutator, "NexuUser")
}

const franchiseInquiriesDDL = `
CREATE SCHEMA IF NOT EXISTS nexus;

CREATE TYPE nexus.franchise_inquiry_status AS ENUM (
	'pending',
	'contacted',
	'archived'
);

CREATE TABLE IF NOT EXISTS nexus.franchise_inquiries
(
    id         UUID PRIMARY KEY         NOT NULL,
    name       VARCHAR(255)             NOT NULL,
    email      VARCHAR(255)             NOT NULL,
    phone      VARCHAR(50)              NOT NULL,
    message    TEXT                     NOT NULL,
    status     nexus.franchise_inquiry_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ(6)           NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ(6)           NOT NULL DEFAULT NOW()
);
`

func TestNexusFranchiseInquiriesEnumIntegration(
	t *testing.T,
) {
	tempDir, err := os.MkdirTemp("", "generator-franchise-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fSys := fstest.MapFS{
		"schema.sql": &fstest.MapFile{
			Data: []byte(franchiseInquiriesDDL),
		},
	}

	err = Run(fSys, "nexus", tempDir)
	require.NoError(t, err)

	// Verify enums.go
	pEnums := filepath.Join(tempDir, "enums.go")
	assert.FileExists(t, pEnums)
	bytesEnums, errEnums := os.ReadFile(pEnums)
	require.NoError(t, errEnums)
	contentEnums := string(bytesEnums)
	assert.Contains(t, contentEnums, "type FranchiseInquiryStatus string")
	assert.Contains(
		t,
		contentEnums,
		`const FranchiseInquiryStatusPending FranchiseInquiryStatus = "pending"`,
	)
	assert.Contains(
		t,
		contentEnums,
		`const FranchiseInquiryStatusContacted FranchiseInquiryStatus = "contacted"`,
	)
	assert.Contains(
		t,
		contentEnums,
		`const FranchiseInquiryStatusArchived FranchiseInquiryStatus = "archived"`,
	)

	// Verify models.go
	pModels := filepath.Join(tempDir, "models.go")
	assert.FileExists(t, pModels)
	bytesModels, errModels := os.ReadFile(pModels)
	require.NoError(t, errModels)
	contentModels := string(bytesModels)
	assert.Regexp(t, `Status\s+FranchiseInquiryStatus`, contentModels)

	// Verify mutator.go
	pMutator := filepath.Join(tempDir, "mutator.go")
	assert.FileExists(t, pMutator)
	bytesMutator, errMutator := os.ReadFile(pMutator)
	require.NoError(t, errMutator)
	contentMutator := string(bytesMutator)
	assert.Regexp(
		t,
		`Status\s+contract\.Set\[FranchiseInquiryStatus\]`,
		contentMutator,
	)

	// Verify schema.go
	pSchema := filepath.Join(tempDir, "schema.go")
	assert.FileExists(t, pSchema)
	bytesSchema, errSchema := os.ReadFile(pSchema)
	require.NoError(t, errSchema)
	contentSchema := string(bytesSchema)
	pattern := `Status\s+column\.StringColumn\[` +
		`NexusFranchiseInquiry,\s*FranchiseInquiryStatus\]`
	assert.Regexp(
		t,
		pattern,
		contentSchema,
	)

	// Verify line length <= 80 characters for all generated files
	files := []string{
		"models.go",
		"mutator.go",
		"schema.go",
		"table.go",
		"hydrate.go",
		"dehydrate.go",
		"relations.go",
		"query.go",
		"enums.go",
	}
	for _, f := range files {
		p := filepath.Join(tempDir, f)
		assert.FileExists(t, p)
		var b []byte
		b, err = os.ReadFile(p)
		require.NoError(t, err)
		lines := strings.Split(string(b), "\n")
		for idx, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "sqlStr :=") {
				continue
			}
			assert.LessOrEqual(
				t,
				len(line),
				80,
				"Line %d in file %s exceeds 80 characters: %q",
				idx+1,
				f,
				line,
			)
		}
	}
}

const fleetDDL = `
CREATE SCHEMA IF NOT EXISTS fleet;

CREATE TABLE IF NOT EXISTS fleet.addresses
(
    id     UUID PRIMARY KEY NOT NULL DEFAULT uuidv7(),
    street VARCHAR(255)     NOT NULL
);

CREATE TABLE IF NOT EXISTS fleet.franchises
(
    id                 UUID PRIMARY KEY NOT NULL DEFAULT uuidv7(),
    name               VARCHAR(255)     NOT NULL UNIQUE,
    address_billing_id UUID             NOT NULL REFERENCES fleet.addresses (id)
);
`

func TestFleetFranchiseCustomRelationIntegration(
	t *testing.T,
) {
	tempDir, err := os.MkdirTemp("", "generator-fleet-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	fSys := fstest.MapFS{
		"schema.sql": &fstest.MapFile{
			Data: []byte(fleetDDL),
		},
	}

	err = Run(fSys, "fleetdb", tempDir)
	require.NoError(t, err)

	// Verify models.go has AddressBilling *FleetAddress
	pModels := filepath.Join(tempDir, "models.go")
	assert.FileExists(t, pModels)
	bytesModels, errModels := os.ReadFile(pModels)
	require.NoError(t, errModels)
	contentModels := string(bytesModels)
	assert.Contains(
		t,
		contentModels,
		"AddressBilling *FleetAddress",
	)
	assert.NotContains(
		t,
		contentModels,
		"FleetAddress *FleetAddress",
	)

	// Verify schema.go has AddressBilling
	pSchema := filepath.Join(tempDir, "schema.go")
	assert.FileExists(t, pSchema)
	bytesSchema, errSchema := os.ReadFile(pSchema)
	require.NoError(t, errSchema)
	contentSchema := string(bytesSchema)
	assert.Regexp(
		t,
		`AddressBilling\s+relation\.BelongsTo\[`,
		contentSchema,
	)

	// Verify relations.go
	pRelations := filepath.Join(tempDir, "relations.go")
	assert.FileExists(t, pRelations)
	bytesRel, errRel := os.ReadFile(pRelations)
	require.NoError(t, errRel)
	contentRel := string(bytesRel)
	assert.Contains(
		t,
		contentRel,
		"Schema.Fleet.FleetFranchise.AddressBilling = relation.BelongsTo[",
	)
}

