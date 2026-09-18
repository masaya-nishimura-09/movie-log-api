package movie

type Cast struct {
	ID           CastID
	Name         CastName
	OriginalName OriginalCastName
	Character    Character
	Role         Role
	Gender       Gender
}

type CastID uint
type CastName string
type OriginalCastName string
type Character string
type Role string
