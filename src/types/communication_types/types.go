package communication_types


type SubmitVotePayload struct {
	SongNumber int
}

type SubmitSong struct {
	Name string
	Artist string
	SongURL string		
	Genre string
	Subgenre string
	Description string
}

/*
Song Order -> Percentage of votes
*/
type TodaysRankingPayload struct {
	RankingMap map[int]int
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

