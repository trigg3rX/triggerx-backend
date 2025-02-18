package services

// CustomError struct to represent an error response
type CustomError struct {
	Error  bool        `json:"error"`
	Message string     `json:"message"`
	Data   interface{} `json:"data"`
}

// NewCustomError creates a new instance of CustomError
func NewCustomError(message string, data interface{}) CustomError {
	return CustomError{
		Error:   true,
		Message: message,
		Data:    data,
	}
}

// CustomResponse struct to represent a successful response
type CustomResponse struct {
	Data   interface{} `json:"data"`
	Error  bool        `json:"error"`
	Message string     `json:"message"`
}

// NewCustomResponse creates a new instance of CustomResponse
func NewCustomResponse(data interface{}, message string) CustomResponse {
	return CustomResponse{
		Data:    data,
		Error:   false,
		Message: message,
	}
}