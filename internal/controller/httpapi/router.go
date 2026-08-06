package httpapi

import (
	"net/http"

	"avatar/internal/controller/httpapi/common/util"

	apimiddleware "avatar/internal/controller/httpapi/common/middleware"
	"avatar/internal/controller/httpapi/health"
	"avatar/internal/controller/httpapi/read"
	"avatar/internal/controller/httpapi/web"
	"avatar/internal/controller/httpapi/write"
	"avatar/internal/logger"
	"avatar/internal/observability"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouterDeps struct {
	Logger      *logger.Logger
	ServiceName string
	Kit         observability.Kit
	AvatarRead  *read.Handler
	AvatarWrite *write.Handler
	Web         *web.Handler
	Health      *health.Handler
}

func NewRouter(deps RouterDeps) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(apimiddleware.Tracing(deps.ServiceName))
	router.Use(apimiddleware.PrometheusMetrics(deps.Kit))
	router.Use(apimiddleware.RequestLogger(deps.Logger))

	router.Get("/health", deps.Health.Health)

	if deps.Web != nil {
		router.Get("/", func(writer http.ResponseWriter, request *http.Request) {
			http.Redirect(writer, request, "/web/upload", http.StatusFound)
		})
		router.Get("/web/upload", deps.Web.UploadPage)
		router.Post("/web/upload", deps.Web.PostUpload)
		router.Get("/web/gallery/{user_id}", deps.Web.GalleryPage)
	}

	router.Route(util.VersionPrefix, func(builder chi.Router) {
		builder.Post("/avatars", deps.AvatarWrite.Upload)
		builder.Get("/avatars/{avatar_id}", deps.AvatarRead.GetAvatar)
		builder.Get("/avatars/{avatar_id}/metadata", deps.AvatarRead.GetMetadata)
		builder.Delete("/avatars/{avatar_id}", deps.AvatarWrite.DeleteAvatar)

		builder.Get("/users/{user_id}/avatar", deps.AvatarRead.GetUserAvatar)
		builder.Delete("/users/{user_id}/avatar", deps.AvatarWrite.DeleteUserAvatar)
		builder.Get("/users/{user_id}/avatars", deps.AvatarRead.GetListUserAvatars)
	})

	return router
}
