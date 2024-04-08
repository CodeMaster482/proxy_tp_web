package models

import "time"

type Header struct {
	Name  string
	Value string
}

type Request struct {
	Id          uint64              `json:"request_id"`
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	Host        string              `json:"host"`
	Get_Params  map[string][]string `json:"query"`
	Headers     map[string][]string `json:"headers"`
	Cookies     map[string]string   `json:"cookies"`
	Post_Params map[string][]string `json:"post_params"`
	Body        string              `json:"body"`
	CreatedAt   time.Time           `json:"created_at"`
}

type Response struct {
	Id        uint64              `json:"response_id"`
	RequestId uint64              `json:"request_id"`
	Code      int                 `json:"code"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
	CreatedAt time.Time           `json:"created_at"`
}

type ErrRequestNotFuound struct{}

func (e *ErrRequestNotFuound) Error() string {
	return "request not found"
}
