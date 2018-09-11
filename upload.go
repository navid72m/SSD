package question_confirmation_controller

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/gocarina/gocsv"

	"encoding/csv"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_model "git.palang.co/qok/qok-server-ng/go/question/entity/model"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
	"git.palang.co/qok/qok-server-ng/go/service/vitess_driver"
	"gopkg.in/gin-gonic/gin.v1"
)

type UploadResponse struct {
	CorruptedCount int64 `json:"corrupted_counter"` // default 0
	AcceptedCount  int64 `json:"accepted_counter"`
	RejectedCount  int64 `json:"rejected_counter"`
}

func UploadQuestionsForConfirmation(c *gin.Context) {

	var response UploadResponse

	uploadAt := c.Request.PostFormValue("upload_at")
	reviewedAt := c.Request.PostFormValue("reviewed_at")
	categoryId, _ := strconv.ParseInt(c.Request.PostFormValue("category_id"), 10, 64)

	// timeUploadAt, _ := time.Parse("2017-01-04 22:33:07", uploadAt)
	// timeReviewedAt, _ := time.Parse("2017-01-04 22:33:07", reviewedAt)

	fmt.Printf("%s %s \n", uploadAt, reviewedAt)

	c.Request.ParseMultipartForm(32 << 20)
	file, header, err := c.Request.FormFile("file")
	filename := header.Filename
	if err != nil {

		error_handler.PrivateError(c, err, "error while reading from request")
	}

	pwd, err := os.Getwd()

	if err != nil {

		error_handler.PrivateError(c, err, "Error while getting current directory!")
	}

	out, err := os.Create(pwd + "/" + filename)

	if err != nil {

		//	log.Fatal(err)
		error_handler.PrivateError(c, err, "Error while creating the confirmation question file!")
	}
	//defer out.Close()

	_, err = io.Copy(out, file)

	if err != nil {

		error_handler.PrivateError(c, err, "Error while creating the confirmation question file!")

	}
	out.Close()
	var status bool
	var corruptedCount int64
	var acceptedCount int64
	var rejectedCount int64
	var corruptedQuestions []*question_repository.QuestionDetailForUpload
	corruptedQuestions, corruptedCount, acceptedCount, rejectedCount, err = ProcessUploadedQuestionsForConfirmation(c, filename, categoryId)
	if err != nil {
		status = false
	} else {
		status = true
	}
	response.CorruptedCount = corruptedCount
	response.AcceptedCount = acceptedCount
	response.RejectedCount = rejectedCount

	csvContent, err2 := gocsv.MarshalString(&corruptedQuestions)
	if err2 != nil {
		error_handler.PrivateError(c, err2, "Error marshaling questions to csv")
		return
	}
	c.JSON(200, gin.H{
		"status":    status,
		"data":      response,
		"corrupted": csvContent,
	})

	// c.Header("Content-Description", "File Transfer")
	// c.Header("Content-Disposition", "attachment; filename=questions.csv")
	// c.Data(http.StatusOK, "text/csv", []byte(csvContent))

}

