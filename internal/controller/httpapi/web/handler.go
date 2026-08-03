package web

import (
	"io/fs"
	"net/http"

	"avatar/internal/controller/httpapi/write"
	siteweb "avatar/web"
)

type Handler struct {
	writeHandler *write.Handler
}

func New(writeHandler *write.Handler) *Handler {
	return &Handler{writeHandler: writeHandler}
}

func (handler *Handler) UploadPage(writer http.ResponseWriter, request *http.Request) {
	handler.serveHTML(writer, "upload.html")
}

func (handler *Handler) GalleryPage(writer http.ResponseWriter, request *http.Request) {
	handler.serveHTML(writer, "gallery.html")
}

// PostUpload handles multipart form POST /web/upload (userId + file fields).
// It maps form userId to X-User-ID and delegates to the REST upload handler.
func (handler *Handler) PostUpload(writer http.ResponseWriter, request *http.Request) {
	if request.Header.Get(write.HeaderUserID) == "" {
		if userID := request.FormValue("userId"); userID != "" {
			request.Header.Set(write.HeaderUserID, userID)
		}
	}
	handler.writeHandler.Upload(writer, request)
}

func (handler *Handler) serveHTML(writer http.ResponseWriter, name string) {
	data, err := siteweb.Static.ReadFile("static/" + name)
	if err != nil {
		http.Error(writer, "page not found", http.StatusNotFound)
		return
	}
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(data)
}

// StaticFS exposes embedded static assets for tests.
func StaticFS() (fs.FS, error) {
	return fs.Sub(siteweb.Static, "static")
}
