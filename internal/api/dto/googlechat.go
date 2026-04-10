package dto

type GoogleChatMessage struct {
	Text  string `json:"text,omitempty"`
	Cards []Card `json:"cards,omitempty"`
}

type Card struct {
	Header   *CardHeader   `json:"header,omitempty"`
	Sections []CardSection `json:"sections,omitempty"`
}

type CardHeader struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle,omitempty"`
}

type CardSection struct {
	Header  string   `json:"header,omitempty"`
	Widgets []Widget `json:"widgets,omitempty"`
}

type Widget struct {
	DecoratedText *DecoratedTextWidget `json:"decoratedText,omitempty"`
	ButtonList    *ButtonListWidget    `json:"buttonList,omitempty"`
}

type DecoratedTextWidget struct {
	TopLabel string `json:"topLabel,omitempty"`
	Text     string `json:"text"`
}

type ButtonListWidget struct {
	Buttons []Button `json:"buttons"`
}

type Button struct {
	Text    string   `json:"text"`
	OnClick *OnClick `json:"onClick"`
}

type OnClick struct {
	OpenLink *OpenLink `json:"openLink,omitempty"`
}

type OpenLink struct {
	URL string `json:"url"`
}
