package page

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type AuthContext interface{ Required(roles ...string) fiber.Handler }

type handler struct{ service *Service }

func RegisterRoutes(app *fiber.App, service *Service, authService AuthContext) {
	h := &handler{service: service}
	app.Get("/pages", authService.Required(), h.list)
	app.Get("/pages/:slug", authService.Required(), h.getBySlug)
	app.Post("/pages", authService.Required("teacher", "admin"), h.create)
	app.Put("/pages/:id", authService.Required("teacher", "admin"), h.update)
	app.Delete("/pages/:id", authService.Required("teacher", "admin"), h.delete)
}

// @Summary Список страниц
// @Description Возвращает список всех статических страниц
// @Tags pages
// @Produce json
// @Success 200 {array} Model
// @Router /pages [get]
func (h *handler) list(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch pages"}) }
	return c.JSON(items)
}
// @Summary Детали страницы
// @Description Возвращает статическую страницу по slug
// @Tags pages
// @Produce json
// @Param slug path string true "Slug страницы"
// @Success 200 {object} Model
// @Failure 404 {string} error
// @Router /pages/{slug} [get]
func (h *handler) getBySlug(c fiber.Ctx) error {
	item, err := h.service.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		if errors.Is(err, ErrPageNotFound) { return c.Status(404).JSON(fiber.Map{"error":"page not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to fetch page"})
	}
	return c.JSON(item)
}
// @Summary Создать страницу
// @Description Создаёт новую статическую страницу. Доступно teacher и admin
// @Tags pages
// @Accept json
// @Produce json
// @Param input body CreateInput true "Данные страницы"
// @Success 201 {object} Model
// @Failure 400 {string} error
// @Failure 409 {string} error
// @Router /pages [post]
func (h *handler) create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.Create(c.Context(), input)
	if err != nil {
		if errors.Is(err, ErrPageConflict) { return c.Status(409).JSON(fiber.Map{"error":"page with this slug already exists"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.Status(201).JSON(item)
}
// @Summary Обновить страницу
// @Description Обновляет статическую страницу. Доступно teacher и admin
// @Tags pages
// @Accept json
// @Produce json
// @Param id path int true "ID страницы"
// @Param input body UpdateInput true "Данные для обновления"
// @Success 200 {object} Model
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 409 {string} error
// @Router /pages/{id} [put]
func (h *handler) update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"),10,64)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid page id"}) }
	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.Update(c.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrPageNotFound) { return c.Status(404).JSON(fiber.Map{"error":"page not found"}) }
		if errors.Is(err, ErrPageConflict) { return c.Status(409).JSON(fiber.Map{"error":"page with this slug already exists"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.JSON(item)
}
// @Summary Удалить страницу
// @Description Удаляет статическую страницу. Доступно teacher и admin
// @Tags pages
// @Param id path int true "ID страницы"
// @Success 204 "No Content"
// @Failure 404 {string} error
// @Router /pages/{id} [delete]
func (h *handler) delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"),10,64)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid page id"}) }
	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, ErrPageNotFound) { return c.Status(404).JSON(fiber.Map{"error":"page not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to delete page"})
	}
	return c.SendStatus(204)
}
