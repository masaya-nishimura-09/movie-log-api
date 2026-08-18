package record

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Credit struct {
	PersonName PersonName
	CreditRole CreditRole
}

type PersonName string

func NewPersonName(value string) (PersonName, error) {
	if value == "" {
		return "", fmt.Errorf("%w: person name is required", exception.ErrInvalid)
	}

	if utf8.RuneCountInString(value) > 100 {
		return "", fmt.Errorf("%w: person name must be at most 100 characters", exception.ErrInvalid)
	}

	return PersonName(value), nil
}

type CreditRole string

const (
	CreditRoleDirector        CreditRole = "director"
	CreditRoleWriter          CreditRole = "writer"
	CreditRoleCinematographer CreditRole = "cinematographer"
	CreditRoleComposer        CreditRole = "composer"
	CreditRoleCast            CreditRole = "cast"
)

func NewCreditRole(value string) (CreditRole, error) {
	if value == "" {
		return "", fmt.Errorf("%w: credit role is required", exception.ErrInvalid)
	}

	switch creditRole := CreditRole(value); creditRole {
	case CreditRoleDirector,
		CreditRoleWriter,
		CreditRoleCinematographer,
		CreditRoleComposer,
		CreditRoleCast:
		return creditRole, nil
	default:
		return "", fmt.Errorf("%w: invalid credit role", exception.ErrInvalid)
	}
}

func NewCredit(personName PersonName, creditRole CreditRole) Credit {
	return Credit{PersonName: personName, CreditRole: creditRole}
}
