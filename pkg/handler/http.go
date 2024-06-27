package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	ping "github.com/go-ping/ping"
	commandUtils "github.com/jkandasa/autoeasy/pkg/utils/command"
	"github.com/jkandasa/iperf3-handler/pkg/types"
	"github.com/jkandasa/iperf3-handler/pkg/version"
	"go.uber.org/zap"
)

type HttpHandler struct{}

func NewHandler() http.Handler {
	hh := &HttpHandler{}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", hh.handleStatus)
	mux.HandleFunc("/api/version", hh.handleVersion)
	mux.HandleFunc("/api/diagnose/network", hh.handleDiagnoseNetwork)

	return mux
}

func (hh *HttpHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	status := map[string]interface{}{
		"status":          "up",
		"timestamp":       now.Unix(),
		"timestampString": now.Format(time.RFC3339),
		"go":              runtime.Version(),
		"platform":        runtime.GOOS,
		"arch":            runtime.GOARCH,
	}
	data, err := json.Marshal(status)
	if err != nil {
		zap.L().Error("error on marshaling json", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		zap.L().Error("error on writing response", zap.Error(err))
	}
}

func (hh *HttpHandler) handleVersion(w http.ResponseWriter, r *http.Request) {
	ver := version.Get()
	data, err := json.Marshal(ver)
	if err != nil {
		zap.L().Error("error on marshaling json", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		zap.L().Error("error on writing response", zap.Error(err))
	}

}

func (hh *HttpHandler) handleDiagnoseNetwork(w http.ResponseWriter, r *http.Request) {
	hostsRaw := r.URL.Query().Get(types.QueryParameterHosts)
	if hostsRaw == "" {
		http.Error(w, "'hosts' query parameter cannot be empty", http.StatusBadRequest)
		return
	}
	hosts := strings.Split(hostsRaw, ",")

	// module enabled
	pingEnabled := strings.ToLower(r.URL.Query().Get(types.QueryParameterPingEnabled)) == "true"
	iperf3Enabled := strings.ToLower(r.URL.Query().Get(types.QueryParameterIPerf3Enabled)) == "true"

	// get user supplied options for ping
	pingCount := r.URL.Query().Get(types.QueryParameterPingCount)
	pingInterval := r.URL.Query().Get(types.QueryParameterPingInterval)

	// get user supplied options for iperf3
	optionsRaw := r.URL.Query().Get(types.QueryParameterIPerf3Options)
	customOptions := []string{}
	if optionsRaw != "" {
		customOptions = strings.Split(optionsRaw, ",")
	}

	networkDiagnoseResponses := []types.NetworkDiagnoseResponse{}

	for _, hostname := range hosts {
		networkResponse := types.NetworkDiagnoseResponse{}
		// perform ping
		if pingEnabled {
			networkResponse.Ping = hh.executePing(hostname, pingCount, pingInterval)
		}

		// perform iperf3
		if iperf3Enabled {
			iperf3Cmd := strings.Replace(types.IPerf3ClientCommand, "localhost", hostname, 1)
			options := strings.Split(iperf3Cmd, " ")
			options = append(options, customOptions...)

			// execute iperf3
			networkResponse.IPerf3 = hh.executeIPerf3(hostname, options[0], options[1:])
		}
		networkDiagnoseResponses = append(networkDiagnoseResponses, networkResponse)
	}

	w.Header().Set("Content-Type", "application/json")

	data, err := json.Marshal(networkDiagnoseResponses)
	if err != nil {
		zap.L().Error("error on marshaling json", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		zap.L().Error("error on writing response", zap.Error(err))
	}
}

// executes the iperf3 throughput test to the given server(hostname)
func (hh *HttpHandler) executeIPerf3(hostname string, command string, args []string) *types.CmdResponse {
	// execute iperf3 command
	iperf3Cmd := commandUtils.Command{
		Name:                 fmt.Sprintf("iperf3-to-%s", hostname),
		Command:              command,
		Args:                 args,
		Timeout:              time.Second * 60,
		StatusUpdateDuration: time.Second * 1,
	}

	response := &types.CmdResponse{
		Hostname:  hostname,
		Options:   args,
		IsSuccess: false,
	}

	zap.L().Debug("running iperf3 command", zap.String("hostname", hostname), zap.String("command", iperf3Cmd.Command), zap.Strings("args", iperf3Cmd.Args))

	err := iperf3Cmd.StartAndWait()
	if err != nil {
		zap.L().Error("error on executing a command", zap.String("command", iperf3Cmd.Command), zap.Strings("args", iperf3Cmd.Args), zap.Error(err))
		response.ErrorMessage = err.Error()
		return response
	}

	status := iperf3Cmd.Status()
	if status.Error != nil {
		zap.L().Error("error on executing a command", zap.String("command", iperf3Cmd.Command), zap.Strings("args", iperf3Cmd.Args), zap.Error(status.Error))
		response.ErrorMessage = status.Error.Error()
		return response
	}

	response.IsSuccess = true
	response.Response = strings.Join(status.Stdout, "")
	return response
}

// executes ping tests and reports the status
func (nn *HttpHandler) executePing(hostname, countStr, intervalStr string) *types.CmdResponse {
	response := &types.CmdResponse{
		Hostname:  hostname,
		IsSuccess: false,
	}
	pinger, err := ping.NewPinger(hostname)
	if err != nil {
		zap.L().Error("error on getting pinger", zap.String("hostname", hostname), zap.Error(err))
		response.ErrorMessage = err.Error()
		return response
	}

	pinger.Count = 3
	// update count
	if countStr != "" {
		count, err := strconv.Atoi(countStr)
		if err != nil {
			zap.L().Error("error on converting count value", zap.Error(err))
			response.ErrorMessage = fmt.Sprintf("error on converting count value: %s", err.Error())
			return response
		}
		pinger.Count = count
	}

	// update interval
	if intervalStr != "" {
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			zap.L().Error("error on converting interval", zap.Error(err))
			response.ErrorMessage = fmt.Sprintf("error on converting interval: %s", err.Error())
			return response
		}
		pinger.Interval = interval
	}

	err = pinger.Run()
	if err != nil {
		zap.L().Error("error on executing pinger", zap.String("hostname", hostname), zap.Error(err))
		response.ErrorMessage = err.Error()
		return response
	}

	pingStatistics := pinger.Statistics()

	data, err := json.Marshal(pingStatistics)
	if err != nil {
		zap.L().Error("error on marshaling", zap.Error(err))
		response.ErrorMessage = err.Error()
		return response
	}

	response.IsSuccess = true
	response.Response = string(data)
	return response
}
