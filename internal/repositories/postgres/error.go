package postgres

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	PgCodeUniqueViolation          = "23505"
	PgCodeCheckConstraintViolation = "23514"
)

var (
	ErrDatabase = errors.New("database error")
)

type Error struct {
	msg string
	Err error
}

func (e Error) Unwrap() error { return e.Err }

func (e Error) Error() string {
	if e.NotFound() {
		return "not found"
	}

	if e.Duplicate() {
		return "duplicate record"
	}

	if e.msg == "" {
		return ErrDatabase.Error()
	}
	return e.msg
}

func (e Error) NotFound() bool {
	return errors.Is(e.Err, gorm.ErrRecordNotFound)
}

func (e Error) Duplicate() bool {
	if e.Err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if ok := errors.As(e.Err, &pgErr); ok {
		if pgErr.Code == PgCodeUniqueViolation {
			return true
		}
	}
	return false
}
