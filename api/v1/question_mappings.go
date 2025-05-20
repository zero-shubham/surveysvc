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

type GetQuestionMappingsQuery struct {
	CampaignID string `form:"campaign_id" binding:"required"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

type GetQuestionMappingsResp struct {
	QuestionMappings []db.QuestionMapping `json:"question_mappings"`
}

func (svc *ApiV1Service) GetQuestionMappings(c *gin.Context) {
	var query GetQuestionMappingsQuery
	if err := c.BindQuery(&query); err != nil {
		svc.logger.Err(err).Msg("invalid query parameters")
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid query parameters"))
		return
	}

	campaignID, err := uuid.Parse(query.CampaignID)
	if err != nil {
		svc.logger.Err(err).Msg("invalid campaign id")
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid query parameters"))
		return
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	orm := db.New(svc.conn)
	mappings, err := orm.GetQuestionMappingsByCampaignID(c.Request.Context(), db.GetQuestionMappingsByCampaignIDParams{
		CampaignID: campaignID,
		Limit:      int32(query.Limit),
		Offset:     int32(query.Offset),
	})
	if err != nil {
		svc.logger.Err(err).Msg("failed to get question mappings")
		c.AbortWithError(http.StatusInternalServerError, errors.New("something went wrong"))
		return
	}

	c.JSON(http.StatusOK, GetQuestionMappingsResp{QuestionMappings: mappings})
}

type UpdateQuestionMappingURI struct {
	ID uuid.UUID `uri:"id" binding:"required"`
}

type UpdateQuestionMappingBody struct {
	QuestionID uuid.UUID `json:"question_id" binding:"required"`
	CampaignID uuid.UUID `json:"campaign_id" binding:"required"`
	OrgID      uuid.UUID `json:"org_id" binding:"required"`
}

func (svc *ApiV1Service) UpdateQuestionMapping(c *gin.Context) {
	orm := db.New(svc.conn)

	var uri UpdateQuestionMappingURI
	if err := c.ShouldBindUri(&uri); err != nil {
		svc.logger.Err(err).Msg("invalid URI parameters")
		c.AbortWithError(http.StatusBadRequest, errors.New("invalid question mapping ID"))
		return
	}

	var in UpdateQuestionMappingBody
	if err := c.BindJSON(&in); err != nil {
		svc.logger.Err(err).Msg("failed to parse request body")
		c.AbortWithError(http.StatusBadRequest, errors.New("bad request body"))
		return
	}

	qm, err := orm.UpdateQuestionMappingsByID(c.Request.Context(), db.UpdateQuestionMappingsByIDParams{
		ID:         uri.ID,
		QuestionID: in.QuestionID,
		CampaignID: in.CampaignID,
		OrgID:      in.OrgID,
	})
	if err != nil {
		svc.logger.Err(err).Msg("failed to update question_mapping")
		c.AbortWithError(http.StatusInternalServerError, errors.New("something went wrong"))
		return
	}

	c.JSON(http.StatusOK, qm)
}
