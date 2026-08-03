package health

import (
	"net/http"

	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/controller/httpapi/health/model"
	domainmodel "avatar/internal/domain/model"
	"avatar/internal/domain/repository"
)

type Handler struct {
	healthRepository repository.HealthRepository
}

func New(health repository.HealthRepository) *Handler {
	return &Handler{healthRepository: health}
}

func (handler *Handler) Health(writer http.ResponseWriter, request *http.Request) {
	report := handler.healthRepository.Check(request.Context())
	resp := model.HealthResponse{
		Status:     report.Status,
		Components: report.Components,
	}
	status := http.StatusOK
	if resp.Status != domainmodel.HealthStatusOK {
		status = http.StatusServiceUnavailable
	}
	util.WriteJSON(writer, status, resp)
}
