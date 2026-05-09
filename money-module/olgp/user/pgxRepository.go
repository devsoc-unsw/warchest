package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgxRepository struct {
	db *pgxpool.Pool
}


func NewPgxRepository(db* pgxpool.Pool) *PgxRepository {
	return &PgxRepository{
		db: db,
	}
}

func (r *PgxRepository) Migrate(ctx context.Context) error {
	query := `
    CREATE TABLE IF NOT EXISTS users(
        id SERIAL PRIMARY KEY,
    );
    `
	_, err := r.db.Exec(ctx, query)
	return err
}
