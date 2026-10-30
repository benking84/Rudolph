package handlers

import (
	"log"
	"net/http"

	"github.com/airbnb/rudolph/internal/handlers/eventupload"
	"github.com/airbnb/rudolph/internal/handlers/health"
	"github.com/airbnb/rudolph/internal/handlers/postflight"
	"github.com/airbnb/rudolph/internal/handlers/preflight"
	"github.com/airbnb/rudolph/internal/handlers/ruledownload"
	"github.com/airbnb/rudolph/internal/handlers/xsrf"
	"github.com/airbnb/rudolph/pkg/request"
	"github.com/airbnb/rudolph/pkg/response"
)

var (
	handlers []HandlerInterface
)

func init() {
	handlers = []HandlerInterface{
		&health.GetHealthHandler{},
		&eventupload.PostEventuploadHandler{},
		&preflight.PostPreflightHandler{},
		&ruledownload.PostRuledownloadHandler{},
		&postflight.PostPostflightHandler{},
		&xsrf.PostXSRFHandler{},
	}
}

func ApiRouter(request request.APIRequest) (*response.ProxyResponse, error) {
	log.Printf("api request %s %s", request.HTTPMethod, request.Resource)

	resp, err := getResponse(request)
	if err != nil {
		log.Printf("api error: %v", err)
	}
	if resp != nil {
		log.Printf("api response %d %s", resp.StatusCode, request.Resource)
	}
	return resp, err
}

func getResponse(request request.APIRequest) (*response.ProxyResponse, error) {
	for _, h := range handlers {
		if h.Handles(request) {
			if h.Boot() != nil {
				return response.APIResponse(http.StatusInternalServerError, nil)
			}
			return h.Handle(request)
		}
	}

	log.Printf("ERROR: ApiRouter failure: unrouteable request: [%+v] %+v", request.HTTPMethod, request.Resource)
	return response.APIResponse(http.StatusMethodNotAllowed, nil)
}

type HandlerInterface interface {
	Handles(request request.APIRequest) bool
	Boot() error
	Handle(request request.APIRequest) (*response.ProxyResponse, error)
}
