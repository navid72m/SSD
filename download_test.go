package question_confirmation_controller

import (
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

func TestDownloadQuestionsForConfirmation(t *testing.T) {
	Convey("Given we have a web server with a download API, we want to make sure it returns the correct results\n",
		t, func() {
			gin.SetMode(gin.ReleaseMode)
			router := gin.Default()
			router.Use(fakevitess.Middleware())
			router.POST("/question/confirmation/download", DownloadQuestionsForConfirmation)

			values := url.Values{
				"page":       {"0"},
				"pagesize":   {"1000"},
				"categoryid": {"2"},
			}

			req, err := http.NewRequest("POST", "/question/confirmation/download", strings.NewReader(values.Encode()))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Add("Content-Length", strconv.Itoa(len(values.Encode())))

			So(err, ShouldEqual, nil)

			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			So(resp.Code, ShouldEqual, http.StatusOK)

			headers := resp.Header()
			contentDescription := headers.Get("Content-Description")
			contentType := headers.Get("Content-Type")
			So(contentDescription, ShouldEqual, "File Transfer")
			So(contentType, ShouldEqual, "text/csv")

			/////////////////////Second test case with wrong request there should be a problem in binidng
			values2 := url.Values{
				"page":                {"0"},
				"pagesdasdfadsfasize": {"1000"},
				"categoryid":          {"2"},
			}

			req2, err2 := http.NewRequest("POST", "/question/confirmation/download", strings.NewReader(values2.Encode()))
			req2.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req2.Header.Add("Content-Length", strconv.Itoa(len(values2.Encode())))

			So(err2, ShouldEqual, nil)

			resp2 := httptest.NewRecorder()

			router.ServeHTTP(resp2, req2)
			So(resp2.Code, ShouldNotEqual, http.StatusOK)

			headers2 := resp2.Header()
			contentDescription2 := headers2.Get("Content-Description")
			contentType2 := headers2.Get("Content-Type")
			So(contentDescription2, ShouldNotEqual, "File Transfer")
			So(contentType2, ShouldNotEqual, "text/csv")

			/////////////////////Second test case with wrong request there should be a problem in binidng (wrong data type)
			values3 := url.Values{
				"page":       {"0"},
				"pagesize":   {"fcsdg"},
				"categoryid": {"2"},
			}

			req3, err3 := http.NewRequest("POST", "/question/confirmation/download", strings.NewReader(values3.Encode()))
			req3.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			req3.Header.Add("Content-Length", strconv.Itoa(len(values3.Encode())))

			So(err3, ShouldEqual, nil)

			resp3 := httptest.NewRecorder()

			router.ServeHTTP(resp3, req3)
			So(resp3.Code, ShouldNotEqual, http.StatusOK)

			headers3 := resp3.Header()
			contentDescription3 := headers3.Get("Content-Description")
			contentType3 := headers3.Get("Content-Type")
			So(contentDescription3, ShouldNotEqual, "File Transfer")
			So(contentType3, ShouldNotEqual, "text/csv")

		})
}
