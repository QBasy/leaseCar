package controllers

import (
    "context"
    "time"

    "github.com/gofiber/fiber/v2"
    "leaseCar/lease-service/internal/dtos"
    "leaseCar/lease-service/internal/services"
)

type LeaseController struct {
    svc *services.LeaseService
}

func NewLeaseController(s *services.LeaseService) *LeaseController {
    return &LeaseController{svc: s}
}

func (c *LeaseController) Create(ctx *fiber.Ctx) error {
    var in dtos.LeaseCreateRequest
    if err := ctx.BodyParser(&in); err != nil {
        return ctx.Status(400).JSON(fiber.Map{"error": "invalid body"})
    }
    if in.StartDate.IsZero() {
        in.StartDate = time.Now()
    }
    if in.EndDate.IsZero() {
        in.EndDate = in.StartDate.AddDate(0, 12, 0)
    }
    id, err := c.svc.Create(context.Background(), &in)
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return ctx.Status(201).JSON(fiber.Map{"id": id})
}

func (c *LeaseController) GetByID(ctx *fiber.Ctx) error {
    id := ctx.Params("id")
    l, err := c.svc.GetByID(context.Background(), id)
    if err != nil {
        return ctx.Status(404).JSON(fiber.Map{"error": "not found"})
    }
    return ctx.JSON(l)
}

func (c *LeaseController) Search(ctx *fiber.Ctx) error {
    q := ctx.Query("q", "")
    res, err := c.svc.Search(context.Background(), q, 20)
    if err != nil {
        return ctx.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return ctx.JSON(res)
}
