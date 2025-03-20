└── chapter12
└── section4
├── README.md
├── api
├── middlewares
│ ├── auth.go
│ ├── logging.go
│ └── traceID.go
└── router.go
├── apperrors
├── error.go
├── errorHandler.go
└── errorcode.go
├── common
└── values.go
├── controllers
├── article_controller.go
├── article_controller_test.go
├── comment_controller.go
├── main_test.go
├── services
│ └── services.go
└── testdata
│ ├── data.go
│ └── mock.go
├── docker-compose.yaml
├── go.mod
├── go.sum
├── main.go
├── models
└── models.go
├── repositories
├── articles.go
├── articles_test.go
├── comment_test.go
├── comments.go
├── main_test.go
└── testdata
│ ├── cleanupDB.sql
│ ├── data.go
│ └── setupDB.sql
└── services
├── article_service.go
├── article_service_test.go
├── comment_service.go
├── errors.go
└── service.go

## /chapter12/section4/README.md:

1 | # 12.4 API にユーザー認証を実装しよう

---

## /chapter12/section4/api/middlewares/auth.go:

1 | package middlewares
2 |
3 | import (
4 | "context"
5 | "errors"
6 | "net/http"
7 | "strings"
8 |
9 | "github.com/yourname/reponame/apperrors"
10 | "github.com/yourname/reponame/common"
11 | "google.golang.org/api/idtoken"
12 | )
13 |
14 | const (
15 | googleClientID = "[yourClientID]"
16 | )
17 |
18 | func AuthMiddleware(next http.Handler) http.Handler {
19 | return http.HandlerFunc(func(w http.ResponseWriter, req \*http.Request) {
20 | // ヘッダを抜き出す
21 | authorization := req.Header.Get("Authorization")
22 |
23 | // ヘッダの妥当性を検証
24 | authHeaders := strings.Split(authorization, " ")
25 | if len(authHeaders) != 2 {
26 | err := apperrors.RequiredAuthorizationHeader.Wrap(errors.New("invalid req header"), "invalid header")
27 | apperrors.ErrorHandler(w, req, err)
28 | return
29 | }
30 |
31 | bearer, idToken := authHeaders[0], authHeaders[1]
32 | if bearer != "Bearer" || idToken == "" {
33 | err := apperrors.RequiredAuthorizationHeader.Wrap(errors.New("invalid req header"), "invalid header")
34 | apperrors.ErrorHandler(w, req, err)
35 | return
36 | }
37 |
38 | // ID トークン検証
39 | tokenValidator, err := idtoken.NewValidator(context.Background())
40 | if err != nil {
41 | err = apperrors.CannotMakeValidator.Wrap(err, "internal auth error")
42 | apperrors.ErrorHandler(w, req, err)
43 | return
44 | }
45 |
46 | payload, err := tokenValidator.Validate(context.Background(), idToken, googleClientID)
47 | if err != nil {
48 | err = apperrors.Unauthorizated.Wrap(err, "invalid id token")
49 | apperrors.ErrorHandler(w, req, err)
50 | return
51 | }
52 |
53 | // name フィールドを payload から抜き出す
54 | name, ok := payload.Claims["name"]
55 | if !ok {
56 | err = apperrors.Unauthorizated.Wrap(err, "invalid id token")
57 | apperrors.ErrorHandler(w, req, err)
58 | return
59 | }
60 |
61 | // context にユーザー名をセット
62 | req = common.SetUserName(req, name.(string))
63 |
64 | // 本物のハンドラへ
65 | next.ServeHTTP(w, req)
66 | })
67 | }
68 |

---

## /chapter12/section4/api/middlewares/logging.go:

1 | package middlewares
2 |
3 | import (
4 | "log"
5 | "net/http"
6 |
7 | "github.com/yourname/reponame/common"
8 | )
9 |
10 | type resLoggingWriter struct {
11 | http.ResponseWriter
12 | code int
13 | }
14 |
15 | func NewResLoggingWriter(w http.ResponseWriter) *resLoggingWriter {
16 | return &resLoggingWriter{ResponseWriter: w, code: http.StatusOK}
17 | }
18 |
19 | func (rsw *resLoggingWriter) WriteHeader(code int) {
20 | rsw.code = code
21 | rsw.ResponseWriter.WriteHeader(code)
22 | }
23 |
24 | func LoggingMiddleware(next http.Handler) http.Handler {
25 | return http.HandlerFunc(func(w http.ResponseWriter, req \*http.Request) {
26 | traceID := newTraceID()
27 |
28 | // リクエスト情報をロギング
29 | log.Printf("[%d]%s %s\n", traceID, req.RequestURI, req.Method)
30 |
31 | ctx := common.SetTraceID(req.Context(), traceID)
32 | req = req.WithContext(ctx)
33 | rlw := NewResLoggingWriter(w)
34 |
35 | next.ServeHTTP(rlw, req)
36 |
37 | log.Printf("[%d]res: %d", traceID, rlw.code)
38 | })
39 | }
40 |

---

## /chapter12/section4/api/middlewares/traceID.go:

1 | package middlewares
2 |
3 | import (
4 | "sync"
5 | )
6 |
7 | var (
8 | logNo int = 1
9 | mu sync.Mutex
10 | )
11 |
12 | func newTraceID() int {
13 | var no int
14 |
15 | mu.Lock()
16 | no = logNo
17 | logNo += 1
18 | mu.Unlock()
19 |
20 | return no
21 | }
22 |

---

## /chapter12/section4/api/router.go:

1 | package api
2 |
3 | import (
4 | "database/sql"
5 | "net/http"
6 |
7 | "github.com/gorilla/mux"
8 | "github.com/yourname/reponame/api/middlewares"
9 | "github.com/yourname/reponame/controllers"
10 | "github.com/yourname/reponame/services"
11 | )
12 |
13 | func NewRouter(db *sql.DB) *mux.Router {
14 | ser := services.NewMyAppService(db)
15 | aCon := controllers.NewArticleController(ser)
16 | cCon := controllers.NewCommentController(ser)
17 |
18 | r := mux.NewRouter()
19 |
20 | r.HandleFunc("/hello", aCon.HelloHandler).Methods(http.MethodGet)
21 |
22 | r.HandleFunc("/article", aCon.PostArticleHandler).Methods(http.MethodPost)
23 | r.HandleFunc("/article/list", aCon.ArticleListHandler).Methods(http.MethodGet)
24 | r.HandleFunc("/article/{id:[0-9]+}", aCon.ArticleDetailHandler).Methods(http.MethodGet)
25 | r.HandleFunc("/article/nice", aCon.PostNiceHandler).Methods(http.MethodPost)
26 |
27 | r.HandleFunc("/comment", cCon.PostCommentHandler).Methods(http.MethodPost)
28 |
29 | r.Use(middlewares.LoggingMiddleware)
30 | r.Use(middlewares.AuthMiddleware)
31 |
32 | return r
33 | }
34 |

---

## /chapter12/section4/apperrors/error.go:

1 | package apperrors
2 |
3 | type MyAppError struct {
4 | // ErrCode -> レスポンスとログに表示するエラーコード
5 | // Message -> レスポンスに表示するエラーメッセージ
6 | // error -> ログに表示する生の内部エラー
7 |
8 | ErrCode
9 | Message string
10 | Err error `json:"-"`
11 | }
12 |
13 | func (myErr *MyAppError) Error() string {
14 | return myErr.Err.Error()
15 | }
16 |
17 | // errors.Is/errors.As を使えるように Unwrap メソッドを定義
18 | func (myErr *MyAppError) Unwrap() error {
19 | return myErr.Err
20 | }
21 |

---

## /chapter12/section4/apperrors/errorHandler.go:

1 | package apperrors
2 |
3 | import (
4 | "encoding/json"
5 | "errors"
6 | "log"
7 | "net/http"
8 |
9 | "github.com/yourname/reponame/common"
10 | )
11 |
12 | // エラーが発生したときのレスポンス処理をここで一括で行う
13 | func ErrorHandler(w http.ResponseWriter, req *http.Request, err error) {
14 | var appErr *MyAppError
15 | if !errors.As(err, &appErr) {
16 | appErr = &MyAppError{
17 | ErrCode: Unknown,
18 | Message: "internal process failed",
19 | Err: err,
20 | }
21 | }
22 |
23 | traceID := common.GetTraceID(req.Context())
24 | log.Printf("[%d]error: %s\n", traceID, appErr)
25 |
26 | var statusCode int
27 |
28 | switch appErr.ErrCode {
29 | case NAData:
30 | statusCode = http.StatusNotFound
31 | case NoTargetData, ReqBodyDecodeFailed, BadParam:
32 | statusCode = http.StatusBadRequest
33 | case RequiredAuthorizationHeader, Unauthorizated:
34 | statusCode = http.StatusUnauthorized
35 | case NotMatchUser:
36 | statusCode = http.StatusForbidden
37 | default:
38 | statusCode = http.StatusInternalServerError
39 | }
40 |
41 | w.WriteHeader(statusCode)
42 | json.NewEncoder(w).Encode(appErr)
43 | }
44 |

---

## /chapter12/section4/apperrors/errorcode.go:

1 | package apperrors
2 |
3 | type ErrCode string
4 |
5 | const (
6 | Unknown ErrCode = "U000"
7 |
8 | InsertDataFailed ErrCode = "S001"
9 | GetDataFailed ErrCode = "S002"
10 | NAData ErrCode = "S003"
11 | NoTargetData ErrCode = "S004"
12 | UpdateDataFailed ErrCode = "S005"
13 |
14 | ReqBodyDecodeFailed ErrCode = "R001"
15 | BadParam ErrCode = "R002"
16 |
17 | RequiredAuthorizationHeader ErrCode = "A001"
18 | CannotMakeValidator ErrCode = "A002"
19 | Unauthorizated ErrCode = "A003"
20 | NotMatchUser ErrCode = "A004"
21 | )
22 |
23 | func (code ErrCode) Wrap(err error, message string) error {
24 | return &MyAppError{ErrCode: code, Message: message, Err: err}
25 | }
26 |

---

## /chapter12/section4/common/values.go:

1 | package common
2 |
3 | import (
4 | "context"
5 | "net/http"
6 | )
7 |
8 | type traceIDKey struct{}
9 |
10 | func SetTraceID(ctx context.Context, traceID int) context.Context {
11 | return context.WithValue(ctx, traceIDKey{}, traceID)
12 | }
13 |
14 | func GetTraceID(ctx context.Context) int {
15 | id := ctx.Value(traceIDKey{})
16 |
17 | if idInt, ok := id.(int); ok {
18 | return idInt
19 | }
20 | return 0
21 | }
22 |
23 | type userNameKey struct{}
24 |
25 | func GetUserName(ctx context.Context) string {
26 | id := ctx.Value(userNameKey{})
27 |
28 | if usernameStr, ok := id.(string); ok {
29 | return usernameStr
30 | }
31 | return ""
32 | }
33 |
34 | func SetUserName(req *http.Request, name string) *http.Request {
35 | ctx := req.Context()
36 |
37 | ctx = context.WithValue(ctx, userNameKey{}, name)
38 | req = req.WithContext(ctx)
39 |
40 | return req
41 | }
42 |

---

## /chapter12/section4/controllers/article_controller.go:

1 | package controllers
2 |
3 | import (
4 | "encoding/json"
5 | "errors"
6 | "io"
7 | "net/http"
8 | "strconv"
9 |
10 | "github.com/gorilla/mux"
11 | "github.com/yourname/reponame/apperrors"
12 | "github.com/yourname/reponame/common"
13 | "github.com/yourname/reponame/controllers/services"
14 | "github.com/yourname/reponame/models"
15 | )
16 |
17 | type ArticleController struct {
18 | service services.ArticleServicer
19 | }
20 |
21 | func NewArticleController(s services.ArticleServicer) *ArticleController {
22 | return &ArticleController{service: s}
23 | }
24 |
25 | // GET /hello のハンドラ
26 | func (c *ArticleController) HelloHandler(w http.ResponseWriter, req *http.Request) {
27 | io.WriteString(w, "Hello, world!\n")
28 | }
29 |
30 | // POST /article のハンドラ
31 | func (c *ArticleController) PostArticleHandler(w http.ResponseWriter, req *http.Request) {
32 | var reqArticle models.Article
33 | if err := json.NewDecoder(req.Body).Decode(&reqArticle); err != nil {
34 | err = apperrors.ReqBodyDecodeFailed.Wrap(err, "bad request body")
35 | apperrors.ErrorHandler(w, req, err)
36 | return
37 | }
38 |
39 | authedUserName := common.GetUserName(req.Context())
40 | if reqArticle.UserName != authedUserName {
41 | err := apperrors.NotMatchUser.Wrap(errors.New("does not match reqBody user and idtoken user"), "invalid parameter")
42 | apperrors.ErrorHandler(w, req, err)
43 | return
44 | }
45 |
46 | article, err := c.service.PostArticleService(reqArticle)
47 | if err != nil {
48 | apperrors.ErrorHandler(w, req, err)
49 | return
50 | }
51 |
52 | json.NewEncoder(w).Encode(article)
53 | }
54 |
55 | // GET /article/list のハンドラ
56 | func (c *ArticleController) ArticleListHandler(w http.ResponseWriter, req *http.Request) {
57 | queryMap := req.URL.Query()
58 |
59 | // クエリパラメータ page を取得
60 | var page int
61 | if p, ok := queryMap["page"]; ok && len(p) > 0 {
62 | var err error
63 | page, err = strconv.Atoi(p[0])
64 | if err != nil {
65 | err = apperrors.BadParam.Wrap(err, "queryparam must be number")
66 | apperrors.ErrorHandler(w, req, err)
67 | return
68 | }
69 | } else {
70 | page = 1
71 | }
72 |
73 | articleList, err := c.service.GetArticleListService(page)
74 | if err != nil {
75 | apperrors.ErrorHandler(w, req, err)
76 | return
77 | }
78 |
79 | json.NewEncoder(w).Encode(articleList)
80 | }
81 |
82 | // GET /article/{id} のハンドラ
83 | func (c *ArticleController) ArticleDetailHandler(w http.ResponseWriter, req *http.Request) {
84 | articleID, err := strconv.Atoi(mux.Vars(req)["id"])
85 | if err != nil {
86 | err = apperrors.BadParam.Wrap(err, "pathparam must be number")
87 | apperrors.ErrorHandler(w, req, err)
88 | return
89 | }
90 |
91 | article, err := c.service.GetArticleService(articleID)
92 | if err != nil {
93 | apperrors.ErrorHandler(w, req, err)
94 | return
95 | }
96 |
97 | json.NewEncoder(w).Encode(article)
98 | }
99 |
100 | // POST /article/nice のハンドラ
101 | func (c *ArticleController) PostNiceHandler(w http.ResponseWriter, req \*http.Request) {
102 | var reqArticle models.Article
103 | if err := json.NewDecoder(req.Body).Decode(&reqArticle); err != nil {
104 | apperrors.ErrorHandler(w, req, err)
105 | http.Error(w, "fail to decode json\n", http.StatusBadRequest)
106 | }
107 |
108 | article, err := c.service.PostNiceService(reqArticle)
109 | if err != nil {
110 | apperrors.ErrorHandler(w, req, err)
111 | return
112 | }
113 |
114 | json.NewEncoder(w).Encode(article)
115 | }
116 |

---

## /chapter12/section4/controllers/article_controller_test.go:

1 | package controllers*test
2 |
3 | import (
4 | "fmt"
5 | "net/http"
6 | "net/http/httptest"
7 | "testing"
8 |
9 | "github.com/gorilla/mux"
10 | )
11 |
12 | func TestArticleListHandler(t \*testing.T) {
13 | var tests = []struct {
14 | name string
15 | query string
16 | resultCode int
17 | }{
18 | {name: "number query", query: "1", resultCode: http.StatusOK},
19 | {name: "alphabet query", query: "aaa", resultCode: http.StatusBadRequest},
20 | }
21 |
22 | for *, tt := range tests {
23 | t.Run(tt.name, func(t *testing.T) {
24 | url := fmt.Sprintf("http://localhost:8080/article/list?page=%s", tt.query)
25 | req := httptest.NewRequest(http.MethodGet, url, nil)
26 |
27 | res := httptest.NewRecorder()
28 |
29 | aCon.ArticleListHandler(res, req)
30 |
31 | if res.Code != tt.resultCode {
32 | t.Errorf("unexpected StatusCode: want %d but %d\n", tt.resultCode, res.Code)
33 | }
34 | })
35 | }
36 | }
37 |
38 | func TestArticleDetailHandler(t *testing.T) {
39 | var tests = []struct {
40 | name string
41 | articleID string
42 | resultCode int
43 | }{
44 | {name: "number pathparam", articleID: "1", resultCode: http.StatusOK},
45 | {name: "alphabet pathparam", articleID: "aaa", resultCode: http.StatusNotFound},
46 | }
47 |
48 | for \_, tt := range tests {
49 | t.Run(tt.name, func(t \*testing.T) {
50 | url := fmt.Sprintf("http://localhost:8080/article/%s", tt.articleID)
51 | req := httptest.NewRequest(http.MethodGet, url, nil)
52 |
53 | res := httptest.NewRecorder()
54 |
55 | r := mux.NewRouter()
56 | r.HandleFunc("/article/{id:[0-9]+}", aCon.ArticleDetailHandler).Methods(http.MethodGet)
57 | r.ServeHTTP(res, req)
58 |
59 | if res.Code != tt.resultCode {
60 | t.Errorf("unexpected StatusCode: want %d but %d\n", tt.resultCode, res.Code)
61 | }
62 | })
63 | }
64 | }
65 |

