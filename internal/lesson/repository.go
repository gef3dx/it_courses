package lesson

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) ListByCourse(ctx context.Context, courseID int64) ([]Model, error) {
	var items []Model
	err := r.db.WithContext(ctx).Preload("Media").Preload("Tests").Where("course_id = ?", courseID).Order("sort_order ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) FindByID(ctx context.Context, courseID, id int64) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Preload("Media").Preload("Tests").Where("course_id = ? AND id = ?", courseID, id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrLessonNotFound }
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Create(ctx context.Context, courseID int64, input CreateInput, authorID int64, slug string) (*Model, error) {
	item := &Model{CourseID: courseID, Title: strings.TrimSpace(input.Title), Slug: slug, Content: input.Content, AuthorID: authorID, SortOrder: input.SortOrder, IsPublished: input.IsPublished}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			if unique(err) { return ErrLessonConflict }
			return err
		}
		for _, media := range input.Media {
			if err := tx.Create(&Media{LessonID:item.ID, MediaType:media.MediaType, URL:media.URL, Caption:media.Caption, SortOrder:media.SortOrder}).Error; err != nil { return err }
		}
		for _, link := range input.Tests {
			if err := tx.Create(&TestLink{LessonID:item.ID, TestID:link.TestID, Description:link.Description}).Error; err != nil { return err }
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindByID(ctx, courseID, item.ID)
}

func (r *Repository) Update(ctx context.Context, courseID, id int64, input UpdateInput, slug string) (*Model, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{}
		if input.Title != nil { updates["title"] = strings.TrimSpace(*input.Title) }
		if input.Slug != nil { updates["slug"] = slug }
		if input.Content != nil { updates["content"] = *input.Content }
		if input.SortOrder != nil { updates["sort_order"] = *input.SortOrder }
		if input.IsPublished != nil { updates["is_published"] = *input.IsPublished }
		if len(updates) > 0 {
			result := tx.Model(&Model{}).Where("course_id = ? AND id = ?", courseID, id).Updates(updates)
			if result.Error != nil {
				if unique(result.Error) { return ErrLessonConflict }
				return result.Error
			}
			if result.RowsAffected == 0 { return ErrLessonNotFound }
		}
		if input.Media != nil {
			if err := tx.Where("lesson_id = ?", id).Delete(&Media{}).Error; err != nil { return err }
			for _, media := range *input.Media {
				if err := tx.Create(&Media{LessonID:id, MediaType:media.MediaType, URL:media.URL, Caption:media.Caption, SortOrder:media.SortOrder}).Error; err != nil { return err }
			}
		}
		if input.Tests != nil {
			if err := tx.Where("lesson_id = ?", id).Delete(&TestLink{}).Error; err != nil { return err }
			for _, link := range *input.Tests {
				if err := tx.Create(&TestLink{LessonID:id, TestID:link.TestID, Description:link.Description}).Error; err != nil { return err }
			}
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindByID(ctx, courseID, id)
}

func (r *Repository) Delete(ctx context.Context, courseID, id int64) error {
	result := r.db.WithContext(ctx).Where("course_id = ? AND id = ?", courseID, id).Delete(&Model{})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return ErrLessonNotFound }
	return nil
}

func (r *Repository) Reorder(ctx context.Context, courseID int64, items []ReorderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			if err := tx.Model(&Model{}).Where("course_id = ? AND id = ?", courseID, item.ID).Update("sort_order", item.SortOrder).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListQuestions(ctx context.Context, lessonID int64) ([]Question, error) {
	var items []Question
	err := r.db.WithContext(ctx).Preload("Options").Where("lesson_id = ?", lessonID).Order("sort_order ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) CreateQuestion(ctx context.Context, lessonID int64, input QuestionInput, authorID int64) (*Question, error) {
	var item Question
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item = Question{LessonID:&lessonID, Text:strings.TrimSpace(input.Text), Solution:input.Solution, AuthorID:authorID, SortOrder:input.SortOrder}
		if err := tx.Table("questions").Create(&item).Error; err != nil { return err }
		for _, option := range input.Options {
			if err := tx.Create(&AnswerOption{QuestionID:item.ID, Text:strings.TrimSpace(option.Text), IsCorrect:option.IsCorrect}).Error; err != nil { return err }
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindQuestionByID(ctx, lessonID, item.ID)
}

func (r *Repository) FindQuestionByID(ctx context.Context, lessonID, id int64) (*Question, error) {
	var item Question
	err := r.db.WithContext(ctx).Preload("Options").Where("lesson_id = ? AND id = ?", lessonID, id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrLessonQuestionNotFound }
		return nil, err
	}
	return &item, nil
}

func (r *Repository) UpdateQuestion(ctx context.Context, lessonID, id int64, input QuestionInput) (*Question, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Table("questions").Where("lesson_id = ? AND id = ?", lessonID, id).Updates(map[string]any{
			"text":strings.TrimSpace(input.Text),"solution":input.Solution,"sort_order":input.SortOrder,
		})
		if result.Error != nil { return result.Error }
		if result.RowsAffected == 0 { return ErrLessonQuestionNotFound }
		if err := tx.Where("question_id = ?", id).Delete(&AnswerOption{}).Error; err != nil { return err }
		for _, option := range input.Options {
			if err := tx.Create(&AnswerOption{QuestionID:id, Text:strings.TrimSpace(option.Text), IsCorrect:option.IsCorrect}).Error; err != nil { return err }
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindQuestionByID(ctx, lessonID, id)
}

func (r *Repository) DeleteQuestion(ctx context.Context, lessonID, id int64) error {
	result := r.db.WithContext(ctx).Table("questions").Where("lesson_id = ? AND id = ?", lessonID, id).Delete(&Question{})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return ErrLessonQuestionNotFound }
	return nil
}

func (r *Repository) LinkTest(ctx context.Context, lessonID int64, input TestLinkInput) (*TestLink, error) {
	item := &TestLink{LessonID:lessonID, TestID:input.TestID, Description:input.Description}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil { return nil, err }
	return item, nil
}

func (r *Repository) UnlinkTest(ctx context.Context, lessonID, testID int64) error {
	result := r.db.WithContext(ctx).Where("lesson_id = ? AND test_id = ?", lessonID, testID).Delete(&TestLink{})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return ErrLessonTestLinkNotFound }
	return nil
}

func (r *Repository) Submit(ctx context.Context, lessonID int64, answers []SubmitAnswerInput) (*SubmitResponse, error) {
	questions, err := r.ListQuestions(ctx, lessonID)
	if err != nil { return nil, err }
	answerMap := map[int64]int64{}
	for _, item := range answers { answerMap[item.QuestionID] = item.SelectedOptionID }
	score := 0
	public := make([]PublicQuestion,0,len(questions))
	for _, question := range questions {
		pq := PublicQuestion{ID:int(question.ID), Text:question.Text, Solution:question.Solution, SortOrder:question.SortOrder}
		for _, option := range question.Options {
			pq.Options = append(pq.Options, PublicAnswerOption{ID:option.ID, Text:option.Text})
			if option.IsCorrect && option.ID == answerMap[question.ID] { score++ }
		}
		public = append(public, pq)
	}
	grade := 0.0
	if len(questions) > 0 { grade = math.Round((float64(score)/float64(len(questions))*100)*100)/100 }
	return &SubmitResponse{Score:score, MaxScore:len(questions), Grade:grade, Questions:public}, nil
}

func unique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
