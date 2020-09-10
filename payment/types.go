package payment

import "fmt"

type authoriseResponse struct {
	Authorisation Authorisation
	Err           error
}

type healthResponse struct {
	Health []Health `json:"health"`
}

type UnmarshalKeyError struct {
	Key  string
	JSON string
}

func (e *UnmarshalKeyError) Error() string {
	return fmt.Sprintf("Cannot unmarshal object key %q from JSON: %s", e.Key, e.JSON)
}
