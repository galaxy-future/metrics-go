package monitoring

import (
	"context"
	"github.com/galaxy-future/metrics-go/aggregate"
	"github.com/galaxy-future/metrics-go/common/logger"
	"github.com/galaxy-future/metrics-go/common/mod"
	"github.com/galaxy-future/metrics-go/internal"
	"github.com/galaxy-future/metrics-go/types"
	"time"
)

const (
	defaultSamplesCache = 100000
)

var (
	metricMaps  map[string]*internal.AsyncMetrics
	sampleCache chan *internal.Sample
)

func NewMetric(metricName string, labels []string, builder aggregate.Builder) types.Metrics {
	metric := internal.NewMetric(metricName, labels, builder, sampleCache)
	metricMaps[metricName] = metric
	return metric
}

func flushMetric() {
	//先获得所有待发送的Series
	var toSendMetrics []*internal.AsyncMetrics
	for _, item := range metricMaps {
		toSendMetrics = append(toSendMetrics, &internal.AsyncMetrics{
			Name:      item.Name,
			Labels:    item.Labels,
			AggValues: item.AggValues,
		})
	}

	//清空现在统计
	for _, metric := range metricMaps {
		metric.AggValues = make(map[string]*internal.AsyncSerial)
	}

	for _, metric := range toSendMetrics {
		batch := metricToBatch(metric)
		err := internal.GatewayClient.SendMetric(batch)
		if err != nil {
			logger.GetLogger().Error(err.Error())
		}
	}

}

func metricToBatch(m *internal.AsyncMetrics) *mod.MetricBatch {
	//timestamp := time.Now().Truncate(internal.AggregateDuration).Unix() * 1000
	timestamp := time.Now().Unix()
	batch := mod.MetricBatch{}
	batch.MetricName = m.Name
	batch.ServiceName = internal.ServiceName
	for _, serial := range m.AggValues {
		message := mod.MetricsMessage{}
		batch.Messages = append(batch.Messages, &message)
		message.Labels = make(map[string]string)
		for index, key := range m.Labels {
			message.Labels[key] = serial.Labels[index]
		}
		message.MetricName = m.Name
		message.ServiceName = internal.ServiceName
		message.ClusterName = internal.ClusterName
		message.ServiceHost = internal.HostName
		message.Timestamp = timestamp
		message.Value = serial.Aggregate.Aggregate()
	}
	return &batch
}

//Start 打点工具会定时上报打点状态
func Start(ctx context.Context) error {
	err := internal.InitGatewayClient()
	if err != nil {
		return err
	}
	tick := time.NewTicker(internal.AggregateDuration)
Looper:
	for {
		select {
		case <-ctx.Done():
			break Looper
		case <-tick.C:
			flushMetric()
		case s, ok := <-sampleCache:
			if ok {
				if metric, exists := metricMaps[s.MetricName]; exists {
					metric.GetSeries(s.LabelValues...).Value(s.Value)
				}
			}
		}
	}
	close(sampleCache)
	for s := range sampleCache {
		if metric, exists := metricMaps[s.MetricName]; exists {
			metric.GetSeries(s.LabelValues...).Value(s.Value)
		}
	}
	flushMetric()
	return nil
}

func InitMonitoringMetrics() {
	metricMaps = make(map[string]*internal.AsyncMetrics)
	sampleCache = make(chan *internal.Sample, defaultSamplesCache)

}
