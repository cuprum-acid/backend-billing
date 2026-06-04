// Package handlers provides HTTP handlers for the billing API.
package handlers

import (
	"net/http"

	"backend-billing/docs"
)

// swaggerUIHTML serves a minimal swagger-ui page that fetches the
// embedded spec from /openapi.yaml. The CDN-hosted swagger-ui assets
// avoid pulling vendored dist files into the repository; users in
// air-gapped environments can swap in a local copy without touching
// the Go source.
const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <title>Billing API – Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
        deepLinking: true,
      });
    };
  </script>
</body>
</html>
`

// SwaggerUI serves the swagger-ui shell HTML.
func SwaggerUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(swaggerUIHTML))
}

// OpenAPISpec serves the embedded OpenAPI 3.0 YAML.
func OpenAPISpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(docs.OpenAPISpec)
}
