package example_test

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/uthereal/scheme-runtime-go/pkg/testutil"
)

// pgContainer maintains the shared PostgreSQL container for integration tests.
var pgContainer *testutil.PostgresContainer

// TestMain manages setup and teardown for the PostgreSQL integration tests.
func TestMain(
	m *testing.M,
) {
	ctx := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	logger := slog.Default()

	var err error
	pgContainer, err = testutil.StartPostgresContainer(ctxTimeout)
	if err != nil {
		logger.ErrorContext(
			ctxTimeout,
			"Failed to start Postgres container.",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	mFiles := []string{
		"migrations/01_init.sql",
		"migrations/02_refactor_renames.sql",
		"migrations/03_add_posts_and_comments.sql",
		"migrations/04_composite_and_multi.sql",
		"migrations/05_add_uuid_devices.sql",
	}

	var ddlParts []string
	for _, f := range mFiles {
		var bytes []byte
		bytes, err = os.ReadFile(f)
		if err != nil {
			logger.ErrorContext(
				ctxTimeout,
				"Failed to read migration file.",
				slog.Any("error", err),
				slog.String("file", f),
			)
			_ = testutil.StopPostgresContainer(ctxTimeout, pgContainer)
			os.Exit(1)
		}
		ddlParts = append(ddlParts, string(bytes))
	}

	fullDDL := strings.Join(ddlParts, "\n")
	lines := strings.Split(fullDDL, "\n")
	var executionLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(strings.ToUpper(line))
		if strings.HasPrefix(trimmed, "CREATE DATABASE") ||
			strings.HasPrefix(trimmed, "ALTER DATABASE") {
			continue
		}
		executionLines = append(executionLines, line)
	}
	cleanDDL := strings.Join(executionLines, "\n")

	err = pgContainer.SetupTemplateDB(ctxTimeout, "example_template", cleanDDL)
	if err != nil {
		logger.ErrorContext(
			ctxTimeout,
			"Failed to setup template DB.",
			slog.Any("error", err),
		)
		_ = testutil.StopPostgresContainer(ctxTimeout, pgContainer)
		os.Exit(1)
	}

	code := m.Run()

	_ = testutil.StopPostgresContainer(ctxTimeout, pgContainer)
	os.Exit(code)
}

// pointerTo returns a pointer to the provided value.
func pointerTo[T any](
	v T,
) *T {
	return &v
}
