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
	ErrDatabase  = errors.New("database error")
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("duplicate record")
)

type Error struct {
	msg string
	Err error
}

func (e Error) Unwrap() error {
	return e.Err
}

func (e Error) Error() string {
	if e.NotFound() {
		return ErrNotFound.Error()
	}

	if e.Duplicate() {
		return ErrDuplicate.Error()
	}

	if e.msg == "" {
		return ErrDatabase.Error()
	}
	return e.msg
}

func (e Error) NotFound() bool {
	if errors.Is(e.Err, gorm.ErrRecordNotFound) {
		return true
	}

	if errors.Is(e.Err, ErrNotFound) {
		return true
	}

	return false
}

func (e Error) Duplicate() bool {
	if e.Err == nil {
		return false
	}

	if errors.Is(e.Err, gorm.ErrDuplicatedKey) {
		return true
	}

	if errors.Is(e.Err, ErrDuplicate) {
		return true
	}

	var pgErr *pgconn.PgError
	if ok := errors.As(e.Err, &pgErr); ok {
		if pgErr.Code == PgCodeUniqueViolation {
			return true
		}
	}
	return false
}
