package record

import (
	"fmt"
	"slices"

	"github.com/masaya-nishimura-09/movie-log-api/internal/domain/exception"
)

type MoodTag string

const (
	MoodTagMoving       MoodTag = "moving"
	MoodTagDark         MoodTag = "dark"
	MoodTagStylish      MoodTag = "stylish"
	MoodTagHardboiled   MoodTag = "hardboiled"
	MoodTagSurreal      MoodTag = "surreal"
	MoodTagPoignant     MoodTag = "poignant"
	MoodTagRefreshing   MoodTag = "refreshing"
	MoodTagNostalgic    MoodTag = "nostalgic"
	MoodTagEpic         MoodTag = "epic"
	MoodTagTense        MoodTag = "tense"
	MoodTagCozy         MoodTag = "cozy"
	MoodTagMinimal      MoodTag = "minimal"
	MoodTagExperimental MoodTag = "experimental"
	MoodTagRomantic     MoodTag = "romantic"
	MoodTagDisturbing   MoodTag = "disturbing"
	MoodTagHeartwarming MoodTag = "heartwarming"
	MoodTagBitterEnding MoodTag = "bitter_ending"
	MoodTagAbsurd       MoodTag = "absurd"
)

func NewMoodTag(value string) (MoodTag, error) {
	if value == "" {
		return "", fmt.Errorf("%w: mood tag is required", exception.ErrInvalid)
	}

	switch moodTag := MoodTag(value); moodTag {
	case MoodTagMoving,
		MoodTagDark,
		MoodTagStylish,
		MoodTagHardboiled,
		MoodTagSurreal,
		MoodTagPoignant,
		MoodTagRefreshing,
		MoodTagNostalgic,
		MoodTagEpic,
		MoodTagTense,
		MoodTagCozy,
		MoodTagMinimal,
		MoodTagExperimental,
		MoodTagRomantic,
		MoodTagDisturbing,
		MoodTagHeartwarming,
		MoodTagBitterEnding,
		MoodTagAbsurd:
		return moodTag, nil
	default:
		return "", fmt.Errorf("%w: invalid mood tag", exception.ErrInvalid)
	}
}

func NewMoodTags(values []string) ([]MoodTag, error) {
	moodTags := make([]MoodTag, 0, len(values))

	for _, value := range values {
		moodTag, err := NewMoodTag(value)
		if err != nil {
			return nil, err
		}

		if slices.Contains(moodTags, moodTag) {
			return nil, fmt.Errorf("%w: duplicate mood tag", exception.ErrInvalid)
		}

		moodTags = append(moodTags, moodTag)
	}

	return moodTags, nil
}
