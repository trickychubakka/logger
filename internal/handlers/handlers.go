// Package handlers -- пакет с реализацией gin-handler-ов logger сервера.
package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	//"logger/cmd/server/initconf"
	"logger/config"
	"logger/internal/database"
	"logger/internal/storage"
	"logger/internal/storage/memstorage"
	"net/http"
	"strconv"
	"strings"
)

// Константа для кодирования смысла полей после парсинга URL на основе их порядкового номера
// Пример: localhost:8080/update/gauge/metric2/7.4
const (
	metricType  = 1
	metricName  = 2
	metricValue = 3
)

// Storager интерфейс с используемыми методами.
type Storager interface {
	UpdateGauge(ctx context.Context, key string, value float64) error
	UpdateCounter(ctx context.Context, key string, value int64) error
	UpdateBatch(ctx context.Context, metrics []storage.Metrics) error
	GetGauge(ctx context.Context, key string) (float64, error)
	GetCounter(ctx context.Context, key string) (int64, error)
	GetValue(ctx context.Context, t string, key string) (any, error)
	GetAllMetrics(ctx context.Context) (any, error)
	Close() error
}

// urlToMap парсинг URL в map по разделителям "/" с предварительным удалением крайних "/".
func urlToMap(url string) ([]string, error) {
	splittedURL := strings.Split(strings.Trim(url, "/"), "/")
	// Если длина разобранного URL не больше 2-х -- недостаток указания метрики/значения, возвращаем StatusNotFound
	if len(splittedURL) <= 3 {
		return splittedURL, errors.New("URL is too short")
	}
	// Если длина разобранного URL больше 4 -- в URL что-то лишнее
	if len(splittedURL) > 4 {
		log.Println("Error in urlToMap: URL is too long")
		return splittedURL, errors.New("URL is too long")
	}
	return splittedURL, nil
}

// MetricsToMemstorage функция конвертации объекта Metrics в Memstorage.
func MetricsToMemstorage(ctx context.Context, metrics []storage.Metrics) (memstorage.MemStorage, error) {
	stor, _ := memstorage.New(ctx)
	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			_ = stor.UpdateGauge(ctx, m.ID, *m.Value)
		case "counter":
			_ = stor.UpdateCounter(ctx, m.ID, *m.Delta)
		}
	}
	log.Println("MetricsToMemstorage: []Metrics :", metrics, " -> stor :", stor)
	return stor, nil
}

// hashBody функция вычисления hash-а body сообщения и подписи сообщения в контексте gin.Context.
func hashBody(body []byte, config *config.Config, c *gin.Context) error {
	if config.Key == "" {
		log.Println("config.Key is empty")
		return nil
	}
	h := hmac.New(sha256.New, []byte(config.Key))
	h.Write(body)
	hash := h.Sum(nil)
	c.Header("HashSHA256", hex.EncodeToString(hash))
	return nil
}

