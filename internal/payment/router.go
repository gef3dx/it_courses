package payment

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/gef3dx/it_courses/internal/course"
)

type AuthContext interface {
	Required(roles ...string) fiber.Handler
}

type handler struct {
	service *Service
}

func RegisterRoutes(app *fiber.App, service *Service, authService AuthContext) {
	h := &handler{service: service}

	app.Post("/courses/:id/payments", authService.Required("student"), h.create)
	app.Get("/payments", authService.Required(), h.list)
	app.Get("/payments/:id", authService.Required(), h.getByID)
	app.Patch("/payments/:id/status", authService.Required("admin"), h.updateStatus)
}

// @Summary Создать платёж за курс
// @Description Создаёт платёж со статусом pending. Только student
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "ID курса"
// @Param input body CreateInput true "Данные платежа"
// @Success 201 {object} Model
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 409 {string} error
// @Router /courses/{id}/payments [post]
func (h *handler) create(c fiber.Ctx) error {
	courseID, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid course id"})
	}

	var input CreateInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.Create(c.Context(), courseID, userID, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentConflict):
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "pending payment already exists"})
		case errors.Is(err, course.ErrCourseNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(http.StatusCreated).JSON(item)
}

// @Summary Список платежей
// @Description Возвращает платежи. Student — свои, admin — все
// @Tags payments
// @Produce json
// @Success 200 {array} Model
// @Router /payments [get]
func (h *handler) list(c fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("userID").(int64)
	items, err := h.service.List(c.Context(), role, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch payments"})
	}
	return c.JSON(items)
}

// @Summary Детали платежа
// @Description Возвращает платёж по ID. Student — только свои, admin — любые
// @Tags payments
// @Produce json
// @Param id path int true "ID платежа"
// @Success 200 {object} Model
// @Failure 403 {string} error
// @Failure 404 {string} error
// @Router /payments/{id} [get]
func (h *handler) getByID(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payment id"})
	}

	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.GetByID(c.Context(), id, role, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "payment not found"})
		case errors.Is(err, ErrPaymentAccessDenied):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "payment access denied"})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch payment"})
		}
	}
	return c.JSON(item)
}

// @Summary Обновить статус платежа
// @Description Меняет статус платежа. При completed автоматически выдаётся доступ к курсу. Только admin
// @Tags payments
// @Accept json
// @Produce json
// @Param id path int true "ID платежа"
// @Param input body UpdateStatusInput true "Новый статус"
// @Success 200 {object} Model
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 409 {string} error
// @Router /payments/{id}/status [patch]
func (h *handler) updateStatus(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid payment id"})
	}

	var input UpdateStatusInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	item, err := h.service.UpdateStatus(c.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "payment not found"})
		case errors.Is(err, ErrPaymentStateConflict):
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "payment status transition is not allowed"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(item)
}
