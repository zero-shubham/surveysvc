package internal

import (
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

func (s *Service) HandleAnswer(message *kafka.Message) error {

}