// MetricsHandler -- Gin handler обработки запросов по изменениям метрик через URL.
// Обрабатывает GET запросы типа /update/:metricType/:metricName/:metricValue.
func MetricsHandler(ctx context.Context, store Storager) gin.HandlerFunc {
	return func(c *gin.Context) {
		splittedURL, err := urlToMap(c.Request.URL.String())
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		// metricHandler Обработка gauge метрики
		if splittedURL[metricType] == "gauge" {
			if val, err := strconv.ParseFloat(splittedURL[metricValue], 64); err == nil {
				if err := store.UpdateGauge(ctx, splittedURL[metricName], val); err != nil {
					c.Status(http.StatusInternalServerError)
					return
				}
			} else {
				log.Println("Error in MetricHandler: There is no metric or wrong metric value type -- must be float64")
				c.Status(http.StatusBadRequest)
				return
			}
			// metricHandler Обработка counter метрик
		} else if splittedURL[metricType] == "counter" {
			if val, err := strconv.ParseInt(splittedURL[metricValue], 10, 64); err == nil {
				if err := store.UpdateCounter(ctx, splittedURL[metricName], val); err != nil {
					c.Status(http.StatusInternalServerError)
					return
				}
			} else {
				log.Println("Error in MetricHandler: There is no metric or wrong metric value type -- must be int64")
				c.Status(http.StatusBadRequest)
				return
			}
			// Неправильный тип метрики
		} else {
			log.Println("Error in MetricHandler: Wrong metric type")
			c.Status(http.StatusBadRequest)
			return
		}
		log.Println("Requested PLAIN metric UPDATE with next metric")
		// Формируем ответ
		c.Header("content-type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
	}
}

// MetricHandlerJSON -- Gin handler обработки запросов по изменениям метрик через JSON в Body.
// Обрабатывает POST запросы на /update/ c JSON-ом метрики в body запроса.
func MetricHandlerJSON(ctx context.Context, store Storager, conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("MetricHandlerJSON START")
		jsn, err := io.ReadAll(c.Request.Body)
		if err != nil {
			http.Error(c.Writer, "Error in json body read", http.StatusInternalServerError)
			return
		}

		//var tmpMetric internal.Metrics
		var tmpMetric storage.Metrics

		err = json.Unmarshal(jsn, &tmpMetric)
		if err != nil {
			log.Println("Error in json body read 2", err, "jsn is:", string(jsn))
			c.Status(http.StatusBadRequest)
			return
		}

		log.Println("Requested JSON metric UPDATE with next metric", tmpMetric)

		if tmpMetric.MType == "gauge" {
			if err := store.UpdateGauge(ctx, tmpMetric.ID, *tmpMetric.Value); err != nil {
				log.Println("Error in UpdateGauge:", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		} else if tmpMetric.MType == "counter" {
			if err := store.UpdateCounter(ctx, tmpMetric.ID, *tmpMetric.Delta); err != nil {
				log.Println("Error in UpdateCounter:", err)
				c.Status(http.StatusInternalServerError)
				return
			}
			// обновляем во временном объекте метрики значение Counter-а для выдачи его в response
			if *tmpMetric.Delta, err = store.GetCounter(ctx, tmpMetric.ID); err != nil {
				log.Println("Error in GetCounter:", err)
				c.Status(http.StatusInternalServerError)
				return
			}
		} else {
			log.Println("Error in MetricHandlerJSON: Wrong metric type")
			c.Status(http.StatusBadRequest)
			return
		}

		j2 := io.NopCloser(bytes.NewBuffer(jsn))
		log.Println("Request from j2:", j2)

		resp, err := json.Marshal(tmpMetric)
		if err != nil {
			log.Println("Error in json.Marshal in handlers:", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := hashBody(resp, conf, c); err != nil {
			log.Println("MetricHandlerBatchUpdate: Error in hashBody:", err)
		}

		c.Header("content-type", "application/json")
		c.Status(http.StatusOK)

		if _, err := c.Writer.Write(resp); err != nil {
			log.Println("GetMetric Writer.Write error:", err)
		}

		log.Println("Initconfig before Save is:", conf)
		log.Println("start SAVE metrics dump to file: ", conf.FileStoragePath, "Store is:", store)
	}
}

// MetricHandlerBatchUpdate -- Gin handler обработки batch запроса по изменениям batch-а метрик через []Metrics в Body.
// Обрабатывает POST запросы на /updates c JSON-ом с несколькими метраками в body запроса.
func MetricHandlerBatchUpdate(ctx context.Context, store Storager, conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsn, err := io.ReadAll(c.Request.Body)
		if err != nil {
			http.Error(c.Writer, "MetricHandlerBatchUpdate: Error in json body read", http.StatusInternalServerError)
			return
		}

		var tmpMetrics []storage.Metrics

		err = json.Unmarshal(jsn, &tmpMetrics)
		if err != nil {
			log.Println("MetricHandlerBatchUpdate: Error in json.Unmarshal", err, "jsn is:", string(jsn))
			c.Status(http.StatusBadRequest)
			return
		}

		log.Println("MetricHandlerBatchUpdate: Requested JSON batch metric UPDATES with next []metric", tmpMetrics)

		j2 := io.NopCloser(bytes.NewBuffer(jsn))
		log.Println("MetricHandlerBatchUpdate: Request from j2:", j2)

		log.Println("MetricHandlerBatchUpdate. Starting storage batch update. Store before update is :", store)
		if err := store.UpdateBatch(ctx, tmpMetrics); err != nil {
			log.Println("MetricHandlerBatchUpdate. Error in UpdateBatch:", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(store)
		if err != nil {
			log.Println("MetricHandlerBatchUpdate: Error in json.Marshal in handlers:", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := hashBody(resp, conf, c); err != nil {
			log.Println("MetricHandlerBatchUpdate: Error in hashBody:", err)
		}

		c.Header("content-type", "application/json")
		c.Status(http.StatusOK)

		if _, err := c.Writer.Write(resp); err != nil {
			log.Println("MetricHandlerBatchUpdate: Writer.Write error:", err)
		}
	}
}

// memStor тип для кастования в него объекта memstorage.MemStorage перед кодированием в JSON из-за приватности исходных полей.
type memStor struct {
	GaugeMap   map[string]float64
	CounterMap map[string]int64
}

// GetAllMetrics Gin handler получения всех метрик.
// Обрабатывает GET запросы на / .
func GetAllMetrics(ctx context.Context, store Storager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var t memStor
		metrics, err := store.GetAllMetrics(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		// metrics может быть типа memstorage.MemStorage. В этом случае из-за приватности map этого типа
		// он не будет правильно кодироваться в JSON. Необходимо кастовать в тип memStor.
		switch v := metrics.(type) {
		case memstorage.MemStorage:
			log.Println("GetAllMetrics: store type", v)
			m, ok := metrics.(memstorage.MemStorage)
			if !ok {
				log.Println("GetAllMetrics error cast metric")
				c.Status(http.StatusInternalServerError)
				return
			}
			t.GaugeMap, _ = m.GetAllGaugesMap(ctx)
			t.CounterMap, _ = m.GetAllCountersMap(ctx)
			c.Header("content-type", "text/html; charset=utf-8")
			c.IndentedJSON(http.StatusOK, t)
		default:
			log.Println("GetAllMetrics: store type", v)
			c.Header("content-type", "text/html; charset=utf-8")
			c.IndentedJSON(http.StatusOK, metrics)
		}
	}
}

// GetMetric Gin handler получения значение метрики.
// Обрабатывает GET запросы типа /value/:metricType/:metricName.
func GetMetric(ctx context.Context, store Storager) gin.HandlerFunc {
	return func(c *gin.Context) {
		splittedURL, err := urlToMap(c.Request.URL.String())
		if err != nil {
			c.Status(http.StatusInternalServerError)
		}
		val, err := store.GetValue(ctx, splittedURL[metricType], splittedURL[metricName])
		if err != nil {
			fmt.Println("Error in GetMetric:", err)
			c.Status(http.StatusNotFound)
		} else {
			switch v := val.(type) {
			case float64:
				{
					c.String(http.StatusOK, fmt.Sprintf("%g", v))
				}
			case int64:
				c.String(http.StatusOK, fmt.Sprintf("%d", v))
			}
		}
	}
}

// GetMetricJSON Gin handler получения значения метрики через POST запрос с JSON с параметрами запрошенной метрики.
// Обрабатывает POST запросы на /value/.
func GetMetricJSON(ctx context.Context, store Storager, conf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsn, err := io.ReadAll(c.Request.Body)

		if err != nil {
			http.Error(c.Writer, "GetMetricJSON: Error in json body read", http.StatusInternalServerError)
			return
		}
		log.Println("GetMetricJSON, jsn after ReadAll:", string(jsn))

		var tmpMetric storage.Metrics

		err = json.Unmarshal(jsn, &tmpMetric)
		if err != nil {
			log.Println("GetMetricJSON: Error json.Unmarshal. Error is:", err)
			c.Status(http.StatusBadRequest)
			return
		}

		if tmpMetric.MType == "gauge" {
			var val float64
			val, err = store.GetGauge(ctx, tmpMetric.ID)
			// Если получили ошибку -- в соответствии со спецификацией возвращаем json запроса
			if err != nil {
				log.Println("GetMetricJSON: Error store.GetGauge", tmpMetric, "Error is", err)
				c.Header("content-type", "application/json")
				c.IndentedJSON(http.StatusNotFound, jsn)
				return
			}
			tmpMetric.Value = &val
		}
		if tmpMetric.MType == "counter" {
			var delta int64
			delta, err = store.GetCounter(ctx, tmpMetric.ID)
			// Если получили ошибку -- в соответствии со спецификацией возвращаем json запроса
			if err != nil {
				log.Println("GetMetricJSON: Error store.GetCounter", tmpMetric, "Error is", err)
				c.Header("content-type", "application/json")
				c.IndentedJSON(http.StatusNotFound, jsn)
				return
			}
			tmpMetric.Delta = &delta
		}

		resp, err := json.Marshal(tmpMetric)
		if err != nil {
			log.Println("GetMetricJSON: Error in json.Marshal with Metric:", tmpMetric, "Error is", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := hashBody(resp, conf, c); err != nil {
			log.Println("GetMetricJSON: Error in hashBody:", err)
		}

		log.Println("GetMetricJSON: Requested metric value with status 200", tmpMetric)
		j2 := io.NopCloser(bytes.NewBuffer(resp))
		log.Println("GetMetricJSON: Request value from j2:", j2)

		c.Header("content-type", "application/json")
		c.Status(http.StatusOK)
		if _, err := c.Writer.Write(resp); err != nil {
			log.Println("GetMetricJSON: Error Writer.Write error:", err)
		}
	}
}

// DBPing Gin handler проверки соединения с БД.
// Обрабатывает GET запрос на /ping.
func DBPing(connStr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Тест коннекта к базе
		log.Println("TEST Connecting to base")
		db := database.Postgresql{}
		err := db.Connect(connStr)
		if err != nil {
			log.Println("Error connecting to database :", err)
		}
		defer func(db *database.Postgresql) {
			err := db.Close()
			if err != nil {
				log.Println("DBPing: db.Close() error", err)
			}
		}(&db)
		err = db.Ping()
		if err != nil {
			log.Println("database connect error")
			c.Status(http.StatusInternalServerError)
			panic(err)
		}
		log.Println("database connected")
		c.Status(http.StatusOK)
		c.Next()
	}
}
