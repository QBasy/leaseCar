package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"leaseCar/payment-service/internal/dtos"
	"leaseCar/payment-service/internal/services"
)

type PaymentController struct {
	svc *services.PaymentService
}

func NewPaymentController(s *services.PaymentService) *PaymentController { return &PaymentController{svc: s} }

func (pc *PaymentController) Create(c *fiber.Ctx) error {
	var req dtos.PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := pc.svc.CreatePayment(ctx, &req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(resp)
}
