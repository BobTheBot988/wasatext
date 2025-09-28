package main

import (
	"net/http"

	"github.com/gorilla/handlers"
)

// ApplyCORSHandler applies a CORS policy to the router. CORS stands for Cross-Origin Resource Sharing: it's a security
// feature present in web browsers that blocks JavaScript requests going across different domains if not specified in a
// policy. This function sends the policy of this API server.
func applyCORSHandler(h http.Handler) http.Handler {
	return handlers.CORS(
		// ! WE NEED TO SPECIFY THE HEADERS WE WANT TO ALLOW
		handlers.AllowedHeaders([]string{"Authorization", "Access-Control-Allow-Origin", "content-type", "multipart/form-data", "blob"}),
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS", "DELETE", "PUT"}),
		// Do not modify the CORS origin and max age, they are used in the evaluation.
		handlers.AllowedOrigins([]string{"*"}),
		handlers.MaxAge(1),
	)(h)
}
