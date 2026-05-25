package lesson

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type AuthContext interface{ Required(roles ...string) fiber.Handler }

type handler struct{ service *Service }

func RegisterRoutes(app *fiber.App, service *Service, authService AuthContext) {
	h := &handler{service: service}
	app.Get("/courses/:courseId/lessons", authService.Required(), h.listByCourse)
	app.Get("/courses/:courseId/lessons/:id", authService.Required(), h.getByID)
	app.Post("/courses/:courseId/lessons", authService.Required("teacher", "admin"), h.create)
	app.Put("/courses/:courseId/lessons/:id", authService.Required("teacher", "admin"), h.update)
	app.Delete("/courses/:courseId/lessons/:id", authService.Required("teacher", "admin"), h.delete)
	app.Patch("/courses/:courseId/lessons/reorder", authService.Required("teacher", "admin"), h.reorder)
	app.Get("/lessons/:id/questions", authService.Required(), h.listQuestions)
	app.Post("/lessons/:lessonId/questions", authService.Required("teacher", "admin"), h.createQuestion)
	app.Put("/lessons/:lessonId/questions/:id", authService.Required("teacher", "admin"), h.updateQuestion)
	app.Delete("/lessons/:lessonId/questions/:id", authService.Required("teacher", "admin"), h.deleteQuestion)
	app.Post("/lessons/:id/submit", authService.Required("student"), h.submit)
	app.Post("/lessons/:lessonId/tests", authService.Required("teacher", "admin"), h.linkTest)
	app.Delete("/lessons/:lessonId/tests/:testId", authService.Required("teacher", "admin"), h.unlinkTest)
}

func paramID(c fiber.Ctx, key, message string) (int64, error) {
	id, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, c.Status(400).JSON(fiber.Map{"error": message})
	}
	return id, nil
}

