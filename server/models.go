package server

// MathRequest is the standard request struct (and its various encodings)
type MathRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// MathOKResponse is returned to the client after a request is properly handled (without errors)
type MathOKResponse struct {
	Action string  `json:"action"`
	X      float64 `json:"x"` // in case our client gets any big ideas
	Y      float64 `json:"y"`
	Answer float64 `json:"answer"`
	Cached bool    `json:"cached"`
}

// MathErrorResponse is returned to the client if there was an error handling their request
type MathErrorResponse struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}
