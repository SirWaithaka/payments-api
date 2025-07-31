package ginzerolog

import (
	"fmt"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/SirWaithaka/payments-api/internal/pkg/logger"
)

type responseWriter struct {
	gin.ResponseWriter
	buf []byte
}

func (rw *responseWriter) Write(b []byte) (n int, err error) {
	copy(rw.buf, b)
	return rw.ResponseWriter.Write(b)
}

func New(cfg Config) gin.HandlerFunc {
	cfg = configDefault(cfg)

	return func(c *gin.Context) {
		// skip uri
		if slices.Contains(cfg.SkipURIs, c.Request.RequestURI) {
			c.Next()
			return
		}

		// get request id from user context
		requestID := xid.New().String()
		reqId := c.Request.Context().Value(logger.LRequestID)
		if reqId != nil {
			requestID = reqId.(string)
		}

		// update logging context with http method and request url
		cfg.Logger.UpdateContext(func(ctx zerolog.Context) zerolog.Context {
			return ctx.Fields(map[string]interface{}{
				logger.LRequestID: requestID,
				"method":          c.Request.Method,
				"url":             c.Request.URL.String(),
			})
		})

		// pass logger to request context
		c.Request = c.Request.WithContext(cfg.Logger.WithContext(c.Request.Context()))

		// capture time before calling next handler
		start := time.Now()

		rw := &responseWriter{buf: make([]byte, 0), ResponseWriter: c.Writer}
		c.Writer = rw

		// call next handler
		c.Next()

		defer func() {
			// calculated the elapsed time
			end := time.Now()
			duration := end.Sub(start)

			elapsed := func(d time.Duration) string {
				s := int(d.Seconds()) % 60
				ms := int(d.Milliseconds()) % 1000
				us := int(d.Microseconds()) % 1000
				return fmt.Sprintf("%ds.%dms,%dus", s, ms, us)
			}

			// change the log level depending on response status code
			statusCode := c.Writer.Status()
			level := zerolog.InfoLevel
			if statusCode >= 400 && statusCode < 500 {
				level = zerolog.WarnLevel
			} else if statusCode >= 500 && statusCode < 600 {
				level = zerolog.ErrorLevel
			}

			lg := *cfg.Logger
			lg.WithLevel(level).
				Int("statusCode", statusCode).
				Str("elapsed", elapsed(duration)).
				Msg("request complete")
		}()
	}
}
