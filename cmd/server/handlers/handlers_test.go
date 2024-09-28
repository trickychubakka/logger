package handlers

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestMetricHandler(t *testing.T) {
	type args struct {
		w *httptest.ResponseRecorder
		r *http.Request
	}
	type want struct {
		code int
		//response    string
		contentType string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Test request -- positive test",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1", nil),
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test request -- negative test, error in urlToMap()",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1/fault/fault", nil),
			},
			want: want{
				code: http.StatusNotFound,
				//contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test request gauge metric update, wrong metric value type -- must be float64",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/gauge/metric1/1ewe", nil),
			},
			want: want{
				code: http.StatusBadRequest,
				//contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test request counter metric update, wrong metric value type -- must be int64",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/counter/metric1/1ewe", nil),
			},
			want: want{
				code: http.StatusBadRequest,
				//contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "Test request counter metric update, wrong metric type",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/bool/metric1/1", nil),
			},
			want: want{
				code: http.StatusBadRequest,
				//contentType: "text/plain; charset=utf-8",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			MetricHandler(tt.args.w, tt.args.r)
			res := tt.args.w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			// получаем и проверяем тело запроса
			defer res.Body.Close()
			//defer res.Body.Close()
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
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
