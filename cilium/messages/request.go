package messages

import (
	"encoding/json"
)

type PowerstripRequest struct {
	powerstripMessage
	Type          string
	ClientRequest ClientRequest
}

type ClientRequest struct {
	Method     string
	Request    string
	Body       string
}

type ServerResponse struct {
	ContentType string
	Body        string
	Code        int
}

func (sr ServerResponse) ConvertTo(i interface{}) error {
	return json.Unmarshal([]byte(sr.Body), i)
}
