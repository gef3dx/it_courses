package course

import "errors"

var (
	ErrCourseNotFound      = errors.New("course not found")
	ErrCourseConflict      = errors.New("course already exists")
	ErrCourseAccessDenied  = errors.New("course access denied")
	ErrCourseAccessExists  = errors.New("course access already exists")
	ErrCourseAccessMissing = errors.New("course access not found")
)
