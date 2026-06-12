// File: sdk/queue/sqlite_store.go
// Purpose: SQLite embedded production implementation of EventStore

package queue

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"meshguard/sdk/types"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements EventStore using embedded SQLite
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore opens or creates the database at the given path
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return store, nil
}

func (s *SQLiteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		status TEXT NOT NULL,
		from_node TEXT NOT NULL,
		to_node TEXT NOT NULL,
		amount_sats INTEGER NOT NULL,
		channel_id TEXT,
		sequence INTEGER NOT NULL UNIQUE,
		timestamp DATETIME NOT NULL,
		payload BLOB,
		signature BLOB,
		htlc_hash TEXT,
		invoice TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_status ON events(status);
	CREATE INDEX IF NOT EXISTS idx_sequence ON events(sequence);
	CREATE INDEX IF NOT EXISTS idx_created ON events(created_at);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *SQLiteStore) Create(ctx context.Context, event *types.MeshGuardEvent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO events (id, type, status, from_node, to_node, amount_sats, channel_id,
			sequence, timestamp, payload, signature, htlc_hash, invoice, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.Type, event.Status, event.FromNode, event.ToNode, event.AmountSats,
		event.ChannelID, event.Sequence, event.Timestamp, event.Payload, event.Signature,
		event.HTLCHash, event.Invoice, event.CreatedAt, event.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Get(ctx context.Context, id string) (*types.MeshGuardEvent, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, type, status, from_node, to_node, amount_sats, channel_id,
			sequence, timestamp, payload, signature, htlc_hash, invoice, created_at, updated_at
		FROM events WHERE id = ?
	`, id)
	return s.scanEvent(row)
}

func (s *SQLiteStore) ListByStatus(ctx context.Context, status types.EventStatus) ([]*types.MeshGuardEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, status, from_node, to_node, amount_sats, channel_id,
			sequence, timestamp, payload, signature, htlc_hash, invoice, created_at, updated_at
		FROM events WHERE status = ? ORDER BY sequence ASC
	`, status)
	if err != nil {
		return nil, fmt.Errorf("query by status: %w", err)
	}
	defer rows.Close()
	return s.scanEvents(rows)
}

func (s *SQLiteStore) ListAll(ctx context.Context, limit int) ([]*types.MeshGuardEvent, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, type, status, from_node, to_node, amount_sats, channel_id,
			sequence, timestamp, payload, signature, htlc_hash, invoice, created_at, updated_at
		FROM events ORDER BY sequence DESC LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query all: %w", err)
	}
	defer rows.Close()
	return s.scanEvents(rows)
}

func (s *SQLiteStore) CountByStatus(ctx context.Context) (map[types.EventStatus]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT status, COUNT(*) FROM events GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("count status: %w", err)
	}
	defer rows.Close()

	counts := make(map[types.EventStatus]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[types.EventStatus(status)] = count
	}
	return counts, nil
}

func (s *SQLiteStore) Update(ctx context.Context, event *types.MeshGuardEvent) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE events SET
			type = ?, status = ?, from_node = ?, to_node = ?, amount_sats = ?,
			channel_id = ?, sequence = ?, timestamp = ?, payload = ?, signature = ?,
			htlc_hash = ?, invoice = ?, updated_at = ?
		WHERE id = ?
	`, event.Type, event.Status, event.FromNode, event.ToNode, event.AmountSats,
		event.ChannelID, event.Sequence, event.Timestamp, event.Payload, event.Signature,
		event.HTLCHash, event.Invoice, time.Now(), event.ID)
	if err != nil {
		return fmt.Errorf("update event: %w", err)
	}
	return nil
}

func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE id = ?`, id)
	return err
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) scanEvent(row *sql.Row) (*types.MeshGuardEvent, error) {
	var e types.MeshGuardEvent
	var ts, created, updated string

	err := row.Scan(
		&e.ID, &e.Type, &e.Status, &e.FromNode, &e.ToNode, &e.AmountSats,
		&e.ChannelID, &e.Sequence, &ts, &e.Payload, &e.Signature,
		&e.HTLCHash, &e.Invoice, &created, &updated,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("scan event: %w", err)
	}

	e.Timestamp, _ = time.Parse(time.RFC3339, ts)
	e.CreatedAt, _ = time.Parse(time.RFC3339, created)
	e.UpdatedAt, _ = time.Parse(time.RFC3339, updated)

	return &e, nil
}

func (s *SQLiteStore) scanEvents(rows *sql.Rows) ([]*types.MeshGuardEvent, error) {
	var events []*types.MeshGuardEvent
	for rows.Next() {
		var e types.MeshGuardEvent
		var ts, created, updated string

		err := rows.Scan(
			&e.ID, &e.Type, &e.Status, &e.FromNode, &e.ToNode, &e.AmountSats,
			&e.ChannelID, &e.Sequence, &ts, &e.Payload, &e.Signature,
			&e.HTLCHash, &e.Invoice, &created, &updated,
		)
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		e.Timestamp, _ = time.Parse(time.RFC3339, ts)
		e.CreatedAt, _ = time.Parse(time.RFC3339, created)
		e.UpdatedAt, _ = time.Parse(time.RFC3339, updated)

		events = append(events, &e)
	}
	return events, nil
}
