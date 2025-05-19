package internal

import (
	"context"

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

func (s *Service) HandleAnswer(ctx context.Context, message *kafka.Message) error {
	return nil
}
