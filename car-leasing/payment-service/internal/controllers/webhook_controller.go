package controllers

import (
	"github.com/gofiber/fiber/v2"
	"leaseCar/payment-service/internal/services"
)

type WebhookController struct{
	svc *services.PaymentService
}

func NewWebhookController(s *services.PaymentService) *WebhookController { return &WebhookController{svc: s} }

func (w *WebhookController) Handle(c *fiber.Ctx) error {
	provider := c.Params("provider")
	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	// delegate to service
	_ = w.svc.HandleProviderWebhook(c.Context(), provider, payload)
	return c.Status(200).JSON(fiber.Map{"status": "ok"})
}
