package compress

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
	"logger/cmd/server/initconf"
	"net/http"
	"net/http/httptest"
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

// gzipBody функция gzip сжатия body для request тестового Gin контекста.
func gzipBody(body []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(body)
	if err != nil {
		log.Println("gzipBody error:", err)
	}
	w.Flush()
	w.Close()
	return b.Bytes()
}

// hashBody функция вычисления hash body.
func hashBody(body []byte, key string) []byte {
	if key == "" {
		log.Println("config.Key is empty")
		return nil
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(body)
	dst := h.Sum(nil)
	fmt.Printf("%x", dst)
	return dst
}

func TestGzipRequestHandle(t *testing.T) {
	type args struct {
		config *initconf.Config
		body   []byte
		w      *httptest.ResponseRecorder
		r      *http.Request
		hash   string
	}
	type want struct {
		code        int
		contentType string
		gzipStatus  string
	}
	// gzipped body fo request.
	body := bytes.NewReader(gzipBody([]byte("test body")))
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Positive test func TestGzipRequestHandle(t *testing.T)",
			args: args{
				//body: []byte("test body"),
				config: &initconf.Config{
					Key: "testkey",
				},
				w:    httptest.NewRecorder(),
				r:    httptest.NewRequest(http.MethodPost, "/update/", body),
				hash: hex.EncodeToString(hashBody(gzipBody([]byte("test body")), "testkey")),
			},
			want: want{
				gzipStatus: "success",
				code:       http.StatusOK,
			},
		},
		{
			name: "checkSign error TestGzipRequestHandle(t *testing.T)",
			args: args{
				body: []byte("test body"),
				config: &initconf.Config{
					Key: "testkey",
				},
				w:    httptest.NewRecorder(),
				r:    httptest.NewRequest(http.MethodPost, "/update/", body),
				hash: "BadHash",
			},
			want: want{
				gzipStatus: "error",
				code:       http.StatusBadRequest,
			},
		},
		{
			name: "gzip error TestGzipRequestHandle(t *testing.T)",
			args: args{
				body: []byte("test body"),
				config: &initconf.Config{
					Key: "testkey",
				},
				w:    httptest.NewRecorder(),
				r:    httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte("wrongBody"))),
				hash: hex.EncodeToString(hashBody([]byte("wrongBody"), "testkey")),
			},
			want: want{
				gzipStatus: "error",
				code:       http.StatusInternalServerError,
			},
		},
		{
			name: "gzip error with non-HASH TestGzipRequestHandle(t *testing.T)",
			args: args{
				body: []byte("test body"),
				config: &initconf.Config{
					Key: "",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader([]byte("wrongBody"))),
			},
			want: want{
				gzipStatus: "error",
				code:       http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
			c.Request.Header.Set(`Content-Encoding`, `compress`)
			if tt.args.config.Key != "" {
				c.Request.Header.Set("HashSHA256", tt.args.hash)
			}
			GzipRequestHandle(context.Background(), tt.args.config)(c)
			keyGot, _ := c.Get("GzipRequestHandle")
			assert.Equal(t, tt.want.gzipStatus, keyGot)
			assert.Equal(t, tt.want.code, c.Writer.Status())
		})
	}
}

func Test_checkSign(t *testing.T) {
	type args struct {
		body   []byte
		hash   string
		config *initconf.Config
		//key    string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Positive test Test_checkSign",
			args: args{
				body: []byte("test body"),
				hash: hex.EncodeToString(hashBody([]byte("test body"), "testkey")),
				config: &initconf.Config{
					Key: "testkey",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty Key error test Test_checkSign",
			args: args{
				body: []byte("test body"),
				//key:  "wrongKey",
				hash: hex.EncodeToString(hashBody([]byte("test body"), "testkey")),
				config: &initconf.Config{
					Key: "",
				},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "hex.DecodeString error test Test_checkSign",
			args: args{
				body: []byte("test body"),
				//key:  "wrongKey",
				hash: "WrongHash",
				config: &initconf.Config{
					Key: "testkey",
				},
			},
			wantErr: true,
		},
		{
			name: "Wrong Signature test Test_checkSign",
			args: args{
				body: []byte("test body"),
				//key:  "wrongKey",
				hash: hex.EncodeToString(hashBody([]byte("test body"), "testkey")),
				config: &initconf.Config{
					Key: "wrongKey",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := checkSign(tt.args.body, tt.args.hash, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkSign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
