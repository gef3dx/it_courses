package test

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type Repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) *Repository { return &Repository{db: db} }

func (r *Repository) List(ctx context.Context) ([]Model, error) {
	var items []Model
	err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) Create(ctx context.Context, input CreateInput, authorID int64) (*Model, error) {
	item := &Model{Title: strings.TrimSpace(input.Title), Description: strings.TrimSpace(input.Description), AuthorID: authorID}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil { return nil, err }
	return item, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*Model, error) {
	var item Model
	err := r.db.WithContext(ctx).Preload("Questions.Options", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC, id ASC")
	}).Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrTestNotFound }
		return nil, err
	}
	return &item, nil
}

func (r *Repository) Update(ctx context.Context, id int64, input UpdateInput) (*Model, error) {
	updates := map[string]any{}
	if input.Title != nil { updates["title"] = strings.TrimSpace(*input.Title) }
	if input.Description != nil { updates["description"] = strings.TrimSpace(*input.Description) }
	result := r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil { return nil, result.Error }
	if result.RowsAffected == 0 { return nil, ErrTestNotFound }
	return r.FindByID(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Model{})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return ErrTestNotFound }
	return nil
}

func (r *Repository) ListQuestionsByTest(ctx context.Context, testID int64) ([]Question, error) {
	var items []Question
	err := r.db.WithContext(ctx).Preload("Options").Where("test_id = ?", testID).Order("sort_order ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) CreateQuestion(ctx context.Context, testID int64, input QuestionInput, authorID int64) (*Question, error) {
	var item Question
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item = Question{TestID: &testID, Text: strings.TrimSpace(input.Text), Solution: input.Solution, AuthorID: authorID, SortOrder: input.SortOrder}
		if err := tx.Create(&item).Error; err != nil { return err }
		for _, option := range input.Options {
			if err := tx.Create(&AnswerOption{
				QuestionID: item.ID, Text: strings.TrimSpace(option.Text), IsCorrect: option.IsCorrect,
			}).Error; err != nil { return err }
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindQuestionByID(ctx, item.ID)
}

func (r *Repository) FindQuestionByID(ctx context.Context, id int64) (*Question, error) {
	var item Question
	err := r.db.WithContext(ctx).Preload("Options").Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrQuestionNotFound }
		return nil, err
	}
	return &item, nil
}

func (r *Repository) UpdateQuestion(ctx context.Context, id int64, input QuestionInput) (*Question, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&Question{}).Where("id = ?", id).Updates(map[string]any{
			"text": strings.TrimSpace(input.Text), "solution": input.Solution, "sort_order": input.SortOrder,
		})
		if result.Error != nil { return result.Error }
		if result.RowsAffected == 0 { return ErrQuestionNotFound }
		if err := tx.Where("question_id = ?", id).Delete(&AnswerOption{}).Error; err != nil { return err }
		for _, option := range input.Options {
			if err := tx.Create(&AnswerOption{QuestionID: id, Text: strings.TrimSpace(option.Text), IsCorrect: option.IsCorrect}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindQuestionByID(ctx, id)
}

func (r *Repository) DeleteQuestion(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&Question{})
	if result.Error != nil { return result.Error }
	if result.RowsAffected == 0 { return ErrQuestionNotFound }
	return nil
}

func (r *Repository) Submit(ctx context.Context, testID, userID int64, answers []SubmitAnswerInput) (*Result, error) {
	var result Result
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var questions []Question
		if err := tx.Preload("Options").Where("test_id = ?", testID).Order("sort_order ASC, id ASC").Find(&questions).Error; err != nil {
			return err
		}
		if len(questions) == 0 {
			return ErrQuestionNotFound
		}

		answerMap := map[int64]int64{}
		for _, answer := range answers { answerMap[answer.QuestionID] = answer.SelectedOptionID }

		score := 0
		items := make([]ResultItem, 0, len(questions))
		for _, question := range questions {
			selectedID := answerMap[question.ID]
			correct := false
			for _, option := range question.Options {
				if option.IsCorrect && option.ID == selectedID { correct = true; score++ }
			}
			items = append(items, ResultItem{QuestionID: question.ID, SelectedOptionID: selectedID, IsCorrect: correct})
		}

		grade := 0.0
		if len(questions) > 0 {
			grade = math.Round((float64(score)/float64(len(questions))*100)*100) / 100
		}
		result = Result{TestID: testID, UserID: userID, Score: score, MaxScore: len(questions), Grade: grade, CompletedAt: time.Now().UTC()}
		if err := tx.Create(&result).Error; err != nil { return err }
		for _, item := range items {
			item.ResultID = result.ID
			if err := tx.Create(&item).Error; err != nil { return err }
		}
		return nil
	})
	if err != nil { return nil, err }
	return r.FindResultByID(ctx, result.ID)
}

func (r *Repository) FindResultByID(ctx context.Context, id int64) (*Result, error) {
	var item Result
	err := r.db.WithContext(ctx).Preload("Answers").Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return nil, ErrResultNotFound }
		return nil, err
	}
	return &item, nil
}

func (r *Repository) ListResultsByUser(ctx context.Context, userID int64) ([]Result, error) {
	var items []Result
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repository) ListResultsByTest(ctx context.Context, testID int64) ([]Result, error) {
	var items []Result
	err := r.db.WithContext(ctx).Where("test_id = ?", testID).Order("id DESC").Find(&items).Error
	return items, err
}

func unique(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
