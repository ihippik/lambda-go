package lambda

import "encoding/json"

type Error struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func NewError(status int, message string) *Error {
	return &Error{Status: status, Message: message}
}

func (e Error) Error() []byte {
	data, err := json.Marshal(e)
	if err != nil {
		return []byte("fatal error")
	}

	return data
}
