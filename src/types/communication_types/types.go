package communication_types

import "time"

// From the 0 indexed list of songs presented to the user, which one will be chosen [0 - 2]
type SubmitVotePayload struct {
	SongOrder int
}

type SubmitSong struct {
	Name string
	Artist string
	SongURL string		
	Genre string
	Subgenre string
}

type SubmitSongSet struct {
	Description string
	Songs []SubmitSong
}

/*
Song Order -> Percentage of votes
*/
type TodaysRankingPayload struct {
	RankingMap map[int]float64
}

type MusicPayloadEntry struct{
	Title string
	Artist string
	PathResource string
	SongOrder int
}

type TodaysMusicPayload struct {
	CuratorName string
	CuratorDescription string
	MusicEntries []MusicPayloadEntry
}


type CalendarPayload struct {

}

type UserDTO struct{
	Username string
	CreationDate string
	Role string
}

type SongsUserVotedOnDTO struct{
	Song SubmitSong
	Date time.Time
}

