package requestid

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const requestIDKey contextKey = "request_id"

const HeaderName = "X-Request-ID"

func FromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(requestIDKey)
	id, ok := v.(string)
	return id, ok && id != ""
}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func GetOrGenerate(r *http.Request) string {
	id := r.Header.Get(HeaderName)
	if id == "" {
		id = uuid.NewString()
	}
	return id
}
