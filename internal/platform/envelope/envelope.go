package envelope

import (
	"encoding/json"
	"io"
)

type Meta struct {
	APIURL         string `json:"apiUrl"`
	Profile        string `json:"profile"`
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
	RequestID      string `json:"requestId,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type Envelope struct {
	Data  any        `json:"data,omitempty"`
	Meta  Meta       `json:"meta"`
	Error *ErrorBody `json:"error,omitempty"`
}

func WriteJSON(w io.Writer, env Envelope) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(env)
}
