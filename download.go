package question_confirmation_controller

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
	"gopkg.in/gin-gonic/gin.v1"
)

type DownloadQuestionsForConfirmationRequest struct {
	Page       int64 `form:"page" json:"page"` // default 0
	PageSize   int64 `form:"pagesize" json:"pagesize" binding:"required"`
	CategoryId int64 `form:"categoryid" json:"categoryid" binding:"required"`
}

func DownloadQuestionsForConfirmation(c *gin.Context) {
	var request DownloadQuestionsForConfirmationRequest
	if err := c.Bind(&request); err != nil {
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	questionRepository := question_repository.QuestionRepository{Context: c}
	questions, err := questionRepository.DownloadQuestions(request.Page, request.PageSize, "WAITING_FOR_CONFIRMATION", request.CategoryId)
	if err != nil {
		error_handler.PrivateError(c, err, "Error while fetching questions for confirmation")
		return
	}
	records := make([][]string, 1001)

	records[0] = append(records[0], "id", "question", "choice1", "choice2", "choice3", "choice4", "correct_answer", "category", "accepted", "reason_id")

	for index, qRow := range questions {
		id := fmt.Sprintf(`%s`, strconv.FormatInt(qRow.Id, 10))
		question := fmt.Sprintf(`%s`, qRow.Question)
		choice1 := fmt.Sprintf(`%s`, qRow.Choice1)
		choice2 := fmt.Sprintf(`%s`, qRow.Choice2)
		choice3 := fmt.Sprintf(`%s`, qRow.Choice3)
		choice4 := fmt.Sprintf(`%s`, qRow.Choice4)
		correctAnswer := fmt.Sprintf(`%s`, strconv.FormatInt(qRow.CorrectAnswer, 10))
		category := fmt.Sprintf(`%s`, qRow.Category)
		accepted := fmt.Sprintf(`%s`, strconv.FormatInt(0, 10))
		reasonId := fmt.Sprintf(`%s`, strconv.FormatInt(0, 10))

		records[index+1] = append(records[index+1], id, question, choice1, choice2, choice3, choice4, correctAnswer, category, accepted, reasonId)
	}
	b := &bytes.Buffer{}   // creates IO Writer
	wr := csv.NewWriter(b) // creates a csv writer that uses the io buffer.
	wr.UseCRLF = true
	for _, qRow := range records {
		wr.Write(qRow)
	}
	wr.Flush()
	// csvContent, err2 := gocsv.MarshalString(&questions)
	// if err2 != nil {
	// 	error_handler.PrivateError(c, err2, "Error marshaling questions to csv")
	// 	return
	// }

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=questions.csv")
	c.Data(http.StatusOK, "text/csv", b.Bytes())
}
