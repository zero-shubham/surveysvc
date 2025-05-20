package internal

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/segmentio/kafka-go"
	db "github.com/zero-shubham/surveysvc/db/orm"
)

type Service struct {
	logger *zerolog.Logger
	conn   db.DBTX
}

func NewService(logger *zerolog.Logger, conn db.DBTX) *Service {
	return &Service{
		logger: logger,
		conn:   conn,
	}
}

type MessageBody struct {
	UserID         uuid.UUID `json:"user_id"`
	QuestionID     uuid.UUID `json:"question_id"`
	QuestionSetID  uuid.UUID `json:"question_set_id"`
	SelectedOption string    `json:"selected_option"`
	AnswerText     string    `json:"answer_text"`
}

func (s *Service) HandleAnswer(ctx context.Context, message *kafka.Message) error {
	var mb MessageBody

	err := json.Unmarshal(message.Value, &mb)
	if err != nil {
		s.logger.Err(err).Msg("failed to parse message body")
		return nil
	}

	orm := db.New(s.conn)

	_, err = orm.CreateAnswer(ctx, db.CreateAnswerParams{
		AnswerText: pgtype.Text{
			String: mb.AnswerText,
			Valid:  true,
		},
		SelectedOption: pgtype.Text{
			String: mb.SelectedOption,
			Valid:  true,
		},
		UserID:        mb.UserID,
		QuestionID:    mb.QuestionID,
		QuestionSetID: mb.QuestionSetID,
	})
	if err != nil {
		s.logger.Err(err).Msg("failed to create answer record")
		return err
	}

	return nil
}
