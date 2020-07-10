package models

type Line struct {
	ID      int64
	VideoID int64
	Text    string
	StartAt int64
	EndAt   int64
}

type Video struct {
	ID           int64
	VideoFile    string
	SubtitleFile string
}
