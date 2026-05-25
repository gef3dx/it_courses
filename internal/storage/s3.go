package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type Object struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}

type Service struct {
	baseURL string
}

func New(baseURL string) *Service {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "http://localhost:3000/uploads"
	}
	return &Service{baseURL: strings.TrimRight(baseURL, "/")}
}

func (s *Service) Upload(_ context.Context, filename string, _ []byte, mimeType string) (*Object, error) {
	name := strings.TrimSpace(filepath.Base(filename))
	if name == "" {
		name = fmt.Sprintf("upload-%d.bin", time.Now().UTC().UnixNano())
	}
	return &Object{
		URL:      s.baseURL + "/" + name,
		MimeType: mimeType,
	}, nil
}

func (s *Service) Delete(context.Context, string) error {
	return nil
}
