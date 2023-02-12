package notify

type Request struct {
	Type        string `json:"type"`
	HTTPRequest string `json:"http_request"`
}