---

## /chapter12/section4/controllers/comment_controller.go:

1 | package controllers
2 |
3 | import (
4 | "encoding/json"
5 | "net/http"
6 |
7 | "github.com/yourname/reponame/apperrors"
8 | "github.com/yourname/reponame/controllers/services"
9 | "github.com/yourname/reponame/models"
10 | )
11 |
12 | type CommentController struct {
13 | service services.CommentServicer
14 | }
15 |
16 | func NewCommentController(s services.CommentServicer) *CommentController {
17 | return &CommentController{service: s}
18 | }
19 |
20 | // POST /comment のハンドラ
21 | func (c *CommentController) PostCommentHandler(w http.ResponseWriter, req \*http.Request) {
22 | var reqComment models.Comment
23 | if err := json.NewDecoder(req.Body).Decode(&reqComment); err != nil {
24 | err = apperrors.ReqBodyDecodeFailed.Wrap(err, "bad request body")
25 | apperrors.ErrorHandler(w, req, err)
26 | }
27 |
28 | comment, err := c.service.PostCommentService(reqComment)
29 | if err != nil {
30 | apperrors.ErrorHandler(w, req, err)
31 | return
32 | }
33 | json.NewEncoder(w).Encode(comment)
34 | }
35 |

---

## /chapter12/section4/controllers/main_test.go:

1 | package controllers_test
2 |
3 | import (
4 | "testing"
5 |
6 | "github.com/yourname/reponame/controllers"
7 | "github.com/yourname/reponame/controllers/testdata"
8 | )
9 |
10 | var aCon *controllers.ArticleController
11 |
12 | func TestMain(m *testing.M) {
13 | ser := testdata.NewServiceMock()
14 | aCon = controllers.NewArticleController(ser)
15 |
16 | m.Run()
17 | }
18 |

---

## /chapter12/section4/controllers/services/services.go:

1 | package services
2 |
3 | import "github.com/yourname/reponame/models"
4 |
5 | // /article 関連を引き受けるサービス
6 | type ArticleServicer interface {
7 | PostArticleService(article models.Article) (models.Article, error)
8 | GetArticleListService(page int) ([]models.Article, error)
9 | GetArticleService(articleID int) (models.Article, error)
10 | PostNiceService(article models.Article) (models.Article, error)
11 | }
12 |
13 | // /comment を引き受けるサービス
14 | type CommentServicer interface {
15 | PostCommentService(comment models.Comment) (models.Comment, error)
16 | }
17 |

---

## /chapter12/section4/controllers/testdata/data.go:

1 | package testdata
2 |
3 | import "github.com/yourname/reponame/models"
4 |
5 | var articleTestData = []models.Article{
6 | models.Article{
7 | ID: 1,
8 | Title: "firstPost",
9 | Contents: "This is my first blog",
10 | UserName: "saki",
11 | NiceNum: 2,
12 | CommentList: commentTestData,
13 | },
14 | models.Article{
15 | ID: 2,
16 | Title: "2nd",
17 | Contents: "Second blog post",
18 | UserName: "saki",
19 | NiceNum: 4,
20 | },
21 | }
22 |
23 | var commentTestData = []models.Comment{
24 | models.Comment{
25 | CommentID: 1,
26 | ArticleID: 1,
27 | Message: "1st comment yeah",
28 | },
29 | models.Comment{
30 | CommentID: 2,
31 | ArticleID: 1,
32 | Message: "welcome",
33 | },
34 | }
35 |

---

## /chapter12/section4/controllers/testdata/mock.go:

1 | package testdata
2 |
3 | import "github.com/yourname/reponame/models"
4 |
5 | type serviceMock struct{}
6 |
7 | func NewServiceMock() *serviceMock {
8 | return &serviceMock{}
9 | }
10 |
11 | func (s *serviceMock) PostArticleService(article models.Article) (models.Article, error) {
12 | return articleTestData[1], nil
13 | }
14 |
15 | func (s *serviceMock) GetArticleListService(page int) ([]models.Article, error) {
16 | return articleTestData, nil
17 | }
18 |
19 | func (s *serviceMock) GetArticleService(articleID int) (models.Article, error) {
20 | return articleTestData[0], nil
21 | }
22 |
23 | func (s *serviceMock) PostNiceService(article models.Article) (models.Article, error) {
24 | return articleTestData[0], nil
25 | }
26 |
27 | func (s *serviceMock) PostCommentService(comment models.Comment) (models.Comment, error) {
28 | return commentTestData[0], nil
29 | }
30 |

---

## /chapter12/section4/docker-compose.yaml:

1 | version: '3.3'
2 | services:
3 |
4 | mysql:
5 | image: mysql:5.7
6 | container_name: db-for-go
7 | command:
8 | - --character-set-server=utf8mb4
9 | - --collation-server=utf8mb4_unicode_ci
10 | - --sql-mode=ONLY_FULL_GROUP_BY,NO_AUTO_CREATE_USER,NO_ENGINE_SUBSTITUTION
11 | environment:
12 | MYSQL_ROOT_USER: ${ROOTUSER}
13 | MYSQL_ROOT_PASSWORD: ${ROOTPASS}
14 | MYSQL_DATABASE: ${DATABASE}
15 | MYSQL_USER: ${USERNAME}
16 | MYSQL_PASSWORD: ${USERPASS}
17 | TZ: 'Asia/Tokyo'
18 | ports:
19 | - "3306:3306"
20 | volumes:
21 | - db-volume:/var/lib/mysql
22 |
23 | volumes:
24 | db-volume:
25 |

---

## /chapter12/section4/go.mod:

1 | module github.com/yourname/reponame
2 |
3 | go 1.17
4 |
5 | require (
6 | cloud.google.com/go v0.99.0 // indirect
7 | github.com/go-sql-driver/mysql v1.6.0 // indirect
8 | github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
9 | github.com/golang/protobuf v1.5.2 // indirect
10 | github.com/gorilla/mux v1.8.0 // indirect
11 | go.opencensus.io v0.23.0 // indirect
12 | golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420 // indirect
13 | golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
14 | golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
15 | golang.org/x/text v0.3.6 // indirect
16 | google.golang.org/api v0.64.0 // indirect
17 | google.golang.org/appengine v1.6.7 // indirect
18 | google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb // indirect
19 | google.golang.org/grpc v1.40.1 // indirect
20 | google.golang.org/protobuf v1.27.1 // indirect
21 | )
22 |

---

## /chapter12/section4/go.sum:

