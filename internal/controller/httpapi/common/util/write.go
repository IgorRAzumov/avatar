package util

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	commonmodel "avatar/internal/controller/httpapi/common/model"
	"avatar/internal/domain/model"
	"avatar/internal/imageformat"
	"avatar/internal/logger"
)

func WriteJSON(writer http.ResponseWriter, status int, v any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(v)
}

func WriteError(writer http.ResponseWriter, status int, body commonmodel.ErrorResponse) {
	WriteJSON(writer, status, body)
}

func WriteServiceError(ctx context.Context, logger *logger.Logger, w http.ResponseWriter, err error, maxSize int64) {
	switch {
	case errors.Is(err, model.ErrNotFound):
		WriteError(w, http.StatusNotFound, commonmodel.ErrorResponse{Error: commonmodel.MsgAvatarNotFound})
	case errors.Is(err, model.ErrForbidden):
		WriteError(w, http.StatusForbidden, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgForbidden,
			Details: commonmodel.MsgForbiddenDeleteDetails,
		})
	case errors.Is(err, model.ErrInvalidFormat):
		WriteError(w, http.StatusBadRequest, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgInvalidFileFormat,
			Details: "Supported formats: " + imageformat.SupportedFormatsLabel(),
		})
	case errors.Is(err, model.ErrFileTooLarge):
		WriteError(w, http.StatusRequestEntityTooLarge, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgFileTooLarge,
			MaxSize: maxSize,
		})
	default:
		logger.Error(ctx, "request failed", "error", err)
		WriteError(w, http.StatusInternalServerError, commonmodel.ErrorResponse{Error: commonmodel.MsgInternalServerError})
	}
}
