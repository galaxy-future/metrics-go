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
	"net/http"
	"time"
)

var GatewayClient *GatewaySender

//GatewaySender 负责将收集到的指标推送到gateway分发
type GatewaySender struct {
	retryTime       int
	timeout         time.Duration
	backoffDuration time.Duration
	gatewayUrl      string
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
	}
	if err := sender.ping(); err != nil {
		return nil, err
	}
	return nil, nil
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
	client := &http.Client{
		Timeout: SendTimeout,
	}
	resp, err := client.Post(fmt.Sprintf("%s/v1/%v/%s/%s", GatewayUrl, method, ServiceName, metricsName), "application/binary", bytes.NewBuffer(data))
	if err != nil {
		return err
	}

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
	client := &http.Client{
		Timeout: SendTimeout,
	}
	resp, err := client.Get(fmt.Sprintf("%s/ping", GatewayUrl))
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
