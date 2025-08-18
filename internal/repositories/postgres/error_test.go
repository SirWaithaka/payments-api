package postgres

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestError_Unwrap(t *testing.T) {
	pgErr := &pgconn.PgError{Code: PgCodeCheckConstraintViolation, Message: "violates check constraint"}

	err := Error{Err: pgErr}
	pe := &pgconn.PgError{}
	if !errors.As(err, &pe) {
		t.Errorf("expected pgconn.PgError, got %T", pe)
	}
}

func TestError_NotFound(t *testing.T) {

	e := Error{Err: gorm.ErrRecordNotFound}
	assert.Equal(t, e.NotFound(), true)

	e = Error{Err: ErrNotFound}
	assert.Equal(t, e.NotFound(), true)

	e = Error{Err: nil}
	assert.Equal(t, e.NotFound(), false)
}

func TestError_Duplicate(t *testing.T) {
	pe := &pgconn.PgError{Code: PgCodeUniqueViolation}
	e := Error{Err: pe}
	assert.Equal(t, e.Duplicate(), true)

	e = Error{Err: gorm.ErrDuplicatedKey}
	assert.Equal(t, e.Duplicate(), true)

	e = Error{Err: ErrDuplicate}
	assert.Equal(t, e.Duplicate(), true)

	e = Error{Err: nil}
	assert.Equal(t, e.Duplicate(), false)
}

func TestError_Error(t *testing.T) {
	e := Error{}
	assert.Equal(t, e.Error(), ErrDatabase.Error())

	e = Error{Err: ErrNotFound}
	assert.Equal(t, e.Error(), ErrNotFound.Error())

	e = Error{Err: ErrDuplicate}
	assert.Equal(t, e.Error(), ErrDuplicate.Error())

	e = Error{Err: nil}
	assert.Equal(t, e.Error(), ErrDatabase.Error())

	e = Error{msg: "test error"}
	assert.Equal(t, e.Error(), "test error")
}
