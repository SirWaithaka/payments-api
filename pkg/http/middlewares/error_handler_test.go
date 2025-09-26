package middlewares_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/SirWaithaka/payments-api/pkg/http/middlewares"
	"github.com/SirWaithaka/payments-api/src/repositories/postgres"
)

func assertEquals(t *testing.T, exp any, val any) {
	if exp != val {
		t.Errorf("expected %v but got %v", exp, val)
	}
}

func TestErrorHandler(t *testing.T) {
	engine := gin.New()
	gin.SetMode(gin.TestMode)
	engine.Use(middlewares.ErrorHandler())

	t.Run("test it catches postgres errors", func(t *testing.T) {
		err := postgres.Error{Err: gorm.ErrRecordNotFound}

		engine.GET("/not-found", func(c *gin.Context) {
			_ = c.Error(err)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/not-found", nil)
		engine.ServeHTTP(w, req)

		assertEquals(t, http.StatusNotFound, w.Code)
	})

	t.Run("test it catches validation errors", func(t *testing.T) {
		type body struct {
			Value string `json:"value" binding:"required"`
		}

		engine.POST("/validation", func(c *gin.Context) {
			var b body
			if err := c.ShouldBind(&b); err != nil {
				_ = c.Error(err)
				return
			}
			c.Status(http.StatusOK)
		})

		w := httptest.NewRecorder()
		b := bytes.NewReader([]byte{})
		req, _ := http.NewRequest(http.MethodPost, "/validation", b)
		engine.ServeHTTP(w, req)

		assertEquals(t, http.StatusBadRequest, w.Code)

	})

	t.Run("test status code is not overwritten", func(t *testing.T) {

		engine.POST("/error", func(c *gin.Context) {
			_ = c.Error(errors.New("test error"))
			c.Status(http.StatusConflict)
		})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/error", nil)
		engine.ServeHTTP(w, req)

		assertEquals(t, http.StatusConflict, w.Code)
	})
}
