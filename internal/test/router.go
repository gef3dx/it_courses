package test

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type AuthContext interface{ Required(roles ...string) fiber.Handler }

type handler struct{ service *Service }

func RegisterRoutes(app *fiber.App, service *Service, authService AuthContext) {
	h := &handler{service: service}
	app.Get("/tests", authService.Required(), h.list)
	app.Get("/tests/:id", authService.Required(), h.getByID)
	app.Post("/tests", authService.Required("teacher", "admin"), h.create)
	app.Put("/tests/:id", authService.Required("teacher", "admin"), h.update)
	app.Delete("/tests/:id", authService.Required("teacher", "admin"), h.delete)
	app.Get("/tests/:id/questions", authService.Required(), h.listQuestions)
	app.Post("/tests/:testId/questions", authService.Required("teacher", "admin"), h.createQuestion)
	app.Put("/tests/:testId/questions/:id", authService.Required("teacher", "admin"), h.updateQuestion)
	app.Delete("/tests/:testId/questions/:id", authService.Required("teacher", "admin"), h.deleteQuestion)
	app.Post("/tests/:id/submit", authService.Required("student"), h.submit)
	app.Get("/results", authService.Required(), h.listResults)
	app.Get("/results/:id", authService.Required(), h.getResultByID)
	app.Get("/tests/:id/results", authService.Required("teacher", "admin"), h.listResultsByTest)
}

func parseID(c fiber.Ctx, key, msg string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, c.Status(400).JSON(fiber.Map{"error": msg})
	}
	return id, nil
}

func (h *handler) list(c fiber.Ctx) error { items, err := h.service.List(c.Context()); if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch tests"}) }; return c.JSON(items) }
func (h *handler) getByID(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid test id"); if err != nil { return nil }
	item, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrTestNotFound) { return c.Status(404).JSON(fiber.Map{"error":"test not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to fetch test"})
	}
	return c.JSON(item)
}
func (h *handler) create(c fiber.Ctx) error {
	var input CreateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.Create(c.Context(), input, userID)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.Status(201).JSON(item)
}
func (h *handler) update(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid test id"); if err != nil { return nil }
	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.Update(c.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrTestNotFound) { return c.Status(404).JSON(fiber.Map{"error":"test not found"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.JSON(item)
}
func (h *handler) delete(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid test id"); if err != nil { return nil }
	if err := h.service.Delete(c.Context(), id); err != nil {
		if errors.Is(err, ErrTestNotFound) { return c.Status(404).JSON(fiber.Map{"error":"test not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to delete test"})
	}
	return c.SendStatus(204)
}
func (h *handler) listQuestions(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid test id"); if err != nil { return nil }
	items, err := h.service.ListPublicQuestions(c.Context(), id)
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch questions"}) }
	return c.JSON(items)
}
func (h *handler) createQuestion(c fiber.Ctx) error {
	testID, err := parseID(c, "testId", "invalid test id"); if err != nil { return nil }
	var input QuestionInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.CreateQuestion(c.Context(), testID, input, userID)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.Status(201).JSON(item)
}
func (h *handler) updateQuestion(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid question id"); if err != nil { return nil }
	var input QuestionInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.UpdateQuestion(c.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrQuestionNotFound) { return c.Status(404).JSON(fiber.Map{"error":"question not found"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.JSON(item)
}
func (h *handler) deleteQuestion(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid question id"); if err != nil { return nil }
	if err := h.service.DeleteQuestion(c.Context(), id); err != nil {
		if errors.Is(err, ErrQuestionNotFound) { return c.Status(404).JSON(fiber.Map{"error":"question not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to delete question"})
	}
	return c.SendStatus(204)
}
func (h *handler) submit(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid test id"); if err != nil { return nil }
	var input SubmitInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.Submit(c.Context(), id, userID, input)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.JSON(item)
}
func (h *handler) listResults(c fiber.Ctx) error {
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("userID").(int64)
	items, err := h.service.ListResults(c.Context(), role, userID)
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch results"}) }
	return c.JSON(items)
}
func (h *handler) getResultByID(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid result id"); if err != nil { return nil }
	role, _ := c.Locals("role").(string)
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.GetResultByID(c.Context(), id, role, userID)
	if err != nil {
		if errors.Is(err, ErrResultNotFound) { return c.Status(404).JSON(fiber.Map{"error":"result not found"}) }
		if errors.Is(err, ErrResultAccessDenied) { return c.Status(403).JSON(fiber.Map{"error":"result access denied"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to fetch result"})
	}
	return c.JSON(item)
}
func (h *handler) listResultsByTest(c fiber.Ctx) error {
	id, err := parseID(c, "id", "invalid test id"); if err != nil { return nil }
	items, err := h.service.ListResultsByTest(c.Context(), id)
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch results"}) }
	return c.JSON(items)
}
