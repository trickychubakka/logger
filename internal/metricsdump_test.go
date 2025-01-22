package internal

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"log"
	"logger/cmd/server/initconf"
	"logger/internal/handlers"
	"logger/internal/storage/memstorage"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func createTestStor(ctx context.Context) Storager {
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

func createTestHandlersStor(ctx context.Context) handlers.Storager {
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

// createDBConfig вспомогательная функция создания временного файла конфигурации
func createFile(filename string) error {
	// Создание файла с именем "greeting.txt" и запись в него данных
	err := os.WriteFile(filename, []byte("Wrong"), 0644)
	if err != nil {
		log.Println("createFile:", err)
		return err
	}
	return nil
}

// deleteFile вспомогательная функция удаления временного файла конфигурации
func deleteFile(filename string) error {
	if err := os.Remove(filename); err != nil {
		return err
	}
	return nil
}

// SetTestGinContext вспомогательная функция создания Gin контекста
func SetTestGinContext(w *httptest.ResponseRecorder, r *http.Request) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(w)
	c.Request = r
	c.Request.Header.Set("Content-Type", "text/plain")
	return c
}

// Fake store для негативного теста, эмуляция "неправильного" хранилища.
type FakeStor struct {
}

func (FakeStor) GetAllMetrics(_ context.Context) (any, error) {
	return nil, fmt.Errorf("fake error")
}

func TestLoad(t *testing.T) {
	type args struct {
		fname string
		//store memstorage.MemStorage
		store Storager
	}
	store := createTestStor(context.Background())
	tests := []struct {
		name string
		args args
		//want    handlers.Storager
		wantErr bool
	}{
		{
			name: "Positive test Load",
			args: args{
				fname: "./testDumpFile.dmp",
				store: store,
			},
			wantErr: false,
		},
		{
			name: "Negative wrong filename test Load",
			args: args{
				fname: "./testDumpFile.dmp",
				store: store,
			},
			wantErr: true,
		},
		{
			name: "Negative wrong store test Load",
			args: args{
				fname: "./wrongfile.dmp",
				store: store,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Для проверки на ошибку загрузки файла с диска отключаем предварительное создание дампа.
			if !tt.wantErr {
				_ = Save(context.Background(), tt.args.store, tt.args.fname)
			}
			if tt.args.fname == "./wrongfile.dmp" {
				createFile(tt.args.fname)
			}
			defer deleteFile(tt.args.fname)
			_, err := Load(tt.args.fname)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			//assert.Equalf(t, tt.want, got, "Load(%v)", tt.args.fname)
		})
	}
}

func TestSave(t *testing.T) {
	type args struct {
		fname string
		//store memstorage.MemStorage
		store Storager
	}
	tests := []struct {
		name string
		args args
		//want    handlers.Storager
		wantErr bool
	}{
		{
			name: "Positive test Save",
			args: args{
				fname: "./testDumpFile.dmp",
				store: createTestStor(context.Background()),
			},
			wantErr: false,
		},
		{
			name: "Negative test Save",
			args: args{
				fname: "./testDumpFile.dmp",
				store: FakeStor{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Save(context.Background(), tt.args.store, tt.args.fname)
			if err != nil {
				log.Println("save error:", err)
			}
			defer deleteFile(tt.args.fname)
			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestSyncDumpUpdate(t *testing.T) {
	type args struct {
		ctx   context.Context
		store handlers.Storager
		conf  *initconf.Config
		w     *httptest.ResponseRecorder
		r     *http.Request
	}
	tests := []struct {
		name string
		args args
		//wantErr  bool
		wantCode string
	}{
		{
			name: "Positive test SyncDumpUpdate",
			args: args{
				ctx:   context.Background(),
				store: createTestHandlersStor(context.Background()),
				conf: &initconf.Config{
					StoreMetricInterval: 0,
					FileStoragePath:     "./testDumpFile.dmp",
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", nil),
			},
			wantCode: "success",
		},
		{
			name: "Negative test SyncDumpUpdate",
			args: args{
				ctx:   context.Background(),
				store: createTestHandlersStor(context.Background()),
				conf: &initconf.Config{
					StoreMetricInterval: 0,
				},
				w: httptest.NewRecorder(),
				r: httptest.NewRequest(http.MethodPost, "/update/", nil),
			},
			wantCode: "fail",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := SetTestGinContext(tt.args.w, tt.args.r)
			SyncDumpUpdate(tt.args.ctx, tt.args.store, tt.args.conf)(c)
			status, _ := c.Get("SyncDumpUpdate")
			log.Println("status is:", status)
			assert.Equalf(t, tt.wantCode, status, "SyncDumpUpdate(%v, %v, %v)", tt.args.ctx, tt.args.store, tt.args.conf)
		})
	}
}
