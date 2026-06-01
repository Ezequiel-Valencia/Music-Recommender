package ranking_table

import (
	"database/sql"
	"music-recommender/config"
	"music-recommender/types/communication_types"
	"music-recommender/types/internal_types"
	"music-recommender/types/internal_types/auth_types"
	"time"

	"github.com/rs/zerolog/log"
)

type RankedDriver struct {
	dbPointer *sql.DB
}

func CreateRankedDriver(db *sql.DB) *RankedDriver {
	return &RankedDriver{dbPointer: db}
}

// Assumes the top song ID is accurate, and is given directly from Calculate Todays Rank
func (rd RankedDriver) InsertAlreadyRankedSongs(topSongId int, rankedSongs []internal_types.RankedSong) {
	voteDate := time.Now().AddDate(0, 0, -1)
	for _, rs := range rankedSongs {
		if _, err := rd.dbPointer.Exec(`INSERT INTO ranked(
			song_id, curator_id, date_ranked, num_votes, winner
		) VALUES($1, $2, $3, $4, $5)`, rs.SongID, rs.CuratorId,
			voteDate.Format(config.StaticEnvs.TimeFormat), rs.NumVotes, rs.SongID == topSongId); err != nil {
			log.Err(err).Msg("Failed to insert ranked song.")
		}
	}
}

func (mdb RankedDriver) GetSongsUserVotedFor(user auth_types.User) []communication_types.SongsUserVotedOnDTO {
	results, err := mdb.dbPointer.Query(`SELECT song_id, date FROM userVotes WHERE user_id = $1 
	ORDER BY "date"::date asc LIMIT 30`, user.UserId)
	if err != nil {
		log.Err(err).Msgf("Problem getting users %s songs they've voted on.", user.Username)
	}
	songSet := []communication_types.SongsUserVotedOnDTO{}
	for results.Next() {
		var songId int
		var date time.Time
		var name, artist, songURL, genre, subgenre string
		if err := results.Scan(&songId, &date); err != nil {
			log.Err(err).Msg("Failed to scan user vote.")
			continue
		}
		if err := mdb.dbPointer.QueryRow(`SELECT name, artist, songURL, genre, subgenre
		FROM music WHERE id = $1`, songId).Scan(&name, &artist, &songURL, &genre, &subgenre); err != nil {
			log.Err(err).Msg("Failed to scan song details.")
			continue
		}
		itemInSongSet := communication_types.SongsUserVotedOnDTO{Title: name, Artist: artist, SongURL: songURL, Date: date}
		songSet = append([]communication_types.SongsUserVotedOnDTO{itemInSongSet}, songSet...)
	}
	return songSet
}

// All the songs ranked, top songs id, error
func (rd TodaysRankingDriver) CalculateTodaysRank() ([]internal_types.RankedSong, int, error) {
	res, err := rd.db.Query("SELECT song_id, num_votes, curator_id FROM todaysRanking")
	if err != nil {
		log.Err(err).Msg("Can't compute todays ranking.")
		return nil, -1, err
	}
	rankedSongs := []internal_types.RankedSong{}
	topSongId := 1
	topVotes := -1
	for res.Next() {
		var song_id, numVotes, curator_id int

		if err := res.Scan(&song_id, &numVotes, &curator_id); err != nil {
			log.Err(err).Msg("Failed to scan ranked song.")
			continue
		}
		if numVotes > topVotes {
			topVotes = numVotes
			topSongId = song_id
		}
		rankedSongs = append(rankedSongs,
			internal_types.RankedSong{SongID: song_id, NumVotes: numVotes, CuratorId: curator_id},
		)
	}
	return rankedSongs, topSongId, nil
}

func (rd TodaysRankingDriver) CleanTodaysRanking() {
	_, err := rd.db.Exec("DELETE FROM todaysRanking")
	if err != nil {
		log.Err(err).Msg("Didn't clean todays ranking.")
	}
}

func (mdb TodaysRankingDriver) GetCalendarsMusic() *communication_types.CalendarPayload {
	return &communication_types.CalendarPayload{}
}
