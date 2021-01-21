package http

import (
	"net/http"

	"github.com/gorilla/handlers"
)

func AllowedHeaders() handlers.CORSOption {
	return handlers.AllowedHeaders([]string{
		"Content-Type",
		"X-Requested-With",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"Content-Language",
		"Origin",
	})
}

func AllowedOrigins() handlers.CORSOption {
	return handlers.AllowedOrigins([]string{"*"})
}

func AllowedMethods() handlers.CORSOption {
	return handlers.AllowedMethods([]string{
		"GET",
		"HEAD",
		"POST",
		"PUT",
		"OPTIONS",
		"DELETE",
	})
}

func CORS(h http.Handler) http.Handler {
	return handlers.CORS(AllowedOrigins(), AllowedHeaders(), AllowedMethods())(h)
}
