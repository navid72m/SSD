package question_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

type CreatedByRequest struct {
	CreatedBy int64 `form:"created_by" json:"created_by" binding:"required"`
}

type CreatedByResponse struct {
	QuestionDetails []*question_repository.QuestionDetailResult `json:"question_details"`
}

func UserQuestionsListView(c *gin.Context) {
	var response CreatedByResponse
	var request CreatedByRequest
	if err := c.Bind(&request); err != nil { // if the input data does not match the requirements of the struct (here user_id is required) it  will return an error
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	questionRepository := question_repository.QuestionRepository{Context: c}
	result, err := questionRepository.GetQuestionDetailsByCreator(request.CreatedBy)
	if err != nil {
		error_handler.PrivateError(c, err, "Error while fetching user details by id")
		return
	}
	response.QuestionDetails = result
	c.JSON(200, gin.H{
		"status": true,
		"data":   response,
	})
}

