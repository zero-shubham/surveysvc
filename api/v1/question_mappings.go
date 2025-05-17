package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/zero-shubham/surveysvc/db/orm"
)

type CreateQuestionMappingBody struct {
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	CampaignID uuid.UUID `json:"campaign_id" binding:"required"`
	OrgID      uuid.UUID `json:"org_id" binding:"required"`
}

func (svc *ApiV1Service) CreateQuestionMapping(c *gin.Context) {
	orm := db.New(svc.conn)
	var in CreateQuestionMappingBody

	// fmt.Println("all okay")
	// body, _ := io.ReadAll(c.Request.Body)
	// svc.logger.Info().Interface("body", string(body)).Msg("in request")
	err := c.BindJSON(&in)
	if err != nil {
		svc.logger.Err(err).Msg("failed to parse request body")
		c.AbortWithError(http.StatusBadRequest, errors.New("bad request body"))
		return
	}

	qm, err := orm.CreateQuestionMapping(c.Request.Context(), db.CreateQuestionMappingParams{
		QuestionID: in.QuestionID,
		CampaignID: in.CampaignID,
		OrgID:      in.OrgID,
	})
	if err != nil {
		svc.logger.Err(err).Msg("failed to create question_mapping")
		c.AbortWithError(http.StatusInternalServerError, errors.New("somethign went wrong"))
	}

	c.JSON(http.StatusOK, qm)
}
