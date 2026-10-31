package handlers

import (
	"compress/gzip"
	"compress/zlib"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/airbnb/rudolph/pkg/request"
	"github.com/airbnb/rudolph/pkg/response"
)

const maxRequestBodyBytes = 10 << 20 // 10 MiB

// Boot initialises every handler once, before the process accepts traffic.
func Boot() error {
	bootMu.Lock()
	defer bootMu.Unlock()
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
		defer r.Body.Close()

		resource := resourceFor(r.URL.Path)
		if os.Getenv("REQUIRE_CLIENT_CERT") == "true" && resource != "/health" {
			if !strings.EqualFold(r.Header.Get("X-Client-Cert-Chain-Verified"), "true") {
				writeProxy(w, &response.ProxyResponse{
					StatusCode: http.StatusUnauthorized,
					Headers:    map[string]string{"Content-Type": "application/json"},
					Body:       `{"error":"client certificate required"}`,
				})
				return
			}
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		body, err := readBoundedBody(r)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
				return
			}
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		req := request.APIRequest{
			HTTPMethod:     r.Method,
			Path:           r.URL.Path,
			Resource:       resource,
			PathParameters: pathParameters(r.URL.Path),
			Headers:        flattenHeaders(r.Header),
			Body:           string(body),
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

func readBoundedBody(r *http.Request) ([]byte, error) {
	encoding := strings.ToLower(strings.TrimSpace(r.Header.Get("Content-Encoding")))
	var reader io.Reader = r.Body
	switch encoding {
	case "", "identity":
	case "deflate", "zlib":
		zr, err := zlib.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer zr.Close()
		reader = io.LimitReader(zr, maxRequestBodyBytes+1)
	case "gzip":
		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		reader = io.LimitReader(gr, maxRequestBodyBytes+1)
	default:
		return nil, errors.New("unsupported content-encoding")
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxRequestBodyBytes {
		return nil, &http.MaxBytesError{Limit: maxRequestBodyBytes}
	}
	return data, nil
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
	if path == "/health" {
		return "/health"
	}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) == 2 && parts[1] != "" {
		switch parts[0] {
		case "preflight":
			return "/preflight/{machine_id}"
		case "eventupload":
			return "/eventupload/{machine_id}"
		case "ruledownload":
			return "/ruledownload/{machine_id}"
		case "postflight":
			return "/postflight/{machine_id}"
		case "xsrf":
			return "/xsrf/{machine_id}"
		}
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
