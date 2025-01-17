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
	"logger/cmd/server/initconf"
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

// urlToMap парсинг URL в map по разделителям "/" с предварительным удалением крайних "/"
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

// MetricsToMemstorage функция конвертации Metrics в Memstorage
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

//func hashBody(body []byte, config *initconf.Config) ([]byte, bool) {
//	if config.Key == "" {
//		log.Println("config.Key is empty")
//		return nil, false
//	}
//	h := hmac.New(sha256.New, []byte(config.Key))
//	h.Write(body)
//	dst := h.Sum(nil)
//	fmt.Printf("%x", dst)
//	return dst, true
//}

// MetricsHandler -- Gin handlers обработки запросов по изменениям метрик через URL
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

// MetricHandlerJSON -- Gin handlers обработки запросов по изменениям метрик через JSON в Body
func MetricHandlerJSON(ctx context.Context, store Storager, conf *initconf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
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

func hashBody(body []byte, config *initconf.Config, c *gin.Context) error {
	if config.Key == "" {
		log.Println("config.Key is empty")
		return nil
	}
	h := hmac.New(sha256.New, []byte(config.Key))
	h.Write(body)
	hash := h.Sum(nil)
	fmt.Printf("%x", hash)
	log.Println("hash is:", hash)
	log.Printf("HashSHA256 is : %x", hash)
	c.Header("HashSHA256", hex.EncodeToString(hash))
	return nil
}

// MetricHandlerBatchUpdate -- Gin handlers обработки batch запроса по изменениям batch-а метрик через []Metrics в Body
func MetricHandlerBatchUpdate(ctx context.Context, store Storager, conf *initconf.Config) gin.HandlerFunc {
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

		log.Println("MetricHandlerBatchUpdate: tmpMetrics : ", tmpMetrics, " -> store :", store)

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

		//hash, keyBool := hashBody(resp, conf, c)
		//log.Println("hash is:", hash, "keyBool is :", keyBool)
		//if err != nil {
		//	log.Println("MetricHandlerBatchUpdate. Error hashing response body:", err)
		//}
		//log.Printf("HashSHA256 is : %x", hash)
		//if keyBool {
		//	c.Header("HashSHA256", hex.EncodeToString(hash))
		//}

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

// GetAllMetrics получить все метрики
func GetAllMetrics(ctx context.Context, store Storager) gin.HandlerFunc {
	return func(c *gin.Context) {

		metrics, err := store.GetAllMetrics(ctx)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("content-type", "text/html; charset=utf-8")
		c.Status(http.StatusOK)
		c.IndentedJSON(http.StatusOK, metrics)
	}
}

// GetMetric получить значение метрики
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

// GetMetricJSON получить значение метрики через JSON
func GetMetricJSON(ctx context.Context, store Storager, conf *initconf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsn, err := io.ReadAll(c.Request.Body)
		log.Println("GetMetricJSON, jsn after ReadAll:", string(jsn))
		if err != nil {
			http.Error(c.Writer, "Error in json body read", http.StatusInternalServerError)
			return
		}

		//var tmpMetric internal.Metrics
		var tmpMetric storage.Metrics

		err = json.Unmarshal(jsn, &tmpMetric)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}

		if tmpMetric.MType == "gauge" {
			var val float64
			val, err = store.GetGauge(ctx, tmpMetric.ID)
			tmpMetric.Value = &val
		}
		if tmpMetric.MType == "counter" {
			var delta int64
			delta, err = store.GetCounter(ctx, tmpMetric.ID)
			tmpMetric.Delta = &delta
		}
		if err != nil {
			log.Println("Requested metric value with status 404", tmpMetric)
			//log.Println("error is", err)
			c.Status(http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(tmpMetric)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}

		if err := hashBody(resp, conf, c); err != nil {
			log.Println("MetricHandlerBatchUpdate: Error in hashBody:", err)
		}

		log.Println("Requested metric value with status 200", tmpMetric)
		j2 := io.NopCloser(bytes.NewBuffer(resp))
		log.Println("Request value from j2:", j2)

		c.Header("content-type", "application/json")
		c.Status(http.StatusOK)
		if _, err := c.Writer.Write(resp); err != nil {
			log.Println("GetMetricJSON Writer.Write error:", err)
		}
	}
}

func DBPing(connStr string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Тест коннекта к базе
		log.Println("TEST Connecting to base")
		db := database.Postgresql{}
		err := db.Connect(connStr)
		if err != nil {
			log.Println("Error connecting to database :", err)
		}
		defer db.Close()
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
