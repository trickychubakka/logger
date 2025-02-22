package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"logger/cmd/server/initconf"
	"logger/internal/storage"
	"logger/internal/storage/memstorage"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

//var store, err = memstorage.New(ctx)

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
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var store, _ = memstorage.New(ctx)
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
			MetricsHandler(ctx, &store)(c)
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
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var store, _ = memstorage.New(ctx)

	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge(ctx, "metric1", 7.77); err != nil {
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
			GetMetric(ctx, &store)(c)
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
	var ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	var store, _ = memstorage.New(ctx)
	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge(ctx, "metric1", 7.77); err != nil {
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
			GetAllMetrics(ctx, &store)(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			// получаем и проверяем тело запроса
			defer tt.args.w.Result().Body.Close()
			//assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
			//require.NoError(t, err)
		})
	}
}

func Test_hashBody(t *testing.T) {
	body := []byte("Test body")
	var c *gin.Context
	config := initconf.Config{
		Key: "superkey",
	}
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	httpArg := args{
		w: httptest.NewRecorder(),
		r: httptest.NewRequest(http.MethodGet, "/value/gauge/metric1", nil),
	}
	c, err := SetTestGinContext(httpArg.w, httpArg.r)
	if err != nil {
		t.Fatal(err)
	}
	if err := hashBody(body, &config, c); err != nil {
		t.Errorf("hashBody() error = %v", err)
	}
}

func Test_MetricsToMemstorage(t *testing.T) {
	var delta int64 = 1
	var value = 1.1

	type args struct {
		metrics []storage.Metrics
	}

	type want struct {
		stor memstorage.MemStorage
	}

	a := args{[]storage.Metrics{
		{ID: "counter1", MType: "counter", Delta: &delta},
		{ID: "gauge1", MType: "gauge", Value: &value},
	},
	}
	var w want
	w.stor, _ = memstorage.New(context.Background())
	w.stor.UpdateCounter(context.Background(), "counter1", 1)
	w.stor.UpdateGauge(context.Background(), "gauge1", 1.1)

	stor, err := MetricsToMemstorage(context.Background(), a.metrics)
	if err != nil {
		t.Errorf("MetricsToMemstorage() error = %v", err)
	}
	assert.Equal(t, w.stor, stor)
}
