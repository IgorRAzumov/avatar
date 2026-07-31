package read

import (
	"net/http"
	"time"

	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/controller/httpapi/read/mapper"
	"avatar/internal/controller/httpapi/read/model"
	"avatar/internal/domain/usecase"
	"avatar/internal/logger"
)

const cacheControl = "max-age=86400"

type Handler struct {
	logger      *logger.Logger
	readUsecase usecase.AvatarReadUsecase
	baseURL     string
}

func New(logger *logger.Logger, readUsecase usecase.AvatarReadUsecase, baseURL string) *Handler {
	return &Handler{logger: logger, readUsecase: readUsecase, baseURL: baseURL}
}

func (handler *Handler) GetAvatar(writer http.ResponseWriter, request *http.Request) {
	avatarID := request.PathValue("avatar_id")
	size := request.URL.Query().Get("size")
	format := request.URL.Query().Get("format")

	image, err := handler.readUsecase.GetImage(request.Context(), avatarID, size, format)
	if err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, 0)
		return
	}

	util.WriteCachedImage(writer, request, image.Data, image.MimeType, cacheControl)
}

func (handler *Handler) GetMetadata(writer http.ResponseWriter, request *http.Request) {
	avatarID := request.PathValue("avatar_id")
	item, err := handler.readUsecase.GetByID(request.Context(), avatarID)
	if err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, 0)
		return
	}
	util.WriteJSON(writer, http.StatusOK, mapper.AvatarMetadataResponse(handler.baseURL, avatarID, item))
}

func (handler *Handler) GetUserAvatar(writer http.ResponseWriter, request *http.Request) {
	userID := util.PathUserID(request)
	size := request.URL.Query().Get("size")
	format := request.URL.Query().Get("format")

	image, err := handler.readUsecase.GetUserAvatarImage(request.Context(), userID, size, format)
	if err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, 0)
		return
	}

	util.WriteCachedImage(writer, request, image.Data, image.MimeType, cacheControl)
}

func (handler *Handler) GetListUserAvatars(writer http.ResponseWriter, request *http.Request) {
	userID := util.PathUserID(request)
	avatars, err := handler.readUsecase.ListByUser(request.Context(), userID)
	if err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, 0)
		return
	}

	items := make([]model.AvatarItem, 0, len(avatars))
	for _, avatar := range avatars {
		items = append(items, model.AvatarItem{
			ID:        avatar.ID,
			UserID:    avatar.UserID,
			FileName:  avatar.FileName,
			Status:    avatar.ProcessingStatus,
			URL:       util.AvatarRelativePath(avatar.ID),
			CreatedAt: avatar.CreatedAt.Format(time.RFC3339),
		})
	}
	util.WriteJSON(writer, http.StatusOK, map[string]any{"avatars": items})
}
