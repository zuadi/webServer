package models

import (
	"encoding/json"
	"net/http"
)

type Context struct {
	request    *http.Request
	response   http.ResponseWriter
	parameters map[string]string
}

func (c *Context) SetRequest(r *http.Request) {
	c.request = r
}
func (c *Context) SetResponseWriter(w http.ResponseWriter) {
	c.response = w
}

func (c *Context) SetParameters(p map[string]string) {
	c.parameters = p
}

func (c *Context) GetRequest() *http.Request {
	return c.request
}
func (c *Context) GetResponseWriter() http.ResponseWriter {
	return c.response
}

func (c *Context) GetParameter(key string) string {
	if value, ok := c.parameters[key]; ok {
		return value
	}
	return ""
}

func (c *Context) RespondString(s string) {
	c.response.Header().Set("Content-Type", "text/plain")
	c.response.Write([]byte(s))
}

func (c *Context) RespondJson(statusCode int, data any) {
	c.response.Header().Set("Content-Type", "application/json; charset=utf-8")

	c.response.WriteHeader(statusCode)

	if err := json.NewEncoder(c.response).Encode(data); err != nil {
		c.response.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(c.response).Encode(map[string]string{"error": err.Error()})
		return
	}
}
