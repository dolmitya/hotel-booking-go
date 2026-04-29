package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func StartPostgres(t *testing.T) *sql.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(
		ctx,
		"postgres:17.7-alpine",
		postgres.WithDatabase("hotel_test"),
		postgres.WithUsername("booking"),
		postgres.WithPassword("booking"),
		postgres.WithSQLDriver("pgx"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() {
		if terminateErr := testcontainers.TerminateContainer(container); terminateErr != nil {
			t.Fatalf("terminate postgres container: %v", terminateErr)
		}
	})

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("build postgres connection string: %v", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open postgres connection: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	if err := applyMigrations(ctx, db); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	return db
}

func applyMigrations(ctx context.Context, db *sql.DB) error {
	bytes, err := os.ReadFile(migrationFilePath())
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	var builder strings.Builder
	for _, line := range strings.Split(string(bytes), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}

	for _, statement := range strings.Split(builder.String(), ";") {
		query := strings.TrimSpace(statement)
		if query == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("exec migration statement %q: %w", query, err)
		}
	}

	return nil
}

func migrationFilePath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean("internal/db/migrations/001_init_schema.sql")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "db", "migrations", "001_init_schema.sql"))
}
