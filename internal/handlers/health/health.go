package health

import (
	"github.com/airbnb/rudolph/pkg/request"
	"net/http"

	"github.com/airbnb/rudolph/pkg/model/globalrules"
	"github.com/airbnb/rudolph/pkg/response"
	"github.com/airbnb/rudolph/pkg/store"
)

type GetHealthHandler struct {
	booted         bool
	dynamodbClient store.Client
}

func (h *GetHealthHandler) Boot() (err error) {
	if h.booted {
		return
	}

	h.dynamodbClient, err = store.FromEnv()
	if err != nil {
		return err
	}

	h.booted = true
	return
}

func (h *GetHealthHandler) Handles(request request.APIRequest) bool {
	return request.Resource == "/health" && request.HTTPMethod == "GET"
}

func (h *GetHealthHandler) Handle(request request.APIRequest) (*response.ProxyResponse, error) {
	return h.healthHandler(request)
}

// HealthHandler to be used by the health function
func (h GetHealthHandler) healthHandler(request request.APIRequest) (*response.ProxyResponse, error) {
	// Test dynamodb connection by querying for 1 global rule at random
	// we don't care about the response, as long as DDB doesn't fail with a permission error or something,
	// our system can be considered "healthy"
	err := globalrules.PingDatabase(h.dynamodbClient)
	if err != nil {
		return response.APIResponse(http.StatusInternalServerError, map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		})
	}

	// Otherwise we gucci, return 200
	return response.APIResponse(http.StatusOK, map[string]string{"status": "healthy"})
}
