package streaming

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
	streamingMaps map[string]*internal.AsyncMetrics
	sampleCache   chan *internal.Sample
)

func NewMetric(metricName string, labels []string) types.Metrics {
	metric := internal.NewMetric(metricName, labels, aggregate.NewStreamingBuilder(), sampleCache)
	streamingMaps[metricName] = metric
	return metric
}

func flushStreaming() {
	//先获得所有待发送的Series
	var toSendMetrics []*internal.AsyncMetrics
	for _, item := range streamingMaps {
		toSendMetrics = append(toSendMetrics, &internal.AsyncMetrics{
			Name:      item.Name,
			Labels:    item.Labels,
			AggValues: item.AggValues,
		})
	}

	//清空现在统计
	for _, metric := range streamingMaps {
		metric.AggValues = make(map[string]*internal.AsyncSerial)
	}

	for _, metric := range toSendMetrics {
		batch := streamingToBatch(metric)
		err := internal.GatewayClient.SendStreaming(batch)
		if err != nil {
			logger.GetLogger().Error(err.Error())
		}
	}

}

func streamingToBatch(m *internal.AsyncMetrics) *mod.StreamingBatch {
	timestamp := time.Now().Truncate(internal.AggregateDuration).Unix() * 1000
	batch := mod.StreamingBatch{}
	batch.MetricName = m.Name
	batch.ServiceName = internal.ServiceName
	for _, serial := range m.AggValues {
		message := mod.StreamingMessage{}
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
		message.Values = serial.Aggregate.Values()
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
			flushStreaming()
		case s, ok := <-sampleCache:
			if ok {
				if metric, exists := streamingMaps[s.MetricName]; exists {
					metric.GetSeries(s.LabelValues...).Value(s.Value)
				}
			} else {
				flushStreaming()
				return nil
			}

		}
	}
	close(sampleCache)
	for s := range sampleCache {
		if metric, exists := streamingMaps[s.MetricName]; exists {
			metric.GetSeries(s.LabelValues...).Value(s.Value)
		}
	}
	flushStreaming()
	return nil
}

func InitStreamingMetrics() {
	streamingMaps = make(map[string]*internal.AsyncMetrics)
	sampleCache = make(chan *internal.Sample, defaultSamplesCache)
}
