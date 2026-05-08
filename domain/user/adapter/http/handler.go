package httpadapter

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"todoe/domain/user/application"
	"todoe/domain/user/port"
)

type Handler struct {
	useCase port.UseCase
}

func NewHandler(useCase port.UseCase) *Handler {
	return &Handler{useCase: useCase}
}

func (h *Handler) Register(c *fiber.Ctx) error {
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	result := h.useCase.Register(c.Context(), body.Name, body.Email)
	if result.IsError() {
		if errors.Is(result.Error(), application.ErrInvalidName) || errors.Is(result.Error(), application.ErrInvalidEmail) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": result.Error().Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.Status(fiber.StatusCreated).JSON(result.MustGet())
}

func (h *Handler) VerifyEmail(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	println(id)
	var body struct {
		Token string `json:"token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	result := h.useCase.VerifyEmail(c.Context(), id, body.Token)
	if result.IsError() {
		if errors.Is(result.Error(), application.ErrInvalidToken) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": result.Error().Error()})
		}
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(result.MustGet())
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	result := h.useCase.GetUser(c.Context(), id)
	if result.IsError() {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(result.MustGet())
}

func (h *Handler) GetHistory(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	result := h.useCase.GetUserHistory(c.Context(), id)
	if result.IsError() {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(result.MustGet())
}

func (h *Handler) ListActivated(c *fiber.Ctx) error {
	result := h.useCase.ListActivatedUsers(c.Context())
	if result.IsError() {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(result.MustGet())
}

func (h *Handler) UpdateContact(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	var body struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Bio   string `json:"bio"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	result := h.useCase.UpdateContact(c.Context(), id, body.Name, body.Email, body.Bio)
	if result.IsError() {
		switch {
		case errors.Is(result.Error(), application.ErrInvalidName),
			errors.Is(result.Error(), application.ErrInvalidEmail):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": result.Error().Error()})
		case errors.Is(result.Error(), application.ErrEmailTaken):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": result.Error().Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(result.MustGet())
}

func (h *Handler) CompleteProfile(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	var body struct {
		Bio string `json:"bio"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	result := h.useCase.CompleteProfile(c.Context(), id, body.Bio)
	if result.IsError() {
		switch {
		case errors.Is(result.Error(), application.ErrCreditDenied):
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": result.Error().Error()})
		case errors.Is(result.Error(), application.ErrCreditNotChecked):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": result.Error().Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal error"})
	}
	return c.JSON(result.MustGet())
}
