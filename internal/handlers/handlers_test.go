package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/storage"
	"logger/internal/storage/memstorage"
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
	log.Println(store)
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

func TestMetricsHandler(t *testing.T) {
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
			MetricsHandler(ctx, &store)(c)
			res := c.Writer
			assert.Equal(t, tt.want.code, res.Status())
			// получаем и проверяем тело запроса
			defer tt.args.w.Result().Body.Close()
			assert.Equal(t, tt.want.contentType, res.Header().Get("Content-Type"))
		})
	}
}

func ExampleMetricsHandler() {

	ctx := context.Background()
	var store, _ = memstorage.New(ctx)

	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge(ctx, "metric1", 7.77); err != nil {
		log.Println("ExampleGetMetric UpdateGauge error", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1", nil)

	c := SetTestGinContext(w, r)
	GetMetric(ctx, &store)(c)
	res := c.Writer
	fmt.Println(res.Status())

	// Output:
	// 200
}

func ExampleMetricsHandler_wrongMetricType() {

	ctx := context.Background()
	var store, _ = memstorage.New(ctx)

	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge(ctx, "metric1", 7.77); err != nil {
		log.Println("ExampleGetMetric UpdateGauge error", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/update/bool/metric1/1", nil)

	c := SetTestGinContext(w, r)
	GetMetric(ctx, &store)(c)
	res := c.Writer
	fmt.Println(res.Status())

	// Output:
	// Error in GetMetric: wrong metric type
	// 404
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

func ExampleGetMetric() {

	ctx := context.Background()
	var store, _ = memstorage.New(ctx)

	// Создадим в Store метрику metric1 со значением 7.77
	if err := store.UpdateGauge(ctx, "metric1", 7.77); err != nil {
		log.Println("ExampleGetMetric UpdateGauge error", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/value/gauge/metric1", nil)

	c := SetTestGinContext(w, r)
	GetMetric(ctx, &store)(c)
	resp := w.Result()
	defer resp.Body.Close()
	res := c.Writer
	// Read and print response.

	jsn, _ := io.ReadAll(resp.Body)

	//if err != nil {
	//	log.Println("io.ReadAll error:", err)
	//}
	fmt.Println(res.Status())
	fmt.Println(string(jsn))

	// Output:
	// 200
	// 7.77
}

//func TestFunc(t *testing.T) {
//	ctx := context.Background()
//	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
//	store := createTestStor(ctx)
//	// Создадим в Store метрику metric1 со значением 7.77
//	if err := store.UpdateGauge(ctx, "metric1", 7.77); err != nil {
//		log.Println("ExampleGetMetric UpdateGauge error", err)
//	}
//	w := httptest.NewRecorder()
//	//defer w.Result().Body.Close()
//	c, engine := gin.CreateTestContext(w)
//
//	//req, _ := http.NewRequest(http.MethodGet, "/value/gauge/metric1", nil)
//	req, _ := http.NewRequest(http.MethodGet, "/", nil)
//
//	// What do I do with `ctx`? Is there a way to inject this into my test?
//	engine.GET("/", GetAllMetrics(ctx, store))
//	engine.GET("/value/:metricType/:metricName", GetMetric(ctx, store))
//	engine.ServeHTTP(w, req)
//
//	w.Result().Body.Close()
//	log.Println("c.Status is", c.Writer.Status())
//	jsn, _ := io.ReadAll(w.Result().Body)
//	fmt.Println(string(jsn))
//	log.Println("Body Closed?", w.Result().Close)
//	//log.Println(w.Result().Body)
//	//assert.Equal(t, 302, w.Result().StatusCode)
//	//assert.Equal(t, "/login", w.Result().Header.Get(HeaderLocation))
//}

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

type tmpMemStorage struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

func ExampleGetAllMetrics() {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	w := httptest.NewRecorder()
	//c, engine := gin.CreateTestContext(w)
	_, engine := gin.CreateTestContext(w)

	req, _ := http.NewRequest(http.MethodGet, "/", nil)

	engine.GET("/", GetAllMetrics(context.Background(), store))
	engine.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	var memStore tmpMemStorage
	err := json.NewDecoder(resp.Body).Decode(&memStore)
	if err != nil {
		log.Println("ExampleGetAllMetrics: json.NewDecoder error:", err)
	}
	fmt.Println(w.Result().Status)
	fmt.Println(memStore)

	// Output:
	// 200 OK
	// {map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]}
}

func ExampleGetAllMetrics_oneMore() {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	r := gin.Default()
	r.GET("/", GetAllMetrics(context.Background(), store))
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		//t.Fatalf("Couldn't create request: %v\n", err)
		log.Println("ExampleGetAllMetrics: http.NewRequest error:", err)
	}
	w := httptest.NewRecorder()
	//Client := &http.Client{}
	r.ServeHTTP(w, req)
	resp := w.Result()
	defer resp.Body.Close()
	w.Result().Body.Close()
	var memStore tmpMemStorage
	err = json.NewDecoder(resp.Body).Decode(&memStore)
	//io.Copy(io.Discard, w.Result().Body) // for close the body?
	if err != nil {
		log.Println("ExampleGetAllMetrics: json.NewDecoder error:", err)
	}
	log.Println("w.Result().Close", w.Result().Close)
	fmt.Println(w.Result().Status)
	fmt.Println(memStore)

	// Output:
	// 200 OK
	// {map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]}
}

func ExampleGetAllMetrics_second() {
	ctx := context.Background()
	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	s, _ := store.GetAllMetrics(ctx)
	log.Println("1", s)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	c := SetTestGinContext(w, r)
	GetAllMetrics(ctx, store)(c)
	res := c.Writer
	resp := w.Result()
	defer resp.Body.Close()
	// Read and print response.
	jsn, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("io.ReadAll error:", err)
	}
	var memStore memstorage.MemStorage
	memstorage.Unmarshal(jsn, &memStore)

	fmt.Println(res.Status())
	fmt.Println(memStore)

	// Output:
	// 200
	// {map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]}
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

func ExampleMetricHandlerBatchUpdate() {

	// Try to update logger server with next batched metrics:
	//	{ID: "counter100", MType: "counter", Delta: 100},
	//	{ID: "gauge1", MType: "gauge", Value: 100.1},
	metrics := createMetricsArray()
	body, _ := json.Marshal(metrics)
	ctx := context.Background()

	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	conf := &initconf.Config{
		Key: "superkey",
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	c := SetTestGinContext(w, r)

	MetricHandlerBatchUpdate(ctx, store, conf)(c)
	// read result from Gin context:
	res := c.Writer
	fmt.Println(res.Status())
	fmt.Println(res.Header().Get("Content-Type"))

	// Output:
	// 200
	// application/json
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

func ExampleMetricHandlerJSON() {
	// Try to update logger server with next metric: {"id": "Counter100", "type": "counter", "delta": 100}
	metric := `{"id": "Counter100", "type": "counter", "delta": 100}`
	ctx := context.Background()
	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	conf := &initconf.Config{
		Key: "superkey",
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte(metric)))
	c := SetTestGinContext(w, r)

	MetricHandlerJSON(ctx, store, conf)(c)
	// read result from Gin context:
	res := c.Writer
	fmt.Println(res.Status())
	fmt.Println(res.Header().Get("Content-Type"))

	// Output:
	// 200
	// application/json
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
				r: httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body)),
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
				r: httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(bodyStatusNotFoundCounter)),
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
				r: httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(bodyStatusNotFoundGauge)),
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
				r: httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader([]byte("wrong"))),
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

func ExampleGetMetricJSON() {
	// Try to request from logger server next metric: {"id": "Counter2", "type": "counter"}
	requestedMetric := `{"id": "Counter2", "type": "counter"}`
	ctx := context.Background()
	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	conf := &initconf.Config{
		Key: "superkey",
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader([]byte(requestedMetric)))
	c := SetTestGinContext(w, r)

	GetMetricJSON(ctx, store, conf)(c)
	// read result from Gin context:
	res := c.Writer
	fmt.Println(res.Status())
	fmt.Println(res.Header().Get("Content-Type"))

	// Read and print response.
	jsn, _ := io.ReadAll(w.Result().Body)
	w.Result().Body.Close()
	//if err != nil {
	//	log.Println("io.ReadAll error:", err)
	//}
	fmt.Println(string(jsn))

	// Output:
	// 200
	// application/json
	// {"id":"Counter2","type":"counter","delta":2}
}

func ExampleGetMetricJSON_second() {
	// Try to request from logger server next metric: {"id": "Counter200", "type": "counter"}
	requestedMetric := `{"id": "Counter200", "type": "counter"}`
	ctx := context.Background()
	// Test store:  map[Gauge1:1.1 Gauge2:2.2 Gauge3:3.3] map[Counter1:1 Counter2:2 Counter3:3]
	store := createTestStor(ctx)
	conf := &initconf.Config{
		Key: "superkey",
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader([]byte(requestedMetric)))
	c := SetTestGinContext(w, r)

	GetMetricJSON(ctx, store, conf)(c)
	// read result from Gin context:
	res := c.Writer
	fmt.Println(res.Status())
	fmt.Println(res.Header().Get("Content-Type"))

	// Output:
	// 404
	// application/json
}
