package streaming

type MusicWrap struct {
	SongName string `json:"name"`
	Artist   string `json:"artist"`
	Duration int    `json:"duration"`
}
