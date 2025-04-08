package controllers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mllb/sampletodo/testutil"
	"github.com/stretchr/testify/assert"
)

func TestCreateTodo(t *testing.T) {
	type want struct {
		status  int
		rspFile string
	}
	tests := map[string]struct {
		reqFile string
		want    want
	}{
		"ok": {
			reqFile: "testdata/todo_controller/ok_req.json.golden",
			want: want{
				status:  http.StatusOK,
				rspFile: "testdata/todo_controller/ok_rsp.json.golden",
			},
		},
		"badRequest": {
			reqFile: "testdata/todo_controller/bad_req.json.golden",
			want: want{
				status:  http.StatusBadRequest,
				rspFile: "testdata/todo_controller/bad_rsp.json.golden",
			},
		},
	}
	for n, tt := range tests {
		tt := tt
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			// Echo のインスタンスとリクエストを作成するのだ
			e := echo.New()
			req := httptest.NewRequest(
				http.MethodPost,
				"/todos",
				bytes.NewReader(testutil.LoadFile(t, tt.reqFile)),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			moq := &TodoServiceMock{}
			moq.CreateTodoFunc = func(
				title string, body string,
			) error {
				if tt.want.status == http.StatusOK {
					return nil
				}
				return errors.New("error from mock")
			}
			sut := NewTodoController(moq)
			if assert.NoError(t, sut.CreateTodo(c)) {
				resp := rec.Result()
				testutil.AssertResponse(t,
					resp, tt.want.status, testutil.LoadFile(t, tt.want.rspFile),
				)
			}
		})
	}
}
