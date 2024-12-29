// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/go-chi/render"
)

type Middleware struct {
  baseUrl string
}

func ValidationMiddleware(baseUrl string) *Middleware {
  return &Middleware{baseUrl: baseUrl}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	swagger, err := GetSwagger()
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		matchedPath, pathParams := matchPath(r, swagger, m.baseUrl)
		if matchedPath != nil {
			pathItem := swagger.Paths.Find(*matchedPath)
			operation := pathItem.Operations()[r.Method]
			if operation != nil {
				requestValidationInput := &openapi3filter.RequestValidationInput{
					Request:    r,
					PathParams: pathParams,
					Route: &routers.Route{
						Spec:      swagger,
						Server:    nil,
						Path:      *matchedPath,
						PathItem:  pathItem,
						Method:    r.Method,
						Operation: operation,
					},
				}
				err := openapi3filter.ValidateRequest(r.Context(), requestValidationInput)
				if err != nil {
					_ = render.Render(w, r, ErrInvalidRequest(err))
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func matchPath(r *http.Request, swagger *openapi3.T, basePath string) (*string, map[string]string) {
	pathParams := make(map[string]string)
  requestPath := strings.TrimPrefix(r.URL.Path, basePath)
	requestElements := strings.Split(requestPath, "/")

	for _, path := range swagger.Paths.InMatchingOrder() {
		match := true
		pathElements := strings.Split(path, "/")
		if len(requestElements) == len(pathElements) {
			for i := 0; i < len(requestElements); i++ {
				pe := pathElements[i]
				if strings.HasPrefix(pe, "{") && strings.HasSuffix(pe, "}") {
					pathParams[pe[1:len(pe)-1]] = requestElements[i]
					continue
				}
				if pe != requestElements[i] {
					match = false
					break
				}
			}

			if match {
				return &path, pathParams
			}
		}
	}

	return nil, nil
}
