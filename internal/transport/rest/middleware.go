package rest

import (
	"context"
	"errors"
	"net/http"

	"github.com/olegtemek/friend-seek-backend/internal/models"
)

const loginHeader = "X-User-Login"

type ctxKey struct{}

var errUnauthorized = errors.New("missing or malformed " + loginHeader + " header")

func auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		login, err := models.NormalizeLogin(r.Header.Get(loginHeader))
		if err != nil {
			fail(w, errUnauthorized)

			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKey{}, login)))
	})
}

func getUserLoginFromCtx(ctx context.Context) string {
	login, _ := ctx.Value(ctxKey{}).(string)
	return login
}