1 | cloud.google.com/go v0.26.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=
2 | cloud.google.com/go v0.34.0/go.mod h1:aQUYkXzVsufM+DwF1aE+0xfcU+56JwCaLick0ClmMTw=
3 | cloud.google.com/go v0.38.0/go.mod h1:990N+gfupTy94rShfmMCWGDn0LpTmnzTp2qbd1dvSRU=
4 | cloud.google.com/go v0.44.1/go.mod h1:iSa0KzasP4Uvy3f1mN/7PiObzGgflwredwwASm/v6AU=
5 | cloud.google.com/go v0.44.2/go.mod h1:60680Gw3Yr4ikxnPRS/oxxkBccT6SA1yMk63TGekxKY=
6 | cloud.google.com/go v0.45.1/go.mod h1:RpBamKRgapWJb87xiFSdk4g1CME7QZg3uwTez+TSTjc=
7 | cloud.google.com/go v0.46.3/go.mod h1:a6bKKbmY7er1mI7TEI4lsAkts/mkhTSZK8w33B4RAg0=
8 | cloud.google.com/go v0.50.0/go.mod h1:r9sluTvynVuxRIOHXQEHMFffphuXHOMZMycpNR5e6To=
9 | cloud.google.com/go v0.52.0/go.mod h1:pXajvRH/6o3+F9jDHZWQ5PbGhn+o8w9qiu/CffaVdO4=
10 | cloud.google.com/go v0.53.0/go.mod h1:fp/UouUEsRkN6ryDKNW/Upv/JBKnv6WDthjR6+vze6M=
11 | cloud.google.com/go v0.54.0/go.mod h1:1rq2OEkV3YMf6n/9ZvGWI3GWw0VoqH/1x2nd8Is/bPc=
12 | cloud.google.com/go v0.56.0/go.mod h1:jr7tqZxxKOVYizybht9+26Z/gUq7tiRzu+ACVAMbKVk=
13 | cloud.google.com/go v0.57.0/go.mod h1:oXiQ6Rzq3RAkkY7N6t3TcE6jE+CIBBbA36lwQ1JyzZs=
14 | cloud.google.com/go v0.62.0/go.mod h1:jmCYTdRCQuc1PHIIJ/maLInMho30T/Y0M4hTdTShOYc=
15 | cloud.google.com/go v0.65.0/go.mod h1:O5N8zS7uWy9vkA9vayVHs65eM1ubvY4h553ofrNHObY=
16 | cloud.google.com/go v0.72.0/go.mod h1:M+5Vjvlc2wnp6tjzE102Dw08nGShTscUx2nZMufOKPI=
17 | cloud.google.com/go v0.74.0/go.mod h1:VV1xSbzvo+9QJOxLDaJfTjx5e+MePCpCWwvftOeQmWk=
18 | cloud.google.com/go v0.78.0/go.mod h1:QjdrLG0uq+YwhjoVOLsS1t7TW8fs36kLs4XO5R5ECHg=
19 | cloud.google.com/go v0.79.0/go.mod h1:3bzgcEeQlzbuEAYu4mrWhKqWjmpprinYgKJLgKHnbb8=
20 | cloud.google.com/go v0.81.0/go.mod h1:mk/AM35KwGk/Nm2YSeZbxXdrNK3KZOYHmLkOqC2V6E0=
21 | cloud.google.com/go v0.83.0/go.mod h1:Z7MJUsANfY0pYPdw0lbnivPx4/vhy/e2FEkSkF7vAVY=
22 | cloud.google.com/go v0.84.0/go.mod h1:RazrYuxIK6Kb7YrzzhPoLmCVzl7Sup4NrbKPg8KHSUM=
23 | cloud.google.com/go v0.87.0/go.mod h1:TpDYlFy7vuLzZMMZ+B6iRiELaY7z/gJPaqbMx6mlWcY=
24 | cloud.google.com/go v0.90.0/go.mod h1:kRX0mNRHe0e2rC6oNakvwQqzyDmg57xJ+SZU1eT2aDQ=
25 | cloud.google.com/go v0.93.3/go.mod h1:8utlLll2EF5XMAV15woO4lSbWQlk8rer9aLOfLh7+YI=
26 | cloud.google.com/go v0.94.1/go.mod h1:qAlAugsXlC+JWO+Bke5vCtc9ONxjQT3drlTTnAplMW4=
27 | cloud.google.com/go v0.97.0/go.mod h1:GF7l59pYBVlXQIBLx3a761cZ41F9bBH3JUlihCt2Udc=
28 | cloud.google.com/go v0.99.0 h1:y/cM2iqGgGi5D5DQZl6D9STN/3dR/Vx5Mp8s752oJTY=
29 | cloud.google.com/go v0.99.0/go.mod h1:w0Xx2nLzqWJPuozYQX+hFfCSI8WioryfRDzkoI/Y2ZA=
30 | cloud.google.com/go/bigquery v1.0.1/go.mod h1:i/xbL2UlR5RvWAURpBYZTtm/cXjCha9lbfbpx4poX+o=
31 | cloud.google.com/go/bigquery v1.3.0/go.mod h1:PjpwJnslEMmckchkHFfq+HTD2DmtT67aNFKH1/VBDHE=
32 | cloud.google.com/go/bigquery v1.4.0/go.mod h1:S8dzgnTigyfTmLBfrtrhyYhwRxG72rYxvftPBK2Dvzc=
33 | cloud.google.com/go/bigquery v1.5.0/go.mod h1:snEHRnqQbz117VIFhE8bmtwIDY80NLUZUMb4Nv6dBIg=
34 | cloud.google.com/go/bigquery v1.7.0/go.mod h1://okPTzCYNXSlb24MZs83e2Do+h+VXtc4gLoIoXIAPc=
35 | cloud.google.com/go/bigquery v1.8.0/go.mod h1:J5hqkt3O0uAFnINi6JXValWIb1v0goeZM77hZzJN/fQ=
36 | cloud.google.com/go/datastore v1.0.0/go.mod h1:LXYbyblFSglQ5pkeyhO+Qmw7ukd3C+pD7TKLgZqpHYE=
37 | cloud.google.com/go/datastore v1.1.0/go.mod h1:umbIZjpQpHh4hmRpGhH4tLFup+FVzqBi1b3c64qFpCk=
38 | cloud.google.com/go/pubsub v1.0.1/go.mod h1:R0Gpsv3s54REJCy4fxDixWD93lHJMoZTyQ2kNxGRt3I=
39 | cloud.google.com/go/pubsub v1.1.0/go.mod h1:EwwdRX2sKPjnvnqCa270oGRyludottCI76h+R3AArQw=
40 | cloud.google.com/go/pubsub v1.2.0/go.mod h1:jhfEVHT8odbXTkndysNHCcx0awwzvfOlguIAii9o8iA=
41 | cloud.google.com/go/pubsub v1.3.1/go.mod h1:i+ucay31+CNRpDW4Lu78I4xXG+O1r/MAHgjpRVR+TSU=
42 | cloud.google.com/go/storage v1.0.0/go.mod h1:IhtSnM/ZTZV8YYJWCY8RULGVqBDmpoyjwiyrjsg+URw=
43 | cloud.google.com/go/storage v1.5.0/go.mod h1:tpKbwo567HUNpVclU5sGELwQWBDZ8gh0ZeosJ0Rtdos=
44 | cloud.google.com/go/storage v1.6.0/go.mod h1:N7U0C8pVQ/+NIKOBQyamJIeKQKkZ+mxpohlUTyfDhBk=
45 | cloud.google.com/go/storage v1.8.0/go.mod h1:Wv1Oy7z6Yz3DshWRJFhqM/UCfaWIRTdp0RXyy7KQOVs=
46 | cloud.google.com/go/storage v1.10.0/go.mod h1:FLPqc6j+Ki4BU591ie1oL6qBQGu2Bl/tZ9ullr3+Kg0=
47 | dmitri.shuralyov.com/gpu/mtl v0.0.0-20190408044501-666a987793e9/go.mod h1:H6x//7gZCb22OMCxBHrMx7a5I7Hp++hsVxbQ4BYO7hU=
48 | github.com/BurntSushi/toml v0.3.1/go.mod h1:xHWCNGjB5oqiDr8zfno3MHue2Ht5sIBksp03qcyfWMU=
49 | github.com/BurntSushi/xgb v0.0.0-20160522181843-27f122750802/go.mod h1:IVnqGOEym/WlBOVXweHU+Q+/VP0lqqI8lqeDx9IjBqo=
50 | github.com/OneOfOne/xxhash v1.2.2/go.mod h1:HSdplMjZKSmBqAxg5vPj2TmRDmfkzw+cTzAElWljhcU=
51 | github.com/antihax/optional v1.0.0/go.mod h1:uupD/76wgC+ih3iEmQUL+0Ugr19nfwCT1kdvxnR2qWY=
52 | github.com/census-instrumentation/opencensus-proto v0.2.1/go.mod h1:f6KPmirojxKA12rnyqOA5BBL4O983OfeGPqjHWSTneU=
53 | github.com/cespare/xxhash v1.1.0/go.mod h1:XrSqR1VqqWfGrhpAt58auRo0WTKS1nRRg3ghfAqPWnc=
54 | github.com/chzyer/logex v1.1.10/go.mod h1:+Ywpsq7O8HXn0nuIou7OrIPyXbp3wmkHB+jjWRnGsAI=
55 | github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e/go.mod h1:nSuG5e5PlCu98SY8svDHJxuZscDgtXS6KTTbou5AhLI=
56 | github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1/go.mod h1:Q3SI9o4m/ZMnBNeIyt5eFwwo7qiLfzFZmjNmxjkiQlU=
57 | github.com/client9/misspell v0.3.4/go.mod h1:qj6jICC3Q7zFZvVWo7KLAzC3yx5G7kyvSDkc90ppPyw=
58 | github.com/cncf/udpa/go v0.0.0-20191209042840-269d4d468f6f/go.mod h1:M8M6+tZqaGXZJjfX53e64911xZQV5JYwmTeXPW+k8Sc=
59 | github.com/cncf/udpa/go v0.0.0-20200629203442-efcf912fb354/go.mod h1:WmhPx2Nbnhtbo57+VJT5O0JRkEi1Wbu0z5j0R8u5Hbk=
60 | github.com/cncf/udpa/go v0.0.0-20201120205902-5459f2c99403/go.mod h1:WmhPx2Nbnhtbo57+VJT5O0JRkEi1Wbu0z5j0R8u5Hbk=
61 | github.com/cncf/xds/go v0.0.0-20210312221358-fbca930ec8ed/go.mod h1:eXthEFrGJvWHgFFCl3hGmgk+/aYT6PnTQLykKQRLhEs=
62 | github.com/davecgh/go-spew v1.1.0/go.mod h1:J7Y8YcW2NihsgmVo/mv3lAwl/skON4iLHjSsI+c5H38=
63 | github.com/envoyproxy/go-control-plane v0.9.0/go.mod h1:YTl/9mNaCwkRvm6d1a2C3ymFceY/DCBVvsKhRF0iEA4=
64 | github.com/envoyproxy/go-control-plane v0.9.1-0.20191026205805-5f8ba28d4473/go.mod h1:YTl/9mNaCwkRvm6d1a2C3ymFceY/DCBVvsKhRF0iEA4=
65 | github.com/envoyproxy/go-control-plane v0.9.4/go.mod h1:6rpuAdCZL397s3pYoYcLgu1mIlRU8Am5FuJP05cCM98=
66 | github.com/envoyproxy/go-control-plane v0.9.7/go.mod h1:cwu0lG7PUMfa9snN8LXBig5ynNVH9qI8YYLbd1fK2po=
67 | github.com/envoyproxy/go-control-plane v0.9.9-0.20201210154907-fd9021fe5dad/go.mod h1:cXg6YxExXjJnVBQHBLXeUAgxn2UodCpnH306RInaBQk=
68 | github.com/envoyproxy/go-control-plane v0.9.9-0.20210217033140-668b12f5399d/go.mod h1:cXg6YxExXjJnVBQHBLXeUAgxn2UodCpnH306RInaBQk=
69 | github.com/envoyproxy/go-control-plane v0.9.9-0.20210512163311-63b5d3c536b0/go.mod h1:hliV/p42l8fGbc6Y9bQ70uLwIvmJyVE5k4iMKlh8wCQ=
70 | github.com/envoyproxy/protoc-gen-validate v0.1.0/go.mod h1:iSmxcyjqTsJpI2R4NaDN7+kN2VEUnK/pcBlmesArF7c=
71 | github.com/ghodss/yaml v1.0.0/go.mod h1:4dBDuWmgqj2HViK6kFavaiC9ZROes6MMH2rRYeMEF04=
72 | github.com/go-gl/glfw v0.0.0-20190409004039-e6da0acd62b1/go.mod h1:vR7hzQXu2zJy9AVAgeJqvqgH9Q5CA+iKCZ2gyEVpxRU=
73 | github.com/go-gl/glfw/v3.3/glfw v0.0.0-20191125211704-12ad95a8df72/go.mod h1:tQ2UAYgL5IevRw8kRxooKSPJfGvJ9fJQFa0TUsXzTg8=
74 | github.com/go-gl/glfw/v3.3/glfw v0.0.0-20200222043503-6f7a984d4dc4/go.mod h1:tQ2UAYgL5IevRw8kRxooKSPJfGvJ9fJQFa0TUsXzTg8=
75 | github.com/go-sql-driver/mysql v1.6.0 h1:BCTh4TKNUYmOmMUcQ3IipzF5prigylS7XXjEkfCHuOE=
76 | github.com/go-sql-driver/mysql v1.6.0/go.mod h1:DCzpHaOWr8IXmIStZouvnhqoel9Qv2LBy8hT2VhHyBg=
77 | github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b/go.mod h1:SBH7ygxi8pfUlaOkMMuAQtPIUF8ecWP5IEl/CR7VP2Q=
78 | github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6/go.mod h1:cIg4eruTrX1D+g88fzRXU5OdNfaM+9IcxsU14FzY7Hc=
79 | github.com/golang/groupcache v0.0.0-20191227052852-215e87163ea7/go.mod h1:cIg4eruTrX1D+g88fzRXU5OdNfaM+9IcxsU14FzY7Hc=
80 | github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e h1:1r7pUrabqp18hOBcwBwiTsbnFeTZHV9eER/QT5JVZxY=
81 | github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e/go.mod h1:cIg4eruTrX1D+g88fzRXU5OdNfaM+9IcxsU14FzY7Hc=
82 | github.com/golang/mock v1.1.1/go.mod h1:oTYuIxOrZwtPieC+H1uAHpcLFnEyAGVDL/k47Jfbm0A=
83 | github.com/golang/mock v1.2.0/go.mod h1:oTYuIxOrZwtPieC+H1uAHpcLFnEyAGVDL/k47Jfbm0A=
84 | github.com/golang/mock v1.3.1/go.mod h1:sBzyDLLjw3U8JLTeZvSv8jJB+tU5PVekmnlKIyFUx0Y=
85 | github.com/golang/mock v1.4.0/go.mod h1:UOMv5ysSaYNkG+OFQykRIcU/QvvxJf3p21QfJ2Bt3cw=
86 | github.com/golang/mock v1.4.1/go.mod h1:UOMv5ysSaYNkG+OFQykRIcU/QvvxJf3p21QfJ2Bt3cw=
87 | github.com/golang/mock v1.4.3/go.mod h1:UOMv5ysSaYNkG+OFQykRIcU/QvvxJf3p21QfJ2Bt3cw=
88 | github.com/golang/mock v1.4.4/go.mod h1:l3mdAwkq5BuhzHwde/uurv3sEJeZMXNpwsxVWU71h+4=
89 | github.com/golang/mock v1.5.0/go.mod h1:CWnOUgYIOo4TcNZ0wHX3YZCqsaM1I1Jvs6v3mP3KVu8=
90 | github.com/golang/mock v1.6.0/go.mod h1:p6yTPP+5HYm5mzsMV8JkE6ZKdX+/wYM6Hr+LicevLPs=
91 | github.com/golang/protobuf v1.2.0/go.mod h1:6lQm79b+lXiMfvg/cZm0SGofjICqVBUtrP5yJMmIC1U=
92 | github.com/golang/protobuf v1.3.1/go.mod h1:6lQm79b+lXiMfvg/cZm0SGofjICqVBUtrP5yJMmIC1U=
93 | github.com/golang/protobuf v1.3.2/go.mod h1:6lQm79b+lXiMfvg/cZm0SGofjICqVBUtrP5yJMmIC1U=
94 | github.com/golang/protobuf v1.3.3/go.mod h1:vzj43D7+SQXF/4pzW/hwtAqwc6iTitCiVSaWz5lYuqw=
95 | github.com/golang/protobuf v1.3.4/go.mod h1:vzj43D7+SQXF/4pzW/hwtAqwc6iTitCiVSaWz5lYuqw=
96 | github.com/golang/protobuf v1.3.5/go.mod h1:6O5/vntMXwX2lRkT1hjjk0nAC1IDOTvTlVgjlRvqsdk=
97 | github.com/golang/protobuf v1.4.0-rc.1/go.mod h1:ceaxUfeHdC40wWswd/P6IGgMaK3YpKi5j83Wpe3EHw8=
98 | github.com/golang/protobuf v1.4.0-rc.1.0.20200221234624-67d41d38c208/go.mod h1:xKAWHe0F5eneWXFV3EuXVDTCmh+JuBKY0li0aMyXATA=
99 | github.com/golang/protobuf v1.4.0-rc.2/go.mod h1:LlEzMj4AhA7rCAGe4KMBDvJI+AwstrUpVNzEA03Pprs=
100 | github.com/golang/protobuf v1.4.0-rc.4.0.20200313231945-b860323f09d0/go.mod h1:WU3c8KckQ9AFe+yFwt9sWVRKCVIyN9cPHBJSNnbL67w=
101 | github.com/golang/protobuf v1.4.0/go.mod h1:jodUvKwWbYaEsadDk5Fwe5c77LiNKVO9IDvqG2KuDX0=
102 | github.com/golang/protobuf v1.4.1/go.mod h1:U8fpvMrcmy5pZrNK1lt4xCsGvpyWQ/VVv6QDs8UjoX8=
103 | github.com/golang/protobuf v1.4.2/go.mod h1:oDoupMAO8OvCJWAcko0GGGIgR6R6ocIYbsSw735rRwI=
104 | github.com/golang/protobuf v1.4.3/go.mod h1:oDoupMAO8OvCJWAcko0GGGIgR6R6ocIYbsSw735rRwI=
105 | github.com/golang/protobuf v1.5.0/go.mod h1:FsONVRAS9T7sI+LIUmWTfcYkHO4aIWwzhcaSAoJOfIk=
106 | github.com/golang/protobuf v1.5.1/go.mod h1:DopwsBzvsk0Fs44TXzsVbJyPhcCPeIwnvohx4u74HPM=
107 | github.com/golang/protobuf v1.5.2 h1:ROPKBNFfQgOUMifHyP+KYbvpjbdoFNs+aK7DXlji0Tw=
108 | github.com/golang/protobuf v1.5.2/go.mod h1:XVQd3VNwM+JqD3oG2Ue2ip4fOMUkwXdXDdiuN0vRsmY=
109 | github.com/golang/snappy v0.0.3/go.mod h1:/XxbfmMg8lxefKM7IXC3fBNl/7bRcc72aCRzEWrmP2Q=
110 | github.com/google/btree v0.0.0-20180813153112-4030bb1f1f0c/go.mod h1:lNA+9X1NB3Zf8V7Ke586lFgjr2dZNuvo3lPJSGZ5JPQ=
111 | github.com/google/btree v1.0.0/go.mod h1:lNA+9X1NB3Zf8V7Ke586lFgjr2dZNuvo3lPJSGZ5JPQ=
112 | github.com/google/go-cmp v0.2.0/go.mod h1:oXzfMopK8JAjlY9xF4vHSVASa0yLyX7SntLO5aqRK0M=
113 | github.com/google/go-cmp v0.3.0/go.mod h1:8QqcDgzrUqlUb/G2PQTWiueGozuR1884gddMywk6iLU=
114 | github.com/google/go-cmp v0.3.1/go.mod h1:8QqcDgzrUqlUb/G2PQTWiueGozuR1884gddMywk6iLU=
115 | github.com/google/go-cmp v0.4.0/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
116 | github.com/google/go-cmp v0.4.1/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
117 | github.com/google/go-cmp v0.5.0/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
118 | github.com/google/go-cmp v0.5.1/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
119 | github.com/google/go-cmp v0.5.2/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
120 | github.com/google/go-cmp v0.5.3/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
121 | github.com/google/go-cmp v0.5.4/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
122 | github.com/google/go-cmp v0.5.5/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
123 | github.com/google/go-cmp v0.5.6/go.mod h1:v8dTdLbMG2kIc/vJvl+f65V22dbkXbowE6jgT/gNBxE=
124 | github.com/google/martian v2.1.0+incompatible/go.mod h1:9I4somxYTbIHy5NJKHRl3wXiIaQGbYVAs8BPL6v8lEs=
125 | github.com/google/martian/v3 v3.0.0/go.mod h1:y5Zk1BBys9G+gd6Jrk0W3cC1+ELVxBWuIGO+w/tUAp0=
126 | github.com/google/martian/v3 v3.1.0/go.mod h1:y5Zk1BBys9G+gd6Jrk0W3cC1+ELVxBWuIGO+w/tUAp0=
127 | github.com/google/martian/v3 v3.2.1/go.mod h1:oBOf6HBosgwRXnUGWUB05QECsc6uvmMiJ3+6W4l/CUk=
128 | github.com/google/pprof v0.0.0-20181206194817-3ea8567a2e57/go.mod h1:zfwlbNMJ+OItoe0UupaVj+oy1omPYYDuagoSzA8v9mc=
129 | github.com/google/pprof v0.0.0-20190515194954-54271f7e092f/go.mod h1:zfwlbNMJ+OItoe0UupaVj+oy1omPYYDuagoSzA8v9mc=
130 | github.com/google/pprof v0.0.0-20191218002539-d4f498aebedc/go.mod h1:ZgVRPoUq/hfqzAqh7sHMqb3I9Rq5C59dIz2SbBwJ4eM=
131 | github.com/google/pprof v0.0.0-20200212024743-f11f1df84d12/go.mod h1:ZgVRPoUq/hfqzAqh7sHMqb3I9Rq5C59dIz2SbBwJ4eM=
132 | github.com/google/pprof v0.0.0-20200229191704-1ebb73c60ed3/go.mod h1:ZgVRPoUq/hfqzAqh7sHMqb3I9Rq5C59dIz2SbBwJ4eM=
133 | github.com/google/pprof v0.0.0-20200430221834-fc25d7d30c6d/go.mod h1:ZgVRPoUq/hfqzAqh7sHMqb3I9Rq5C59dIz2SbBwJ4eM=
134 | github.com/google/pprof v0.0.0-20200708004538-1a94d8640e99/go.mod h1:ZgVRPoUq/hfqzAqh7sHMqb3I9Rq5C59dIz2SbBwJ4eM=
135 | github.com/google/pprof v0.0.0-20201023163331-3e6fc7fc9c4c/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
136 | github.com/google/pprof v0.0.0-20201203190320-1bf35d6f28c2/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
137 | github.com/google/pprof v0.0.0-20210122040257-d980be63207e/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
138 | github.com/google/pprof v0.0.0-20210226084205-cbba55b83ad5/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
139 | github.com/google/pprof v0.0.0-20210601050228-01bbb1931b22/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
140 | github.com/google/pprof v0.0.0-20210609004039-a478d1d731e9/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
141 | github.com/google/pprof v0.0.0-20210720184732-4bb14d4b1be1/go.mod h1:kpwsk12EmLew5upagYY7GY0pfYCcupk39gWOCRROcvE=
142 | github.com/google/renameio v0.1.0/go.mod h1:KWCgfxg9yswjAJkECMjeO8J8rahYeXnNhOm40UhjYkI=
143 | github.com/google/uuid v1.1.2/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
144 | github.com/googleapis/gax-go/v2 v2.0.4/go.mod h1:0Wqv26UfaUD9n4G6kQubkQ+KchISgw+vpHVxEJEs9eg=
145 | github.com/googleapis/gax-go/v2 v2.0.5/go.mod h1:DWXyrwAJ9X0FpwwEdw+IPEYBICEFu5mhpdKc/us6bOk=
146 | github.com/googleapis/gax-go/v2 v2.1.0/go.mod h1:Q3nei7sK6ybPYH7twZdmQpAd1MKb7pfu6SK+H1/DsU0=
147 | github.com/googleapis/gax-go/v2 v2.1.1/go.mod h1:hddJymUZASv3XPyGkUpKj8pPO47Rmb0eJc8R6ouapiM=
148 | github.com/gorilla/mux v1.8.0 h1:i40aqfkR1h2SlN9hojwV5ZA91wcXFOvkdNIeFDP5koI=
149 | github.com/gorilla/mux v1.8.0/go.mod h1:DVbg23sWSpFRCP0SfiEN6jmj59UnW/n46BH5rLB71So=
150 | github.com/grpc-ecosystem/grpc-gateway v1.16.0/go.mod h1:BDjrQk3hbvj6Nolgz8mAMFbcEtjT1g+wF4CSlocrBnw=
151 | github.com/hashicorp/golang-lru v0.5.0/go.mod h1:/m3WP610KZHVQ1SGc6re/UDhFvYD7pJ4Ao+sR/qLZy8=
152 | github.com/hashicorp/golang-lru v0.5.1/go.mod h1:/m3WP610KZHVQ1SGc6re/UDhFvYD7pJ4Ao+sR/qLZy8=
153 | github.com/ianlancetaylor/demangle v0.0.0-20181102032728-5e5cf60278f6/go.mod h1:aSSvb/t6k1mPoxDqO4vJh6VOCGPwU4O0C2/Eqndh1Sc=
154 | github.com/ianlancetaylor/demangle v0.0.0-20200824232613-28f6c0f3b639/go.mod h1:aSSvb/t6k1mPoxDqO4vJh6VOCGPwU4O0C2/Eqndh1Sc=
155 | github.com/jstemmer/go-junit-report v0.0.0-20190106144839-af01ea7f8024/go.mod h1:6v2b51hI/fHJwM22ozAgKL4VKDeJcHhJFhtBdhmNjmU=
156 | github.com/jstemmer/go-junit-report v0.9.1/go.mod h1:Brl9GWCQeLvo8nXZwPNNblvFj/XSXhF0NWZEnDohbsk=
157 | github.com/kisielk/gotool v1.0.0/go.mod h1:XhKaO+MFFWcvkIS/tQcRk01m1F5IRFswLeQ+oQHNcck=
158 | github.com/kr/pretty v0.1.0/go.mod h1:dAy3ld7l9f0ibDNOQOHHMYYIIbhfbHSm3C4ZsoJORNo=
159 | github.com/kr/pty v1.1.1/go.mod h1:pFQYn66WHrOpPYNljwOMqo10TkYh1fy3cYio2l3bCsQ=
160 | github.com/kr/text v0.1.0/go.mod h1:4Jbv+DJW3UT/LiOwJeYQe1efqtUx/iVham/4vfdArNI=
161 | github.com/pmezard/go-difflib v1.0.0/go.mod h1:iKH77koFhYxTK1pcRnkKkqfTogsbg7gZNVY4sRDYZ/4=
162 | github.com/prometheus/client_model v0.0.0-20190812154241-14fe0d1b01d4/go.mod h1:xMI15A0UPsDsEKsMN9yxemIoYk6Tm2C1GtYGdfGttqA=
163 | github.com/rogpeppe/fastuuid v1.2.0/go.mod h1:jVj6XXZzXRy/MSR5jhDC/2q6DgLz+nrA6LYCDYWNEvQ=
164 | github.com/rogpeppe/go-internal v1.3.0/go.mod h1:M8bDsm7K2OlrFYOpmOWEs/qY81heoFRclV5y23lUDJ4=
165 | github.com/spaolacci/murmur3 v0.0.0-20180118202830-f09979ecbc72/go.mod h1:JwIasOWyU6f++ZhiEuf87xNszmSA2myDM2Kzu9HwQUA=
166 | github.com/stretchr/objx v0.1.0/go.mod h1:HFkY916IF+rwdDfMAkV7OtwuqBVzrE8GR6GFx+wExME=
167 | github.com/stretchr/testify v1.4.0/go.mod h1:j7eGeouHqKxXV5pUuKE4zz7dFj8WfuZ+81PSLYec5m4=
168 | github.com/stretchr/testify v1.5.1/go.mod h1:5W2xD1RspED5o8YsWQXVCued0rvSQ+mT+I5cxcmMvtA=
169 | github.com/stretchr/testify v1.6.1/go.mod h1:6Fq8oRcR53rry900zMqJjRRixrwX3KX962/h/Wwjteg=
170 | github.com/yuin/goldmark v1.1.25/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
171 | github.com/yuin/goldmark v1.1.27/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
172 | github.com/yuin/goldmark v1.1.32/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
173 | github.com/yuin/goldmark v1.2.1/go.mod h1:3hX8gzYuyVAZsxl0MRgGTJEmQBFcNTphYh9decYSb74=
174 | github.com/yuin/goldmark v1.3.5/go.mod h1:mwnBkeHKe2W/ZEtQ+71ViKU8L12m81fl3OWwC1Zlc8k=
175 | go.opencensus.io v0.21.0/go.mod h1:mSImk1erAIZhrmZN+AvHh14ztQfjbGwt4TtuofqLduU=
176 | go.opencensus.io v0.22.0/go.mod h1:+kGneAE2xo2IficOXnaByMWTGM9T73dGwxeWcUqIpI8=
177 | go.opencensus.io v0.22.2/go.mod h1:yxeiOL68Rb0Xd1ddK5vPZ/oVn4vY4Ynel7k9FzqtOIw=
178 | go.opencensus.io v0.22.3/go.mod h1:yxeiOL68Rb0Xd1ddK5vPZ/oVn4vY4Ynel7k9FzqtOIw=
179 | go.opencensus.io v0.22.4/go.mod h1:yxeiOL68Rb0Xd1ddK5vPZ/oVn4vY4Ynel7k9FzqtOIw=
180 | go.opencensus.io v0.22.5/go.mod h1:5pWMHQbX5EPX2/62yrJeAkowc+lfs/XD7Uxpq3pI6kk=
181 | go.opencensus.io v0.23.0 h1:gqCw0LfLxScz8irSi8exQc7fyQ0fKQU/qnC/X8+V/1M=
182 | go.opencensus.io v0.23.0/go.mod h1:XItmlyltB5F7CS4xOC1DcqMoFqwtC6OG2xF7mCv7P7E=
183 | go.opentelemetry.io/proto/otlp v0.7.0/go.mod h1:PqfVotwruBrMGOCsRd/89rSnXhoiJIqeYNgFYFoEGnI=
184 | golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2/go.mod h1:djNgcEr1/C05ACkg1iLfiJU5Ep61QUkGW8qpdssI0+w=
185 | golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529/go.mod h1:yigFU9vqHzYiE8UmvKecakEJjdnWj3jj499lnFckfCI=
186 | golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5/go.mod h1:yigFU9vqHzYiE8UmvKecakEJjdnWj3jj499lnFckfCI=
187 | golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550/go.mod h1:yigFU9vqHzYiE8UmvKecakEJjdnWj3jj499lnFckfCI=
188 | golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9/go.mod h1:LzIPMQfyMNhhGPhUkYOs5KpL4U8rLKemX1yGLhDgUto=
189 | golang.org/x/exp v0.0.0-20190121172915-509febef88a4/go.mod h1:CJ0aWSM057203Lf6IL+f9T1iT9GByDxfZKAQTCR3kQA=
190 | golang.org/x/exp v0.0.0-20190306152737-a1d7652674e8/go.mod h1:CJ0aWSM057203Lf6IL+f9T1iT9GByDxfZKAQTCR3kQA=
191 | golang.org/x/exp v0.0.0-20190510132918-efd6b22b2522/go.mod h1:ZjyILWgesfNpC6sMxTJOJm9Kp84zZh5NQWvqDGG3Qr8=
192 | golang.org/x/exp v0.0.0-20190829153037-c13cbed26979/go.mod h1:86+5VVa7VpoJ4kLfm080zCjGlMRFzhUhsZKEZO7MGek=
193 | golang.org/x/exp v0.0.0-20191030013958-a1ab85dbe136/go.mod h1:JXzH8nQsPlswgeRAPE3MuO9GYsAcnJvJ4vnMwN/5qkY=
194 | golang.org/x/exp v0.0.0-20191129062945-2f5052295587/go.mod h1:2RIsYlXP63K8oxa1u096TMicItID8zy7Y6sNkU49FU4=
195 | golang.org/x/exp v0.0.0-20191227195350-da58074b4299/go.mod h1:2RIsYlXP63K8oxa1u096TMicItID8zy7Y6sNkU49FU4=
196 | golang.org/x/exp v0.0.0-20200119233911-0405dc783f0a/go.mod h1:2RIsYlXP63K8oxa1u096TMicItID8zy7Y6sNkU49FU4=
197 | golang.org/x/exp v0.0.0-20200207192155-f17229e696bd/go.mod h1:J/WKrq2StrnmMY6+EHIKF9dgMWnmCNThgcyBT1FY9mM=
198 | golang.org/x/exp v0.0.0-20200224162631-6cc2880d07d6/go.mod h1:3jZMyOhIsHpP37uCMkUooju7aAi5cS1Q23tOzKc+0MU=
199 | golang.org/x/image v0.0.0-20190227222117-0694c2d4d067/go.mod h1:kZ7UVZpmo3dzQBMxlp+ypCbDeSB+sBbTgSJuh5dn5js=
200 | golang.org/x/image v0.0.0-20190802002840-cff245a6509b/go.mod h1:FeLwcggjj3mMvU+oOTbSwawSJRM1uh48EjtB4UJZlP0=
201 | golang.org/x/lint v0.0.0-20181026193005-c67002cb31c3/go.mod h1:UVdnD1Gm6xHRNCYTkRU2/jEulfH38KcIWyp/GAMgvoE=
202 | golang.org/x/lint v0.0.0-20190227174305-5b3e6a55c961/go.mod h1:wehouNa3lNwaWXcvxsM5YxQ5yQlVC4a0KAMCusXpPoU=
203 | golang.org/x/lint v0.0.0-20190301231843-5614ed5bae6f/go.mod h1:UVdnD1Gm6xHRNCYTkRU2/jEulfH38KcIWyp/GAMgvoE=
204 | golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3/go.mod h1:6SW0HCj/g11FgYtHlgUYUwCkIfeOF89ocIRzGO/8vkc=
205 | golang.org/x/lint v0.0.0-20190409202823-959b441ac422/go.mod h1:6SW0HCj/g11FgYtHlgUYUwCkIfeOF89ocIRzGO/8vkc=
206 | golang.org/x/lint v0.0.0-20190909230951-414d861bb4ac/go.mod h1:6SW0HCj/g11FgYtHlgUYUwCkIfeOF89ocIRzGO/8vkc=
207 | golang.org/x/lint v0.0.0-20190930215403-16217165b5de/go.mod h1:6SW0HCj/g11FgYtHlgUYUwCkIfeOF89ocIRzGO/8vkc=
208 | golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f/go.mod h1:5qLYkcX4OjUUV8bRuDixDT3tpyyb+LUpUlRWLxfhWrs=
209 | golang.org/x/lint v0.0.0-20200130185559-910be7a94367/go.mod h1:3xt1FjdF8hUf6vQPIChWIBhFzV8gjjsPE/fR3IyQdNY=
210 | golang.org/x/lint v0.0.0-20200302205851-738671d3881b/go.mod h1:3xt1FjdF8hUf6vQPIChWIBhFzV8gjjsPE/fR3IyQdNY=
211 | golang.org/x/lint v0.0.0-20201208152925-83fdc39ff7b5/go.mod h1:3xt1FjdF8hUf6vQPIChWIBhFzV8gjjsPE/fR3IyQdNY=
212 | golang.org/x/lint v0.0.0-20210508222113-6edffad5e616/go.mod h1:3xt1FjdF8hUf6vQPIChWIBhFzV8gjjsPE/fR3IyQdNY=
213 | golang.org/x/mobile v0.0.0-20190312151609-d3739f865fa6/go.mod h1:z+o9i4GpDbdi3rU15maQ/Ox0txvL9dWGYEHz965HBQE=
214 | golang.org/x/mobile v0.0.0-20190719004257-d2bd2a29d028/go.mod h1:E/iHnbuqvinMTCcRqshq8CkpyQDoeVncDDYHnLhea+o=
215 | golang.org/x/mod v0.0.0-20190513183733-4bf6d317e70e/go.mod h1:mXi4GBBbnImb6dmsKGUJ2LatrhH/nqhxcFungHvyanc=
216 | golang.org/x/mod v0.1.0/go.mod h1:0QHyrYULN0/3qlju5TqG8bIK38QM8yzMo5ekMj3DlcY=
217 | golang.org/x/mod v0.1.1-0.20191105210325-c90efee705ee/go.mod h1:QqPTAvyqsEbceGzBzNggFXnrqF1CaUcvgkdR5Ot7KZg=
218 | golang.org/x/mod v0.1.1-0.20191107180719-034126e5016b/go.mod h1:QqPTAvyqsEbceGzBzNggFXnrqF1CaUcvgkdR5Ot7KZg=
219 | golang.org/x/mod v0.2.0/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
220 | golang.org/x/mod v0.3.0/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
221 | golang.org/x/mod v0.4.0/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
222 | golang.org/x/mod v0.4.1/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
223 | golang.org/x/mod v0.4.2/go.mod h1:s0Qsj1ACt9ePp/hMypM3fl4fZqREWJwdYDEqhRiZZUA=
224 | golang.org/x/net v0.0.0-20180724234803-3673e40ba225/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
225 | golang.org/x/net v0.0.0-20180826012351-8a410e7b638d/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
226 | golang.org/x/net v0.0.0-20190108225652-1e06a53dbb7e/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
227 | golang.org/x/net v0.0.0-20190213061140-3a22650c66bd/go.mod h1:mL1N/T3taQHkDXs73rZJwtUhF3w3ftmwwsq0BUmARs4=
228 | golang.org/x/net v0.0.0-20190311183353-d8887717615a/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
229 | golang.org/x/net v0.0.0-20190404232315-eb5bcb51f2a3/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
230 | golang.org/x/net v0.0.0-20190501004415-9ce7a6920f09/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
231 | golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c/go.mod h1:t9HGtf8HONx5eT2rtn7q6eTqICYqUVnKs3thJo3Qplg=
232 | golang.org/x/net v0.0.0-20190603091049-60506f45cf65/go.mod h1:HSz+uSET+XFnRR8LxR5pz3Of3rY3CfYBVs4xY44aLks=
233 | golang.org/x/net v0.0.0-20190620200207-3b0461eec859/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
234 | golang.org/x/net v0.0.0-20190628185345-da137c7871d7/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
235 | golang.org/x/net v0.0.0-20190724013045-ca1201d0de80/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
236 | golang.org/x/net v0.0.0-20191209160850-c0dbc17a3553/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
237 | golang.org/x/net v0.0.0-20200114155413-6afb5195e5aa/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
238 | golang.org/x/net v0.0.0-20200202094626-16171245cfb2/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
239 | golang.org/x/net v0.0.0-20200222125558-5a598a2470a0/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
240 | golang.org/x/net v0.0.0-20200226121028-0de0cce0169b/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
241 | golang.org/x/net v0.0.0-20200301022130-244492dfa37a/go.mod h1:z5CRVTTTmAJ677TzLLGU+0bjPO0LkuOLi4/5GtJWs/s=
242 | golang.org/x/net v0.0.0-20200324143707-d3edc9973b7e/go.mod h1:qpuaurCH72eLCgpAm/N6yyVIVM9cpaDIP3A8BGJEC5A=
243 | golang.org/x/net v0.0.0-20200501053045-e0ff5e5a1de5/go.mod h1:qpuaurCH72eLCgpAm/N6yyVIVM9cpaDIP3A8BGJEC5A=
244 | golang.org/x/net v0.0.0-20200506145744-7e3656a0809f/go.mod h1:qpuaurCH72eLCgpAm/N6yyVIVM9cpaDIP3A8BGJEC5A=
245 | golang.org/x/net v0.0.0-20200513185701-a91f0712d120/go.mod h1:qpuaurCH72eLCgpAm/N6yyVIVM9cpaDIP3A8BGJEC5A=
246 | golang.org/x/net v0.0.0-20200520182314-0ba52f642ac2/go.mod h1:qpuaurCH72eLCgpAm/N6yyVIVM9cpaDIP3A8BGJEC5A=
247 | golang.org/x/net v0.0.0-20200625001655-4c5254603344/go.mod h1:/O7V0waA8r7cgGh81Ro3o1hOxt32SMVPicZroKQ2sZA=
248 | golang.org/x/net v0.0.0-20200707034311-ab3426394381/go.mod h1:/O7V0waA8r7cgGh81Ro3o1hOxt32SMVPicZroKQ2sZA=
249 | golang.org/x/net v0.0.0-20200822124328-c89045814202/go.mod h1:/O7V0waA8r7cgGh81Ro3o1hOxt32SMVPicZroKQ2sZA=
250 | golang.org/x/net v0.0.0-20201021035429-f5854403a974/go.mod h1:sp8m0HH+o8qH0wwXwYZr8TS3Oi6o0r6Gce1SSxlDquU=
251 | golang.org/x/net v0.0.0-20201031054903-ff519b6c9102/go.mod h1:sp8m0HH+o8qH0wwXwYZr8TS3Oi6o0r6Gce1SSxlDquU=
252 | golang.org/x/net v0.0.0-20201110031124-69a78807bb2b/go.mod h1:sp8m0HH+o8qH0wwXwYZr8TS3Oi6o0r6Gce1SSxlDquU=
253 | golang.org/x/net v0.0.0-20201209123823-ac852fbbde11/go.mod h1:m0MpNAwzfU5UDzcl9v0D8zg8gWTRqZa9RBIspLL5mdg=
254 | golang.org/x/net v0.0.0-20210119194325-5f4716e94777/go.mod h1:m0MpNAwzfU5UDzcl9v0D8zg8gWTRqZa9RBIspLL5mdg=
255 | golang.org/x/net v0.0.0-20210226172049-e18ecbb05110/go.mod h1:m0MpNAwzfU5UDzcl9v0D8zg8gWTRqZa9RBIspLL5mdg=
256 | golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4/go.mod h1:RBQZq4jEuRlivfhVLdyRGr576XBO4/greRjx4P4O3yc=
257 | golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4/go.mod h1:p54w0d4576C0XHj96bSt6lcn1PtDYWL6XObtHCRCNQM=
258 | golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420 h1:a8jGStKg0XqKDlKqjLrXn0ioF5MH36pT7Z0BRTqLhbk=
259 | golang.org/x/net v0.0.0-20210503060351-7fd8e65b6420/go.mod h1:9nx3DQGgdP8bBQD5qxJ1jj9UTztislL4KSBs9R2vV5Y=
260 | golang.org/x/oauth2 v0.0.0-20180821212333-d2e6202438be/go.mod h1:N/0e6XlmueqKjAGxoOufVs8QHGRruUQn6yWY3a++T0U=
261 | golang.org/x/oauth2 v0.0.0-20190226205417-e64efc72b421/go.mod h1:gOpvHmFTYa4IltrdGE7lF6nIHvwfUNPOp7c8zoXwtLw=
262 | golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45/go.mod h1:gOpvHmFTYa4IltrdGE7lF6nIHvwfUNPOp7c8zoXwtLw=
263 | golang.org/x/oauth2 v0.0.0-20191202225959-858c2ad4c8b6/go.mod h1:gOpvHmFTYa4IltrdGE7lF6nIHvwfUNPOp7c8zoXwtLw=
264 | golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d/go.mod h1:gOpvHmFTYa4IltrdGE7lF6nIHvwfUNPOp7c8zoXwtLw=
265 | golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
266 | golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
267 | golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
268 | golang.org/x/oauth2 v0.0.0-20210218202405-ba52d332ba99/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
269 | golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
270 | golang.org/x/oauth2 v0.0.0-20210313182246-cd4f82c27b84/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
271 | golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
272 | golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
273 | golang.org/x/oauth2 v0.0.0-20210805134026-6f1e6394065a/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
274 | golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
275 | golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 h1:RerP+noqYHUQ8CMRcPlC2nvTa4dcBIjegkuWdcUDuqg=
276 | golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8/go.mod h1:KelEdhl1UZF7XfJ4dDtk6s++YSgaE7mD/BuKKDLBl4A=
277 | golang.org/x/sync v0.0.0-20180314180146-1d60e4601c6f/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
278 | golang.org/x/sync v0.0.0-20181108010431-42b317875d0f/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
279 | golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
280 | golang.org/x/sync v0.0.0-20190227155943-e225da77a7e6/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
281 | golang.org/x/sync v0.0.0-20190423024810-112230192c58/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
282 | golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
283 | golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
284 | golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
285 | golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
286 | golang.org/x/sync v0.0.0-20201207232520-09787c993a3a/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
287 | golang.org/x/sync v0.0.0-20210220032951-036812b2e83c/go.mod h1:RxMgew5VJxzue5/jJTE5uejpjVlOe/izrB70Jof72aM=
288 | golang.org/x/sys v0.0.0-20180830151530-49385e6e1522/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
289 | golang.org/x/sys v0.0.0-20190215142949-d0b11bdaac8a/go.mod h1:STP8DvDyc/dI5b8T5hshtkjS+E42TnysNCUPdjciGhY=
290 | golang.org/x/sys v0.0.0-20190312061237-fead79001313/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
291 | golang.org/x/sys v0.0.0-20190412213103-97732733099d/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
292 | golang.org/x/sys v0.0.0-20190502145724-3ef323f4f1fd/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
293 | golang.org/x/sys v0.0.0-20190507160741-ecd444e8653b/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
294 | golang.org/x/sys v0.0.0-20190606165138-5da285871e9c/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
295 | golang.org/x/sys v0.0.0-20190624142023-c5567b49c5d0/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
296 | golang.org/x/sys v0.0.0-20190726091711-fc99dfbffb4e/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
297 | golang.org/x/sys v0.0.0-20191001151750-bb3f8db39f24/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
298 | golang.org/x/sys v0.0.0-20191204072324-ce4227a45e2e/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
299 | golang.org/x/sys v0.0.0-20191228213918-04cbcbbfeed8/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
300 | golang.org/x/sys v0.0.0-20200113162924-86b910548bc1/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
301 | golang.org/x/sys v0.0.0-20200122134326-e047566fdf82/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
302 | golang.org/x/sys v0.0.0-20200202164722-d101bd2416d5/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
303 | golang.org/x/sys v0.0.0-20200212091648-12a6c2dcc1e4/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
304 | golang.org/x/sys v0.0.0-20200223170610-d5e6a3e2c0ae/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
305 | golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
306 | golang.org/x/sys v0.0.0-20200323222414-85ca7c5b95cd/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
307 | golang.org/x/sys v0.0.0-20200331124033-c3d80250170d/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
308 | golang.org/x/sys v0.0.0-20200501052902-10377860bb8e/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
309 | golang.org/x/sys v0.0.0-20200511232937-7e40ca221e25/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
310 | golang.org/x/sys v0.0.0-20200515095857-1151b9dac4a9/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
311 | golang.org/x/sys v0.0.0-20200523222454-059865788121/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
312 | golang.org/x/sys v0.0.0-20200803210538-64077c9b5642/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
313 | golang.org/x/sys v0.0.0-20200905004654-be1d3432aa8f/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
314 | golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
315 | golang.org/x/sys v0.0.0-20201119102817-f84b799fce68/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
316 | golang.org/x/sys v0.0.0-20201201145000-ef89a241ccb3/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
317 | golang.org/x/sys v0.0.0-20210104204734-6f8348627aad/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
318 | golang.org/x/sys v0.0.0-20210119212857-b64e53b001e4/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
319 | golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
320 | golang.org/x/sys v0.0.0-20210305230114-8fe3ee5dd75b/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
321 | golang.org/x/sys v0.0.0-20210315160823-c6e025ad8005/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
322 | golang.org/x/sys v0.0.0-20210320140829-1e4c9ba3b0c4/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
323 | golang.org/x/sys v0.0.0-20210330210617-4fbd30eecc44/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
324 | golang.org/x/sys v0.0.0-20210423082822-04245dca01da/go.mod h1:h1NjWce9XRLGQEsW7wpKNCjG9DtNlClVuFLEZdDNbEs=
325 | golang.org/x/sys v0.0.0-20210510120138-977fb7262007/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
326 | golang.org/x/sys v0.0.0-20210514084401-e8d321eab015/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
327 | golang.org/x/sys v0.0.0-20210603125802-9665404d3644/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
328 | golang.org/x/sys v0.0.0-20210616094352-59db8d763f22/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
329 | golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
330 | golang.org/x/sys v0.0.0-20210806184541-e5e7981a1069/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
331 | golang.org/x/sys v0.0.0-20210823070655-63515b42dcdf/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
332 | golang.org/x/sys v0.0.0-20210908233432-aa78b53d3365/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
333 | golang.org/x/sys v0.0.0-20211124211545-fe61309f8881/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
334 | golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e h1:fLOSk5Q00efkSvAm+4xcoXD+RRmLmmulPn5I3Y9F2EM=
335 | golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e/go.mod h1:oPkhp1MJrh7nUepCBck5+mAzfO9JrbApNNgaTdGDITg=
336 | golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1/go.mod h1:bj7SfCRtBDWHUb9snDiAeCFNEtKQo2Wmx5Cou7ajbmo=
337 | golang.org/x/text v0.0.0-20170915032832-14c0d48ead0c/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
338 | golang.org/x/text v0.3.0/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
339 | golang.org/x/text v0.3.1-0.20180807135948-17ff2d5776d2/go.mod h1:NqM8EUOU14njkJ3fqMW+pc6Ldnwhi/IjpwHt7yyuwOQ=
340 | golang.org/x/text v0.3.2/go.mod h1:bEr9sfX3Q8Zfm5fL9x+3itogRgK3+ptLWKqgva+5dAk=
341 | golang.org/x/text v0.3.3/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
342 | golang.org/x/text v0.3.4/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
343 | golang.org/x/text v0.3.5/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
344 | golang.org/x/text v0.3.6 h1:aRYxNxv6iGQlyVaZmk6ZgYEDa+Jg18DxebPSrd6bg1M=
345 | golang.org/x/text v0.3.6/go.mod h1:5Zoc/QRtKVWzQhOtBMvqHzDpF6irO9z98xDceosuGiQ=
346 | golang.org/x/time v0.0.0-20181108054448-85acf8d2951c/go.mod h1:tRJNPiyCQ0inRvYxbN9jk5I+vvW/OXSQhTDSoE431IQ=
347 | golang.org/x/time v0.0.0-20190308202827-9d24e82272b4/go.mod h1:tRJNPiyCQ0inRvYxbN9jk5I+vvW/OXSQhTDSoE431IQ=
348 | golang.org/x/time v0.0.0-20191024005414-555d28b269f0/go.mod h1:tRJNPiyCQ0inRvYxbN9jk5I+vvW/OXSQhTDSoE431IQ=
349 | golang.org/x/tools v0.0.0-20180917221912-90fa682c2a6e/go.mod h1:n7NCudcB/nEzxVGmLbDWY5pfWTLqBcC2KZ6jyYvM4mQ=
350 | golang.org/x/tools v0.0.0-20190114222345-bf090417da8b/go.mod h1:n7NCudcB/nEzxVGmLbDWY5pfWTLqBcC2KZ6jyYvM4mQ=
351 | golang.org/x/tools v0.0.0-20190226205152-f727befe758c/go.mod h1:9Yl7xja0Znq3iFh3HoIrodX9oNMXvdceNzlUR8zjMvY=
352 | golang.org/x/tools v0.0.0-20190311212946-11955173bddd/go.mod h1:LCzVGOaR6xXOjkQ3onu1FJEFr0SW1gC7cKk1uF8kGRs=
353 | golang.org/x/tools v0.0.0-20190312151545-0bb0c0a6e846/go.mod h1:LCzVGOaR6xXOjkQ3onu1FJEFr0SW1gC7cKk1uF8kGRs=
354 | golang.org/x/tools v0.0.0-20190312170243-e65039ee4138/go.mod h1:LCzVGOaR6xXOjkQ3onu1FJEFr0SW1gC7cKk1uF8kGRs=
355 | golang.org/x/tools v0.0.0-20190425150028-36563e24a262/go.mod h1:RgjU9mgBXZiqYHBnxXauZ1Gv1EHHAz9KjViQ78xBX0Q=
356 | golang.org/x/tools v0.0.0-20190506145303-2d16b83fe98c/go.mod h1:RgjU9mgBXZiqYHBnxXauZ1Gv1EHHAz9KjViQ78xBX0Q=
357 | golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135/go.mod h1:RgjU9mgBXZiqYHBnxXauZ1Gv1EHHAz9KjViQ78xBX0Q=
358 | golang.org/x/tools v0.0.0-20190606124116-d0a3d012864b/go.mod h1:/rFqwRUd4F7ZHNgwSSTFct+R/Kf4OFW1sUzUTQQTgfc=
359 | golang.org/x/tools v0.0.0-20190621195816-6e04913cbbac/go.mod h1:/rFqwRUd4F7ZHNgwSSTFct+R/Kf4OFW1sUzUTQQTgfc=
360 | golang.org/x/tools v0.0.0-20190628153133-6cdbf07be9d0/go.mod h1:/rFqwRUd4F7ZHNgwSSTFct+R/Kf4OFW1sUzUTQQTgfc=
361 | golang.org/x/tools v0.0.0-20190816200558-6889da9d5479/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
362 | golang.org/x/tools v0.0.0-20190911174233-4f2ddba30aff/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
363 | golang.org/x/tools v0.0.0-20191012152004-8de300cfc20a/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
364 | golang.org/x/tools v0.0.0-20191113191852-77e3bb0ad9e7/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
365 | golang.org/x/tools v0.0.0-20191115202509-3a792d9c32b2/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
366 | golang.org/x/tools v0.0.0-20191119224855-298f0cb1881e/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
367 | golang.org/x/tools v0.0.0-20191125144606-a911d9008d1f/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
368 | golang.org/x/tools v0.0.0-20191130070609-6e064ea0cf2d/go.mod h1:b+2E5dAYhXwXZwtnZ6UAqBI28+e2cm9otk0dWdXHAEo=
369 | golang.org/x/tools v0.0.0-20191216173652-a0e659d51361/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
370 | golang.org/x/tools v0.0.0-20191227053925-7b8e75db28f4/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
371 | golang.org/x/tools v0.0.0-20200117161641-43d50277825c/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
372 | golang.org/x/tools v0.0.0-20200122220014-bf1340f18c4a/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
373 | golang.org/x/tools v0.0.0-20200130002326-2f3ba24bd6e7/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
374 | golang.org/x/tools v0.0.0-20200204074204-1cc6d1ef6c74/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
375 | golang.org/x/tools v0.0.0-20200207183749-b753a1ba74fa/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
376 | golang.org/x/tools v0.0.0-20200212150539-ea181f53ac56/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
377 | golang.org/x/tools v0.0.0-20200224181240-023911ca70b2/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
378 | golang.org/x/tools v0.0.0-20200227222343-706bc42d1f0d/go.mod h1:TB2adYChydJhpapKDTa4BR/hXlZSLoq2Wpct/0txZ28=
379 | golang.org/x/tools v0.0.0-20200304193943-95d2e580d8eb/go.mod h1:o4KQGtdN14AW+yjsvvwRTJJuXz8XRtIHtEnmAXLyFUw=
380 | golang.org/x/tools v0.0.0-20200312045724-11d5b4c81c7d/go.mod h1:o4KQGtdN14AW+yjsvvwRTJJuXz8XRtIHtEnmAXLyFUw=
381 | golang.org/x/tools v0.0.0-20200331025713-a30bf2db82d4/go.mod h1:Sl4aGygMT6LrqrWclx+PTx3U+LnKx/seiNR+3G19Ar8=
382 | golang.org/x/tools v0.0.0-20200501065659-ab2804fb9c9d/go.mod h1:EkVYQZoAsY45+roYkvgYkIh4xh/qjgUK9TdY2XT94GE=
383 | golang.org/x/tools v0.0.0-20200512131952-2bc93b1c0c88/go.mod h1:EkVYQZoAsY45+roYkvgYkIh4xh/qjgUK9TdY2XT94GE=
384 | golang.org/x/tools v0.0.0-20200515010526-7d3b6ebf133d/go.mod h1:EkVYQZoAsY45+roYkvgYkIh4xh/qjgUK9TdY2XT94GE=
385 | golang.org/x/tools v0.0.0-20200618134242-20370b0cb4b2/go.mod h1:EkVYQZoAsY45+roYkvgYkIh4xh/qjgUK9TdY2XT94GE=
386 | golang.org/x/tools v0.0.0-20200729194436-6467de6f59a7/go.mod h1:njjCfa9FT2d7l9Bc6FUM5FLjQPp3cFF28FI3qnDFljA=
387 | golang.org/x/tools v0.0.0-20200804011535-6c149bb5ef0d/go.mod h1:njjCfa9FT2d7l9Bc6FUM5FLjQPp3cFF28FI3qnDFljA=
388 | golang.org/x/tools v0.0.0-20200825202427-b303f430e36d/go.mod h1:njjCfa9FT2d7l9Bc6FUM5FLjQPp3cFF28FI3qnDFljA=
389 | golang.org/x/tools v0.0.0-20200904185747-39188db58858/go.mod h1:Cj7w3i3Rnn0Xh82ur9kSqwfTHTeVxaDqrfMjpcNT6bE=
390 | golang.org/x/tools v0.0.0-20201110124207-079ba7bd75cd/go.mod h1:emZCQorbCU4vsT4fOWvOPXz4eW1wZW4PmDk9uLelYpA=
391 | golang.org/x/tools v0.0.0-20201201161351-ac6f37ff4c2a/go.mod h1:emZCQorbCU4vsT4fOWvOPXz4eW1wZW4PmDk9uLelYpA=
392 | golang.org/x/tools v0.0.0-20201208233053-a543418bbed2/go.mod h1:emZCQorbCU4vsT4fOWvOPXz4eW1wZW4PmDk9uLelYpA=
393 | golang.org/x/tools v0.0.0-20210105154028-b0ab187a4818/go.mod h1:emZCQorbCU4vsT4fOWvOPXz4eW1wZW4PmDk9uLelYpA=
394 | golang.org/x/tools v0.1.0/go.mod h1:xkSsbof2nBLbhDlRMhhhyNLN/zl3eTqcnHD5viDpcZ0=
395 | golang.org/x/tools v0.1.1/go.mod h1:o0xws9oXOQQZyjljx8fwUC0k7L1pTE6eaCbjGeHmOkk=
396 | golang.org/x/tools v0.1.2/go.mod h1:o0xws9oXOQQZyjljx8fwUC0k7L1pTE6eaCbjGeHmOkk=
397 | golang.org/x/tools v0.1.3/go.mod h1:o0xws9oXOQQZyjljx8fwUC0k7L1pTE6eaCbjGeHmOkk=
398 | golang.org/x/tools v0.1.4/go.mod h1:o0xws9oXOQQZyjljx8fwUC0k7L1pTE6eaCbjGeHmOkk=
399 | golang.org/x/tools v0.1.5/go.mod h1:o0xws9oXOQQZyjljx8fwUC0k7L1pTE6eaCbjGeHmOkk=
400 | golang.org/x/xerrors v0.0.0-20190717185122-a985d3407aa7/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
401 | golang.org/x/xerrors v0.0.0-20191011141410-1b5146add898/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
402 | golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
403 | golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1/go.mod h1:I/5z698sn9Ka8TeJc9MKroUUfqBBauWjQqLJ2OPfmY0=
404 | google.golang.org/api v0.4.0/go.mod h1:8k5glujaEP+g9n7WNsDg8QP6cUVNI86fCNMcbazEtwE=
405 | google.golang.org/api v0.7.0/go.mod h1:WtwebWUNSVBH/HAw79HIFXZNqEvBhG+Ra+ax0hx3E3M=
406 | google.golang.org/api v0.8.0/go.mod h1:o4eAsZoiT+ibD93RtjEohWalFOjRDx6CVaqeizhEnKg=
407 | google.golang.org/api v0.9.0/go.mod h1:o4eAsZoiT+ibD93RtjEohWalFOjRDx6CVaqeizhEnKg=
408 | google.golang.org/api v0.13.0/go.mod h1:iLdEw5Ide6rF15KTC1Kkl0iskquN2gFfn9o9XIsbkAI=
409 | google.golang.org/api v0.14.0/go.mod h1:iLdEw5Ide6rF15KTC1Kkl0iskquN2gFfn9o9XIsbkAI=
410 | google.golang.org/api v0.15.0/go.mod h1:iLdEw5Ide6rF15KTC1Kkl0iskquN2gFfn9o9XIsbkAI=
411 | google.golang.org/api v0.17.0/go.mod h1:BwFmGc8tA3vsd7r/7kR8DY7iEEGSU04BFxCo5jP/sfE=
412 | google.golang.org/api v0.18.0/go.mod h1:BwFmGc8tA3vsd7r/7kR8DY7iEEGSU04BFxCo5jP/sfE=
413 | google.golang.org/api v0.19.0/go.mod h1:BwFmGc8tA3vsd7r/7kR8DY7iEEGSU04BFxCo5jP/sfE=
414 | google.golang.org/api v0.20.0/go.mod h1:BwFmGc8tA3vsd7r/7kR8DY7iEEGSU04BFxCo5jP/sfE=
415 | google.golang.org/api v0.22.0/go.mod h1:BwFmGc8tA3vsd7r/7kR8DY7iEEGSU04BFxCo5jP/sfE=
416 | google.golang.org/api v0.24.0/go.mod h1:lIXQywCXRcnZPGlsd8NbLnOjtAoL6em04bJ9+z0MncE=
417 | google.golang.org/api v0.28.0/go.mod h1:lIXQywCXRcnZPGlsd8NbLnOjtAoL6em04bJ9+z0MncE=
418 | google.golang.org/api v0.29.0/go.mod h1:Lcubydp8VUV7KeIHD9z2Bys/sm/vGKnG1UHuDBSrHWM=
419 | google.golang.org/api v0.30.0/go.mod h1:QGmEvQ87FHZNiUVJkT14jQNYJ4ZJjdRF23ZXz5138Fc=
420 | google.golang.org/api v0.35.0/go.mod h1:/XrVsuzM0rZmrsbjJutiuftIzeuTQcEeaYcSk/mQ1dg=
421 | google.golang.org/api v0.36.0/go.mod h1:+z5ficQTmoYpPn8LCUNVpK5I7hwkpjbcgqA7I34qYtE=
422 | google.golang.org/api v0.40.0/go.mod h1:fYKFpnQN0DsDSKRVRcQSDQNtqWPfM9i+zNPxepjRCQ8=
423 | google.golang.org/api v0.41.0/go.mod h1:RkxM5lITDfTzmyKFPt+wGrCJbVfniCr2ool8kTBzRTU=
424 | google.golang.org/api v0.43.0/go.mod h1:nQsDGjRXMo4lvh5hP0TKqF244gqhGcr/YSIykhUk/94=
425 | google.golang.org/api v0.47.0/go.mod h1:Wbvgpq1HddcWVtzsVLyfLp8lDg6AA241LmgIL59tHXo=
426 | google.golang.org/api v0.48.0/go.mod h1:71Pr1vy+TAZRPkPs/xlCf5SsU8WjuAWv1Pfjbtukyy4=
427 | google.golang.org/api v0.50.0/go.mod h1:4bNT5pAuq5ji4SRZm+5QIkjny9JAyVD/3gaSihNefaw=
428 | google.golang.org/api v0.51.0/go.mod h1:t4HdrdoNgyN5cbEfm7Lum0lcLDLiise1F8qDKX00sOU=
429 | google.golang.org/api v0.54.0/go.mod h1:7C4bFFOvVDGXjfDTAsgGwDgAxRDeQ4X8NvUedIt6z3k=
430 | google.golang.org/api v0.55.0/go.mod h1:38yMfeP1kfjsl8isn0tliTjIb1rJXcQi4UXlbqivdVE=
431 | google.golang.org/api v0.56.0/go.mod h1:38yMfeP1kfjsl8isn0tliTjIb1rJXcQi4UXlbqivdVE=
432 | google.golang.org/api v0.57.0/go.mod h1:dVPlbZyBo2/OjBpmvNdpn2GRm6rPy75jyU7bmhdrMgI=
433 | google.golang.org/api v0.61.0/go.mod h1:xQRti5UdCmoCEqFxcz93fTl338AVqDgyaDRuOZ3hg9I=
434 | google.golang.org/api v0.64.0 h1:l3pi8ncrQgB9+ncFw3A716L8lWujnXniBYbxWqqy6tE=
435 | google.golang.org/api v0.64.0/go.mod h1:931CdxA8Rm4t6zqTFGSsgwbAEZ2+GMYurbndwSimebM=
436 | google.golang.org/appengine v1.1.0/go.mod h1:EbEs0AVv82hx2wNQdGPgUI5lhzA/G0D9YwlJXL52JkM=
437 | google.golang.org/appengine v1.4.0/go.mod h1:xpcJRLb0r/rnEns0DIKYYv+WjYCduHsrkT7/EB5XEv4=
438 | google.golang.org/appengine v1.5.0/go.mod h1:xpcJRLb0r/rnEns0DIKYYv+WjYCduHsrkT7/EB5XEv4=
439 | google.golang.org/appengine v1.6.1/go.mod h1:i06prIuMbXzDqacNJfV5OdTW448YApPu5ww/cMBSeb0=
440 | google.golang.org/appengine v1.6.5/go.mod h1:8WjMMxjGQR8xUklV/ARdw2HLXBOI7O7uCIDZVag1xfc=
441 | google.golang.org/appengine v1.6.6/go.mod h1:8WjMMxjGQR8xUklV/ARdw2HLXBOI7O7uCIDZVag1xfc=
442 | google.golang.org/appengine v1.6.7 h1:FZR1q0exgwxzPzp/aF+VccGrSfxfPpkBqjIIEq3ru6c=
443 | google.golang.org/appengine v1.6.7/go.mod h1:8WjMMxjGQR8xUklV/ARdw2HLXBOI7O7uCIDZVag1xfc=
444 | google.golang.org/genproto v0.0.0-20180817151627-c66870c02cf8/go.mod h1:JiN7NxoALGmiZfu7CAH4rXhgtRTLTxftemlI0sWmxmc=
445 | google.golang.org/genproto v0.0.0-20190307195333-5fe7a883aa19/go.mod h1:VzzqZJRnGkLBvHegQrXjBqPurQTc5/KpmUdxsrq26oE=
446 | google.golang.org/genproto v0.0.0-20190418145605-e7d98fc518a7/go.mod h1:VzzqZJRnGkLBvHegQrXjBqPurQTc5/KpmUdxsrq26oE=
447 | google.golang.org/genproto v0.0.0-20190425155659-357c62f0e4bb/go.mod h1:VzzqZJRnGkLBvHegQrXjBqPurQTc5/KpmUdxsrq26oE=
448 | google.golang.org/genproto v0.0.0-20190502173448-54afdca5d873/go.mod h1:VzzqZJRnGkLBvHegQrXjBqPurQTc5/KpmUdxsrq26oE=
449 | google.golang.org/genproto v0.0.0-20190801165951-fa694d86fc64/go.mod h1:DMBHOl98Agz4BDEuKkezgsaosCRResVns1a3J2ZsMNc=
450 | google.golang.org/genproto v0.0.0-20190819201941-24fa4b261c55/go.mod h1:DMBHOl98Agz4BDEuKkezgsaosCRResVns1a3J2ZsMNc=
451 | google.golang.org/genproto v0.0.0-20190911173649-1774047e7e51/go.mod h1:IbNlFCBrqXvoKpeg0TB2l7cyZUmoaFKYIwrEpbDKLA8=
452 | google.golang.org/genproto v0.0.0-20191108220845-16a3f7862a1a/go.mod h1:n3cpQtvxv34hfy77yVDNjmbRyujviMdxYliBSkLhpCc=
453 | google.golang.org/genproto v0.0.0-20191115194625-c23dd37a84c9/go.mod h1:n3cpQtvxv34hfy77yVDNjmbRyujviMdxYliBSkLhpCc=
454 | google.golang.org/genproto v0.0.0-20191216164720-4f79533eabd1/go.mod h1:n3cpQtvxv34hfy77yVDNjmbRyujviMdxYliBSkLhpCc=
455 | google.golang.org/genproto v0.0.0-20191230161307-f3c370f40bfb/go.mod h1:n3cpQtvxv34hfy77yVDNjmbRyujviMdxYliBSkLhpCc=
456 | google.golang.org/genproto v0.0.0-20200115191322-ca5a22157cba/go.mod h1:n3cpQtvxv34hfy77yVDNjmbRyujviMdxYliBSkLhpCc=
457 | google.golang.org/genproto v0.0.0-20200122232147-0452cf42e150/go.mod h1:n3cpQtvxv34hfy77yVDNjmbRyujviMdxYliBSkLhpCc=
458 | google.golang.org/genproto v0.0.0-20200204135345-fa8e72b47b90/go.mod h1:GmwEX6Z4W5gMy59cAlVYjN9JhxgbQH6Gn+gFDQe2lzA=
459 | google.golang.org/genproto v0.0.0-20200212174721-66ed5ce911ce/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
460 | google.golang.org/genproto v0.0.0-20200224152610-e50cd9704f63/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
461 | google.golang.org/genproto v0.0.0-20200228133532-8c2c7df3a383/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
462 | google.golang.org/genproto v0.0.0-20200305110556-506484158171/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
463 | google.golang.org/genproto v0.0.0-20200312145019-da6875a35672/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
464 | google.golang.org/genproto v0.0.0-20200331122359-1ee6d9798940/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
465 | google.golang.org/genproto v0.0.0-20200430143042-b979b6f78d84/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
466 | google.golang.org/genproto v0.0.0-20200511104702-f5ebc3bea380/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
467 | google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884/go.mod h1:55QSHmfGQM9UVYDPBsyGGes0y52j32PQ3BqQfXhyH3c=
468 | google.golang.org/genproto v0.0.0-20200515170657-fc4c6c6a6587/go.mod h1:YsZOwe1myG/8QRHRsmBRE1LrgQY60beZKjly0O1fX9U=
469 | google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013/go.mod h1:NbSheEEYHJ7i3ixzK3sjbqSGDJWnxyFXZblF3eUsNvo=
470 | google.golang.org/genproto v0.0.0-20200618031413-b414f8b61790/go.mod h1:jDfRM7FcilCzHH/e9qn6dsT145K34l5v+OpcnNgKAAA=
471 | google.golang.org/genproto v0.0.0-20200729003335-053ba62fc06f/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
472 | google.golang.org/genproto v0.0.0-20200804131852-c06518451d9c/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
473 | google.golang.org/genproto v0.0.0-20200825200019-8632dd797987/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
474 | google.golang.org/genproto v0.0.0-20200904004341-0bd0a958aa1d/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
475 | google.golang.org/genproto v0.0.0-20201109203340-2640f1f9cdfb/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
476 | google.golang.org/genproto v0.0.0-20201201144952-b05cb90ed32e/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
477 | google.golang.org/genproto v0.0.0-20201210142538-e3217bee35cc/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
478 | google.golang.org/genproto v0.0.0-20201214200347-8c77b98c765d/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
479 | google.golang.org/genproto v0.0.0-20210222152913-aa3ee6e6a81c/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
480 | google.golang.org/genproto v0.0.0-20210303154014-9728d6b83eeb/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
481 | google.golang.org/genproto v0.0.0-20210310155132-4ce2db91004e/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
482 | google.golang.org/genproto v0.0.0-20210319143718-93e7006c17a6/go.mod h1:FWY/as6DDZQgahTzZj3fqbO1CbirC29ZNUFHwi0/+no=
483 | google.golang.org/genproto v0.0.0-20210402141018-6c239bbf2bb1/go.mod h1:9lPAdzaEmUacj36I+k7YKbEc5CXzPIeORRgDAUOu28A=
484 | google.golang.org/genproto v0.0.0-20210513213006-bf773b8c8384/go.mod h1:P3QM42oQyzQSnHPnZ/vqoCdDmzH28fzWByN9asMeM8A=
485 | google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c/go.mod h1:UODoCrxHCcBojKKwX1terBiRUaqAsFqJiF615XL43r0=
486 | google.golang.org/genproto v0.0.0-20210604141403-392c879c8b08/go.mod h1:UODoCrxHCcBojKKwX1terBiRUaqAsFqJiF615XL43r0=
487 | google.golang.org/genproto v0.0.0-20210608205507-b6d2f5bf0d7d/go.mod h1:UODoCrxHCcBojKKwX1terBiRUaqAsFqJiF615XL43r0=
488 | google.golang.org/genproto v0.0.0-20210624195500-8bfb893ecb84/go.mod h1:SzzZ/N+nwJDaO1kznhnlzqS8ocJICar6hYhVyhi++24=
489 | google.golang.org/genproto v0.0.0-20210713002101-d411969a0d9a/go.mod h1:AxrInvYm1dci+enl5hChSFPOmmUF1+uAa/UsgNRWd7k=
490 | google.golang.org/genproto v0.0.0-20210716133855-ce7ef5c701ea/go.mod h1:AxrInvYm1dci+enl5hChSFPOmmUF1+uAa/UsgNRWd7k=
491 | google.golang.org/genproto v0.0.0-20210728212813-7823e685a01f/go.mod h1:ob2IJxKrgPT52GcgX759i1sleT07tiKowYBGbczaW48=
492 | google.golang.org/genproto v0.0.0-20210805201207-89edb61ffb67/go.mod h1:ob2IJxKrgPT52GcgX759i1sleT07tiKowYBGbczaW48=
493 | google.golang.org/genproto v0.0.0-20210813162853-db860fec028c/go.mod h1:cFeNkxwySK631ADgubI+/XFU/xp8FD5KIVV4rj8UC5w=
494 | google.golang.org/genproto v0.0.0-20210821163610-241b8fcbd6c8/go.mod h1:eFjDcFEctNawg4eG61bRv87N7iHBWyVhJu7u1kqDUXY=
495 | google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71/go.mod h1:eFjDcFEctNawg4eG61bRv87N7iHBWyVhJu7u1kqDUXY=
496 | google.golang.org/genproto v0.0.0-20210831024726-fe130286e0e2/go.mod h1:eFjDcFEctNawg4eG61bRv87N7iHBWyVhJu7u1kqDUXY=
497 | google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83/go.mod h1:eFjDcFEctNawg4eG61bRv87N7iHBWyVhJu7u1kqDUXY=
498 | google.golang.org/genproto v0.0.0-20210909211513-a8c4777a87af/go.mod h1:eFjDcFEctNawg4eG61bRv87N7iHBWyVhJu7u1kqDUXY=
499 | google.golang.org/genproto v0.0.0-20210924002016-3dee208752a0/go.mod h1:5CzLGKJ67TSI2B9POpiiyGha0AjJvZIUgRMt1dSmuhc=
500 | google.golang.org/genproto v0.0.0-20211118181313-81c1377c94b1/go.mod h1:5CzLGKJ67TSI2B9POpiiyGha0AjJvZIUgRMt1dSmuhc=
501 | google.golang.org/genproto v0.0.0-20211206160659-862468c7d6e0/go.mod h1:5CzLGKJ67TSI2B9POpiiyGha0AjJvZIUgRMt1dSmuhc=
502 | google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb h1:ZrsicilzPCS/Xr8qtBZZLpy4P9TYXAfl49ctG1/5tgw=
503 | google.golang.org/genproto v0.0.0-20211223182754-3ac035c7e7cb/go.mod h1:5CzLGKJ67TSI2B9POpiiyGha0AjJvZIUgRMt1dSmuhc=
504 | google.golang.org/grpc v1.19.0/go.mod h1:mqu4LbDTu4XGKhr4mRzUsmM4RtVoemTSY81AxZiDr8c=
505 | google.golang.org/grpc v1.20.1/go.mod h1:10oTOabMzJvdu6/UiuZezV6QK5dSlG84ov/aaiqXj38=
506 | google.golang.org/grpc v1.21.1/go.mod h1:oYelfM1adQP15Ek0mdvEgi9Df8B9CZIaU1084ijfRaM=
507 | google.golang.org/grpc v1.23.0/go.mod h1:Y5yQAOtifL1yxbo5wqy6BxZv8vAUGQwXBOALyacEbxg=
508 | google.golang.org/grpc v1.25.1/go.mod h1:c3i+UQWmh7LiEpx4sFZnkU36qjEYZ0imhYfXVyQciAY=
509 | google.golang.org/grpc v1.26.0/go.mod h1:qbnxyOmOxrQa7FizSgH+ReBfzJrCY1pSN7KXBS8abTk=
510 | google.golang.org/grpc v1.27.0/go.mod h1:qbnxyOmOxrQa7FizSgH+ReBfzJrCY1pSN7KXBS8abTk=
511 | google.golang.org/grpc v1.27.1/go.mod h1:qbnxyOmOxrQa7FizSgH+ReBfzJrCY1pSN7KXBS8abTk=
512 | google.golang.org/grpc v1.28.0/go.mod h1:rpkK4SK4GF4Ach/+MFLZUBavHOvF2JJB5uozKKal+60=
513 | google.golang.org/grpc v1.29.1/go.mod h1:itym6AZVZYACWQqET3MqgPpjcuV5QH3BxFS3IjizoKk=
514 | google.golang.org/grpc v1.30.0/go.mod h1:N36X2cJ7JwdamYAgDz+s+rVMFjt3numwzf/HckM8pak=
515 | google.golang.org/grpc v1.31.0/go.mod h1:N36X2cJ7JwdamYAgDz+s+rVMFjt3numwzf/HckM8pak=
516 | google.golang.org/grpc v1.31.1/go.mod h1:N36X2cJ7JwdamYAgDz+s+rVMFjt3numwzf/HckM8pak=
517 | google.golang.org/grpc v1.33.1/go.mod h1:fr5YgcSWrqhRRxogOsw7RzIpsmvOZ6IcH4kBYTpR3n0=
518 | google.golang.org/grpc v1.33.2/go.mod h1:JMHMWHQWaTccqQQlmk3MJZS+GWXOdAesneDmEnv2fbc=
519 | google.golang.org/grpc v1.34.0/go.mod h1:WotjhfgOW/POjDeRt8vscBtXq+2VjORFy659qA51WJ8=
520 | google.golang.org/grpc v1.35.0/go.mod h1:qjiiYl8FncCW8feJPdyg3v6XW24KsRHe+dy9BAGRRjU=
521 | google.golang.org/grpc v1.36.0/go.mod h1:qjiiYl8FncCW8feJPdyg3v6XW24KsRHe+dy9BAGRRjU=
522 | google.golang.org/grpc v1.36.1/go.mod h1:qjiiYl8FncCW8feJPdyg3v6XW24KsRHe+dy9BAGRRjU=
523 | google.golang.org/grpc v1.37.0/go.mod h1:NREThFqKR1f3iQ6oBuvc5LadQuXVGo9rkm5ZGrQdJfM=
524 | google.golang.org/grpc v1.37.1/go.mod h1:NREThFqKR1f3iQ6oBuvc5LadQuXVGo9rkm5ZGrQdJfM=
525 | google.golang.org/grpc v1.38.0/go.mod h1:NREThFqKR1f3iQ6oBuvc5LadQuXVGo9rkm5ZGrQdJfM=
526 | google.golang.org/grpc v1.39.0/go.mod h1:PImNr+rS9TWYb2O4/emRugxiyHZ5JyHW5F+RPnDzfrE=
527 | google.golang.org/grpc v1.39.1/go.mod h1:PImNr+rS9TWYb2O4/emRugxiyHZ5JyHW5F+RPnDzfrE=
528 | google.golang.org/grpc v1.40.0/go.mod h1:ogyxbiOoUXAkP+4+xa6PZSE9DZgIHtSpzjDTB9KAK34=
529 | google.golang.org/grpc v1.40.1 h1:pnP7OclFFFgFi4VHQDQDaoXUVauOFyktqTsqqgzFKbc=
530 | google.golang.org/grpc v1.40.1/go.mod h1:ogyxbiOoUXAkP+4+xa6PZSE9DZgIHtSpzjDTB9KAK34=
531 | google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0/go.mod h1:6Kw0yEErY5E/yWrBtf03jp27GLLJujG4z/JK95pnjjw=
532 | google.golang.org/protobuf v0.0.0-20200109180630-ec00e32a8dfd/go.mod h1:DFci5gLYBciE7Vtevhsrf46CRTquxDuWsQurQQe4oz8=
533 | google.golang.org/protobuf v0.0.0-20200221191635-4d8936d0db64/go.mod h1:kwYJMbMJ01Woi6D6+Kah6886xMZcty6N08ah7+eCXa0=
534 | google.golang.org/protobuf v0.0.0-20200228230310-ab0ca4ff8a60/go.mod h1:cfTl7dwQJ+fmap5saPgwCLgHXTUD7jkjRqWcaiX5VyM=
535 | google.golang.org/protobuf v1.20.1-0.20200309200217-e05f789c0967/go.mod h1:A+miEFZTKqfCUM6K7xSMQL9OKL/b6hQv+e19PK+JZNE=
536 | google.golang.org/protobuf v1.21.0/go.mod h1:47Nbq4nVaFHyn7ilMalzfO3qCViNmqZ2kzikPIcrTAo=
537 | google.golang.org/protobuf v1.22.0/go.mod h1:EGpADcykh3NcUnDUJcl1+ZksZNG86OlYog2l/sGQquU=
538 | google.golang.org/protobuf v1.23.0/go.mod h1:EGpADcykh3NcUnDUJcl1+ZksZNG86OlYog2l/sGQquU=
539 | google.golang.org/protobuf v1.23.1-0.20200526195155-81db48ad09cc/go.mod h1:EGpADcykh3NcUnDUJcl1+ZksZNG86OlYog2l/sGQquU=
540 | google.golang.org/protobuf v1.24.0/go.mod h1:r/3tXBNzIEhYS9I1OUVjXDlt8tc493IdKGjtUeSXeh4=
541 | google.golang.org/protobuf v1.25.0/go.mod h1:9JNX74DMeImyA3h4bdi1ymwjUzf21/xIlbajtzgsN7c=
542 | google.golang.org/protobuf v1.26.0-rc.1/go.mod h1:jlhhOSvTdKEhbULTjvd4ARK9grFBp09yW+WbY/TyQbw=
543 | google.golang.org/protobuf v1.26.0/go.mod h1:9q0QmTI4eRPtz6boOQmLYwt+qCgq0jsYwAQnmE0givc=
544 | google.golang.org/protobuf v1.27.1 h1:SnqbnDw1V7RiZcXPx5MEeqPv2s79L9i7BJUlG/+RurQ=
545 | google.golang.org/protobuf v1.27.1/go.mod h1:9q0QmTI4eRPtz6boOQmLYwt+qCgq0jsYwAQnmE0givc=
546 | gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
547 | gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
548 | gopkg.in/errgo.v2 v2.1.0/go.mod h1:hNsd1EY+bozCKY1Ytp96fpM3vjJbqLJn88ws8XvfDNI=
549 | gopkg.in/yaml.v2 v2.2.2/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
550 | gopkg.in/yaml.v2 v2.2.3/go.mod h1:hI93XBmqTisBFMUTm0b8Fm+jr3Dg1NNxqwp+5A1VGuI=
551 | gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
552 | honnef.co/go/tools v0.0.0-20190102054323-c2f93a96b099/go.mod h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4=
553 | honnef.co/go/tools v0.0.0-20190106161140-3f1c8253044a/go.mod h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4=
554 | honnef.co/go/tools v0.0.0-20190418001031-e561f6794a2a/go.mod h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4=
555 | honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc/go.mod h1:rf3lG4BRIbNafJWhAfAdb/ePZxsR/4RtNHQocxwk9r4=
556 | honnef.co/go/tools v0.0.1-2019.2.3/go.mod h1:a3bituU0lyd329TUQxRnasdCoJDkEUEAqEt0JzvZhAg=
557 | honnef.co/go/tools v0.0.1-2020.1.3/go.mod h1:X/FiERA/W4tHapMX5mGpAtMSVEeEUOyHaw9vFzvIQ3k=
558 | honnef.co/go/tools v0.0.1-2020.1.4/go.mod h1:X/FiERA/W4tHapMX5mGpAtMSVEeEUOyHaw9vFzvIQ3k=
559 | rsc.io/binaryregexp v0.2.0/go.mod h1:qTv7/COck+e2FymRvadv62gMdZztPaShugOCi3I+8D8=
560 | rsc.io/quote/v3 v3.1.0/go.mod h1:yEA65RcK8LyAZtP9Kv3t0HmxON59tX3rD+tICJqUlj0=
561 | rsc.io/sampler v1.3.0/go.mod h1:T1hPZKmBbMNahiBKFy5HrXp6adAjACjK9JXDnKaTXpA=
562 |

