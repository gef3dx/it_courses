package page

import "errors"

var (
	ErrPageNotFound = errors.New("page not found")
	ErrPageConflict = errors.New("page already exists")
)
