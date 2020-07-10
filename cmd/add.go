package cmd

import (
	"database/sql"
	"fmt"
	"github.com/asticode/go-astisub"
	"github.com/buckket/subkacker/database"
	"github.com/buckket/subkacker/models"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"path/filepath"
)

func init() {
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add [video file] [subtitle file]",
	Short: "Add new subtitles for video",
	Args:  cobra.ExactArgs(2),
	Run:   add,
}

func add(cmd *cobra.Command, args []string) {
	db, err := database.New(viper.GetString("DATABASE_FILE"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.AutoMigrate()
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateTableVideos()
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateTableLines()
	if err != nil {
		log.Fatal(err)
	}

	sub, err := astisub.OpenFile(args[1])
	if err != nil {
		log.Fatal(err)
	}

	video, err := db.VideoByVideoFile(args[0])
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatalf("error while fetching video from database: %s", err)
		}
		video = &models.Video{
			VideoFile:    args[0],
			SubtitleFile: args[1],
		}
		video.ID, err = db.InsertVideo(video)
		if err != nil {
			log.Fatalf("error while creating video in database: %s", err)
		}
	} else {
		// Video already exists, delete existing lines
		err := db.DeleteLines(video)
		if err != nil {
			log.Fatalf("error while deleting lines: %s", err)
		}
	}

	sub.Optimize()

	var lines []models.Line
	var lastLine string
	for _, item := range sub.Items {
		for _, line := range item.Lines {
			for _, li := range line.Items {
				if li.Text != lastLine && li.Text != "" {
					lastLine = li.Text
					lines = append(lines, models.Line{
						VideoID: video.ID,
						Text:    li.Text,
						StartAt: item.StartAt.Nanoseconds(),
						EndAt:   item.EndAt.Nanoseconds(),
					})
				}
			}
		}
	}

	for _, line := range lines {
		_, err := db.InsertLine(&line)
		if err != nil {
			log.Printf("error while add line to database: %s", err)
		}
	}

	fmt.Printf("Added %d lines for video '%s'\n", len(lines), filepath.Base(video.VideoFile))
}
