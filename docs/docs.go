// Package docs embeds the OpenAPI specification at build time so the
// REST handlers can serve it without depending on a particular working
// directory at runtime.
package docs

import _ "embed"

// OpenAPISpec is the OpenAPI 3.0 description of the public REST surface.
//
//go:embed openapi.yaml
var OpenAPISpec []byte
