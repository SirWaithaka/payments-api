package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/SirWaithaka/payments-api/internal/api/rest/handlers"
	dipkg "github.com/SirWaithaka/payments-api/internal/di"
)

func routes(router *gin.Engine, di *dipkg.DI) {
	webhookRoutes(router, di)

	paymentHandlers := handlers.NewPaymentHandlers(di.Wallets)

	group := router.Group("/api")

	group.POST("/charge", paymentHandlers.Charge)
	group.POST("/payout", paymentHandlers.Payout)
}

func webhookRoutes(router *gin.Engine, di *dipkg.DI) {
	webhookGroup := router.Group("/webhooks")

	webhookHandlers := handlers.NewWebhookHandlers(di.Webhook)
	webhookGroup.POST("/daraja/:action", webhookHandlers.Daraja)
}
