package utils

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type RsSp1OnlyParsedOut struct {
	Type byte
	Val  interface{}
}
