package rest

import (
	"github.com/gin-gonic/gin"

	"github.com/SirWaithaka/payments-api/src/api/rest/handlers"
	dipkg "github.com/SirWaithaka/payments-api/src/di"
)

func routes(router *gin.Engine, di *dipkg.DI) {
	webhookRoutes(router, di)

	mpesaHandlers := handlers.NewMpesaHandlers(di.Mpesa, di.ShortCode)

	group := router.Group("/api")

	// mpesa routes
	mpesaGroup := group.Group("/mpesa")
	mpesaGroup.POST("/charge", mpesaHandlers.Charge)
	mpesaGroup.POST("/payout", mpesaHandlers.Payout)
	mpesaGroup.POST("/transfer", mpesaHandlers.Transfer)
	mpesaGroup.POST("/status", mpesaHandlers.PaymentStatus)

	mpesaGroup.POST("/shortcode", mpesaHandlers.AddShortCode)
}

func webhookRoutes(router *gin.Engine, di *dipkg.DI) {
	webhookGroup := router.Group("/webhooks")

	webhookHandlers := handlers.NewWebhookHandlers(di.Webhook)
	webhookGroup.POST("/daraja/:action", webhookHandlers.Daraja)
	webhookGroup.POST("/quikk/mpesa/:action", webhookHandlers.QuikkMpesa)
}
