package movie

import (
	"fmt"
	"unicode/utf8"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type Person struct {
	TMDBID PersonTMDBID
	Name   PersonName
}
type PersonTMDBID uint
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

func NewPerson(tmdbID PersonTMDBID, name PersonName) Person {
	return Person{TMDBID: tmdbID, Name: name}
}
