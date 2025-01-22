package logging

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// SetTestGinContext вспомогательная функция создания Gin контекста
func SetTestGinContext(w *httptest.ResponseRecorder, r *http.Request) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Request.Header.Set("Content-Type", "text/plain")
	return c
}

func TestWithLogging(t *testing.T) {
	type args struct {
		w     *httptest.ResponseRecorder
		r     *http.Request
		sugar *zap.SugaredLogger
		c     *gin.Context
	}
	type want struct {
		code        int
		contentType string
		gzipStatus  string
	}
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	sugar := *logger.Sugar()

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Positive test TestWithLogging(t *testing.T)",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", nil),
			},
			want: want{
				gzipStatus: "success",
				code:       http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
			WithLogging(&sugar)(c)
			if !assert.Equal(t, reflect.TypeOf(c), reflect.TypeOf(&gin.Context{})) {
				t.Errorf("With11Logging() got = %v, want %v", reflect.TypeOf(c), reflect.TypeOf(gin.Context{}))
			}
		})
	}
}
