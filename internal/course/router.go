package course

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type AuthContext interface {
	Required(roles ...string) fiber.Handler
}

type handler struct {
	service *Service
}

func RegisterRoutes(app *fiber.App, service *Service, authService AuthContext) {
	h := &handler{service: service}

	app.Get("/courses", authService.Required(), h.list)
	app.Get("/courses/:slug", authService.Required(), h.getBySlug)
	app.Get("/my/courses", authService.Required(), h.myCourses)
	app.Post("/courses", authService.Required("teacher", "admin"), h.create)
	app.Put("/courses/:id", authService.Required("teacher", "admin"), h.update)
	app.Delete("/courses/:id", authService.Required("teacher", "admin"), h.delete)
	app.Get("/courses/:id/access", authService.Required("admin"), h.listAccesses)
	app.Post("/courses/:id/access", authService.Required("admin"), h.grantAccess)
	app.Delete("/courses/:id/access/:userId", authService.Required("admin"), h.revokeAccess)
}

func (h *handler) list(c fiber.Ctx) error {
	userID, _ := c.Locals("userID").(int64)
	role, _ := c.Locals("role").(string)

	items, err := h.service.List(c.Context(), role, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch courses"})
	}

	return c.JSON(items)
}

func (h *handler) myCourses(c fiber.Ctx) error {
	userID, _ := c.Locals("userID").(int64)
	items, err := h.service.MyCourses(c.Context(), userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch courses"})
	}
	return c.JSON(items)
}

func (h *handler) getBySlug(c fiber.Ctx) error {
	userID, _ := c.Locals("userID").(int64)
	role, _ := c.Locals("role").(string)

	item, err := h.service.GetBySlug(c.Context(), c.Params("slug"), role, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCourseNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
		case errors.Is(err, ErrCourseAccessDenied):
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"error": "course access denied"})
		default:
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch course"})
		}
	}

	return c.JSON(item)
}

func (h *handler) create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.Create(c.Context(), input, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCourseConflict):
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "course with this slug already exists"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(http.StatusCreated).JSON(item)
}

func (h *handler) update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid course id"})
	}

	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	item, err := h.service.Update(c.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, ErrCourseNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
		case errors.Is(err, ErrCourseConflict):
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "course with this slug already exists"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.JSON(item)
}

func (h *handler) delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid course id"})
	}

	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to delete course"})
	}

	return c.SendStatus(http.StatusNoContent)
}

func (h *handler) listAccesses(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid course id"})
	}

	items, err := h.service.ListAccesses(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch course accesses"})
	}

	return c.JSON(items)
}

func (h *handler) grantAccess(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid course id"})
	}

	var input GrantAccessInput
	if err := c.Bind().Body(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	adminID, _ := c.Locals("userID").(int64)
	item, err := h.service.GrantAccess(c.Context(), id, input, adminID)
	if err != nil {
		switch {
		case errors.Is(err, ErrCourseNotFound):
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course not found"})
		case errors.Is(err, ErrCourseAccessExists):
			return c.Status(http.StatusConflict).JSON(fiber.Map{"error": "course access already exists"})
		default:
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
	}

	return c.Status(http.StatusCreated).JSON(item)
}

func (h *handler) revokeAccess(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid course id"})
	}
	userID, err := strconv.ParseInt(c.Params("userId"), 10, 64)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid user id"})
	}

	if err := h.service.RevokeAccess(c.Context(), id, userID); err != nil {
		if errors.Is(err, ErrCourseAccessMissing) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "course access not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "failed to revoke access"})
	}

	return c.SendStatus(http.StatusNoContent)
}
