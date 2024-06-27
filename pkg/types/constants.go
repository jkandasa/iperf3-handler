package types

const (
	IPerf3ServerCommand = "iperf3 --server --port=5201"
	IPerf3ClientCommand = "iperf3 --client localhost --port=5201"

	QueryParameterHosts = "hosts"

	QueryParameterIPerf3Enabled = "iperf3_enabled"
	QueryParameterIPerf3Options = "iperf3_options"

	QueryParameterPingEnabled  = "ping_enabled"
	QueryParameterPingCount    = "ping_count"
	QueryParameterPingInterval = "ping_interval"
)
