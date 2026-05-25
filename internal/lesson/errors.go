package lesson

import "errors"

var (
	ErrLessonNotFound       = errors.New("lesson not found")
	ErrLessonConflict       = errors.New("lesson already exists")
	ErrLessonQuestionNotFound = errors.New("lesson question not found")
	ErrLessonTestLinkNotFound = errors.New("lesson test link not found")
)
