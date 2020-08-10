package prometheus

import (
	"encoding/json"
	"github.com/yinjk/go-utils/pkg/utils/httpclient"
	"github.com/prometheus/common/log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/common/model"
	"strings"
)

const (
	KeyDefault = "default"
	KeySolace  = "solace"
	KeyPass    = "pass"
	KeyHost    = "host"
	KeyK8s     = "k8s"
)

type apiResponse struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrorType ErrorType       `json:"errorType"`
	Error     string          `json:"error"`
}

// NewAPI returns a new API for the client.
//
// It is safe to use the returned API from multiple goroutines.
func NewAPI(config *Config) API {
	config.key = KeyDefault
	return &promAPI{client: config}
}

type Config struct {
	key  string
	Urls map[string]string
}

func (pc Config) URL(ep string, args map[string]string) string {
	if pc.key == "" {
		pc.key = KeyDefault
	}
	if host := pc.Urls[pc.key]; host == "" {
		pc.Urls[pc.key] = pc.Urls[KeyDefault]
	}
	url := "http://" + pc.Urls[pc.key]
	if ep != "" {
		for k, v := range args { //替换url中的restful风格参数
			k = ":" + k
			ep = strings.Replace(ep, k, v, -1)
		}
		url = url + ep
	}
	return url
}

type promAPI struct {
	client *Config
}

func (p *promAPI) NewPromQLHandler(promQL string) *PromQLHandler {
	return &PromQLHandler{Prom: promQL}
}

func (p *promAPI) Key(key string) API {
	client := *p.client
	client.key = key
	return &promAPI{client: &client}
}

func (p *promAPI) AlertManagers() (AlertManagersResult, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(EP_ALERT_MANAGERS, nil), &ares)
	if err != nil || response.Code != http.StatusOK {
		return AlertManagersResult{}, err
	}
	if ares.Status == PromError {
		return AlertManagersResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var amr AlertManagersResult
	if err = json.Unmarshal(ares.Data, &amr); err != nil {
		log.Errorf("promAPI.AlertManagers call failed, err = %v", err)
		return AlertManagersResult{}, err
	}
	return amr, nil
}

func (p *promAPI) CleanTombstones() error {
	_, err := httpclient.PostForm(p.client.URL(epCleanTombstones, nil), nil, nil, nil)
	return err
}

func (p *promAPI) Config() (ConfigResult, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(epConfig, nil), &ares)
	if err != nil || response.Code != http.StatusOK {
		return ConfigResult{}, err
	}
	if ares.Status == PromError {
		return ConfigResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var cr ConfigResult
	if err = json.Unmarshal(ares.Data, &cr); err != nil {
		log.Errorf("promAPI.AlertManagers call failed, err = %v", err)
		return ConfigResult{}, err
	}
	return cr, nil
}

func (p *promAPI) DeleteSeries(matches []string, startTime time.Time, endTime time.Time) (err error) {
	value := httpclient.NewFormValue()
	for _, m := range matches {
		value.Add("match[]", m)
	}
	if !startTime.IsZero() {
		value.Set("cronjob", startTime.Format(time.RFC3339Nano))
	}
	if !endTime.IsZero() {
		value.Set("end", endTime.Format(time.RFC3339Nano))
	}

	response, err := httpclient.PostForm(httpclient.EncodeUrl(p.client.URL(epDeleteSeries, nil), value), nil, nil, "")
	if err != nil {
		return err
	}
	if response.Code != http.StatusNoContent {
		return &Error{}
	}
	return err
}

func (p *promAPI) Flags() (FlagsResult, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(epFlags, nil), &ares)
	if err != nil {
		return FlagsResult{}, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return FlagsResult{}, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return FlagsResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var res FlagsResult
	err = json.Unmarshal(ares.Data, &res)
	return res, err
}

func (p *promAPI) LabelValues(label string) (model.LabelValues, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(epLabelValues, map[string]string{"name": label}), &ares)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return nil, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return nil, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var labelValues model.LabelValues
	err = json.Unmarshal(ares.Data, &labelValues)
	return labelValues, err
}

