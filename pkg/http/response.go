package http

import (
	"io"
	"net/http"
	"proxy/internal/models"
	"strings"
)

func ParseResponse(r http.Response) models.Response {
	ri := models.Response{
		Code: r.StatusCode,
	}

	headers := make(http.Header)
	for k, values := range r.Header {
		headers[k] = append(headers[k], values...)
	}
	ri.Headers = headers

	body := &strings.Builder{}
	defer r.Body.Close()
	if _, err := io.Copy(body, r.Body); err == nil {
		ri.Body = body.String()
	}

	return ri
}
