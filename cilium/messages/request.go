package messages

import ()

type PowerstripRequest struct {
	powerstripMessage
	Type          string
	ClientRequest ClientRequest
}

type ClientRequest struct {
	Method  string
	Request string
	Body    string
}
