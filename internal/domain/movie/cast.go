package movie

type Cast struct {
	ID           CastID
	Name         CastName
	OriginalName OriginalCastName
	Character    Character
	Department   Department
	Gender       Gender
}

type CastID uint
type CastName string
type OriginalCastName string
type Character string
type Department string
type Gender string

const (
	GenderFemale Gender = "female"
	GenderMale   Gender = "male"
	GenderOther  Gender = "other"
)
