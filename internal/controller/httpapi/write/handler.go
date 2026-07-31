package write

import (
	"io"
	"net/http"

	commonmodel "avatar/internal/controller/httpapi/common/model"
	"avatar/internal/controller/httpapi/common/util"
	"avatar/internal/controller/httpapi/write/mapper"
	"avatar/internal/domain/usecase"
	"avatar/internal/logger"
)

const (
	multipartFormName = "file"
	multipartOverhead = 1024
	HeaderUserID      = "X-User-ID"
)

type Handler struct {
	logger       *logger.Logger
	writeUsecase usecase.AvatarWriteUsecase
	maxSize      int64
	baseURL      string
}

func New(logger *logger.Logger, writeUsecase usecase.AvatarWriteUsecase, maxSize int64, baseURL string) *Handler {
	return &Handler{logger: logger, writeUsecase: writeUsecase, maxSize: maxSize, baseURL: baseURL}
}

func (handler *Handler) Upload(writer http.ResponseWriter, request *http.Request) {
	userID := request.Header.Get(HeaderUserID)
	if userID == "" {
		util.WriteError(writer, http.StatusBadRequest, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgMissingHeader,
			Details: commonmodel.MsgMissingUserIDHeader,
		})
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, handler.maxSize+multipartOverhead)
	if err := request.ParseMultipartForm(handler.maxSize); err != nil {
		util.WriteError(writer, http.StatusRequestEntityTooLarge, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgFileTooLarge,
			MaxSize: handler.maxSize,
		})
		return
	}

	file, header, err := request.FormFile(multipartFormName)
	if err != nil {
		util.WriteError(writer, http.StatusBadRequest, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgInvalidRequest,
			Details: commonmodel.MsgFileFieldRequired,
		})
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		util.WriteError(writer, http.StatusBadRequest, commonmodel.ErrorResponse{Error: commonmodel.MsgFailedToReadFile})
		return
	}

	item, err := handler.writeUsecase.Upload(request.Context(), userID, header.Filename, data)
	if err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, handler.maxSize)
		return
	}

	util.WriteJSON(writer, http.StatusCreated, mapper.UploadResponse(handler.baseURL, item))
}

func (handler *Handler) DeleteAvatar(writer http.ResponseWriter, request *http.Request) {
	userID := request.Header.Get(HeaderUserID)
	if userID == "" {
		util.WriteError(writer, http.StatusBadRequest, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgMissingHeader,
			Details: commonmodel.MsgMissingUserIDHeader,
		})
		return
	}

	avatarID := request.PathValue("avatar_id")
	if err := handler.writeUsecase.Delete(request.Context(), avatarID, userID); err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, handler.maxSize)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) DeleteUserAvatar(writer http.ResponseWriter, request *http.Request) {
	userID := request.Header.Get(HeaderUserID)
	if userID == "" {
		util.WriteError(writer, http.StatusBadRequest, commonmodel.ErrorResponse{
			Error:   commonmodel.MsgMissingHeader,
			Details: commonmodel.MsgMissingUserIDHeader,
		})
		return
	}

	targetUserID := util.PathUserID(request)
	if err := handler.writeUsecase.DeleteByUser(request.Context(), targetUserID, userID); err != nil {
		util.WriteServiceError(request.Context(), handler.logger, writer, err, handler.maxSize)
		return
	}
	writer.WriteHeader(http.StatusNoContent)
}
