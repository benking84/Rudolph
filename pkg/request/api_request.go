package request

import (
	"log"
	"net/http"
	"strings"

	"github.com/airbnb/rudolph/pkg/response"
	"github.com/airbnb/rudolph/pkg/types"
	"github.com/google/uuid"
)

// APIRequest is the subset of an HTTP request the sync handlers look at.
// Resource is the route template, for example /preflight/{machine_id}.
type APIRequest struct {
	HTTPMethod     string
	Path           string
	Resource       string
	PathParameters map[string]string
	Headers        map[string]string
	Body           string
}

// IsValidUUID reports whether machineID is a canonical 36-char UUID.
func IsValidUUID(machineID string) bool {
	if err := types.ValidateMachineID(machineID); err != nil {
		return false
	}
	_, err := uuid.Parse(machineID)
	return err == nil
}

// GetMachineID reads and validates the machine_id path parameter.
func GetMachineID(req APIRequest) (machineID string, errorResponse *response.ProxyResponse, err error) {
	machineID, found := req.PathParameters["machine_id"]

	if !found || len(machineID) == 0 {
		log.Printf("ASSERTION FAILED: Received blank {machine_id}")
		errorResponse, err = response.APIResponse(http.StatusBadRequest, response.ErrBlankPathParameterResponse)
		return
	}

	if !(IsValidUUID(machineID)) {
		log.Printf("ASSERTION FAILED: Received invalid {machine_id}")
		errorResponse, err = response.APIResponse(http.StatusBadRequest, response.ErrInvalidPathParameterResponse)
		return
	}

	machineID = strings.ToUpper(machineID)
	return
}
