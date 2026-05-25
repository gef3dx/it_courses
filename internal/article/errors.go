package article

import "errors"

var (
	ErrArticleNotFound = errors.New("article not found")
	ErrArticleConflict = errors.New("article already exists")
)
