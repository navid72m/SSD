package question_controller

import (
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
)

type GetQuestionCategoryStatRequest struct {
	CategoryId int64 `form:"category_id" json:"category_id" binding:"required"`
}

type GetQuestionCategoryStatResponse struct {
	Statistics []*question_repository.CategoryStat `json:"statistics"` // default 0
}

func GetQuestionCategoryStat(c *gin.Context) {
	var request GetQuestionCategoryStatRequest
	var response GetQuestionCategoryStatResponse
	var err error
	if err = c.Bind(&request); err != nil {
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	var stats []*question_repository.CategoryStat
	questionRepository := question_repository.QuestionRepository{Context: c}

	if request.CategoryId != -1 && request.CategoryId != 1 {
		stats, err = questionRepository.GetCategoryStatsById(request.CategoryId)
		if err != nil {
			error_handler.PrivateError(c, err, "Error while fetching questions size")
			return
		}
	} else if request.CategoryId == -1 { //total count for all questions
		stats, err = questionRepository.GetCategoryStatsTotal()
		if err != nil {
			error_handler.PrivateError(c, err, "Error while fetching questions size")
			return
		}

	} else if request.CategoryId == 1 { //total count for all categories
		stats, err = questionRepository.GetAllCategoryStats()
		if err != nil {
			error_handler.PrivateError(c, err, "Error while fetching questions size")
			return
		}
	}

	response.Statistics = stats

	c.JSON(200, gin.H{
		"status": true,
		"data":   response,
	})
}
