package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
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
func SetTestGinContext(w *httptest.ResponseRecorder, r *http.Request) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Request.Header.Set("Content-Type", "text/plain")
	return c
}

func createTestStor(ctx context.Context) memstorage.MemStorage {
	store, err := memstorage.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	store.UpdateCounter(ctx, "Counter1", 1)
	store.UpdateCounter(ctx, "Counter2", 2)
	store.UpdateCounter(ctx, "Counter3", 3)
	store.UpdateGauge(ctx, "Gauge1", 1.1)
	store.UpdateGauge(ctx, "Gauge2", 2.2)
	store.UpdateGauge(ctx, "Gauge3", 3.3)
	return store
}

func createMetricsArray() []storage.Metrics {
	var c int64 = 100
	var g = 100.1
	tmpMetrics := []storage.Metrics{
		{ID: "counter100", MType: "counter", Delta: &c},
		{ID: "gauge1", MType: "gauge", Value: &g},
	}
	return tmpMetrics
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
			c := SetTestGinContext(tt.args.w, tt.args.r)
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
			c := SetTestGinContext(tt.args.w, tt.args.r)
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
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
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
	c = SetTestGinContext(httpArg.w, httpArg.r)
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

func TestMetricHandlerBatchUpdate(t *testing.T) {
	type args struct {
		ctx   context.Context
		store Storager
		conf  *initconf.Config
		w     *httptest.ResponseRecorder
		r     *http.Request
	}
	type want struct {
		code        int
		contentType string
	}
	tmpMetrics := createMetricsArray()
	body, _ := json.Marshal(tmpMetrics)
	tests := []struct {
		name       string
		args       args
		want       want
		updateJSON string
	}{
		{
			name: "Positive test MetricHandlerBatchUpdate",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body)),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "StatusBadRequest test MetricHandlerBatchUpdate ",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte("wrong"))),
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
			MetricHandlerBatchUpdate(tt.args.ctx, tt.args.store, tt.args.conf)(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
		})
	}
}

func TestMetricHandlerJSON(t *testing.T) {
	type args struct {
		ctx   context.Context
		store Storager
		conf  *initconf.Config
		w     *httptest.ResponseRecorder
		r     *http.Request
	}
	type want struct {
		code        int
		contentType string
	}
	body := []byte(`{"id": "Counter100", "type": "counter", "delta": 100}`)
	bodyStatusBadRequest := []byte(`{"id": "Counter100", "type": "wrongType", "delta": 100}`)
	tests := []struct {
		name       string
		args       args
		want       want
		updateJSON string
	}{
		{
			name: "Positive test MetricHandlerJSON",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body)),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "StatusBadRequest test MetricHandlerJSON ",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyStatusBadRequest)),
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
			MetricHandlerJSON(tt.args.ctx, tt.args.store, tt.args.conf)(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
		})
	}
}

func TestDBPing(t *testing.T) {
	type args struct {
		connStr string
	}
	tests := []struct {
		name string
		args args
		want gin.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, DBPing(tt.args.connStr), "DBPing(%v)", tt.args.connStr)
		})
	}
}

func TestGetMetricJSON(t *testing.T) {
	type args struct {
		ctx   context.Context
		store Storager
		conf  *initconf.Config
		w     *httptest.ResponseRecorder
		r     *http.Request
	}
	type want struct {
		code        int
		contentType string
	}
	body := []byte(`{"id": "Counter2", "type": "counter"}`)
	bodyStatusNotFoundCounter := []byte(`{"id": "Counter200", "type": "counter"}`)
	bodyStatusNotFoundGauge := []byte(`{"id": "Gauge200", "type": "gauge"}`)
	tests := []struct {
		name       string
		args       args
		want       want
		updateJSON string
	}{
		{
			name: "Positive test GetMetricJSON",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body)),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
			},
		},
		{
			name: "StatusNotFound counter test GetMetricJSON",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyStatusNotFoundCounter)),
			},
			want: want{
				code:        http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name: "StatusNotFound gauge test GetMetricJSON",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(bodyStatusNotFoundGauge)),
			},
			want: want{
				code:        http.StatusNotFound,
				contentType: "application/json",
			},
		},
		{
			name: "StatusBadRequest test GetMetricJSON ",
			args: args{
				ctx:   context.Background(),
				store: createTestStor(context.Background()),
				conf: &initconf.Config{
					FileStoragePath: "",
					Key:             "superkey",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte("wrong"))),
			},
			want: want{
				code:        http.StatusBadRequest,
				contentType: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
			GetMetricJSON(tt.args.ctx, tt.args.store, tt.args.conf)(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
		})
	}
}