---

## /chapter12/section4/main.go:

1 | package main
2 |
3 | import (
4 | "database/sql"
5 | "fmt"
6 | "log"
7 | "net/http"
8 | "os"
9 |
10 | \_ "github.com/go-sql-driver/mysql"
11 | "github.com/yourname/reponame/api"
12 | )
13 |
14 | var (
15 | dbUser = os.Getenv("DB_USER")
16 | dbPassword = os.Getenv("DB_PASSWORD")
17 | dbDatabase = os.Getenv("DB_NAME")
18 | dbConn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser, dbPassword, dbDatabase)
19 | )
20 |
21 | func main() {
22 | db, err := sql.Open("mysql", dbConn)
23 | if err != nil {
24 | log.Println("fail to connect DB")
25 | return
26 | }
27 |
28 | r := api.NewRouter(db)
29 |
30 | log.Println("server start at port 8080")
31 | log.Fatal(http.ListenAndServe(":8080", r))
32 | }
33 |

---

## /chapter12/section4/models/models.go:

1 | package models
2 |
3 | import "time"
4 |
5 | type Comment struct {
6 | CommentID int `json:"comment_id"`
7 | ArticleID int `json:"article_id"`
8 | Message string `json:"message"`
9 | CreatedAt time.Time `json:"created_at"`
10 | }
11 |
12 | type Article struct {
13 | ID int `json:"article_id"`
14 | Title string `json:"title"`
15 | Contents string `json:"contents"`
16 | UserName string `json:"user_name"`
17 | NiceNum int `json:"nice"`
18 | CommentList []Comment `json:"comments"`
19 | CreatedAt time.Time `json:"created_at"`
20 | }
21 |