// @Summary Список уроков курса
// @Description Возвращает все уроки указанного курса
// @Tags lessons
// @Produce json
// @Param courseId path int true "ID курса"
// @Success 200 {array} Model
// @Failure 500 {string} error
// @Router /courses/{courseId}/lessons [get]
func (h *handler) listByCourse(c fiber.Ctx) error {
	courseID, err := paramID(c,"courseId","invalid course id"); if err != nil { return nil }
	items, err := h.service.ListByCourse(c.Context(), courseID)
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch lessons"}) }
	return c.JSON(items)
}
// @Summary Детали урока
// @Description Возвращает урок с контентом, медиа и ссылками на тесты
// @Tags lessons
// @Produce json
// @Param courseId path int true "ID курса"
// @Param id path int true "ID урока"
// @Success 200 {object} Model
// @Failure 404 {string} error
// @Router /courses/{courseId}/lessons/{id} [get]
func (h *handler) getByID(c fiber.Ctx) error {
	courseID, err := paramID(c,"courseId","invalid course id"); if err != nil { return nil }
	id, err := paramID(c,"id","invalid lesson id"); if err != nil { return nil }
	item, err := h.service.GetByID(c.Context(), courseID, id)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) { return c.Status(404).JSON(fiber.Map{"error":"lesson not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to fetch lesson"})
	}
	return c.JSON(item)
}
// @Summary Создать урок
// @Description Создаёт новый урок в курсе. Доступно teacher и admin
// @Tags lessons
// @Accept json
// @Produce json
// @Param courseId path int true "ID курса"
// @Param input body CreateInput true "Данные урока"
// @Success 201 {object} Model
// @Failure 400 {string} error
// @Failure 409 {string} error
// @Router /courses/{courseId}/lessons [post]
func (h *handler) create(c fiber.Ctx) error {
	courseID, err := paramID(c,"courseId","invalid course id"); if err != nil { return nil }
	var input CreateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.Create(c.Context(), courseID, input, userID)
	if err != nil {
		if errors.Is(err, ErrLessonConflict) { return c.Status(409).JSON(fiber.Map{"error":"lesson with this slug already exists"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.Status(201).JSON(item)
}
// @Summary Обновить урок
// @Description Обновляет данные урока. Доступно teacher и admin
// @Tags lessons
// @Accept json
// @Produce json
// @Param courseId path int true "ID курса"
// @Param id path int true "ID урока"
// @Param input body UpdateInput true "Данные для обновления"
// @Success 200 {object} Model
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Failure 409 {string} error
// @Router /courses/{courseId}/lessons/{id} [put]
func (h *handler) update(c fiber.Ctx) error {
	courseID, err := paramID(c,"courseId","invalid course id"); if err != nil { return nil }
	id, err := paramID(c,"id","invalid lesson id"); if err != nil { return nil }
	var input UpdateInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.Update(c.Context(), courseID, id, input)
	if err != nil {
		if errors.Is(err, ErrLessonNotFound) { return c.Status(404).JSON(fiber.Map{"error":"lesson not found"}) }
		if errors.Is(err, ErrLessonConflict) { return c.Status(409).JSON(fiber.Map{"error":"lesson with this slug already exists"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.JSON(item)
}
// @Summary Удалить урок
// @Description Удаляет урок из курса. Доступно teacher и admin
// @Tags lessons
// @Param courseId path int true "ID курса"
// @Param id path int true "ID урока"
// @Success 204 "No Content"
// @Failure 404 {string} error
// @Router /courses/{courseId}/lessons/{id} [delete]
func (h *handler) delete(c fiber.Ctx) error {
	courseID, err := paramID(c,"courseId","invalid course id"); if err != nil { return nil }
	id, err := paramID(c,"id","invalid lesson id"); if err != nil { return nil }
	if err := h.service.Delete(c.Context(), courseID, id); err != nil {
		if errors.Is(err, ErrLessonNotFound) { return c.Status(404).JSON(fiber.Map{"error":"lesson not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to delete lesson"})
	}
	return c.SendStatus(204)
}
// @Summary Изменить порядок уроков
// @Description Меняет порядок уроков в курсе. Доступно teacher и admin
// @Tags lessons
// @Accept json
// @Produce json
// @Param courseId path int true "ID курса"
// @Param input body []ReorderItem true "Массив ID и sort_order"
// @Success 200 {string} error
// @Failure 400 {string} error
// @Router /courses/{courseId}/lessons/reorder [patch]
func (h *handler) reorder(c fiber.Ctx) error {
	courseID, err := paramID(c,"courseId","invalid course id"); if err != nil { return nil }
	var input []ReorderItem
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	if err := h.service.Reorder(c.Context(), courseID, input); err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.JSON(fiber.Map{"message":"reordered"})
}
// @Summary Вопросы урока
// @Description Возвращает вопросы урока для прохождения (без правильных ответов)
// @Tags lessons
// @Produce json
// @Param id path int true "ID урока"
// @Success 200 {array} PublicQuestion
// @Router /lessons/{id}/questions [get]
func (h *handler) listQuestions(c fiber.Ctx) error {
	id, err := paramID(c,"id","invalid lesson id"); if err != nil { return nil }
	items, err := h.service.ListPublicQuestions(c.Context(), id)
	if err != nil { return c.Status(500).JSON(fiber.Map{"error":"failed to fetch questions"}) }
	return c.JSON(items)
}
// @Summary Создать вопрос урока
// @Description Создаёт вопрос с вариантами ответов для урока. Доступно teacher и admin
// @Tags lessons
// @Accept json
// @Produce json
// @Param lessonId path int true "ID урока"
// @Param input body QuestionInput true "Вопрос и варианты ответов"
// @Success 201 {object} Question
// @Failure 400 {string} error
// @Router /lessons/{lessonId}/questions [post]
func (h *handler) createQuestion(c fiber.Ctx) error {
	lessonID, err := paramID(c,"lessonId","invalid lesson id"); if err != nil { return nil }
	var input QuestionInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	userID, _ := c.Locals("userID").(int64)
	item, err := h.service.CreateQuestion(c.Context(), lessonID, input, userID)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.Status(201).JSON(item)
}
// @Summary Обновить вопрос урока
// @Description Обновляет вопрос и варианты ответов. Доступно teacher и admin
// @Tags lessons
// @Accept json
// @Produce json
// @Param lessonId path int true "ID урока"
// @Param id path int true "ID вопроса"
// @Param input body QuestionInput true "Новые данные вопроса"
// @Success 200 {object} Question
// @Failure 400 {string} error
// @Failure 404 {string} error
// @Router /lessons/{lessonId}/questions/{id} [put]
func (h *handler) updateQuestion(c fiber.Ctx) error {
	lessonID, err := paramID(c,"lessonId","invalid lesson id"); if err != nil { return nil }
	id, err := paramID(c,"id","invalid question id"); if err != nil { return nil }
	var input QuestionInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.UpdateQuestion(c.Context(), lessonID, id, input)
	if err != nil {
		if errors.Is(err, ErrLessonQuestionNotFound) { return c.Status(404).JSON(fiber.Map{"error":"question not found"}) }
		return c.Status(400).JSON(fiber.Map{"error":err.Error()})
	}
	return c.JSON(item)
}
// @Summary Удалить вопрос урока
// @Description Удаляет вопрос из урока. Доступно teacher и admin
// @Tags lessons
// @Param lessonId path int true "ID урока"
// @Param id path int true "ID вопроса"
// @Success 204 "No Content"
// @Failure 404 {string} error
// @Router /lessons/{lessonId}/questions/{id} [delete]
func (h *handler) deleteQuestion(c fiber.Ctx) error {
	lessonID, err := paramID(c,"lessonId","invalid lesson id"); if err != nil { return nil }
	id, err := paramID(c,"id","invalid question id"); if err != nil { return nil }
	if err := h.service.DeleteQuestion(c.Context(), lessonID, id); err != nil {
		if errors.Is(err, ErrLessonQuestionNotFound) { return c.Status(404).JSON(fiber.Map{"error":"question not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to delete question"})
	}
	return c.SendStatus(204)
}
// @Summary Отправить ответы на вопросы урока
// @Description Принимает ответы на вопросы урока и возвращает результат. Только student
// @Tags lessons
// @Accept json
// @Produce json
// @Param id path int true "ID урока"
// @Param input body SubmitInput true "Ответы на вопросы"
// @Success 200 {object} SubmitResponse
// @Failure 400 {string} error
// @Router /lessons/{id}/submit [post]
func (h *handler) submit(c fiber.Ctx) error {
	id, err := paramID(c,"id","invalid lesson id"); if err != nil { return nil }
	var input SubmitInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	result, err := h.service.Submit(c.Context(), id, input)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.JSON(result)
}
// @Summary Привязать тест к уроку
// @Description Добавляет ссылку на тест в урок. Доступно teacher и admin
// @Tags lessons
// @Accept json
// @Produce json
// @Param lessonId path int true "ID урока"
// @Param input body TestLinkInput true "ID теста"
// @Success 201 {object} TestLink
// @Failure 400 {string} error
// @Router /lessons/{lessonId}/tests [post]
func (h *handler) linkTest(c fiber.Ctx) error {
	lessonID, err := paramID(c,"lessonId","invalid lesson id"); if err != nil { return nil }
	var input TestLinkInput
	if err := c.Bind().Body(&input); err != nil { return c.Status(400).JSON(fiber.Map{"error":"invalid request body"}) }
	item, err := h.service.LinkTest(c.Context(), lessonID, input)
	if err != nil { return c.Status(400).JSON(fiber.Map{"error":err.Error()}) }
	return c.Status(201).JSON(item)
}
// @Summary Отвязать тест от урока
// @Description Удаляет ссылку на тест из урока. Доступно teacher и admin
// @Tags lessons
// @Param lessonId path int true "ID урока"
// @Param testId path int true "ID теста"
// @Success 204 "No Content"
// @Failure 404 {string} error
// @Router /lessons/{lessonId}/tests/{testId} [delete]
func (h *handler) unlinkTest(c fiber.Ctx) error {
	lessonID, err := paramID(c,"lessonId","invalid lesson id"); if err != nil { return nil }
	testID, err := paramID(c,"testId","invalid test id"); if err != nil { return nil }
	if err := h.service.UnlinkTest(c.Context(), lessonID, testID); err != nil {
		if errors.Is(err, ErrLessonTestLinkNotFound) { return c.Status(404).JSON(fiber.Map{"error":"lesson test link not found"}) }
		return c.Status(500).JSON(fiber.Map{"error":"failed to unlink test"})
	}
	return c.SendStatus(204)
}
