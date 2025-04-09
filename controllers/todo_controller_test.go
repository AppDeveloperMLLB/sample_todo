package controllers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	appjwt "github.com/mllb/sampletodo/jwt"
	"github.com/mllb/sampletodo/testutil"
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

			config := echojwt.Config{
				NewClaimsFunc: func(c echo.Context) jwt.Claims {
					return new(appjwt.JwtCustomClaims)
				},
				SigningKey: []byte(appjwt.SigningKey),
			}

			handler := func(c echo.Context) error {
				moq := &TodoServiceMock{}
				moq.CreateTodoFunc = func(
					uid uint, title string, body string,
				) error {
					if tt.want.status == http.StatusOK {
						return nil
					}
					return errors.New("error from mock")
				}
				sut := NewTodoController(moq)
				return sut.CreateTodo(c)
			}

			h := echojwt.WithConfig(config)(handler)

			req := httptest.NewRequest(
				http.MethodPost,
				"/todos",
				bytes.NewReader(testutil.LoadFile(t, tt.reqFile)),
			)
			// テスト用のJWTトークンを作成するのだ
			token, err := appjwt.GenerateToken(1, "test@example.com")
			if err != nil {
				t.Fatalf("failed to generate token: %v", err)
			}
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = h(c)
			if err != nil {
				e.HTTPErrorHandler(err, c)
				return
			}

			resp := rec.Result()
			testutil.AssertResponse(t,
				resp, tt.want.status, testutil.LoadFile(t, tt.want.rspFile),
			)
		})
	}
}
