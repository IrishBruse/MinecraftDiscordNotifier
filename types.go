package main

type DiscordWebhook struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds"`
}

type Embed struct {
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Color       int64   `json:"color"`
	Fields      []Field `json:"fields"`
	Thumbnail   Image   `json:"thumbnail"`
	Image       Image   `json:"image"`
	Footer      Footer  `json:"footer"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline *bool  `json:"inline,omitempty"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type Image struct {
	URL string `json:"url"`
}

//

type Status struct {
	Version            Version `json:"version"`
	EnforcesSecureChat bool    `json:"enforcesSecureChat"`
	Description        string  `json:"description"`
	Players            Players `json:"players"`
}

type Players struct {
	Max    int64    `json:"max"`
	Online int64    `json:"online"`
	Sample []Player `json:"sample"`
}

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Version struct {
	Name     string `json:"name"`
	Protocol int64  `json:"protocol"`
}
