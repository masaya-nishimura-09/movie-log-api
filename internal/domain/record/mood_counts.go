package record

type MoodCount uint

type MoodTagCount struct {
	Tag   MoodTag
	Count MoodCount
}

type MoodTagCounts []MoodTagCount