func ProcessUploadedQuestionsForConfirmation(c *gin.Context, filename string, category_id int64) (corrupted []*question_repository.QuestionDetailForUpload, notEqualwithDb int64, acceptedCount int64, rejectedCount int64, err error) {

	start := time.Now()
	var err_id, err_corAns, err_accepted, err_reason error
	pwd, err := os.Getwd()

	if err != nil {

		error_handler.PrivateError(c, err, "Error while getting current directory!")
	}

	questions := make([]*question_repository.QuestionDetailForUpload, 1000)
	var CorruptedQuestions []*question_repository.QuestionDetailForUpload
	questionRepository := question_repository.QuestionRepository{Context: c}
	// var CorruptedCount int64 = 0
	var NotEqualWithDbCount int64 = 0

	// m := make(map[string][]int64)
	var accepted []int64
	var acceptedCategoryId []int64
	var rejected []int64
	var rejectedCategoryId []int64

	var AcceptedQuestionIds string = ""
	var RejectedQuestionIds string = ""

	var AcceptedCount int64 = 0
	var RejectedCount int64 = 0

	var questionArrIds []int64
	var statuses []string
	var previous_statuses []string
	var reasonIds []int64

	var UploadStatusQuestionStatusChanged string = "QuestionStatusChanged"
	var UploadStatusAddedToRedis string = "QuestionsAddedToRedis"
	// var UploadStatusCoinsRewarded string = "CoinsRewarded"
	// var UploadStatusFinished string = "Finished"

	status, err := questionRepository.GetUploadStat(filename)
	fmt.Printf("status is : %s", status)
	elapsed := time.Since(start)
	fmt.Printf("time after getting stat: %s\n", elapsed)
	conn := vitess_driver.FromContext(c)
	transactionTrue := true
	txConn, err := conn.Begin(c)
	fmt.Println("time")
	fmt.Println("2.transaction started")

	if err != nil {
		fmt.Println("2.1 transaction ruined")
		error_handler.PrivateError(c, err, "Error while creating transaction")
	}
	if status == "" {

		questionsFile, errr := os.OpenFile(pwd+"/"+filename, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if errr != nil {

			error_handler.PrivateError(c, err, "Error while creating file!")
		}

		fmt.Println("filename is:")
		fmt.Println(filename)
		reader := csv.NewReader(questionsFile)
		record, err := reader.ReadAll()

		if err != nil {
			error_handler.PrivateError(c, err, "Error while reading from file!")
		}
		questionIds := ""
		for index, row := range record {
			if index == 1 {
				questionIds = questionIds + row[0]
			} else if index > 1 {
				questionIds = questionIds + ", " + row[0]
			}
		}
		// fmt.Println("this is question Ids")
		// fmt.Println(questionIds)
		fmt.Println("Getting question Details")
		elapsed = time.Since(start)
		questionDB, errDB := questionRepository.GetQuestionsDetailsById(questionIds)
		tElapsed := time.Since(start)
		tElapsed = tElapsed - elapsed
		fmt.Printf("time for getting question details %s\n", tElapsed)

		if errDB != nil {
			// fmt.Println("error in database is happening")
			// fmt.Println(errDB)
			error_handler.PrivateError(c, err, "Error while fetching from DB!")
		}
		// CorruptedCount := 0

		// lineCount := 0

		for index, value := range record {

			if index > 0 {

				// fmt.Printf("this is index %d \n", index)
				// fmt.Printf("this is index %s \n", value)
				// fmt.Printf("this is id %s \n", value[0])

				question := new(question_repository.QuestionDetailForUpload)

				question.Id, err_id = strconv.ParseInt(value[0], 10, 64)

				question.Question = string(value[1])

				question.Choice1 = string(value[2])

				question.Choice2 = string(value[3])

				question.Choice3 = string(value[4])

				question.Choice4 = string(value[5])

				question.CorrectAnswer, err_corAns = strconv.ParseInt(value[6], 10, 64)

				question.Category = string(value[7])

				// fmt.Printf("this is category id %d \n", questionDB[index-1].CategoryId)

				question.CategoryId = questionDB[index-1].CategoryId

				// fmt.Printf("this is question category id %d \n", question.CategoryId)

				question.Accepted, err_accepted = strconv.ParseInt(value[8], 10, 64)

				question.ReasonId, err_reason = strconv.ParseInt(value[9], 10, 64)

				questionTextDB, _ := strconv.Unquote(questionDB[index-1].Question)
				questionTextFile, _ := strconv.Unquote(question.Question)

				choice1TextDB, _ := strconv.Unquote(questionDB[index-1].Choice1)
				choice1File, _ := strconv.Unquote(question.Choice1)

				choice2TextDB, _ := strconv.Unquote(questionDB[index-1].Choice2)
				choice2File, _ := strconv.Unquote(question.Choice2)

				choice3TextDB, _ := strconv.Unquote(questionDB[index-1].Choice3)
				choice3File, _ := strconv.Unquote(question.Choice3)

				choice4TextDB, _ := strconv.Unquote(questionDB[index-1].Choice4)
				choice4File, _ := strconv.Unquote(question.Choice4)

				if err_id == nil && err_corAns == nil && err_accepted == nil && err_reason == nil && question.Accepted == 0 {

					if question.Id != questionDB[index-1].Id || questionTextFile != questionTextDB || choice1TextDB != choice1File || choice2File != choice2TextDB || choice3File != choice3TextDB || choice4File != choice4TextDB || question.CorrectAnswer != questionDB[index-1].CorrectAnswer {
						NotEqualWithDbCount++
						questions[index-1] = nil
						CorruptedQuestions = append(CorruptedQuestions, question)
					} else {

						questions[index-1] = question

					}

				} else if err_id == nil && err_corAns == nil && err_accepted == nil && question.Accepted == 1 {

					if question.Id != questionDB[index-1].Id || questionTextFile != questionTextDB || choice1TextDB != choice1File || choice2File != choice2TextDB || choice3File != choice3TextDB || choice4File != choice4TextDB || question.CorrectAnswer != questionDB[index-1].CorrectAnswer {
						NotEqualWithDbCount++
						questions[index-1] = nil
						CorruptedQuestions = append(CorruptedQuestions, question)
						fmt.Printf("id is:%d\n", question.Id)
					} else {
						questions[index-1] = question

					}

				} else if err_id != nil && err_corAns != nil && err_accepted != nil {

					fmt.Printf("id is:%d\n", question.Id)
					questions[index-1] = nil
					CorruptedQuestions = append(CorruptedQuestions, question)
					NotEqualWithDbCount++

				}
			}

		}

		for _, value := range questions {

			if value != nil {

				if value.Accepted == 1 {

					questionArrIds = append(questionArrIds, value.Id)
					accepted = append(accepted, value.Id)
					acceptedCategoryId = append(acceptedCategoryId, value.CategoryId)

					if AcceptedQuestionIds == "" {
						AcceptedQuestionIds = AcceptedQuestionIds + strconv.FormatInt(value.Id, 10)
					} else {
						AcceptedQuestionIds = AcceptedQuestionIds + "," + strconv.FormatInt(value.Id, 10)
					}

					statuses = append(statuses, question_model.QuestionStatusAccepted)
					previous_statuses = append(previous_statuses, question_model.QuestionStatusWaitingForConfirmation)
					reasonIds = append(reasonIds, -1)
					AcceptedCount++

				} else if value.Accepted == 0 {

					questionArrIds = append(questionArrIds, value.Id)
					rejected = append(rejected, value.Id)
					rejectedCategoryId = append(rejectedCategoryId, value.CategoryId)

					if RejectedQuestionIds == "" {
						RejectedQuestionIds = RejectedQuestionIds + strconv.FormatInt(value.Id, 10)
					} else {
						RejectedQuestionIds = RejectedQuestionIds + "," + strconv.FormatInt(value.Id, 10)
					}

					statuses = append(statuses, question_model.QuestionStatusRejected)
					previous_statuses = append(previous_statuses, question_model.QuestionStatusWaitingForConfirmation)
					reasonIds = append(reasonIds, value.ReasonId)
					RejectedCount++
				}
			}
		}

		elapsed = time.Since(start)
		fmt.Println("starting question change status")
		for i, questionId := range questionArrIds {
			err = questionRepository.MultipleChangeStatus(txConn, questionId, previous_statuses[i], statuses[i], reasonIds[i], 1)

			if err != nil {
				error_handler.PrivateError(c, err, "Error while changing status of questions")
				transactionTrue = false
			}

		}

		if transactionTrue == true {
			fmt.Println("Change status transaction committed")
			txConn.Commit(c)
		} else {
			fmt.Println("Change status transaction rollbacked")
			txConn.Rollback(c)
		}
		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("time for changing question status %s\n", tElapsed)

		err = questionRepository.SetUploadStat(filename, UploadStatusQuestionStatusChanged)
		if err != nil {
			fmt.Println("Error While setting upload status after  changing status")
			fmt.Println(err)
		}

	}

	status, err = questionRepository.GetUploadStat(filename)

	if err != nil {
		fmt.Println("Error While getting upload status after  changing status")
		fmt.Println(err)
	}

	if status == UploadStatusQuestionStatusChanged {

		for index, questionId := range accepted {
			result_redis, err_redis := questionRepository.AddAcceptedQuestionToRedis(questionId, acceptedCategoryId[index])
			if err_redis != nil || result_redis != true {
				error_handler.PrivateError(c, err, "Error while adding accepted question to redis")
			}
		}

		for index, questionId := range rejected {
			result_redis, err_redis := questionRepository.DeleteRejectedQuestionFromRedis(questionId, rejectedCategoryId[index])
			if err_redis != nil || result_redis != true {
				error_handler.PrivateError(c, err, "Error while deleting question from redis")
			}
		}

		err = questionRepository.SetUploadStat(filename, UploadStatusAddedToRedis)
		if err != nil {
			fmt.Println("Error While setting upload status after adding to redis")
			fmt.Println(err)
		}

	}

	status, err = questionRepository.GetUploadStat(filename)

	if err != nil {
		fmt.Println("Error While getting upload status after adding to redis")
		fmt.Println(err)
	}

	if status == UploadStatusAddedToRedis {

		for _, questionId := range questionArrIds {
			err = questionRepository.PushQuestionToRedis(questionId)

			if err != nil {
				fmt.Printf("Error happend while pushing questions to reward queue %s\n", err)
			}

		}

		err = questionRepository.SetHamisunLogs(filename, category_id, NotEqualWithDbCount, AcceptedCount, RejectedCount)
		fmt.Println(err)

	}

	return CorruptedQuestions, NotEqualWithDbCount, AcceptedCount, RejectedCount, err

}
