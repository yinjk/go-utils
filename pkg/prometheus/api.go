package prometheus

import (
	"github.com/prometheus/common/model"
	"time"
)

// API provides bindings for Prometheus's v1 API.
type API interface {
	Key(key string) API
	// new promQL handler to hand promQL
	NewPromQLHandler(promQL string) *PromQLHandler
	// prometheus发现的所有AlertManagers的状态
	AlertManagers() (AlertManagersResult, error)
	// 当前prometheus的配置
	Config() (ConfigResult, error)
	// 返回prometheus启动时的flag
	Flags() (FlagsResult, error)
	// 查询指定标签的所有值
	LabelValues(label string) (model.LabelValues, error)
	// 返回prometheus当前发现的所有Targets的状态
	Targets() (TargetsResult, error)
	// 查询指定时间的时间序列值（瞬时值）
	Query(query string, ts time.Time) (model.Vector, error)
	// 根据特定的时间区间查询时序数据 （向量）
	// 参数：
	//   - query: 查询promQL语句，如"up"
	//   - range:
	//       Start: 查询时间范围的开始时间
	//       End：  查询时间范围的结束时间
	//       Step:  查询的粒度，比如每15秒一条数据，每30秒一条数据，粒度分别为15 * time.Second 和 30 * time.Second
	QueryRange(query string, r Range) (model.Value, error)
	// 获取label匹配的所有时间序列
	Series(matches []string, startTime time.Time, endTime time.Time) ([]model.LabelSet, error)
	// 获取所有处于活动状态下的告警
	Alerts() (AlertsResult, error)
	// 获取所有的rules
	Rules() (RulesResult, error)
	// 获取metric标准元数据
	// 参数：
	//   - match:  匹配的标签，如：{job="prometheus"}
	//   - metric: 度量名，如：up、http_requests_total
	//   - limit:  返回的数据最大条数，若传-1表示返回所有
	MetaData(match, metric string, limit int) ([]TargetMetaData, error)

	// Snapshot creates a snapshot of all current data into snapshots/<datetime>-<rand>
	// under the TSDB's data directory and returns the directory as response.
	// 调用以下3个方法会操作TSDB时序数据库，需要在启动prometheus时设置添加--web.enable-admin-api flag，
	// prometheus-operator暂时不支持该设置，作者认为该设置不安全，没有开启的必要，但后续可能开放，关于该设置的讨论，详见 https://github.com/coreos/prometheus-operator/issues/1215
	//Deprecated
	Snapshot(skipHead bool) (SnapshotResult, error)
	// CleanTombstones removes the deleted data from disk and cleans up the existing tombstones.
	//Deprecated
	CleanTombstones() error
	// DeleteSeries deletes data for a selection of series in a time range.
	//Deprecated
	DeleteSeries(matches []string, startTime time.Time, endTime time.Time) error
}