---

## /chapter12/section4/repositories/articles.go:

1 | package repositories
2 |
3 | import (
4 | "database/sql"
5 |
6 | "github.com/yourname/reponame/models"
7 | )
8 |
9 | const (
10 | articleNumPerPage = 5
11 | )
12 |
13 | // 新規投稿を DB に insert する関数
14 | func InsertArticle(db *sql.DB, article models.Article) (models.Article, error) {
15 | const sqlStr = `
 16 | 	insert into articles (title, contents, username, nice, created_at) values
 17 | 	(?, ?, ?, 0, now());
 18 | 	`
19 |
20 | var newArticle models.Article
21 | newArticle.Title, newArticle.Contents, newArticle.UserName = article.Title, article.Contents, article.UserName
22 |
23 | result, err := db.Exec(sqlStr, article.Title, article.Contents, article.UserName)
24 | if err != nil {
25 | return models.Article{}, err
26 | }
27 |
28 | id, \_ := result.LastInsertId()
29 |
30 | newArticle.ID = int(id)
31 |
32 | return newArticle, nil
33 | }
34 |
35 | // 投稿一覧を DB から取得する関数
36 | func SelectArticleList(db *sql.DB, page int) ([]models.Article, error) {
37 | const sqlStr = `
 38 | 		select article_id, title, contents, username, nice
 39 | 		from articles
 40 | 		limit ? offset ?;
 41 | 	`
42 |
43 | rows, err := db.Query(sqlStr, articleNumPerPage, ((page - 1) * articleNumPerPage))
44 | if err != nil {
45 | return nil, err
46 | }
47 | defer rows.Close()
48 |
49 | articleArray := make([]models.Article, 0)
50 | for rows.Next() {
51 | var article models.Article
52 | rows.Scan(&article.ID, &article.Title, &article.Contents, &article.UserName, &article.NiceNum)
53 |
54 | articleArray = append(articleArray, article)
55 | }
56 |
57 | return articleArray, nil
58 | }
59 |
60 | // 投稿 ID を指定して、記事データを取得する関数
61 | func SelectArticleDetail(db *sql.DB, articleID int) (models.Article, error) {
62 | const sqlStr = `
 63 | 		select *
 64 | 		from articles
 65 | 		where article_id = ?;
 66 | 	`
67 | row := db.QueryRow(sqlStr, articleID)
68 | if err := row.Err(); err != nil {
69 | return models.Article{}, err
70 | }
71 |
72 | var article models.Article
73 | var createdTime sql.NullTime
74 | err := row.Scan(&article.ID, &article.Title, &article.Contents, &article.UserName, &article.NiceNum, &createdTime)
75 | if err != nil {
76 | return models.Article{}, err
77 | }
78 |
79 | if createdTime.Valid {
80 | article.CreatedAt = createdTime.Time
81 | }
82 |
83 | return article, nil
84 | }
85 |
86 | // いいねの数を update する関数
87 | func UpdateNiceNum(db \*sql.DB, articleID int) error {
88 | tx, err := db.Begin()
89 | if err != nil {
90 | return err
91 | }
92 |
93 | const sqlGetNice = `
 94 | 		select nice
 95 | 		from articles
 96 | 		where article_id = ?;
 97 | 	`
98 | row := tx.QueryRow(sqlGetNice, articleID)
99 | if err := row.Err(); err != nil {
100 | tx.Rollback()
101 | return err
102 | }
103 |
104 | var nicenum int
105 | err = row.Scan(&nicenum)
106 | if err != nil {
107 | tx.Rollback()
108 | return err
109 | }
110 |
111 | const sqlUpdateNice = `update articles set nice = ? where article_id = ?`
112 | \_, err = tx.Exec(sqlUpdateNice, nicenum+1, articleID)
113 | if err != nil {
114 | tx.Rollback()
115 | return err
116 | }
117 |
118 | if err := tx.Commit(); err != nil {
119 | return err
120 | }
121 | return nil
122 | }
123 |

