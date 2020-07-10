# subkacker [![Go Report Card](https://goreportcard.com/badge/github.com/buckket/subkacker)](https://goreportcard.com/report/github.com/buckket/subkacker)  [![GoDoc](https://godoc.org/github.com/buckket/subkacker?status.svg)](https://godoc.org/github.com/buckketsubkacker)

**subkacker** is a tool designed to help in the production of so called [*YouTube Poops*](https://en.wikipedia.org/wiki/YouTube_Poop),
which make heavy use of *word remixing*.

YouTube automatically generates subtitle files for its videos. We can make use of this to easily search for words or
phonetic patterns, which we than can extract and remix.

This tool parses a given .srt file and stores every line in a SQLite3 database. The database can then be searched via a comfortable TUI:

![subkacker usage example][screenshot]

When *mpv* is installed the selected segment can also be previewed!

[screenshot]: subkacker.png "subkacker TUI"

## Installation

### From source

    go get -u github.com/buckket/subkacker

## Usage

1) Add subtitle file to database and reference related video:

        subkacker add "test_video.srt" "test_video.mp4"
        
2) Add as many subtitles/videos as you like
    
3) Start TUI and enter keyword in searchbar:

        subkacker search
        
4) By pressing Enter the selected segment can pe previewed via *mpv*
        
### Shortcuts

- \<TAB\> Switch between video selection and search bar
- \<Up\>/\<Down\> Scroll through results/videos
- \<Enter\> Play selected video segment via *mpv* OR enable/disable search for selected video when in video selection mode

## License

 GNU GPLv3+
 