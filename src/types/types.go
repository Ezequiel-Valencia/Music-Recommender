package types


type SubmitVotePayload struct {
	SongName string
}

type SubmitSong struct {
	Name string
	SongURL string		
	Genre string
	Subgenre string
	Description string
}

type TodaysRankingPayload struct {
	RankingMap map[string]int
}


type TodaysMusicPayload struct {

}


type CalendarPayload struct {

}

type UserDTO struct{
	Username string
	CreationDate string
	Role string
}

