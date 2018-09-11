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

type TestUpdateResponse struct {
	Status bool            `json:"status"`
	Data   *UpdateResponse `json:"data"`
}

func TestUpdate(t *testing.T) {
	Convey("Given we have a web server with a search API, we want to make sure it returns the correct results\n",
		t, func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.Use(fakevitess.Middleware())
			router.POST("/question/update", Update)

			values := url.Values{
				"id":             {"1"},
				"question":       {"Who's your Daddy?"},
				"choice1":        {"A"},
				"choice2":        {"B"},
				"choice3":        {"C"},
				"choice4":        {"D"},
				"correct_answer": {"2"},
				"status":         {"REJECTED"},
			}

			req, err := http.NewRequest("POST", "/question/update", strings.NewReader(values.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))

			So(err, ShouldEqual, nil)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			// read the JSON and check the correctness
			bytes, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldEqual, nil)

			var jsonResponse TestUpdateResponse
			err = json.Unmarshal(bytes, &jsonResponse)
			So(err, ShouldEqual, nil)
			So(jsonResponse.Status, ShouldEqual, true)
			So(jsonResponse.Data.Status, ShouldEqual, true)
		})
}
