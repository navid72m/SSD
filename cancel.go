package question_confirmation_controller

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"encoding/csv"

	"git.palang.co/qok/qok-server-ng/go/qokapp/errors"
	question_model "git.palang.co/qok/qok-server-ng/go/question/entity/model"
	question_repository "git.palang.co/qok/qok-server-ng/go/question/entity/repository"
	"git.palang.co/qok/qok-server-ng/go/service/vitess_driver"
	user_repository "git.palang.co/qok/qok-server-ng/go/user/entity/repository"
	"gopkg.in/gin-gonic/gin.v1"
)

// type UploadResponse struct {
// 	CorruptedCount      int64 `json:"corrupted_counter"` // default 0
// 	NotEqualWithDbCount int64 `json:"db_corrupted_counter"`
// }
type CancelRequest struct {
	FileName string `json:"file_name"`
}

func CancelQuestionsForConfirmation(c *gin.Context) {

	var request CancelRequest
	var response UploadResponse

	if err := c.Bind(&request); err != nil { // if the input data does not match the requirements of the struct (here user_id is required) it  will return an error
		error_handler.LogicError(c, error_handler.WrongInput)
		return
	}

	filename := request.FileName

	var err error
	var status bool
	var corruptedCount int64
	// var notEqualwithDbCount int64
	corruptedCount, _, err = CancelConfirmation(c, filename)
	if err != nil {
		status = false
	} else {
		status = true
	}
	response.CorruptedCount = corruptedCount
	// response.NotEqualWithDbCount = notEqualwithDbCount
	c.JSON(200, gin.H{
		"status": status,
		"data":   response,
	})

}

