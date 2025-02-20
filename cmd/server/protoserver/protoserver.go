package protoserver

import (
	"context"
	"fmt"
	"log"
	pb "logger/cmd/proto"
	"logger/internal/handlers"
	"logger/internal/storage"
)

//type Metrics struct {
//	ID    string   `json:"id"`
//	MType string   `json:"type"`
//	Delta *int64   `json:"delta,omitempty"`
//	Value *float64 `json:"value,omitempty"`
//}

type MetricsBunchServer struct {
	pb.UnimplementedMetricsBunchServer
	Store handlers.Storager
}

// BunchToMemstorage конвертация хранилища метрик типа pb.Bunch в MetricsStorage.
func BunchToMemstorage(p *pb.Bunch) ([]storage.Metrics, error) {
	var metrics []storage.Metrics
	var tmpMetric storage.Metrics
	for _, v := range p.GetMetric() {
		tmpMetric.ID = v.Id
		tmpMetric.MType = v.Type
		tmpMetric.Value = &v.Value
		tmpMetric.Delta = &v.Delta
		fmt.Printf("BunchToMemstorage -- ID: %s, Type: %s, Value is: %f, Delta is: %d ", tmpMetric.ID, tmpMetric.MType, *tmpMetric.Value, *tmpMetric.Delta)
		metrics = append(metrics, tmpMetric)
	}
	return metrics, nil
}

func (m *MetricsBunchServer) AddBunch(ctx context.Context, in *pb.AddBunchRequest) (*pb.AddBunchResponse, error) {
	var response pb.AddBunchResponse
	mBunch, err := BunchToMemstorage(in.Bunch)
	if err != nil {
		log.Println("AddBunch error:", err)
	}

	log.Println("AddBunch: mBunch :", mBunch)

	if err := m.Store.UpdateBatch(ctx, mBunch); err != nil {
		log.Println("AddBunch. Error in AddBunch:", err)
		response.Error = fmt.Sprintf("AddBunch. Error in AddBunch %s", err)
	}
	return &response, nil
}
