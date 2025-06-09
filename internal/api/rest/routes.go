package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/SirWaithaka/payments-api/internal/api/rest/handlers"
)

func routes(router *gin.Engine) {
	paymentHandlers := handlers.PaymentHandlers{}

	router.POST("/deposits", paymentHandlers.Deposit)
}
