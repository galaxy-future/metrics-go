package internal

import (
	"github.com/galaxy-future/metrics-go/aggregate"
	"github.com/galaxy-future/metrics-go/types"
)

type AsyncSerial struct {
	Labels    []string
	Aggregate aggregate.Function
}

func (series *AsyncSerial) Value(value float64) {
	series.Aggregate.Accumulate(value)
}

func newAggregateValues(labelValues []string, aggregate aggregate.Function) *AsyncSerial {
	return &AsyncSerial{
		Labels:    labelValues,
		Aggregate: aggregate,
	}
}

func newSampleSeriesBuilder(metricName string, labelValues []string, sampleCache chan *Sample) types.Series {
	return &seriesBuilder{
		metricName:  metricName,
		labelValues: labelValues,
		sampleCache: sampleCache,
	}
}

type seriesBuilder struct {
	metricName  string
	labelValues []string
	sampleCache chan *Sample
}

func (s *seriesBuilder) Value(value float64) {
	newSample := Sample{
		MetricName:  s.metricName,
		LabelValues: s.labelValues,
		Value:       value,
	}
	s.sampleCache <- &newSample
}

type Sample struct {
	MetricName  string
	LabelValues []string
	Value       float64
}
