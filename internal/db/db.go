package db

import (
	"context"
	"rr/web/internal/models"

	"github.com/jackc/pgx/v5"
)

type Database struct {
	conn *pgx.Conn
}

func New(connStr string) (*Database, error) {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		return nil, err
	}

	if err := initSchema(conn); err != nil {
		conn.Close(context.Background())
		return nil, err
	}

	return &Database{conn: conn}, nil
}

func initSchema(conn *pgx.Conn) error {
	_, err := conn.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS playing_with_neon(
			id SERIAL PRIMARY KEY, 
			name TEXT NOT NULL, 
			value REAL
		);

		CREATE TABLE IF NOT EXISTS users(
			id SERIAL PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			name TEXT,
			picture TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		-- Insert sample data if table is empty
		INSERT INTO playing_with_neon(name, value) 
		SELECT LEFT(md5(i::TEXT), 10), random() 
		FROM generate_series(1, 10) s(i)
		WHERE NOT EXISTS (SELECT 1 FROM playing_with_neon LIMIT 1);
	`)
	return err
}

func (db *Database) Close() error {
	return db.conn.Close(context.Background())
}

func (db *Database) GetRecords(ctx context.Context) ([]models.Record, error) {
	rows, err := db.conn.Query(ctx, "SELECT id, name, value FROM playing_with_neon")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Record
	for rows.Next() {
		var r models.Record
		if err := rows.Scan(&r.ID, &r.Name, &r.Value); err != nil {
			return nil, err
		}
		records = append(records, r)
	}

	return records, rows.Err()
}

func (db *Database) UpsertUser(ctx context.Context, user *models.User) error {
	_, err := db.conn.Exec(ctx, `
		INSERT INTO users (email, name, picture)
		VALUES ($1, $2, $3)
		ON CONFLICT (email) 
		DO UPDATE SET name = $2, picture = $3
	`, user.Email, user.Name, user.Picture)
	return err
}
//
// In db/db.go
func (db *Database) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := db.conn.QueryRow(ctx, `
        SELECT email, name, picture 
        FROM users 
        WHERE email = $1
    `, email).Scan(&user.Email, &user.Name, &user.Picture)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
