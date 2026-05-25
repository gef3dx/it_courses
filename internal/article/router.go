package article

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type AuthContext interface{ Required(roles ...string) fiber.Handler }

type handler struct{ service *Service }

func RegisterRoutes(app *fiber.App, service *Service, authService AuthContext) {
	h := &handler{service: service}
	app.Get("/articles", authService.Required(), h.list)
	app.Get("/articles/:slug", authService.Required(), h.getBySlug)
	app.Post("/articles", authService.Required("teacher", "admin"), h.create)
	app.Put("/articles/:id", authService.Required("teacher", "admin"), h.update)
	app.Delete("/articles/:id", authService.Required("teacher", "admin"), h.delete)
}

// @Summary Список статей
// @Description Возвращает список всех образовательных статей
// @Tags articles
// @Produce json
// @Success 200 {array} Model
// @Router /articles [get]
func (h *handler) list(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch articles"}) }
	return c.JSON(items)
}
// @Summary Детали статьи
// @Description Возвращает статью с медиа и ссылками на тесты по slug
// @Tags articles
// @Produce json
// @Param slug path string true "Slug статьи"
// @Success 200 {object} Model
// @Failure 404 {string} error
// @Router /articles/{slug} [get]
func (h *handler) getBySlug(c fiber.Ctx) error {
	item, err := h.service.GetBySlug(c.Context(), c.Params("slug"))
	if err != nil {
		if errors.Is(err, ErrArticleNotFound) { return c.Status(404).JSON(fiber.Map{"error":"article not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to fetch article"})
	}
	return c.JSON(item)
}
// @Summary Создать статью
// @Description Создаёт новую образовательную статью. Доступно teacher и admin
// @Tags articles
// @Accept json
// @Produce json
// @Param input body CreateInput true "Данные статьи"
// @Success 201 {object} Model
// @Failure 400 {string} error
// @Failure 409 {string} error
// @Router /articles [post]
func (h *handler) create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	authorID, _ := c.Locals("userID").(int64)
	item, err := h.service.Create(c.Context(), input, authorID)
	if err != nil {
		if errors.Is(err, ErrArticleConflict) { return c.Status(409).JSON(fiber.Map{"error":"article with this slug already exists"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.Status(201).JSON(item)
}
// @Summary Обновить статью
// @Description Обновляет данные статьи. Доступно teacher и admin
// @Tags articles
// @Accept json
// @Produce json
// @Param id path int true "ID статьи"
// @Param input body UpdateInput true "Данные для обновления"
// @Success 200 {object} Model
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 409 {string} error
// @Router /articles/{id} [put]
func (h *handler) update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"),10,64)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid article id"}) }
	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.Update(c.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrArticleNotFound) { return c.Status(404).JSON(fiber.Map{"error":"article not found"}) }
		if errors.Is(err, ErrArticleConflict) { return c.Status(409).JSON(fiber.Map{"error":"article with this slug already exists"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.JSON(item)
}
// @Summary Удалить статью
// @Description Удаляет статью по ID. Доступно teacher и admin
// @Tags articles
// @Param id path int true "ID статьи"
// @Success 204 "No Content"
// @Failure 404 {string} error
// @Router /articles/{id} [delete]
func (h *handler) delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"),10,64)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid article id"}) }
	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, ErrArticleNotFound) { return c.Status(404).JSON(fiber.Map{"error":"article not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to delete article"})
	}
	return c.SendStatus(204)
}
