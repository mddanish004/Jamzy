package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mddanish004/jamzy/server/internal/logging"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const (
	migrationDir   = "migrations"
	migrationLock  = int64(773411902)
	prismaTable    = "_prisma_migrations"
	versionedTable = "schema_migrations"
)

func Migrate(ctx context.Context, pool *pgxpool.Pool, log logging.Logger) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer conn.Release()

	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationLock); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", migrationLock)
	}()

	if _, err := conn.Exec(ctx, createVersionedTable); err != nil {
		return fmt.Errorf("create %s: %w", versionedTable, err)
	}

	applied, err := appliedVersions(ctx, conn)
	if err != nil {
		return err
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		version := strings.TrimSuffix(name, ".sql")
		if applied[version] {
			continue
		}

		script, err := migrationFiles.ReadFile(migrationDir + "/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		if err := applyMigration(ctx, conn, version, string(script)); err != nil {
			return err
		}

		log.Info("Applied migration", version)
	}

	return nil
}

const createVersionedTable = `
CREATE TABLE IF NOT EXISTS ` + versionedTable + ` (
    version    TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`

func appliedVersions(ctx context.Context, conn *pgxpool.Conn) (map[string]bool, error) {
	applied := map[string]bool{}

	rows, err := conn.Query(ctx, "SELECT version FROM "+versionedTable)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", versionedTable, err)
	}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			return nil, err
		}
		applied[version] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	prisma, err := prismaVersions(ctx, conn)
	if err != nil {
		return nil, err
	}
	for version := range prisma {
		applied[version] = true
	}

	return applied, nil
}

func prismaVersions(ctx context.Context, conn *pgxpool.Conn) (map[string]bool, error) {
	versions := map[string]bool{}

	var exists bool
	if err := conn.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", prismaTable).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return versions, nil
	}

	rows, err := conn.Query(ctx, "SELECT migration_name FROM "+prismaTable+" WHERE rolled_back_at IS NULL")
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", prismaTable, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		versions[name] = true
	}

	return versions, rows.Err()
}

func migrationNames() ([]string, error) {
	entries, err := migrationFiles.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func applyMigration(ctx context.Context, conn *pgxpool.Conn, version string, script string) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", version, err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()

	results := tx.Conn().PgConn().Exec(ctx, script)
	if _, err := results.ReadAll(); err != nil {
		return fmt.Errorf("run migration %s: %w", version, err)
	}

	if _, err := tx.Exec(ctx, "INSERT INTO "+versionedTable+" (version) VALUES ($1)", version); err != nil {
		return fmt.Errorf("record migration %s: %w", version, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit migration %s: %w", version, err)
	}

	return nil
}
