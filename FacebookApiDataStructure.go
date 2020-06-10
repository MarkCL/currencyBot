package main

type Button struct {
	Type  string `json:"type,omitempty"`
	URL   string `json:"url,omitempty"`
	Title string `json:"title,omitempty"`
}

type DefaultAction struct {
	Type                string `json:"type,omitempty"`
	URL                 string `json:"url,omitempty"`
	MessengerExtensions bool   `json:"messenger_extensions,omitempty"`
	WebviewHeightRatio  string `json:"webview_height_ratio,omitempty"`
	FallbackURL         string `json:"fallback_url,omitempty"`
}

type Element struct {
	Title         string        `json:"title,omitempty"`
	Subtitle      string        `json:"subtitle,omitempty"`
	ImageURL      string        `json:"image_url,omitempty"`
	DefaultAction DefaultAction `json:"default_action,omitempty"`
	Buttons       []Button      `json:"buttons,omitempty"`
}

type Payload struct {
	URL               string    `json:"url,omitempty"`
	TemplateType      string    `json:"template_type,omitempty"`
	Sharable          bool      `json:"sharable,omitempty"`
	ImageAspectRation string    `json:"image_aspect_ratio,omitempty"`
	Elements          []Element `json:"elements,omitempty"`
}

type Attachment struct {
	Type    string  `json:"type,omitempty"`
	Payload Payload `json:"payload,omitempty"`
}

type Message struct {
	Text       string `json:"text,omitempty"`
	QuickReply *struct {
		Payload string `json:"payload,omitempty"`
	} `json:"quick_reply,omitempty"`
	Attachment  *Attachment   `json:"attachment,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

type User struct {
	ID string `json:"id,omitempty"`
}

type Response struct {
	Recipient User    `json:"recipient,omitempty"`
	Message   Message `json:"message,omitempty"`
}

type Messaging struct {
	Sender    User    `json:"sender,omitempty"`
	Recipient User    `json:"recipient,omitempty"`
	Timestamp int     `json:"timestamp,omitempty"`
	Message   Message `json:"message,omitempty"`
}

type Entry struct {
	ID        string      `json:"id,omitempty"`
	Time      int         `json:"time,omitempty"`
	Messaging []Messaging `json:"messaging,omitempty"`
}

type Callback struct {
	Object string  `json:"object,omitempty"`
	Entry  []Entry `json:"entry,omitempty"`
}
