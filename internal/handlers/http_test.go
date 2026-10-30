package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPHandlerUnknownRoute(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/not-a-sync-route", nil)
	rr := httptest.NewRecorder()
	HTTPHandler().ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestResourceFor(t *testing.T) {
	cases := map[string]string{
		"/health":          "/health",
		"/preflight/abc":   "/preflight/{machine_id}",
		"/postflight/abc/": "/postflight/{machine_id}",
		"/ruledownload/u":  "/ruledownload/{machine_id}",
		"/eventupload/u":   "/eventupload/{machine_id}",
		"/xsrf/u":          "/xsrf/{machine_id}",
	}
	for path, want := range cases {
		if got := resourceFor(path); got != want {
			t.Fatalf("%s: got %s want %s", path, got, want)
		}
	}
	params := pathParameters("/preflight/AAAAAAAA-A00A-1234-1234-5864377B4831")
	if params["machine_id"] != "AAAAAAAA-A00A-1234-1234-5864377B4831" {
		t.Fatalf("machine id %q", params["machine_id"])
	}
}
