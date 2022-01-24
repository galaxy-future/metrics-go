package internal

import (
	"github.com/galaxy-future/metrics-go/aggregate"
	"github.com/galaxy-future/metrics-go/types"
	"strings"
)

const (
	seriesUniqueToken = "\t"
)

type AsyncMetrics struct {
	Name             string
	Labels           []string
	AggregateBuilder aggregate.Builder
	AggValues        map[string]*AsyncSerial
	SampleCache      chan *Sample
}

func NewMetric(metricName string, labels []string, builder aggregate.Builder, sampleCache chan *Sample) *AsyncMetrics {
	m := &AsyncMetrics{
		Labels:           labels,
		Name:             metricName,
		AggregateBuilder: builder,
		SampleCache:      sampleCache,
		AggValues:        make(map[string]*AsyncSerial),
	}
	return m
}

func (m *AsyncMetrics) MetricName() string {
	return m.Name
}

func (m *AsyncMetrics) With(labelValues ...string) types.Series {
	return newSampleSeriesBuilder(m.MetricName(), labelValues, m.SampleCache)
}

func (m *AsyncMetrics) GetSeries(labelValues ...string) *AsyncSerial {
	uniqueKey := CalcSeriesKey(labelValues)

	series, ok := m.AggValues[uniqueKey]
	if !ok {
		series = newAggregateValues(labelValues, m.AggregateBuilder.Build())
		m.AggValues[uniqueKey] = series
	}
	return series
}

func CalcSeriesKey(values []string) string {
	return strings.Join(values, seriesUniqueToken)
}
