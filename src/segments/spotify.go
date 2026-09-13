package segments

type Spotify struct {
	Base

	MusicPlayer
}

func (s *Spotify) Template() string {
	return " {{ .Icon }}{{ if ne .Status \"stopped\" }}{{ .Artist }} - {{ .Track }}{{ end }} "
}
