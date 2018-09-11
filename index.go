package question_confirmation_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

type IndexResponse struct {
	First int64 `json:"first_id"` // default 0
	Last  int64 `json:"last_id"`
}

type IndexRequest struct {
	Page       int64 `form:"page" json:"page"` // default 0
	PageSize   int64 `form:"pagesize" json:"pagesize" binding:"required"`
	CategoryId int64 `form:"categoryid" json:"categoryid" binding:"required"`
}

func Index(c *gin.Context) {
	var response IndexResponse
	var request IndexRequest

	if err := c.Bind(&request); err != nil {
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	questionRepository := question_repository.QuestionRepository{Context: c}
	first_id, last_id, err := questionRepository.GetIdDownloadedQuestions(request.Page, request.PageSize, "WAITING_FOR_CONFIRMATION", request.CategoryId)
	if err != nil {
		error_handler.PrivateError(c, err, "Error while fetching questions size")
		return
	}

	response.First = first_id
	response.Last = last_id

	c.JSON(200, gin.H{
		"status": true,
		"data":   response,
	})
}
