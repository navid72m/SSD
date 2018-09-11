package question_confirmation_controller

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
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/gin-gonic/gin.v1"
)

type TestIndexResponse struct {
	Status bool           `json:"status"`
	Data   *IndexResponse `json:"data"`
}

func TestIndex(t *testing.T) {
	Convey("Given we have a web server with a search API, we want to make sure it returns the correct results\n",
		t, func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.Use(fakevitess.Middleware())
			router.POST("/question/index", Index)

			values := url.Values{
				"page":       {"0"},
				"pagesize":   {"1000"},
				"categoryid": {"2"},
			}

			req, err := http.NewRequest("POST", "/question/index", strings.NewReader(values.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))

			So(err, ShouldEqual, nil)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			// read the JSON and check the correctness
			bytes, err := ioutil.ReadAll(resp.Body)
			So(err, ShouldBeNil)

			var jsonResponse TestIndexResponse
			err = json.Unmarshal(bytes, &jsonResponse)
			So(err, ShouldBeNil)
			So(jsonResponse.Status, ShouldEqual, true)
			So(jsonResponse.Data, ShouldNotBeNil)
			So(jsonResponse.Data.First, ShouldEqual, 1)
			So(jsonResponse.Data.Last, ShouldEqual, 2)
		})
}
