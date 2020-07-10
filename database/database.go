package database

import (
	"database/sql"
	"fmt"
	"github.com/buckket/subkacker/models"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	*sql.DB
}

func (db *Database) AutoMigrate() error {
	var version int
	err := db.QueryRow(`PRAGMA user_version`).Scan(&version)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) CreateTableLines() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS lines(
			id			INTEGER PRIMARY KEY AUTOINCREMENT,
			video_id	INTEGER NOT NULL,
			text		TEXT,
			start_at	INTEGER NOT NULL,
			end_at		INTEGER NOT NULL,
			FOREIGN KEY(video_id) REFERENCES videos(id)
		);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func (db *Database) CreateTableVideos() error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS videos(
			id 				INTEGER PRIMARY KEY AUTOINCREMENT,
			video_file		VARCHAR(255) UNIQUE,
			subtitle_file	VARCHAR(255) UNIQUE
		);
	`
	_, err := db.Exec(sqlStmt)
	return err
}

func (db *Database) InsertVideo(video *models.Video) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO videos(video_file, subtitle_file) VALUES(?, ?);`,
		video.VideoFile, video.SubtitleFile)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (db *Database) InsertLine(line *models.Line) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`INSERT INTO lines(video_id, text, start_at, end_at) VALUES(?, ?, ?, ?);`,
		line.VideoID, line.Text, line.StartAt, line.EndAt)
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return lastID, nil
}

func (db *Database) DeleteLines(video *models.Video) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`DELETE FROM lines WHERE video_id = ?;`, video.ID)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (db *Database) Videos() ([]models.Video, error) {
	var videos []models.Video

	rows, err := db.Query("SELECT id, video_file, subtitle_file FROM videos ORDER BY id ASC;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		v := models.Video{}
		err = rows.Scan(&v.ID, &v.VideoFile, &v.SubtitleFile)
		if err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}

	return videos, nil
}

func (db *Database) VideoByVideoFile(videoFile string) (*models.Video, error) {
	video := models.Video{}
	err := db.QueryRow("SELECT id, video_file, subtitle_file FROM videos WHERE video_file = ?", videoFile).Scan(
		&video.ID, &video.VideoFile, &video.SubtitleFile)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (db *Database) VideoByID(ID int64) (*models.Video, error) {
	video := models.Video{}
	err := db.QueryRow("SELECT id, video_file, subtitle_file FROM videos WHERE id = ?", ID).Scan(
		&video.ID, &video.VideoFile, &video.SubtitleFile)
	if err != nil {
		return nil, err
	}
	return &video, nil
}

func (db *Database) SearchLine(videos []int64, keyword string) ([]models.Line, error) {
	var lines []models.Line

	if keyword == "" {
		return lines, nil
	}

	for _, video := range videos {
		rows, err := db.Query("SELECT id, video_id, text, start_at, end_at FROM lines WHERE video_id = ? AND text LIKE ? ORDER BY video_id ASC;", video, "%"+keyword+"%")
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			l := models.Line{}
			err = rows.Scan(&l.ID, &l.VideoID, &l.Text, &l.StartAt, &l.EndAt)
			if err != nil {
				rows.Close()
				return nil, err
			}
			lines = append(lines, l)
		}
		rows.Close()
	}

	return lines, nil
}

func New(target string) (*Database, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?&_fk=true", target))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Database{db}, nil
}
