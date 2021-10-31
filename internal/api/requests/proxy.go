package requests

import (
	"encoding/json"
	"net/http"
)

type ProxyRequest struct {
	Address string `json:"address"`
}

func NewProxyRequest(r *http.Request) (*ProxyRequest, error) {
	var request ProxyRequest

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		return nil, err
	}

	return &request, nil
}