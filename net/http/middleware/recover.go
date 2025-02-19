package middleware

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/foomo/keel/log"
)

type RecoverConfig struct {
	DisablePrintStack bool
}

var DefaultRecoverConfig = RecoverConfig{
	DisablePrintStack: false,
}

func Recover() Middleware {
	return RecoverWithConfig(DefaultRecoverConfig)
}

func RecoverWithConfig(config RecoverConfig) Middleware {
	return func(l *zap.Logger, next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if e := recover(); e != nil {
					err, ok := e.(error)
					if !ok {
						err = fmt.Errorf("%v", e)
					}
					l = log.WithError(l, err)
					if !config.DisablePrintStack {
						l = l.With(log.FStackSkip(3))
					}
					l.Error("recovering from panic")
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
