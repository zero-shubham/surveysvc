package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/zero-shubham/surveysvc/db/orm"
)

type CreateAnswerBody struct {
	SelectedOption string    `json:"selected_option"`
	AnswerText     string    `json:"answer_text"`
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	QuestionID     uuid.UUID `json:"question_id" binding:"required"`
	QuestionSetID  uuid.UUID `json:"question_set_id" binding:"required"`
}

func (svc *ApiV1Service) CreateAnswer(c *gin.Context) {
	var in CreateAnswerBody
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	orm := db.New(svc.conn)
	answer, err := orm.CreateAnswer(c.Request.Context(), db.CreateAnswerParams{
		SelectedOption: pgtype.Text{String: in.SelectedOption, Valid: in.SelectedOption != ""},
		AnswerText:     pgtype.Text{String: in.AnswerText, Valid: in.AnswerText != ""},
		UserID:         in.UserID,
		QuestionID:     in.QuestionID,
		QuestionSetID:  in.QuestionSetID,
	})
	if err != nil {
		svc.logger.Err(err).Msg("failed to create answer")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create answer"})
		return
	}

	c.JSON(http.StatusCreated, answer)
}

type GetAnswersQuery struct {
	QuestionID string `form:"question_id" binding:"required"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

func (svc *ApiV1Service) GetAnswers(c *gin.Context) {
	var query GetAnswersQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(400, gin.H{"error": "invalid query parameters"})
		return
	}

	// Set default limit if not provided
	if query.Limit == 0 {
		query.Limit = 10
	}

	questionID, err := uuid.Parse(query.QuestionID)
	if err != nil {
		svc.logger.Err(err).Msg("invalid question id")
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid query parameters"))
		return
	}

	orm := db.New(svc.conn)
	answers, err := orm.GetAnswersByQuestionID(c.Request.Context(), db.GetAnswersByQuestionIDParams{
		QuestionID: questionID,
		Limit:      int32(query.Limit),
		Offset:     int32(query.Offset),
	})
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to fetch answers"})
		return
	}

	c.JSON(200, answers)
}