---

## /chapter12/section4/repositories/articles_test.go:

1 | package repositories*test
2 |
3 | import (
4 | "testing"
5 |
6 | "github.com/yourname/reponame/models"
7 | "github.com/yourname/reponame/repositories"
8 | "github.com/yourname/reponame/repositories/testdata"
9 |
10 | * "github.com/go-sql-driver/mysql"
11 | )
12 |
13 | // SelectArticleList 関数のテスト
14 | func TestSelectArticleList(t *testing.T) {
15 | expectedNum := len(testdata.ArticleTestData)
16 | got, err := repositories.SelectArticleList(testDB, 1)
17 | if err != nil {
18 | t.Fatal(err)
19 | }
20 |
21 | if num := len(got); num != expectedNum {
22 | t.Errorf("want %d but got %d articles\n", expectedNum, num)
23 | }
24 | }
25 |
26 | // SelectArticleDetail 関数のテスト
27 | func TestSelectArticleDetail(t *testing.T) {
28 | tests := []struct {
29 | testTitle string
30 | expected models.Article
31 | }{
32 | {
33 | testTitle: "subtest1",
34 | expected: testdata.ArticleTestData[0],
35 | }, {
36 | testTitle: "subtest2",
37 | expected: testdata.ArticleTestData[1],
38 | },
39 | }
40 |
41 | for _, test := range tests {
42 | t.Run(test.testTitle, func(t *testing.T) {
43 | got, err := repositories.SelectArticleDetail(testDB, test.expected.ID)
44 | if err != nil {
45 | t.Fatal(err)
46 | }
47 |
48 | if got.ID != test.expected.ID {
49 | t.Errorf("ID: get %d but want %d\n", got.ID, test.expected.ID)
50 | }
51 | if got.Title != test.expected.Title {
52 | t.Errorf("Title: get %s but want %s\n", got.Title, test.expected.Title)
53 | }
54 | if got.Contents != test.expected.Contents {
55 | t.Errorf("Content: get %s but want %s\n", got.Contents, test.expected.Contents)
56 | }
57 | if got.UserName != test.expected.UserName {
58 | t.Errorf("UserName: get %s but want %s\n", got.UserName, test.expected.UserName)
59 | }
60 | if got.NiceNum != test.expected.NiceNum {
61 | t.Errorf("NiceNum: get %d but want %d\n", got.NiceNum, test.expected.NiceNum)
62 | }
63 | })
64 | }
65 | }
66 |
67 | // InsertArticle 関数のテスト
68 | func TestInsertArticle(t *testing.T) {
69 | article := models.Article{
70 | Title: "insertTest",
71 | Contents: "testest",
72 | UserName: "saki",
73 | }
74 |
75 | expectedArticleNum := 3
76 | newArticle, err := repositories.InsertArticle(testDB, article)
77 | if err != nil {
78 | t.Error(err)
79 | }
80 | if newArticle.ID != expectedArticleNum {
81 | t.Errorf("new article id is expected %d but got %d\n", expectedArticleNum, newArticle.ID)
82 | }
83 |
84 | t.Cleanup(func() {
85 | const sqlStr = `
 86 | 			delete from articles
 87 | 			where title = ? and contents = ? and username = ?
 88 | 		`
89 | testDB.Exec(sqlStr, article.Title, article.Contents, article.UserName)
90 | })
91 | }
92 |
93 | // UpdateNiceNum 関数のテスト
94 | func TestUpdateNiceNum(t \*testing.T) {
95 | articleID := 1
96 | err := repositories.UpdateNiceNum(testDB, articleID)
97 | if err != nil {
98 | t.Fatal(err)
99 | }
100 |
101 | got, _ := repositories.SelectArticleDetail(testDB, articleID)
102 |
103 | if got.NiceNum-testdata.ArticleTestData[articleID-1].NiceNum != 1 {
104 | t.Errorf("fail to update nice num: expected %d but got %d\n",
105 | testdata.ArticleTestData[articleID].NiceNum,
106 | got.NiceNum)
107 | }
108 | }
109 |

