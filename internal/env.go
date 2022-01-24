package internal

import (
	"github.com/galaxy-future/metrics-go/common/logger"
	"github.com/galaxy-future/metrics-go/common/utils"
	"os"
	"time"
)

const (
	AggregateDurationEnv = "CUDGX_AGG_DURATION"
	SendTimeoutEnv       = "CUDGX_SEND_TIMEOUT"
	ResendTimesEnv       = "CUDGX_RESEND_TIMES"
	ResendBackoffEnv     = "CUDGX_BACKOFF_DURATION"
	GatewayUrlEnv        = "CUDGX_GATEWAY_URL"
	ServiceNameEnv       = "CUDGX_SERVICE_NAME"
	ClusterNameEnv       = "CUDGX_CLUSTER_NAME"
)

var (
	AggregateDuration = time.Second
	SendTimeout       = time.Second * 5
	ResendTimes       = 1
	DurationBackoff   = 100 * time.Millisecond
	GatewayUrl        = "http://cudgx-gateway.internal.galaxy-future.org"
	HostName          = ""
	ServiceName       = ""
	ClusterName       = ""
)

func InitEnvironment() {
	AggregateDuration = utils.TryGetDurationEnvironment(AggregateDurationEnv, AggregateDuration)
	ResendTimes = utils.TryGetIntEnvironment(ResendTimesEnv, ResendTimes)
	DurationBackoff = utils.TryGetDurationEnvironment(ResendBackoffEnv, DurationBackoff)
	GatewayUrl = utils.TryGetStringEnvironment(GatewayUrlEnv, GatewayUrl)
	SendTimeout = utils.TryGetDurationEnvironment(SendTimeoutEnv, SendTimeout)
	ServiceName = utils.TryGetStringEnvironment(ServiceNameEnv, ServiceName)
	ClusterName = utils.TryGetStringEnvironment(ClusterNameEnv, ClusterName)

	if host, err := os.Hostname(); err == nil {
		HostName = host
	} else {
		logger.GetLogger().Warn("can not get host Name from system, use empty string")
	}
}
