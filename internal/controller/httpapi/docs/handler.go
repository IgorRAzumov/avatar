package docs

import (
	"fmt"
	"net/http"

	"avatar/internal/controller/httpapi/common/util"

	sitedocs "avatar/docs"
)

type Handler struct{}

func New() *Handler {
	return &Handler{}
}

func (handler *Handler) Spec(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/yaml")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write(sitedocs.OpenAPI)
}

func (handler *Handler) UI(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(swaggerHTML))
}

var swaggerHTML = fmt.Sprintf(swaggerTemplate, util.OpenAPIPath)

const swaggerTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8"/>
  <title>GophProfile API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: %q,
      dom_id: "#swagger-ui"
    });
  </script>
</body>
</html>
`
