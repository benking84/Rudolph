package response

import "encoding/json"

// ProxyResponse is the HTTP response produced by a sync handler.
type ProxyResponse struct {
	Headers    map[string]string
	StatusCode int
	Body       string
}

// APIResponse constructs a JSON response. A body that cannot be encoded
// becomes a 500 with a fixed error string.
func APIResponse(status int, rawBody interface{}) (*ProxyResponse, error) {
	resp := &ProxyResponse{
		Headers:    map[string]string{"Content-Type": "application/json"},
		StatusCode: status,
	}

	body, err := json.Marshal(rawBody)
	if err != nil {
		resp.StatusCode = 500
		resp.Body = "value could not be serialized to JSON"
		return resp, err
	}

	resp.Body = string(body)
	return resp, nil
}
