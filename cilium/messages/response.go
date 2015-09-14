package messages

import ()

type PowerstripResponse struct {
	powerstripMessage
}

type Response interface {
	GetPowerstripHookResponse() interface{}
}
