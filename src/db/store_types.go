package db

type MusicEntry struct {
	id int
	name string
	insert_date string
	songURL string
	genre string
	subgenre string
	description string
	submitterID int
	rank_id string
	num_ranks int
}


type RankEntry struct {
	id int
	date_ranked string
	ranking string
}


