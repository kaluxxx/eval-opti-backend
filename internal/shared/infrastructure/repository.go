package infrastructure

import (
	"context"
	"database/sql"
)

// QueryRepository interface de base pour les opérations de lecture
type QueryRepository interface {
	// WithContext permet d'ajouter un contexte pour l'annulation/timeout
	WithContext(ctx context.Context) QueryRepository
}

// CommandRepository interface de base pour les opérations d'écriture
type CommandRepository interface {
	// WithContext permet d'ajouter un contexte
	WithContext(ctx context.Context) CommandRepository
	// WithTx permet d'exécuter dans une transaction
	WithTx(tx *sql.Tx) CommandRepository
}

// UnitOfWork gère les transactions pour les opérations d'écriture
type UnitOfWork interface {
	Begin() (*sql.Tx, error)
	Commit(tx *sql.Tx) error
	Rollback(tx *sql.Tx) error
	Execute(fn func(tx *sql.Tx) error) error
}

// DBUnitOfWork implémentation de UnitOfWork avec sql.DB
type DBUnitOfWork struct {
	db *sql.DB
}

// NewUnitOfWork crée une nouvelle instance de UnitOfWork
func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &DBUnitOfWork{db: db}
}

// Begin démarre une transaction
func (uow *DBUnitOfWork) Begin() (*sql.Tx, error) {
	return uow.db.Begin()
}

// Commit valide une transaction
func (uow *DBUnitOfWork) Commit(tx *sql.Tx) error {
	return tx.Commit()
}

// Rollback annule une transaction
func (uow *DBUnitOfWork) Rollback(tx *sql.Tx) error {
	return tx.Rollback()
}

// Execute exécute une fonction dans une transaction
func (uow *DBUnitOfWork) Execute(fn func(tx *sql.Tx) error) error {
	tx, err := uow.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = uow.Rollback(tx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := uow.Rollback(tx); rbErr != nil {
			return rbErr
		}
		return err
	}

	return uow.Commit(tx)
}

// Specification pattern pour les requêtes complexes
type Specification interface {
	ToSQL() (string, []interface{})
}

// BaseRepository structure de base pour les repositories
type BaseRepository struct {
	db  *sql.DB
	tx  *sql.Tx
	ctx context.Context
}

// NewBaseRepository crée un nouveau repository de base
func NewBaseRepository(db *sql.DB) BaseRepository {
	return BaseRepository{
		db:  db,
		ctx: context.Background(),
	}
}

// Executor retourne l'exécuteur approprié (DB ou Tx)
func (r *BaseRepository) Executor() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// Context retourne le contexte actuel
func (r *BaseRepository) Context() context.Context {
	return r.ctx
}

// Query exécute une requête de lecture
func (r *BaseRepository) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return r.Executor().QueryContext(r.ctx, query, args...)
}

// QueryRow exécute une requête de lecture pour une seule ligne
func (r *BaseRepository) QueryRow(query string, args ...interface{}) *sql.Row {
	return r.Executor().QueryRowContext(r.ctx, query, args...)
}

// Exec exécute une requête d'écriture
func (r *BaseRepository) Exec(query string, args ...interface{}) (sql.Result, error) {
	return r.Executor().ExecContext(r.ctx, query, args...)
}