---

## /chapter12/section4/repositories/comment_test.go:

1 | package repositories*test
2 |
3 | import (
4 | "testing"
5 |
6 | "github.com/yourname/reponame/models"
7 | "github.com/yourname/reponame/repositories"
8 | )
9 |
10 | // SelectCommentList 関数のテスト
11 | func TestSelectCommentList(t \*testing.T) {
12 | articleID := 1
13 | got, err := repositories.SelectCommentList(testDB, articleID)
14 | if err != nil {
15 | t.Fatal(err)
16 | }
17 |
18 | for *, comment := range got {
19 | if comment.ArticleID != articleID {
20 | t.Errorf("want comment of articleID %d but got ID %d\n", articleID, comment.ArticleID)
21 | }
22 | }
23 | }
24 |
25 | // InsertComment 関数のテスト
26 | func TestInsertComment(t \*testing.T) {
27 | comment := models.Comment{
28 | ArticleID: 1,
29 | Message: "CommentInsertTest",
30 | }
31 |
32 | expectedCommentID := 3
33 | newComment, err := repositories.InsertComment(testDB, comment)
34 | if err != nil {
35 | t.Error(err)
36 | }
37 | if newComment.CommentID != expectedCommentID {
38 | t.Errorf("new comment id is expected %d but got %d\n", expectedCommentID, newComment.CommentID)
39 | }
40 |
41 | t.Cleanup(func() {
42 | const sqlStr = `
43 | 			delete from comments
44 | 			where message = ?
45 | 		`
46 | testDB.Exec(sqlStr, comment.Message)
47 | })
48 | }
49 |

