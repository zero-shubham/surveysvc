package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	db "github.com/zero-shubham/surveysvc/db/orm"
)

type ApiV1Service struct {
	conn   db.DBTX
	logger *zerolog.Logger
}

type RouterGroupCreator interface {
	Group(relativePath string, handlers ...gin.HandlerFunc) *gin.RouterGroup
}

func NewApiV1Service(rgc RouterGroupCreator, conn db.DBTX, logger *zerolog.Logger) *ApiV1Service {
	v1Api := ApiV1Service{
		conn:   conn,
		logger: logger,
	}

	v1 := rgc.Group("/v1")
	v1.POST("/question-mappings", v1Api.CreateQuestionMapping)
	v1.GET("/question-mappings", v1Api.GetQuestionMappings)
	v1.PATCH("/question-mappings/:id", v1Api.UpdateQuestionMapping)

	v1.GET("/answers", v1Api.GetAnswers)

	return &v1Api
}
