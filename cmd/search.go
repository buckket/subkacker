package cmd

import (
	"fmt"
	"github.com/buckket/subkacker/database"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func init() {
	rootCmd.AddCommand(searchCmd)
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search subtitles",
	Run:   search,
}

func search(cmd *cobra.Command, args []string) {
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

	app := tview.NewApplication()

	videos := tview.NewList()
	videos.ShowSecondaryText(false)
	videos.SetSelectedFocusOnly(true)
	videos.SetBorder(true).SetTitle("Videos")

	lines := tview.NewTable()
	lines.SetBorders(true)
	lines.SetBorder(true)
	lines.SetSelectable(true, false)
	lines.SetFixed(1, 4)
	lines.SetTitle("Lines")

	input := tview.NewInputField()
	input.SetLabel("Search: ")

	videos.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(input)
			return nil
		}
		return event
	})

	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			app.SetFocus(videos)
			return nil
		case tcell.KeyUp:
			row, _ := lines.GetSelection()
			if row > 0 {
				lines.Select(row-1, 0)
			}
			return nil
		case tcell.KeyDown:
			row, _ := lines.GetSelection()
			max := lines.GetRowCount()
			if row < max-1 {
				lines.Select(row+1, 0)
			}
			return nil
		case tcell.KeyEnter:
			row, _ := lines.GetSelection()
			video, ok := lines.GetCell(row, 0).GetReference().(int64)
			if !ok {
				return nil
			}
			dbVideo, err := db.VideoByID(video)
			if err != nil {
				log.Fatalf("could not get video from database: %s", err)
			}
			start := time.Duration(lines.GetCell(row, 1).GetReference().(int64)).Seconds()
			end := time.Duration(lines.GetCell(row, 2).GetReference().(int64)).Seconds()
			cmd := exec.Command("mpv", fmt.Sprintf("--start=%f", start), fmt.Sprintf("--ab-loop-a=%f", start), fmt.Sprintf("--ab-loop-b=%f", end), dbVideo.VideoFile)
			err = cmd.Run()
			if err != nil {
				log.Printf("could not play video file: %s", err)
			}
			return nil
		}
		return event
	})

	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(videos, 0, 2, false)
	flex.AddItem(lines, 0, 5, false)
	flex.AddItem(input, 1, 1, true)
	app.SetRoot(flex, true)

	videoIDs := make(map[int64]bool)
	searchFunc := func(text string) {
		lines.Clear()

		var activeVideos []int64
		for k, v := range videoIDs {
			if v {
				activeVideos = append(activeVideos, k)
			}
		}

		dbLines, err := db.SearchLine(activeVideos, text)
		if err != nil {
			log.Fatalf("could not search database: %s", err)
		}
		lines.SetCell(0, 0, &tview.TableCell{Text: "ID", Align: tview.AlignCenter, NotSelectable: true, Color: tcell.ColorYellow})
		lines.SetCell(0, 1, &tview.TableCell{Text: "[", Align: tview.AlignCenter, NotSelectable: true, Color: tcell.ColorYellow})
		lines.SetCell(0, 2, &tview.TableCell{Text: "]", Align: tview.AlignCenter, NotSelectable: true, Color: tcell.ColorYellow})
		lines.SetCell(0, 3, &tview.TableCell{Text: "Text", Align: tview.AlignCenter, NotSelectable: true, Color: tcell.ColorYellow, Expansion: 1})
		for row, line := range dbLines {
			lines.SetCell(row+1, 0, tview.NewTableCell(strconv.FormatInt(line.VideoID, 10)).SetReference(line.VideoID))
			lines.SetCell(row+1, 1, tview.NewTableCell(fmt.Sprintf("%s", time.Duration(line.StartAt))).SetReference(line.StartAt))
			lines.SetCell(row+1, 2, tview.NewTableCell(fmt.Sprintf("%s", time.Duration(line.EndAt))).SetReference(line.EndAt))
			highlightedText := strings.ReplaceAll(line.Text, text, fmt.Sprintf("[red]%s[-]", text))
			lines.SetCell(row+1, 3, tview.NewTableCell(highlightedText).SetExpansion(1).SetReference(line.Text))
		}
		lines.SetTitle(fmt.Sprintf("Lines (%d)", len(dbLines)))
		lines.Select(0, 0)
		lines.ScrollToBeginning()
	}
	input.SetChangedFunc(searchFunc)

	dbVideos, err := db.Videos()
	if err != nil {
		log.Fatalf("error while fetching videos from database: %s", err)
	}
	for _, video := range dbVideos {
		videos.AddItem(tview.Escape(fmt.Sprintf("[X] (%d) %s", video.ID, filepath.Base(video.VideoFile))), strconv.FormatInt(video.ID, 10), 0, func() {
			currItem := videos.GetCurrentItem()
			primary, secondary := videos.GetItemText(currItem)
			if strings.HasPrefix(primary, "[X[]") {
				primary = tview.Escape("[ ]" + primary[4:])
				videos.SetItemText(currItem, primary, secondary)
				vid, _ := strconv.ParseInt(secondary, 10, 64)
				videoIDs[vid] = false
			} else {
				primary = tview.Escape("[X]" + primary[4:])
				videos.SetItemText(currItem, primary, secondary)
				vid, _ := strconv.ParseInt(secondary, 10, 64)
				videoIDs[vid] = true
			}
			searchFunc(input.GetText())
		})
		videoIDs[video.ID] = true
		videos.SetTitle(fmt.Sprintf("Videos (%d)", len(dbVideos)))
	}

	searchFunc("")
	if err := app.Run(); err != nil {
		log.Fatalf("Error running application: %s\n", err)
	}
}
