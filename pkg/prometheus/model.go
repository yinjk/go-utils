package prometheus

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/model"
	"time"
)

const (
	apiPrefix = "/api/v1"

	EP_ALERT_MANAGERS = apiPrefix + "/alertmanagers"
	epQuery           = apiPrefix + "/query"
	epQueryRange      = apiPrefix + "/query_range"
	epLabelValues     = apiPrefix + "/label/:name/values"
	epSeries          = apiPrefix + "/series"
	epTargets         = apiPrefix + "/targets"
	epMetaData        = apiPrefix + "/targets/metadata"
	epAlerts          = apiPrefix + "/alerts"
	epRules           = apiPrefix + "/rules"
	epSnapshot        = apiPrefix + "/admin/tsdb/snapshot"
	epDeleteSeries    = apiPrefix + "/admin/tsdb/delete_series"
	epCleanTombstones = apiPrefix + "/admin/tsdb/clean_tombstones"
	epConfig          = apiPrefix + "/status/config"
	epFlags           = apiPrefix + "/status/flags"
)

// ErrorType model the different API error types.
type ErrorType string

// HealthStatus model the health status of a scrape target.
type HealthStatus string

const (
	// Possible values for ErrorType.
	ErrBadData     ErrorType = "bad_data"
	ErrTimeout     ErrorType = "timeout"
	ErrCanceled    ErrorType = "canceled"
	ErrExec        ErrorType = "execution"
	ErrBadResponse ErrorType = "bad_response"
	ErrServer      ErrorType = "server_error"
	ErrClient      ErrorType = "client_error"

	// Possible values for HealthStatus.
	HealthGood    HealthStatus = "up"
	HealthUnknown HealthStatus = "unknown"
	HealthBad     HealthStatus = "down"
)

// Error is an error returned by the API.
type Error struct {
	Type   ErrorType
	Msg    string
	Detail string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Msg)
}

// Range represents a sliced time range.
type Range struct {
	// The boundaries of the time range.
	Start, End time.Time
	// The maximum time between two slices within the boundaries.
	Step time.Duration
}

const (
	promSuccess = "success"
	PromError   = "error"
)

type PromResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

// AlertManagersResult contains the result from querying the alertmanagers endpoint.
type AlertManagersResult struct {
	Active  []AlertManager `json:"activeAlertManagers"`
	Dropped []AlertManager `json:"droppedAlertManagers"`
}

// AlertManager model a configured Alert Manager.
type AlertManager struct {
	URL string `json:"url"`
}

// ConfigResult contains the result from querying the client endpoint.
type ConfigResult struct {
	YAML string `json:"yaml"`
}

// FlagsResult contains the result from querying the flag endpoint.
type FlagsResult map[string]string

// SnapshotResult contains the result from querying the snapshot endpoint.
type SnapshotResult struct {
	Name string `json:"name"`
}

// TargetsResult contains the result from querying the targets endpoint.
type TargetsResult struct {
	Active  []ActiveTarget  `json:"activeTargets"`
	Dropped []DroppedTarget `json:"droppedTargets"`
}

// ActiveTarget model an active Prometheus scrape target.
type ActiveTarget struct {
	DiscoveredLabels model.LabelSet `json:"discoveredLabels"`
	Labels           model.LabelSet `json:"labels"`
	ScrapeURL        string         `json:"scrapeUrl"`
	LastError        string         `json:"lastError"`
	LastScrape       time.Time      `json:"lastScrape"`
	Health           HealthStatus   `json:"health"`
}

//-------处于活动状态下的告警--------
type AlertsResult struct {
	Alerts []Alert `json:"alerts"`
}

type AlertState string

const (
	PENDING AlertState = "pending"
	FIRING  AlertState = "firing"
)

type Alert struct {
	State       AlertState        `json:"state"`       //pending|firing
	ActiveAt    time.Time         `json:"activeAt"`    //告警触发时间
	Value       float64           `json:"value"`       //告警值
	Labels      map[string]string `json:"labels"`      //告警label
	Annotations map[string]string `json:"annotations"` //告警的附加说明
}

//-------rules--------
type RulesResult struct {
	Groups []RuleGroup `json:"groups"`
}

type RuleGroup struct {
	Name     string `json:"name"`
	File     string `json:"file"`
	Interval int    `json:"interval"`
	Rules    []Rule `json:"rules"`
}

type Rule struct {
	Name        string            `json:"name"`   //规则名
	Query       string            `json:"query"`  //promQL
	Health      string            `json:"health"` //ok
	Type        RuleType          `json:"type"`   //规则类型:[recording|alerting]，当是recording类型时，没有下面三个字段
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Alerts      []Alert           `json:"alerts"` //该规则触发的告警
}

type RuleType string

const (
	ALERTING  RuleType = "alerting"
	RECORDING RuleType = "recording"
)

//-------metadata--------
type TargetMetaData struct {
	Metric string            `json:"metric"`
	Type   MetricType        `json:"type"`
	Help   string            `json:"help"`
	Target map[string]string `json:"target"`
}

type MetricType string

const (
	MetricType_COUNTER   MetricType = "counter"
	MetricType_GAUGE     MetricType = "gauge"
	MetricType_SUMMARY   MetricType = "summary"
	MetricType_UNTYPED   MetricType = "untyped"
	MetricType_HISTOGRAM MetricType = "histogram"
)

// DroppedTarget model a dropped Prometheus scrape target.
type DroppedTarget struct {
	DiscoveredLabels model.LabelSet `json:"discoveredLabels"`
}

// queryResult contains result data for a query.
type queryResult struct {
	Type   model.ValueType `json:"resultType"`
	Result interface{}     `json:"result"`

	// The decoded value.
	v model.Value
}

func (qr *queryResult) UnmarshalJSON(b []byte) error {
	v := struct {
		Type   model.ValueType `json:"resultType"`
		Result json.RawMessage `json:"result"`
	}{}

	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	switch v.Type {
	case model.ValScalar:
		var sv model.Scalar
		err = json.Unmarshal(v.Result, &sv)
		qr.v = &sv

	case model.ValVector:
		var vv model.Vector
		err = json.Unmarshal(v.Result, &vv)
		qr.v = vv

	case model.ValMatrix:
		var mv model.Matrix
		err = json.Unmarshal(v.Result, &mv)
		qr.v = mv

	default:
		err = fmt.Errorf("unexpected value type %q", v.Type)
	}
	return err
}
