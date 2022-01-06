## Metrics-Go 

`metrics-go` 是cudgx指标打点工具，它集成了监控和数据分析指标能力。


### 数据流程

指标数据流程为： 
1. 用户代码调用打点 
2. SDK指标聚合 ，SDK会将用户的打点数据按照指定时间周期聚合（默认是1s） 
3. batch推送 ，每个聚合周期会将指标推送到Gateway中
4. Gateway 将数据分发到Kafka
5. 消费数据，存储到Clickhouse
6. 用户基于Clickhouse查询指标

### 如何使用

指标分为两类： 监控指标和流式指标

监控指标： sdk聚合之后存储链路将数据直接存储在clickhouse中
流式指标： sdk收集指标详细数据，不做聚合将数据传输到clickhouse，用户可以基于收集到的数据对数据流失计算

**新建监控指标**

```go
latencyMin = metricGo.NewMonitoringMetric("latencyMin", []string{}, aggregate.NewMinBuilder())
latencyMax = metricGo.NewMonitoringMetric("latencyMax", []string{}, aggregate.NewMaxBuilder())
```

**新建流式指标**

```go
latency = metricGo.NewStreamingMetric("latency", []string{})
```

**打点**

```go
latencyMin.With().Value(float64(cost))
latencyMax.With().Value(float64(cost))
latency.With().Value(float64(cost))
```

行为准则
------
[贡献者公约](https://github.com/galaxy-future/cudgx/blob/master/CODE_OF_CONDUCT.md)

授权
-----

Metrics-Go使用[Elastic License 2.0](https://github.com/galaxy-future/cudgx/blob/master/LICENSE)授权协议进行授权

