## Metrics-Go 

`metrics-go` is the dotting tool of CudgX metrics


### Data Flow

The matric data flow is as follows:
1.	User integrates code with SDK
2.	SDK matric integration: SDK aggregates user data in a specified period (1s by default).
3.	Batch Push: Each aggregation cycle pushes metrics to CudgX-gateway
4.	CudgX-gateway distributes data to Kafka 
5.	CudgX-consumer consumes data and stores it to clickhouse
6.	Query metrics based on clickhouse data 



### How to use SDK complete metrics data collection 
Metrics are classified into two categories: Monitoring Metrics and Streaming Metrics
- Monitoring Metrics: After SDK Aggregation, data is reported to CudgX-gateway. 
- Streaming Metrics: SDK collects detailed data of metrics and reports data to CudgX-gateway without aggregation Create Monitoring Metrics.


**1.Create  Metrics**

（1）Create Monitoring Metrics

QPS metrics data collection, for example:
```go
qps = metricGo.NewMonitoringMetric("qps", []string{}, aggregate.NewCountBuilder())
```

（2）Create Streaming Metrics 
 latency data collection, for example:
```go
latency = metricGo.NewStreamingMetric("latency", []string{})
```

**2.	Dotting (Collect Metrics Data)**

For example:
```go
begin := time.Now()
// Business Interface/Method

cost := time.Now().Sub(begin).Milliseconds()
qps.With().Value(1)
latency.With().Value(float64(cost))
```

### How to customize metrics
metric_qps metrics access example:


**1.Implement Function and Builder interfaces in metrics-go**

![base](./images/base.png)

（1）Implement Function intf, for example：

![function](./images/function.png)

（2）Implement Builder intf,for example：

![builder](./images/builder.png)

**2.Create Monitoring Metrics**

for example：

![new](./images/new.png)

**3.Dotting (Collect Metrics Data)**

for example：

![with](./images/with.png)

### Metrics Access Example Application: `cudgx-sample-pi`

click to view [cudgx-sample-pi](https://github.com/galaxy-future/cudgx/blob/master/sample/pi/main.go) sample program


Code of Conduct
------
[Contributor Convention](https://github.com/galaxy-future/cudgx/blob/master/CODE_OF_CONDUCT.md)

Authorization
-----

Metrics-Go uses [Elastic License 2.0](https://github.com/galaxy-future/cudgx/blob/master/LICENSE) Agreement for Authorization.

