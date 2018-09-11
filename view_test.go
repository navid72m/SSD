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

	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/gin-gonic/gin.v1"

	"git.palang.co/qok/qok-server-ng/go/mock/fakevitess"
)

type TestViewResponse struct {
	Status bool          `json:"status"`
	Data   *ViewResponse `json:"data"`
}

func TestView(t *testing.T) {

	Convey("Given we have a web server with a search API, we want to make sure it returns the correct results\n",
		t, func() {

			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.Use(fakevitess.Middleware())
			router.POST("/question/view", View)

			// Convey("We try an empty request")

			values := url.Values{
				"id": {"1"},
			}

			req, err := http.NewRequest("POST", "/question/view", strings.NewReader(values.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))

			So(err, ShouldEqual, nil)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			// read the JSON and check the correctness
			bytes, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			var jsonResponse TestViewResponse
			err = json.Unmarshal(bytes, &jsonResponse)
			So(err, ShouldBeNil)
			So(jsonResponse.Status, ShouldEqual, true)
			So(jsonResponse.Data, ShouldNotBeNil)
			So(jsonResponse.Data.QuestionDetails.Id, ShouldEqual, 1)
			So(jsonResponse.Data.QuestionDetails.Question, ShouldEqual, "Who's your Daddy?")
			So(jsonResponse.Data.QuestionDetails.Choice1, ShouldEqual, "A")
			So(jsonResponse.Data.QuestionDetails.Choice2, ShouldEqual, "B")
			So(jsonResponse.Data.QuestionDetails.Choice3, ShouldEqual, "C")
			So(jsonResponse.Data.QuestionDetails.Choice4, ShouldEqual, "D")
			So(jsonResponse.Data.QuestionDetails.CorrectAnswer, ShouldEqual, 3)
			So(jsonResponse.Data.QuestionDetails.Status, ShouldEqual, "ACCEPTED")
		})
}
