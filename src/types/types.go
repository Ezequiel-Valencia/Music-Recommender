package types


type SubmitVotePayload struct {
	SongNumber int
}

type SubmitSong struct {
	Name string
	SongURL string		
	Genre string
	Subgenre string
	Description string
}

/*
Order -> Percentage of votes
*/
type TodaysRankingPayload struct {
	RankingMap map[int]int
}

type MusicPayloadEntry struct{
	Title string
	Artist string
	PathResource string
	Order int
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

