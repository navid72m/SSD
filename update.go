package question_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_model "git.palang.co/qok/qok-server-ng/go/question/entity/model"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

///fix-me-please-In progress
type UpdateRequest struct {
	Id             int64  `form:"id" json:"id" binding:"required"`
	Question       string `form:"question" json:"question"`
	Choice1        string `form:"choice1" json:"choice1"`
	Choice2        string `form:"choice2" json:"choice2"`
	Choice3        string `form:"choice3" json:"choice3"`
	Choice4        string `form:"choice4" json:"choice4"`
	Correct_Answer int64  `form:"correct_answer" json:"correct_answer"`
	Status         string `form:"status" json:"status"`
	CategoryId     int64  `form:"category_id" json:"category_id"`
}

type UpdateResponse struct {
	Status bool `json:"status"`
}

func Update(c *gin.Context) {
	var response UpdateResponse
	var request UpdateRequest
	var result bool
	var err error

	if err = c.Bind(&request); err != nil { // if the input data does not match the requirements of the struct (here user_id is required) it  will return an error
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}
	questionRepository := question_repository.QuestionRepository{Context: c}

	result, err = questionRepository.UpdateQuestionDetails(request.Id, request.Question, request.Choice1, request.Choice2, request.Choice3, request.Choice4, request.Correct_Answer, request.Status)

	if err != nil {
		error_handler.PrivateError(c, err, "Error while updating details!")
		return
	}

	if request.Status == question_model.QuestionStatusAccepted {
		_, err = questionRepository.AddAcceptedQuestionToRedis(request.Id, request.CategoryId)
		if err != nil {
			error_handler.PrivateError(c, err, "Error while adding accepted question to redis")
		}
	} else if request.Status == question_model.QuestionStatusRejected {
		_, err = questionRepository.DeleteRejectedQuestionFromRedis(request.Id, request.CategoryId)
		if err != nil {
			error_handler.PrivateError(c, err, "Error while deleting question from redis")
		}
	}

	response.Status = result
	c.JSON(200, gin.H{
		"status": true,
		"data":   response,
	})
}
