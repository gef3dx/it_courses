package upload

import (
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/gef3dx/it_courses/internal/queue"
	"github.com/gef3dx/it_courses/internal/storage"
)

type AuthContext interface{ Required(roles ...string) fiber.Handler }

type handler struct {
	storage   *storage.Service
	publisher queue.Publisher
}

func RegisterRoutes(app *fiber.App, authService AuthContext, storageService *storage.Service, publisher queue.Publisher) {
	h := &handler{storage: storageService, publisher: publisher}
	app.Post("/upload", authService.Required(), h.upload)
}

func (h *handler) upload(c fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "file is required"})
	}

	contentType := fileHeader.Header.Get("Content-Type")
	if !supportedContentType(fileHeader.Filename, contentType) {
		return c.Status(400).JSON(fiber.Map{"error": "unsupported content type"})
	}

	content, err := readFile(fileHeader)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to read upload"})
	}

	object, err := h.storage.Upload(c.Context(), fileHeader.Filename, content, contentType)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to store upload"})
	}

	_ = h.publisher.Publish(c.Context(), queue.Message{
		Type: "media_uploaded",
		Payload: map[string]any{
			"url":       object.URL,
			"mime_type": object.MimeType,
		},
	})

	return c.Status(201).JSON(object)
}

func readFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	return io.ReadAll(file)
}

func supportedContentType(filename, value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "image/jpeg", "image/png", "image/webp", "video/mp4":
		return true
	case "application/octet-stream", "":
		switch strings.ToLower(filepath.Ext(filename)) {
		case ".jpg", ".jpeg", ".png", ".webp", ".mp4":
			return true
		}
		return false
	default:
		return false
	}
}
