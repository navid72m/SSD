package question_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_model "git.palang.co/qok/qok-server-ng/go/question/entity/model"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

type GetQuestionCategorySizeRequest struct {
	Status string `form:"status" json:"status" binding:"required"`
}

type GetQuestionCategorySizeResponse struct {
	CountList []*question_repository.CategorySize `json:"counter_list"` // default 0
}

func GetQuestionCategorySize(c *gin.Context) {
	var response GetQuestionCategorySizeResponse
	var request GetQuestionCategorySizeRequest

	if err := c.Bind(&request); err != nil {
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	questionRepository := question_repository.QuestionRepository{Context: c}
	counterList, err := questionRepository.GetSizeCategory(question_model.QuestionStatusWaitingForConfirmation)
	if err != nil {
		error_handler.PrivateError(c, err, "Error while fetching questions size")
		return
	}

	response.CountList = counterList

	c.JSON(200, gin.H{
		"status": true,
		"data":   response,
	})
}
