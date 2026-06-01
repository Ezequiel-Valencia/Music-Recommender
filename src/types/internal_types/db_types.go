package internal_types

type TodaysRankingSubmission struct {
	SongIDs     []int
	Description_Id int
	CuratorId int
}

type RankedSong struct {
	SongID    int
	CuratorId int
	NumVotes  int
}
