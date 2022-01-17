package metrics_go

import (
	"context"
	"github.com/galaxy-future/metrics-go/aggregate"
	"github.com/galaxy-future/metrics-go/common/logger"
	"github.com/galaxy-future/metrics-go/internal"
	"github.com/galaxy-future/metrics-go/internal/monitoring"
	"github.com/galaxy-future/metrics-go/internal/streaming"
	"github.com/galaxy-future/metrics-go/types"
	"go.uber.org/zap"
)

//NewMonitoringMetric 新建指标打点
func NewMonitoringMetric(metricName string, labels []string, builder aggregate.Builder) types.Metrics {
	return monitoring.NewMetric(metricName, labels, builder)
}

//NewStreamingMetric 新建指标流收集
func NewStreamingMetric(metricName string, labels []string) types.Metrics {
	return streaming.NewMetric(metricName, labels)
}

func init() {
	internal.InitEnvironment()
	monitoring.InitMonitoringMetrics()
	go func() {
		err := monitoring.Start(context.Background())
		if err != nil {
			logger.GetLogger().Error("can not start monitoring/metric", zap.Error(err))
		}
	}()

	streaming.InitStreamingMetrics()
	go func() {
		err := streaming.Start(context.Background())
		if err != nil {
			logger.GetLogger().Error("can not start monitoring/metric", zap.Error(err))
		}
	}()
}
