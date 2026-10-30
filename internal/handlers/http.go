package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/airbnb/rudolph/pkg/request"
	"github.com/airbnb/rudolph/pkg/response"
)

// Boot initialises every handler once, before the process accepts traffic.
func Boot() error {
	for _, h := range handlers {
		if err := h.Boot(); err != nil {
			return err
		}
	}
	return nil
}

// HTTPHandler serves the Santa sync API.
func HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		req := request.APIRequest{
			HTTPMethod:     r.Method,
			Path:           r.URL.Path,
			Resource:       resourceFor(r.URL.Path),
			PathParameters: pathParameters(r.URL.Path),
			Headers:        flattenHeaders(r.Header),
			Body:           string(body),
		}
		if os.Getenv("REQUIRE_CLIENT_CERT") == "true" && req.Resource != "/health" {
			if !strings.EqualFold(r.Header.Get("X-Client-Cert-Chain-Verified"), "true") {
				writeProxy(w, &response.ProxyResponse{
					StatusCode: http.StatusUnauthorized,
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"error":"client certificate required"}`,
				})
				return
			}
		}
		resp, err := ApiRouter(req)
		if err != nil {
			log.Printf("handler error: %v", err)
		}
		if resp == nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		writeProxy(w, resp)
	})
}

func writeProxy(w http.ResponseWriter, resp *response.ProxyResponse) {
	for k, v := range resp.Headers {
		w.Header().Set(k, v)
	}
	if resp.StatusCode == 0 {
		resp.StatusCode = http.StatusOK
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write([]byte(resp.Body))
}

func resourceFor(path string) string {
	path = strings.TrimSuffix(path, "/")
	switch {
	case path == "/health" || path == "":
		if path == "/health" {
			return "/health"
		}
	case strings.HasPrefix(path, "/preflight/"):
		return "/preflight/{machine_id}"
	case strings.HasPrefix(path, "/eventupload/"):
		return "/eventupload/{machine_id}"
	case strings.HasPrefix(path, "/ruledownload/"):
		return "/ruledownload/{machine_id}"
	case strings.HasPrefix(path, "/postflight/"):
		return "/postflight/{machine_id}"
	case strings.HasPrefix(path, "/xsrf/"):
		return "/xsrf/{machine_id}"
	}
	return path
}

func pathParameters(path string) map[string]string {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 2 {
		return map[string]string{"machine_id": parts[1]}
	}
	return map[string]string{}
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[k] = v[0]
			out[strings.ToLower(k)] = v[0]
		}
	}
	return out
}
