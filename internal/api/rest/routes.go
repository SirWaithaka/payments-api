package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/SirWaithaka/payments-api/internal/api/rest/handlers"
	dipkg "github.com/SirWaithaka/payments-api/internal/di"
)

func routes(router *gin.Engine, di *dipkg.DI) {
	webhookRoutes(router, di)

	paymentHandlers := handlers.NewPaymentHandlers(di.Payments)

	router.POST("/deposits", paymentHandlers.Deposit)
}

func webhookRoutes(router *gin.Engine, di *dipkg.DI) {
	webhookGroup := router.Group("/webhooks")

	webhookHandlers := handlers.NewWebhookHandlers()
	webhookGroup.POST("/daraja/:action", webhookHandlers.Daraja)
}
