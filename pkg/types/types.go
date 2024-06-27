package types

type NetworkDiagnoseResponse struct {
	IPerf3 *CmdResponse `json:"iperf3,omitempty"`
	Ping   *CmdResponse `json:"ping,omitempty"`
}

type CmdResponse struct {
	Hostname     string   `json:"hostname"`
	Response     string   `json:"response"`
	ErrorMessage string   `json:"error"`
	IsSuccess    bool     `json:"isSuccess"`
	Options      []string `json:"options"`
}
