package question_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

type GetQuestionSizeRequest struct {
	Status     string `form:"status" json:"status" binding:"required"`
	CategoryId int64  `form:"categoryid" json:"categoryid" binding:"required"`
}

type GetQuestionSizeResponse struct {
	Counter int64 `json:"counter"` // default 0
}

func GetQuestionSize(c *gin.Context) {
	var response GetQuestionSizeResponse
	var request GetQuestionSizeRequest

	if err := c.Bind(&request); err != nil {
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	questionRepository := question_repository.QuestionRepository{Context: c}
	counter, err := questionRepository.CountQuestionsByStatus(request.Status, request.CategoryId)
	if err != nil {
		error_handler.PrivateError(c, err, "Error while fetching questions size")
		return
	}

	response.Counter = counter

	c.JSON(200, gin.H{
		"status": true,
		"data":   response,
	})
}
