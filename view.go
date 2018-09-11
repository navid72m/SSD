package question_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

type ViewRequest struct {
	Id int64 `form:"id" json:"id" binding:"required"`
}

type ViewResponse struct {
	QuestionDetails *question_repository.QuestionDetailResult `json:"question_details"`
}

func View(c *gin.Context) {
	var response ViewResponse
	var request ViewRequest
	if err := c.Bind(&request); err != nil { // if the input data does not match the requirements of the struct (here user_id is required) it  will return an error
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	questionRepository := question_repository.QuestionRepository{Context: c}
	result, err := questionRepository.GetQuestionDetailsById(request.Id)
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

