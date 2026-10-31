package eventupload

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/airbnb/rudolph/pkg/eventsink"
	"github.com/airbnb/rudolph/pkg/request"
	"github.com/airbnb/rudolph/pkg/response"
)

type PostEventuploadHandler struct {
	booted    bool
	publisher eventsink.Publisher
	enabled   bool
}

func (h *PostEventuploadHandler) Boot() (err error) {
	if h.booted {
		return
	}
	publisher, err := eventsink.FromEnv(context.Background())
	if err != nil {
		return err
	}
	h.publisher = publisher
	_, disabled := publisher.(eventsink.Nop)
	h.enabled = !disabled
	h.booted = true
	return
}

func (h *PostEventuploadHandler) Handles(request request.APIRequest) bool {
	return request.Resource == "/eventupload/{machine_id}" && request.HTTPMethod == "POST"
}

func (h *PostEventuploadHandler) Handle(request request.APIRequest) (*response.ProxyResponse, error) {
	log.Printf("EventUploadHandler %s %s", request.HTTPMethod, request.Path)

	machineID, eventsRequest, errorResponse, err := parseRequest(request)
	if errorResponse != nil || err != nil {
		return errorResponse, err
	}

	if !h.enabled {
		log.Printf("No eventupload sink is configured")
		return response.APIResponse(http.StatusOK, map[string]string{"status": "ok"})
	}

	forwarded := convertRequestEventsToUploadEvents(machineID, eventsRequest.Events)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err = h.publisher.Publish(ctx, machineID, forwarded); err != nil {
		return response.APIResponse(http.StatusInternalServerError, response.ErrInternalServerErrorResponse)
	}
	return response.APIResponse(http.StatusOK, map[string]string{"status": "ok"})
}