func (p *promAPI) Query(query string, ts time.Time) (model.Vector, error) {
	var ares apiResponse
	value := httpclient.NewFormValue()
	value.Set("query", query)
	if !ts.IsZero() {
		value.Set("time", ts.Format(time.RFC3339Nano))
	}
	response, err := httpclient.GetParam(p.client.URL(epQuery, nil), value, &ares)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return nil, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return nil, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var qRes queryResult
	err = json.Unmarshal(ares.Data, &qRes)

	return qRes.v.(model.Vector), err
}

func (p *promAPI) QueryRange(query string, r Range) (model.Value, error) {
	var ares apiResponse
	value := httpclient.NewFormValue()
	var (
		start = r.Start.Format(time.RFC3339Nano)
		end   = r.End.Format(time.RFC3339Nano)
		step  = strconv.FormatFloat(r.Step.Seconds(), 'f', 3, 64)
	)
	value.Set("query", query)
	value.Set("start", start)
	value.Set("end", end)
	value.Set("step", step)
	response, err := httpclient.GetParam(p.client.URL(epQueryRange, nil), value, &ares)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return nil, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return nil, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}

	var qRes queryResult
	err = json.Unmarshal(ares.Data, &qRes)
	return qRes.v, err
}

func (p *promAPI) Series(matches []string, startTime time.Time, endTime time.Time) ([]model.LabelSet, error) {
	var ares apiResponse
	value := httpclient.NewFormValue()
	for _, m := range matches {
		value.Add("match[]", m)
	}
	if !startTime.IsZero() {
		value.Set("cronjob", startTime.Format(time.RFC3339Nano))
	}
	if !endTime.IsZero() {
		value.Set("end", endTime.Format(time.RFC3339Nano))
	}
	response, err := httpclient.GetParam(p.client.URL(epSeries, nil), value, &ares)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return nil, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return nil, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}

	var mSet []model.LabelSet
	err = json.Unmarshal(ares.Data, &mSet)
	return mSet, err
}

func (p *promAPI) Snapshot(skipHead bool) (SnapshotResult, error) {
	var ares apiResponse
	value := httpclient.NewFormValue()
	value.Set("skip_head", strconv.FormatBool(skipHead))
	response, err := httpclient.PostForm(httpclient.EncodeUrl(p.client.URL(epSnapshot, nil), value), nil, nil, &ares)
	if err != nil {
		return SnapshotResult{}, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return SnapshotResult{}, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return SnapshotResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}

	var res SnapshotResult
	err = json.Unmarshal(ares.Data, &res)
	return res, err
}

func (p *promAPI) Targets() (TargetsResult, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(epTargets, nil), &ares)
	if err != nil {
		return TargetsResult{}, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return TargetsResult{}, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return TargetsResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var res TargetsResult
	err = json.Unmarshal(ares.Data, &res)
	return res, err
}

func (p *promAPI) Alerts() (AlertsResult, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(epAlerts, nil), &ares)
	if err != nil {
		return AlertsResult{}, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return AlertsResult{}, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return AlertsResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var res AlertsResult
	err = json.Unmarshal(ares.Data, &res)
	return res, err
}

func (p *promAPI) Rules() (RulesResult, error) {
	var ares apiResponse
	response, err := httpclient.Get(p.client.URL(epRules, nil), &ares)
	if err != nil {
		return RulesResult{}, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return RulesResult{}, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return RulesResult{}, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var res RulesResult
	err = json.Unmarshal(ares.Data, &res)
	return res, err
}

func (p *promAPI) MetaData(match, metric string, limit int) ([]TargetMetaData, error) {
	var ares apiResponse
	value := httpclient.NewFormValue()
	if metric != "" {
		value.Set("metric", metric)
	}
	if match != "" {
		value.Set("match_target", match)
	}
	if limit > 0 {
		value.Set("limit", strconv.Itoa(limit))
	}
	response, err := httpclient.GetParam(p.client.URL(epMetaData, nil), value, &ares)
	if err != nil {
		return nil, err
	}
	if response.Code != http.StatusOK { //prometheus 服务端返回错误响应码
		return nil, &Error{Type: errorTypeByCode(response.Code), Msg: response.Message}
	}
	if ares.Status == PromError {
		return nil, &Error{Type: ares.ErrorType, Msg: ares.Error}
	}
	var res []TargetMetaData
	err = json.Unmarshal(ares.Data, &res)
	return res, err
}

func errorTypeByCode(code int) ErrorType {
	switch code / 100 {
	case 4:
		return ErrClient
	case 5:
		return ErrServer
	}
	return ErrBadResponse
}
