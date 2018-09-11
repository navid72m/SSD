package question_controller

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"git.palang.co/qok/qok-server-ng/go/mock/fakevitess"
	question_model "git.palang.co/qok/qok-server-ng/go/question/entity/model"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/gin-gonic/gin.v1"
)

type TestGetQuestionCategorySizeResponse struct {
	Status bool                             `json:"status"`
	Data   *GetQuestionCategorySizeResponse `json:"data"`
}

func TestGetCategorySize(t *testing.T) {
	Convey("Given we have a web server with a search API, we want to make sure it returns the correct results\n",
		t, func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.Use(fakevitess.Middleware())
			router.POST("/question/categorysize", GetQuestionCategorySize)

			values := url.Values{
				"status": {question_model.QuestionStatusWaitingForConfirmation},
			}

			req, err := http.NewRequest("POST", "/question/categorysize", strings.NewReader(values.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))

			So(err, ShouldEqual, nil)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			// read the JSON and check the correctness
			bytes, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			var jsonResponse TestGetQuestionCategorySizeResponse
			err = json.Unmarshal(bytes, &jsonResponse)
			So(err, ShouldBeNil)
			So(jsonResponse.Status, ShouldEqual, true)
			So(jsonResponse.Data, ShouldNotBeNil)
			So(jsonResponse.Data.CountList[0].CategoryId, ShouldEqual, 2)
			So(jsonResponse.Data.CountList[0].Count, ShouldEqual, 400)
			So(jsonResponse.Data.CountList[0].Title, ShouldEqual, "Cinema")

			So(jsonResponse.Data.CountList[1].CategoryId, ShouldEqual, 19)
			So(jsonResponse.Data.CountList[1].Count, ShouldEqual, 700)
			So(jsonResponse.Data.CountList[1].Title, ShouldEqual, "TV")
		})
}
