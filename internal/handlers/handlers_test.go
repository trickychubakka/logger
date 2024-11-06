package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	//"logger/cmd/server/initconf"
	"logger/internal/storage/memstorage"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var store = memstorage.New()

// SetTestGinContext вспомогательная функция создания Gin контекста
func SetTestGinContext(w *httptest.ResponseRecorder, r *http.Request) (*gin.Context, error) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Request.Header.Set("Content-Type", "text/plain")
	return c, nil
}

func TestMetricHandler(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Positive request test",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1", nil),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name: "Negative request test, error in urlToMap()",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1/fault/fault", nil),
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "Negative test request gauge metric update, wrong metric value type",
			args: args{
				// c: *gin.Context,
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1ewe", nil),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Negative test request counter metric update, wrong metric value type",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/counter/metric1/1ewe", nil),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Negative test request counter metric update, wrong metric type",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/bool/metric1/1", nil),
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := SetTestGinContext(tt.args.w, tt.args.r)
			if err != nil {
				t.Fatal(err)
			}
			//MetricsHandler(c)
			MetricsHandler(&store)(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			// получаем и проверяем тело запроса
			//defer res.Body.Close()
			defer tt.args.w.Result().Body.Close()
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			//require.NoError(t, err)
		})
	}
}

func Test_urlToMap(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name:    "Positive URL test",
			args:    args{url: "/update/counter/metric1/1"},
			want:    []string{"update", "counter", "metric1", "1"},
			wantErr: false,
		},
		{
			name:    "Short URL negative test",
			args:    args{url: "/update/counter/metr"},
			want:    []string{"update", "counter", "metr"},
			wantErr: true,
		},
		{
			name:    "Long URL negative test",
			args:    args{url: "/update/counter/metric1/1/err/err"},
			want:    []string{"update", "counter", "metric1", "1", "err", "err"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := urlToMap(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("urlToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("urlToMap() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
		//metricName string
	}
	type want struct {
		code        int
		contentType string
	}

	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge("metric1", 7.77); err != nil {
		t.Error(err)
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Positive test get gauge metric",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/value/gauge/metric1", nil),
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "Negative test get gauge metric",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/value/gauge/metric111", nil),
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := SetTestGinContext(tt.args.w, tt.args.r)
			if err != nil {
				t.Fatal(err)
			}
			GetMetric(&store)(c)
			//GetMetric(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			// получаем и проверяем тело запроса
			//defer res.Body.Close()
			defer tt.args.w.Result().Body.Close()
			//assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			//require.NoError(t, err)

		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type want struct {
		code        int
		contentType string
	}

	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge("metric1", 7.77); err != nil {
		//t.Fatal(err)
		t.Error(err)
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Positive test get all metrics",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodGet, "/", nil),
			},
			want: want{
				code: http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := SetTestGinContext(tt.args.w, tt.args.r)
			if err != nil {
				t.Fatal(err)
			}
			GetAllMetrics(&store)(c)
			//GetAllMetrics(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			// получаем и проверяем тело запроса
			defer tt.args.w.Result().Body.Close()
			//assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			//require.NoError(t, err)
		})
	}
}
