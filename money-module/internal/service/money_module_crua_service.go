package service

import (
	"context"
	"fmt"
	db "money-module/db/postgres/generated"

	"github.com/jackc/pgx/v5/pgxpool"
)


type BudgetService interface {
	CreateBudget()
	ReadBudget()
	UpdateBudget()
	ArchiveBudget()
}

type Service struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, q: db.New(pool)}
}

type NewAccount struct {
	id uint32;
}

func (s *Service) CreateAccount(ctx context.Context, in NewAccount) (int32, error) {
	// tx is a database transaction
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	q := s.q.WithTx(tx)

	soc, err := q.CreateAccount(ctx,
		db.CreateSocietyParams{Name: in.Name, Slug: in.Slug})
	if pgerr.IsUniqueViolation(err) {
		return 0, ErrSlugTaken
	}
	if err != nil {
		return 0, fmt.Errorf("create society: %w", err)
	}


	err = tx.Commit(ctx)
	if err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}

	return soc.ID, nil
}
