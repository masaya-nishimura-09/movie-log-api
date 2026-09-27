package movie

type Credit struct {
	PersonName PersonName
	CreditRole CreditRole
}

type PersonName string
type CreditRole string

const (
	CreditRoleDirector        CreditRole = "director"
	CreditRoleWriter          CreditRole = "writer"
	CreditRoleCinematographer CreditRole = "cinematographer"
	CreditRoleComposer        CreditRole = "composer"
	CreditRoleCast            CreditRole = "cast"
)
