package dto

type APIErrorResponse struct {
	Description      string   `json:"description"`
	Code             string   `json:"code"`
	ExceptionName    string   `json:"exceptionName"`
	ExceptionMessage string   `json:"exceptionMessage"`
	Stacktrace       []string `json:"stacktrace"`
}
