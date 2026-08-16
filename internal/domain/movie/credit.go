package movie

import (
	"fmt"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Credit struct {
	Person Person
	Role   CreditRole
}
type CreditRole string

const (
	CreditRoleDirector        CreditRole = "director"
	CreditRoleWriter          CreditRole = "writer"
	CreditRoleCinematographer CreditRole = "cinematographer"
	CreditRoleComposer        CreditRole = "composer"
	CreditRoleCast            CreditRole = "cast"
)

func NewCredits(values []Credit) ([]Credit, error) {
	credits := make([]Credit, 0, len(values))

	for _, value := range values {
		name, err := NewPersonName(string(value.Person.Name))
		if err != nil {
			return nil, err
		}

		role, err := NewCreditRole(string(value.Role))
		if err != nil {
			return nil, err
		}

		credit := Credit{
			Person: Person{TMDBID: value.Person.TMDBID, Name: name},
			Role:   role,
		}

		for _, existing := range credits {
			if existing.sameAs(credit) {
				return nil, fmt.Errorf("%w: duplicate credit", exception.ErrInvalid)
			}
		}

		credits = append(credits, credit)
	}

	return credits, nil
}

func (c Credit) sameAs(other Credit) bool {
	if c.Role != other.Role {
		return false
	}

	if c.Person.TMDBID != nil && other.Person.TMDBID != nil {
		return *c.Person.TMDBID == *other.Person.TMDBID
	}

	return c.Person.Name == other.Person.Name
}

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
