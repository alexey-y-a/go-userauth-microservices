package http

import (
	"net/http"

	"github.com/alexey-y-a/go-userauth-microservices/libs/requestid"
	"github.com/rs/zerolog"
)

func RequestIDMiddleware(log zerolog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := requestid.GetOrGenerate(r)
		ctx := requestid.WithContext(r.Context(), id)

		w.Header().Set(requestid.HeaderName, id)

		reqLog := log.With().Str("request_id", id).Logger()
		_ = reqLog

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
