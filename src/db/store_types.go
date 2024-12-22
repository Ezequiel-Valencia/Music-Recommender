package db

import "time"

type MusicEntry struct {
	id          int
	name        string
	insert_date string
	songURL     string
	genre       string
	subgenre    string
	description string
	submitterID int
	rank_id     string
	num_ranks   int
}

type RankEntry struct {
	id          int
	date_ranked string
	ranking     string
}

type User struct {
	UserId          int
	Username        string
	CreationSource UserCreationSource
	CreationDate   time.Time
	UserRole       UserRoles
	UserPrivileges UserPrivileges
}