---

## /chapter12/section4/repositories/comments.go:

1 | package repositories
2 |
3 | import (
4 | "database/sql"
5 |
6 | "github.com/yourname/reponame/models"
7 | )
8 |
9 | // 新規投稿を DB に insert する関数
10 | func InsertComment(db *sql.DB, comment models.Comment) (models.Comment, error) {
11 | const sqlStr = `
12 | 		insert into comments (article_id, message, created_at) values
13 | 		(?, ?, now());
14 | 	`
15 | var newComment models.Comment
16 | newComment.ArticleID, newComment.Message = comment.ArticleID, comment.Message
17 |
18 | result, err := db.Exec(sqlStr, comment.ArticleID, comment.Message)
19 | if err != nil {
20 | return models.Comment{}, err
21 | }
22 |
23 | id, \_ := result.LastInsertId()
24 | newComment.CommentID = int(id)
25 |
26 | return newComment, nil
27 | }
28 |
29 | // 指定 ID の記事についたコメント一覧を取得する関数
30 | func SelectCommentList(db *sql.DB, articleID int) ([]models.Comment, error) {
31 | const sqlStr = `
32 | 		select *
33 | 		from comments
34 | 		where article_id = ?;
35 | 	`
36 |
37 | rows, err := db.Query(sqlStr, articleID)
38 | if err != nil {
39 | return nil, err
40 | }
41 | defer rows.Close()
42 |
43 | commentArray := make([]models.Comment, 0)
44 | for rows.Next() {
45 | var comment models.Comment
46 | var createdTime sql.NullTime
47 | rows.Scan(&comment.CommentID, &comment.ArticleID, &comment.Message, &createdTime)
48 |
49 | if createdTime.Valid {
50 | comment.CreatedAt = createdTime.Time
51 | }
52 |
53 | commentArray = append(commentArray, comment)
54 | }
55 |
56 | return commentArray, nil
57 | }
58 |

---

## /chapter12/section4/repositories/main_test.go:

1 | package repositories_test
2 |
3 | import (
4 | "database/sql"
5 | "fmt"
6 | "os"
7 | "os/exec"
8 | "testing"
9 | )
10 |
11 | var testDB *sql.DB
12 |
13 | var (
14 | dbUser = "docker"
15 | dbPassword = "docker"
16 | dbDatabase = "sampledb"
17 | dbConn = fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser, dbPassword, dbDatabase)
18 | )
19 |
20 | func connectDB() error {
21 | var err error
22 | testDB, err = sql.Open("mysql", dbConn)
23 | if err != nil {
24 | return err
25 | }
26 | return nil
27 | }
28 |
29 | func setupTestData() error {
30 | cmd := exec.Command("mysql", "-h", "127.0.0.1", "-u", "docker", "sampledb", "--password=docker", "-e", "source ./testdata/setupDB.sql")
31 | err := cmd.Run()
32 | if err != nil {
33 | return err
34 | }
35 | return nil
36 | }
37 |
38 | func cleanupDB() error {
39 | cmd := exec.Command("mysql", "-h", "127.0.0.1", "-u", "docker", "sampledb", "--password=docker", "-e", "source ./testdata/cleanupDB.sql")
40 | err := cmd.Run()
41 | if err != nil {
42 | return err
43 | }
44 | return nil
45 | }
46 |
47 | // 全テスト共通の前処理を書く
48 | func setup() error {
49 | if err := connectDB(); err != nil {
50 | return err
51 | }
52 | if err := cleanupDB(); err != nil {
53 | fmt.Println("cleanup", err)
54 | return err
55 | }
56 | if err := setupTestData(); err != nil {
57 | fmt.Println("setup")
58 | return err
59 | }
60 | return nil
61 | }
62 |
63 | // 前テスト共通の後処理を書く
64 | func teardown() {
65 | cleanupDB()
66 | testDB.Close()
67 | }
68 |
69 | func TestMain(m *testing.M) {
70 | err := setup()
71 | if err != nil {
72 | os.Exit(1)
73 | }
74 |
75 | m.Run()
76 |
77 | teardown()
78 | }
79 |

---

## /chapter12/section4/repositories/testdata/cleanupDB.sql:

1 | drop table if exists comments;
2 |
3 | drop table if exists articles;
4 |

---

## /chapter12/section4/repositories/testdata/data.go:

1 | package testdata
2 |
3 | import "github.com/yourname/reponame/models"
4 |
5 | var ArticleTestData = []models.Article{
6 | models.Article{
7 | ID: 1,
8 | Title: "firstPost",
9 | Contents: "This is my first blog",
10 | UserName: "saki",
11 | NiceNum: 2,
12 | },
13 | models.Article{
14 | ID: 2,
15 | Title: "2nd",
16 | Contents: "Second blog post",
17 | UserName: "saki",
18 | NiceNum: 4,
19 | },
20 | }
21 |
22 | var CommentTestData = []models.Comment{
23 | models.Comment{
24 | CommentID: 1,
25 | ArticleID: 1,
26 | Message: "1st comment yeah",
27 | },
28 | models.Comment{
29 | CommentID: 2,
30 | ArticleID: 1,
31 | Message: "welcome",
32 | },
33 | }
34 |

---

## /chapter12/section4/repositories/testdata/setupDB.sql:

1 | create table if not exists articles (
2 | article_id integer unsigned auto_increment primary key,
3 | title varchar(100) not null,
4 | contents text not null,
5 | username varchar(100) not null,
6 | nice integer not null,
7 | created_at datetime
8 | );
9 |
10 | create table if not exists comments (
11 | comment_id integer unsigned auto_increment primary key,
12 | article_id integer unsigned not null,
13 | message text not null,
14 | created_at datetime,
15 | foreign key (article_id) references articles(article_id)
16 | );
17 |
18 | insert into articles (title, contents, username, nice, created_at) values
19 | ('firstPost', 'This is my first blog', 'saki', 2, now());
20 |
21 | insert into articles (title, contents, username, nice) values
22 | ('2nd', 'Second blog post', 'saki', 4);
23 |
24 | insert into comments (article_id, message, created_at) values
25 | (1, '1st comment yeah', now());
26 |
27 | insert into comments (article_id, message) values
28 | (1, 'welcome');
29 |

---

## /chapter12/section4/services/article_service.go:

1 | package services
2 |
3 | import (
4 | "database/sql"
5 | "errors"
6 | "sync"
7 |
8 | "github.com/yourname/reponame/apperrors"
9 | "github.com/yourname/reponame/models"
10 | "github.com/yourname/reponame/repositories"
11 | )
12 |
13 | // PostArticleHandler で使うことを想定したサービス
14 | // 引数の情報をもとに新しい記事を作り、結果を返却
15 | func (s *MyAppService) PostArticleService(article models.Article) (models.Article, error) {
16 | newArticle, err := repositories.InsertArticle(s.db, article)
17 | if err != nil {
18 | err = apperrors.InsertDataFailed.Wrap(err, "fail to record data")
19 | return models.Article{}, err
20 | }
21 | return newArticle, nil
22 | }
23 |
24 | // ArticleListHandler で使うことを想定したサービス
25 | // 指定 page の記事一覧を返却
26 | func (s *MyAppService) GetArticleListService(page int) ([]models.Article, error) {
27 | articleList, err := repositories.SelectArticleList(s.db, page)
28 | if err != nil {
29 | err = apperrors.GetDataFailed.Wrap(err, "fail to get data")
30 | return nil, err
31 | }
32 |
33 | if len(articleList) == 0 {
34 | err := apperrors.NAData.Wrap(ErrNoData, "no data")
35 | return nil, err
36 | }
37 |
38 | return articleList, nil
39 | }
40 |
41 | // ArticleDetailHandler で使うことを想定したサービス
42 | // 指定 ID の記事情報を返却
43 | func (s *MyAppService) GetArticleService(articleID int) (models.Article, error) {
44 | var article models.Article
45 | var commentList []models.Comment
46 | var articleGetErr, commentGetErr error
47 |
48 | var wg sync.WaitGroup
49 | wg.Add(2)
50 |
51 | var amu sync.Mutex
52 | var cmu sync.Mutex
53 |
54 | go func(db *sql.DB, articleID int) {
55 | defer wg.Done()
56 | newarticle, err := repositories.SelectArticleDetail(db, articleID)
57 | amu.Lock()
58 | article, articleGetErr = newarticle, err
59 | amu.Unlock()
60 | }(s.db, articleID)
61 |
62 | go func(db *sql.DB, articleID int) {
63 | defer wg.Done()
64 | newcommentList, err := repositories.SelectCommentList(db, articleID)
65 | cmu.Lock()
66 | commentList, commentGetErr = newcommentList, err
67 | cmu.Unlock()
68 | }(s.db, articleID)
69 |
70 | wg.Wait()
71 |
72 | if articleGetErr != nil {
73 | if errors.Is(articleGetErr, sql.ErrNoRows) {
74 | err := apperrors.NAData.Wrap(articleGetErr, "no data")
75 | return models.Article{}, err
76 | }
77 | err := apperrors.GetDataFailed.Wrap(articleGetErr, "fail to get data")
78 | return models.Article{}, err
79 | }
80 |
81 | if commentGetErr != nil {
82 | err := apperrors.GetDataFailed.Wrap(commentGetErr, "fail to get data")
83 | return models.Article{}, err
84 | }
85 |
86 | article.CommentList = append(article.CommentList, commentList...)
87 |
88 | return article, nil
89 | }
90 |
91 | // PostNiceHandler で使うことを想定したサービス
92 | // 指定 ID の記事のいいね数を+1 して、結果を返却
93 | func (s *MyAppService) PostNiceService(article models.Article) (models.Article, error) {
94 | err := repositories.UpdateNiceNum(s.db, article.ID)
95 | if err != nil {
96 | if errors.Is(err, sql.ErrNoRows) {
97 | err = apperrors.NoTargetData.Wrap(err, "does not exist target article")
98 | return models.Article{}, err
99 | }
100 | err = apperrors.UpdateDataFailed.Wrap(err, "fail to update nice count")
101 | return models.Article{}, err
102 | }
103 |
104 | return models.Article{
105 | ID: article.ID,
106 | Title: article.Title,
107 | Contents: article.Contents,
108 | UserName: article.UserName,
109 | NiceNum: article.NiceNum + 1,
110 | CreatedAt: article.CreatedAt,
111 | }, nil
112 | }
113 |

---

## /chapter12/section4/services/article_service_test.go:

1 | package services*test
2 |
3 | import (
4 | "database/sql"
5 | "fmt"
6 | "os"
7 | "testing"
8 |
9 | "github.com/yourname/reponame/services"
10 |
11 | * "github.com/go-sql-driver/mysql"
12 | )
13 |
14 | var aSer *services.MyAppService
15 |
16 | func TestMain(m *testing.M) {
17 | dbUser := "docker"
18 | dbPassword := "docker"
19 | dbDatabase := "sampledb"
20 | dbConn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s?parseTime=true", dbUser, dbPassword, dbDatabase)
21 |
22 | db, err := sql.Open("mysql", dbConn)
23 | if err != nil {
24 | fmt.Println(err)
25 | os.Exit(1)
26 | }
27 |
28 | aSer = services.NewMyAppService(db)
29 |
30 | m.Run()
31 | }
32 |
33 | func BenchmarkGetArticleService(b \*testing.B) {
34 | articleID := 1
35 |
36 | b.ResetTimer()
37 | for i := 0; i < b.N; i++ {
38 | \_, err := aSer.GetArticleService(articleID)
39 | if err != nil {
40 | b.Error(err)
41 | break
42 | }
43 | }
44 | }
45 |

---

## /chapter12/section4/services/comment_service.go:

1 | package services
2 |
3 | import (
4 | "github.com/yourname/reponame/apperrors"
5 | "github.com/yourname/reponame/models"
6 | "github.com/yourname/reponame/repositories"
7 | )
8 |
9 | // PostCommentHandler で使用することを想定したサービス
10 | // 引数の情報をもとに新しいコメントを作り、結果を返却
11 | func (s \*MyAppService) PostCommentService(comment models.Comment) (models.Comment, error) {
12 | newComment, err := repositories.InsertComment(s.db, comment)
13 | if err != nil {
14 | err = apperrors.InsertDataFailed.Wrap(err, "fail to record data")
15 | return models.Comment{}, err
16 | }
17 |
18 | return newComment, nil
19 | }
20 |

---

## /chapter12/section4/services/errors.go:

1 | package services
2 |
3 | import "errors"
4 |
5 | var ErrNoData = errors.New("get 0 record from db.Query")
6 |

---

## /chapter12/section4/services/service.go:

1 | package services
2 |
3 | import "database/sql"
4 |
5 | type MyAppService struct {
6 | db *sql.DB
7 | }
8 |
9 | func NewMyAppService(db *sql.DB) \*MyAppService {
10 | return &MyAppService{db: db}
11 | }
12 |

---
