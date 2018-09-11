package question_confirmation_controller

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/mock/fakevitess"
	"git.palang.co/qok/qok-server-ng/go/service/redis_driver"
	"git.palang.co/qok/qok-server-ng/go/test/testredis"
)

func TestUploadQuestionsForConfirmation(t *testing.T) {
	Convey("Given we have a web server with a download API, we want to make sure it returns the correct results\n",
		t, func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.Use(fakevitess.Middleware())
			router.Use(testredis.Middleware(redis_driver.DEFAULT_REDIS))
			router.POST("/question/confirmation/upload", UploadQuestionsForConfirmation)

			bodyBuf := &bytes.Buffer{}
			bodyWriter := multipart.NewWriter(bodyBuf)

			// this step is very important
			filename := "./question.csv"
			fileWriter, err := bodyWriter.CreateFormFile("file", "question.csv")

			So(err, ShouldBeNil)
			// open file handle
			fh, err := os.Open(filename)

			So(err, ShouldBeNil)

			//iocopy
			_, err = io.Copy(fileWriter, fh)

			So(err, ShouldBeNil)

			contentType := bodyWriter.FormDataContentType()

			req, err := http.NewRequest("POST", "/question/confirmation/upload", bodyBuf)
			req.Header.Add("Content-Type", contentType)
			req.Header.Add("Content-Length", strconv.Itoa(bodyBuf.Len()))
			bodyWriter.Close()
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			So(err, ShouldBeNil)
			So(resp.Code, ShouldEqual, http.StatusOK)

		})
}

func TestProcessUploadedQuestionsForConfirmation(t *testing.T) {
	Convey("Given we have a function, we want to make sure it returns the correct results\n",
		t, func() {
			gin.SetMode(gin.ReleaseMode)
			context := gin.Context{}
			conn := fakevitess.New()
			context.Set("vitess", conn)
			redis := testredis.New()
			context.Set(redis_driver.DEFAULT_REDIS, redis)

			filename := "question.csv"
			var categoryId int64 = 1

			var uploadAt time.Time
			uploadAt = time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)

			corruptedCount, corrupted_db, err := ProcessUploadedQuestionsForConfirmation(&context, filename, categoryId, uploadAt, uploadAt)

			So(err, ShouldBeNil)
			So(corruptedCount, ShouldEqual, 0)
			So(corrupted_db, ShouldEqual, 0)
			//So(resp.Code, ShouldEqual, http.StatusOK)

		})
}