func CancelConfirmation(c *gin.Context, filename string) (corrupted int64, notEqualwithDb int64, err error) {

	start := time.Now()
	var err_id, err_corAns, err_accepted, err_reason error
	pwd, err := os.Getwd()

	if err != nil {

		error_handler.PrivateError(c, err, "Error while getting current directory!")
	}

	questions := make([]*question_repository.QuestionDetailForUpload, 1000)
	questionRepository := question_repository.QuestionRepository{Context: c}
	var CorruptedCount int64 = 0
	var NotEqualWithDbCount int64 = 0
	userRepository := user_repository.UserRepository{Context: c}
	userStatRepository := user_repository.UserStatsRepository{Context: c}

	// m := make(map[string][]int64)
	var accepted []int64
	var acceptedCategoryId []int64
	var rejected []int64
	var rejectedCategoryId []int64

	var AcceptedQuestionIds string = ""
	var RejectedQuestionIds string = ""

	var questionArrIds []int64
	var statuses []string
	var previous_statuses []string
	var reasonIds []int64

	var creator_coins int64 = -30
	var reviewer_coins int64 = -3

	var UploadStatusQuestionStatusChanged string = "QuestionStatusChanged"
	var UploadStatusAddedToRedis string = "QuestionsAddedToRedis"
	var UploadStatusCoinsRewarded string = "CoinsRewarded"
	var UploadStatusFinished string = "Finished"
	var UploadStatusCancelled string = "Cancelled"

	userChangeStats := make(map[int64]user_repository.UserQuestionStat)
	userChangeCoins := make(map[int64]int64)

	counter := 0

	status, err := questionRepository.GetUploadStat(filename)
	fmt.Printf("status is : %s", status)

	err = questionRepository.SetUploadStat(filename, UploadStatusCancelled)
	if err != nil {
		fmt.Println(err)
	}

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

	if status == UploadStatusFinished || status == UploadStatusCoinsRewarded || status == UploadStatusAddedToRedis || status == UploadStatusQuestionStatusChanged {
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

				if err_id == nil && err_corAns == nil && err_accepted == nil && err_reason == nil && question.Accepted == 0 {

					if question.Id != questionDB[index-1].Id || question.Question != questionDB[index-1].Question || question.Choice1 != questionDB[index-1].Choice1 || question.Choice2 != questionDB[index-1].Choice2 || question.Choice3 != questionDB[index-1].Choice3 || question.Choice4 != questionDB[index-1].Choice4 || question.CorrectAnswer != questionDB[index-1].CorrectAnswer {
						NotEqualWithDbCount++
						questions[index-1] = nil
					} else {
						// fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
						questions[index-1] = question
						// fmt.Println(index - 1)
						// fmt.Print(questions[index-1])
					}

				} else if err_id == nil && err_corAns == nil && err_accepted == nil && question.Accepted == 1 {

					if question.Id != questionDB[index-1].Id || question.Question != questionDB[index-1].Question || question.Choice1 != questionDB[index-1].Choice1 || question.Choice2 != questionDB[index-1].Choice2 || question.Choice3 != questionDB[index-1].Choice3 || question.Choice4 != questionDB[index-1].Choice4 || question.CorrectAnswer != questionDB[index-1].CorrectAnswer {
						NotEqualWithDbCount++
						questions[index-1] = nil
						fmt.Printf("id is:%d\n", question.Id)
					} else {
						questions[index-1] = question
						// fmt.Println("BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
						// fmt.Println(index - 1)
						// fmt.Print(questions[index-1])
					}

				} else if err_id != nil && err_corAns != nil && err_accepted != nil {
					// fmt.Println("CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC")
					fmt.Printf("id is:%d\n", question.Id)
					questions[index-1] = nil
					CorruptedCount++

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
					reasonIds = append(reasonIds, int64(0))

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
					reasonIds = append(reasonIds, int64(0))
				}
			}
		}

		elapsed = time.Since(start)
		fmt.Println("starting question change status")
		for i, questionId := range questionArrIds {
			err = questionRepository.MultipleChangeStatus(txConn, questionId, statuses[i], previous_statuses[i], reasonIds[i], 0)

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

	}
	// for index, questionId := range accepted {
	// 	result_redis, err_redis := questionRepository.AddAcceptedQuestionToRedis(questionId, acceptedCategoryId[index])
	// 	if err_redis != nil || result_redis != true {
	// 		error_handler.PrivateError(c, err, "Error while adding accepted question to redis")
	// 	}
	// }

	if status == UploadStatusFinished || status == UploadStatusCoinsRewarded || status == UploadStatusAddedToRedis {

		for index, questionId := range accepted {
			result_redis, err_redis := questionRepository.DeleteRejectedQuestionFromRedis(questionId, acceptedCategoryId[index])
			if err_redis != nil || result_redis != true {
				error_handler.PrivateError(c, err, "Error while deleting question from redis")
			}
		}

	}

	if status == UploadStatusFinished || status == UploadStatusCoinsRewarded {
		elapsed = time.Since(start)
		RejecteCreatorIds, err := questionRepository.GetCreatedByIds(RejectedQuestionIds)

		if err != nil {
			error_handler.PrivateError(c, err, "Error while getting Creator Ids")
		}

		tElapsed := time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Getting rejected question creators %s\n", tElapsed)

		elapsed = time.Since(start)
		AcceptedCreatorIds, err := questionRepository.GetCreatedByIds(AcceptedQuestionIds)

		if err != nil {
			error_handler.PrivateError(c, err, "Error while getting Creator Ids")
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Getting Accepted question creators %s\n", tElapsed)

		for creatorId, questionIds := range AcceptedCreatorIds {
			for _ = range questionIds {
				if _, ok := userChangeCoins[creatorId]; ok {
					userChangeCoins[creatorId] += creator_coins
				} else {
					userChangeCoins[creatorId] = creator_coins

				}

				if _, ok := userChangeStats[creatorId]; ok {
					userStat := userChangeStats[creatorId]
					userStat.AcceptedCount--
					userStat.InprogressCount++
					userChangeStats[creatorId] = userStat
				} else {
					var zero int64 = 0
					userStat := user_repository.UserQuestionStat{
						AcceptedCount:           zero,
						RejectedCount:           zero,
						InprogressCount:         zero,
						WrongReviewsCount:       zero,
						CorrectReviewsCount:     zero,
						TotalReviewsCount:       zero,
						LastWrongReviewsCount:   zero,
						LastCorrectReviewsCount: zero}

					userStat.AcceptedCount--
					userStat.InprogressCount++
					userChangeStats[creatorId] = userStat
				}

			}
		}

		for creatorId, questionIds := range RejecteCreatorIds {
			for _ = range questionIds {
				if _, ok := userChangeStats[creatorId]; ok {
					// userChangeCoins[creatorId] += creator_coins
					userStat := userChangeStats[creatorId]
					userStat.RejectedCount--
					userStat.InprogressCount++
					userChangeStats[creatorId] = userStat
				} else {
					// userChangeCoins[creatorId] = creator_coins
					var zero int64 = 0
					userStat := user_repository.UserQuestionStat{
						AcceptedCount:           zero,
						RejectedCount:           zero,
						InprogressCount:         zero,
						WrongReviewsCount:       zero,
						CorrectReviewsCount:     zero,
						TotalReviewsCount:       zero,
						LastWrongReviewsCount:   zero,
						LastCorrectReviewsCount: zero}
					userStat.RejectedCount--
					userStat.InprogressCount++
					userChangeStats[creatorId] = userStat
				}
			}

		}

		elapsed = time.Since(start)
		AcceptedReviewersIds, err := questionRepository.GetReviewersId(AcceptedQuestionIds, question_model.QuestionStatusAccepted)

		if err != nil {
			error_handler.PrivateError(c, err, "Error while fetching Reviewers ")
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Get accepted reviewersIds %s \n", tElapsed)

		elapsed = time.Since(start)
		RejectedReviewersIds, err := questionRepository.GetReviewersId(RejectedQuestionIds, question_model.QuestionStatusRejected)

		if err != nil {
			error_handler.PrivateError(c, err, "Error while fetching Reviewers ")
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Get rejected reviewersIds %s \n", tElapsed)

		for _, reviewStats := range AcceptedReviewersIds {
			for _, reviewStat := range reviewStats {

				if _, ok := userChangeStats[reviewStat.ReviewerId]; ok {
					// userChangeCoins[creatorId] += creator_coins
					userStat := userChangeStats[reviewStat.ReviewerId]
					userStat.CorrectReviewsCount--
					userStat.LastCorrectReviewsCount--
					userStat.TotalReviewsCount--
					// userStat.InprogressCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				} else {
					// userChangeCoins[creatorId] = creator_coins
					var zero int64 = 0
					userStat := user_repository.UserQuestionStat{
						AcceptedCount:           zero,
						RejectedCount:           zero,
						InprogressCount:         zero,
						WrongReviewsCount:       zero,
						CorrectReviewsCount:     zero,
						TotalReviewsCount:       zero,
						LastWrongReviewsCount:   zero,
						LastCorrectReviewsCount: zero}

					userStat.CorrectReviewsCount--
					userStat.LastCorrectReviewsCount--
					userStat.TotalReviewsCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				}

			}
		}

		// for userId, userStat := range userChangeStats {
		// 	fmt.Printf("userId is : %d, correctReviewCount is : %d\n", userId, userStat.CorrectReviewsCount)
		// }

		for _, reviewStats := range RejectedReviewersIds {
			for _, reviewStat := range reviewStats {
				if _, ok := userChangeStats[reviewStat.ReviewerId]; ok {
					// userChangeCoins[creatorId] += creator_coins
					userStat := userChangeStats[reviewStat.ReviewerId]
					userStat.CorrectReviewsCount--
					userStat.LastCorrectReviewsCount--
					userStat.TotalReviewsCount--
					// userStat.InprogressCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				} else {
					// userChangeCoins[creatorId] = creator_coins
					var zero int64 = 0
					userStat := user_repository.UserQuestionStat{
						AcceptedCount:           zero,
						RejectedCount:           zero,
						InprogressCount:         zero,
						WrongReviewsCount:       zero,
						CorrectReviewsCount:     zero,
						TotalReviewsCount:       zero,
						LastWrongReviewsCount:   zero,
						LastCorrectReviewsCount: zero}

					userStat.CorrectReviewsCount--
					userStat.LastCorrectReviewsCount--
					userStat.TotalReviewsCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				}
			}

		}

		for _, ReviewRecords := range AcceptedReviewersIds {
			for _, reviewRecord := range ReviewRecords {
				if _, ok := userChangeCoins[reviewRecord.ReviewerId]; ok {
					userChangeCoins[reviewRecord.ReviewerId] += reviewer_coins
				} else {
					userChangeCoins[reviewRecord.ReviewerId] = reviewer_coins
				}
			}
		}

		for _, ReviewRecords := range RejectedReviewersIds {
			for _, reviewRecord := range ReviewRecords {
				if _, ok := userChangeCoins[reviewRecord.ReviewerId]; ok {
					userChangeCoins[reviewRecord.ReviewerId] += reviewer_coins
				} else {
					userChangeCoins[reviewRecord.ReviewerId] = reviewer_coins
				}
			}
		}

		txConn, err = conn.Begin(c)
		// fmt.Println("2.transaction started")
		transactionTrue = true
		if err != nil {
			fmt.Println("2.1 transaction ruined")
			error_handler.PrivateError(c, err, "Error while creating transaction")
		}

		elapsed = time.Since(start)
		var offset int64 = 0
		var redisOffset int64 = 0
		redisOffset, err = questionRepository.GetUploadOffset(filename, 0)
		if err != nil {
			fmt.Println("Error While getting offset coins")
			fmt.Printf("redis offset %d\n", redisOffset)
			fmt.Println(err)
		}

		for userId, userCoins := range userChangeCoins {

			if counter < 500 {
				err = userRepository.IncreaseUserCoinsArray(txConn, userCoins, userId)

				if err != nil {
					// error_handler.PrivateError(c, err, "Error while increasing user coins")
					fmt.Println("3 transaction problem")
					fmt.Printf("Problem happening for changing user:%d coins.\n", userId)
					transactionTrue = false
				}
			} else if counter == 500 {

				if transactionTrue == true {
					// fmt.Println("5 transaction committed")

					if offset <= redisOffset {
						txConn.Commit(c)

					}
					offset++
				} else {
					// fmt.Println("5 transaction rollbacked")
					txConn.Rollback(c)
				}
				transactionTrue = true
				counter = 0

				txConn, err = conn.Begin(c)
				// fmt.Println("6.transaction started")
				if err != nil {
					// fmt.Println("6.1 transaction ruined")
					error_handler.PrivateError(c, err, "Error while creating transaction")
				}

			}

			counter += 1

		}

		if counter > 0 {
			if transactionTrue == true {

				// fmt.Println("5 transaction committed")
				redisOffset, err = questionRepository.GetUploadOffset(filename, 0)
				if err != nil {
					fmt.Println(err)
				}

				if offset <= redisOffset {
					txConn.Commit(c)

				}

			} else {
				// fmt.Println("5 transaction rollbacked")
				txConn.Rollback(c)
			}
			transactionTrue = true
			counter = 0
		}
		err = questionRepository.SetUploadOffset(filename, -1, 0)
		if err != nil {
			fmt.Println("Error While getting offset coins")
			fmt.Println(err)
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Giving userCoins time %s \n", tElapsed)

		elapsed = time.Since(start)
		WrongReviewersIdsA, err := questionRepository.GetReviewersId(AcceptedQuestionIds, question_model.QuestionStatusRejected)

		if err != nil {
			// fmt.Println(err)
			error_handler.PrivateError(c, err, "Error while fetching Reviewers")
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Get wrong reviewersIds %s \n", tElapsed)

		for _, reviewStats := range WrongReviewersIdsA {
			for _, reviewStat := range reviewStats {
				if _, ok := userChangeStats[reviewStat.ReviewerId]; ok {
					// userChangeCoins[creatorId] += creator_coins
					userStat := userChangeStats[reviewStat.ReviewerId]
					userStat.WrongReviewsCount--
					userStat.LastWrongReviewsCount--
					userStat.TotalReviewsCount--
					// userStat.InprogressCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				} else {
					// userChangeCoins[creatorId] = creator_coins
					var zero int64 = 0
					userStat := user_repository.UserQuestionStat{
						AcceptedCount:           zero,
						RejectedCount:           zero,
						InprogressCount:         zero,
						WrongReviewsCount:       zero,
						CorrectReviewsCount:     zero,
						TotalReviewsCount:       zero,
						LastWrongReviewsCount:   zero,
						LastCorrectReviewsCount: zero}

					userStat.WrongReviewsCount--
					userStat.LastWrongReviewsCount--
					userStat.TotalReviewsCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				}
			}

		}

		elapsed = time.Since(start)

		WrongReviewersIdsB, err := questionRepository.GetReviewersId(RejectedQuestionIds, question_model.QuestionStatusAccepted)

		if err != nil {
			error_handler.PrivateError(c, err, "Error while fetching Reviewers")
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Get wrong reviewersIds %s \n", tElapsed)

		for _, reviewStats := range WrongReviewersIdsB {
			for _, reviewStat := range reviewStats {
				if _, ok := userChangeStats[reviewStat.ReviewerId]; ok {
					// userChangeCoins[creatorId] += creator_coins
					userStat := userChangeStats[reviewStat.ReviewerId]
					userStat.WrongReviewsCount--
					userStat.LastWrongReviewsCount--
					userStat.TotalReviewsCount--
					// userStat.InprogressCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				} else {
					// userChangeCoins[creatorId] = creator_coins
					var zero int64 = 0
					userStat := user_repository.UserQuestionStat{
						AcceptedCount:           zero,
						RejectedCount:           zero,
						InprogressCount:         zero,
						WrongReviewsCount:       zero,
						CorrectReviewsCount:     zero,
						TotalReviewsCount:       zero,
						LastWrongReviewsCount:   zero,
						LastCorrectReviewsCount: zero}

					userStat.WrongReviewsCount--
					userStat.LastWrongReviewsCount--
					userStat.TotalReviewsCount--
					userChangeStats[reviewStat.ReviewerId] = userStat
				}
			}

		}

		elapsed = time.Since(start)

		txConn, err = conn.Begin(c)
		// fmt.Println("2.transaction started")
		transactionTrue = true
		if err != nil {
			// fmt.Println("2.1 transaction ruined")
			error_handler.PrivateError(c, err, "Error while creating transaction")
		}

		offset = 0
		redisOffset, err = questionRepository.GetUploadOffset(filename, 1)
		if err != nil {
			fmt.Println("Error While getting offset stats")
			fmt.Printf("redis offset %d\n", redisOffset)
			fmt.Println(err)
		}

		for userId, userStat := range userChangeStats {

			if counter < 500 {
				err = userStatRepository.UpdateUserQuestionStats(txConn, userStat.AcceptedCount, userStat.RejectedCount, userStat.InprogressCount, userStat.WrongReviewsCount, userStat.CorrectReviewsCount, userStat.TotalReviewsCount, userStat.LastWrongReviewsCount, userStat.LastCorrectReviewsCount, userId)

				if err != nil {
					// fmt.Println("4 transaction problem")
					// fmt.Printf("Problem happening for changing user:%d stats.\n", userId)

					transactionTrue = false
				}
			} else if counter == 500 {

				if transactionTrue == true {

					if offset <= redisOffset {
						txConn.Commit(c)

					}
					offset++
				} else {
					// fmt.Println("5 transaction rollbacked")
					txConn.Rollback(c)
				}
				transactionTrue = true
				counter = 0

				txConn, err = conn.Begin(c)
				// fmt.Println("6.transaction started")
				if err != nil {
					// fmt.Println("6.1 transaction ruined")
					error_handler.PrivateError(c, err, "Error while creating transaction")
				}

			}

			counter += 1
		}

		if counter > 0 {
			if transactionTrue == true {
				// fmt.Println("5 transaction committed")
				redisOffset, err = questionRepository.GetUploadOffset(filename, 1)
				if err != nil {
					fmt.Println(err)
				}

				if offset <= redisOffset {
					txConn.Commit(c)

				}

			} else {
				// fmt.Println("5 transaction rollbacked")
				txConn.Rollback(c)
			}
			transactionTrue = true
			counter = 0
		}

		err = questionRepository.SetUploadOffset(filename, -1, 1)
		if err != nil {
			fmt.Println("Error While getting offset coins")
			fmt.Println(err)
		}

		tElapsed = time.Since(start)
		tElapsed = tElapsed - elapsed

		fmt.Printf("Updating user stats %s \n", tElapsed)

	}

	return CorruptedCount, NotEqualWithDbCount, err

}
