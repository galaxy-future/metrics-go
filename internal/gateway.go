package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/galaxy-future/metrics-go/common/logger"
	"github.com/galaxy-future/metrics-go/common/mod"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	MaxIdleConns        int = 10
	MaxIdleConnsPerHost int = 10
	IdleConnTimeout     int = 90
)

var GatewayClient *GatewaySender

//GatewaySender 负责将收集到的指标推送到gateway分发
type GatewaySender struct {
	retryTime       int
	timeout         time.Duration
	backoffDuration time.Duration
	gatewayUrl      string
	httpclient      *http.Client
}

func InitGatewayClient() error {
	client, err := newGatewaySender()
	if err != nil {
		return err
	}
	GatewayClient = client
	return nil
}

func newGatewaySender() (*GatewaySender, error) {
	sender := &GatewaySender{
		retryTime:       ResendTimes,
		timeout:         SendTimeout,
		backoffDuration: DurationBackoff,
		gatewayUrl:      GatewayUrl,
		httpclient:      createHTTPClient(),
	}
	if err := sender.ping(); err != nil {
		return nil, err
	}
	return sender, nil
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        MaxIdleConns,
			MaxIdleConnsPerHost: MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Duration(IdleConnTimeout) * time.Second,
		},
	}
	return client
}

func (gateway *GatewaySender) SendMetric(batch *mod.MetricBatch) error {
	data, err := proto.Marshal(batch)
	if err != nil {
		return errors.Wrap(err, "can not marshal message")
	}

	err = sendBatch("monitoring", data, batch.MetricName)
	if err != nil {
		logger.GetLogger().Error("failed to send metric", zap.String("error", err.Error()))
	} else {
		return nil
	}

	for i := 0; i < ResendTimes; i++ {
		err = sendBatch("monitoring", data, batch.MetricName)
		if err != nil {
			logger.GetLogger().Error("failed to send metric", zap.String("error", err.Error()))
			time.Sleep(DurationBackoff)
			continue
		}
		return nil
	}

	return fmt.Errorf("send metric after %d times", ResendTimes)
}

func (gateway *GatewaySender) SendStreaming(batch *mod.StreamingBatch) error {
	data, err := proto.Marshal(batch)
	if err != nil {
		return errors.Wrap(err, "can not marshal message")
	}

	err = sendBatch("streaming", data, batch.MetricName)
	if err != nil {
		logger.GetLogger().Error("failed to send metric", zap.String("error", err.Error()))
	} else {
		return nil
	}

	for i := 0; i < ResendTimes; i++ {
		err = sendBatch("streaming", data, batch.MetricName)
		if err != nil {
			logger.GetLogger().Error("failed to send metric", zap.String("error", err.Error()))
			time.Sleep(DurationBackoff)
			continue
		}
		return nil
	}

	return fmt.Errorf("send metric after %d times", ResendTimes)
}

func sendBatch(method string, data []byte, metricsName string) error {

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/%v/%s/%s", GatewayUrl, method, ServiceName, metricsName), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/binary")
	req.Header.Set("Connection", "keep-alive")

	resp, err := GatewayClient.httpclient.Do(req)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode/100 == 2 {
		return nil
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("failed to send metric to gateway message: %s", string(respData))
}

func (gateway *GatewaySender) ping() error {

	resp, err := gateway.httpclient.Get(fmt.Sprintf("%s/ping", GatewayUrl))
	if err != nil {
		return err
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var pingResult mod.GatewayPingResult
	err = json.Unmarshal(respData, &pingResult)
	if err != nil {
		return err
	}

	if pingResult.Module != mod.GatewayModuleName || pingResult.Status != mod.GatewayStatusSuccess {
		return fmt.Errorf("gateway internal error, ping result: %v", string(respData))
	}
	return nil

}
