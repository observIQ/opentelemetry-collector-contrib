// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package varnishreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/varnishreceiver"

import (
	"go.opentelemetry.io/collector/model/pdata"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/varnishreceiver/internal/metadata"
)

// Stats contain an autogenerated json to struct fields.
type Stats struct {
	Version   int    `json:"version"`
	Timestamp string `json:"timestamp"`
	Counters  struct {
		MAINBackendConn struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_conn"`
		MAINBackendUnhealthy struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_unhealthy"`
		MAINBackendBusy struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_busy"`
		MAINBackendFail struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_fail"`
		MAINBackendReuse struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_reuse"`
		MAINBackendRecycle struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_recycle"`
		MAINBackendRetry struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_retry"`
		MAINCacheHit struct {
			Value int64 `json:"value"`
		} `json:"MAIN.cache_hit"`
		MAINCacheHitpass struct {
			Value int64 `json:"value"`
		} `json:"MAIN.cache_hitpass"`
		MAINCacheMiss struct {
			Value int64 `json:"value"`
		} `json:"MAIN.cache_miss"`
		MAINThreadsCreated struct {
			Value int64 `json:"value"`
		} `json:"MAIN.threads_created"`
		MAINThreadsDestroyed struct {
			Value int64 `json:"value"`
		} `json:"MAIN.threads_destroyed"`
		MAINThreadsFailed struct {
			Value int64 `json:"value"`
		} `json:"MAIN.threads_failed"`
		MAINSessConn struct {
			Value int64 `json:"value"`
		} `json:"MAIN.sess_conn"`
		MAINSessFail struct {
			Value int64 `json:"value"`
		} `json:"MAIN.sess_fail"`
		MAINSessDropped struct {
			Value int64 `json:"value"`
		} `json:"MAIN.sess_dropped"`
		MAINNObject struct {
			Value int64 `json:"value"`
		} `json:"MAIN.n_object"`
		MAINNExpired struct {
			Value int64 `json:"value"`
		} `json:"MAIN.n_expired"`
		MAINNLruNuked struct {
			Value int64 `json:"value"`
		} `json:"MAIN.n_lru_nuked"`
		MAINNLruMoved struct {
			Value int64 `json:"value"`
		} `json:"MAIN.n_lru_moved"`
		MAINClientReq struct {
			Value int64 `json:"value"`
		} `json:"MAIN.client_req"`
		MAINBackendReq struct {
			Value int64 `json:"value"`
		} `json:"MAIN.backend_req"`
	} `json:"counters"`
}

// Stats6_4 contain an autogenerated json to struct fields.
type Stats6_4 struct {
	MAINBackendConn struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_conn"`
	MAINBackendUnhealthy struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_unhealthy"`
	MAINBackendBusy struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_busy"`
	MAINBackendFail struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_fail"`
	MAINBackendReuse struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_reuse"`
	MAINBackendRecycle struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_recycle"`
	MAINBackendRetry struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_retry"`
	MAINCacheHit struct {
		Value int64 `json:"value"`
	} `json:"MAIN.cache_hit"`
	MAINCacheHitpass struct {
		Value int64 `json:"value"`
	} `json:"MAIN.cache_hitpass"`
	MAINCacheMiss struct {
		Value int64 `json:"value"`
	} `json:"MAIN.cache_miss"`
	MAINThreadsCreated struct {
		Value int64 `json:"value"`
	} `json:"MAIN.threads_created"`
	MAINThreadsDestroyed struct {
		Value int64 `json:"value"`
	} `json:"MAIN.threads_destroyed"`
	MAINThreadsFailed struct {
		Value int64 `json:"value"`
	} `json:"MAIN.threads_failed"`
	MAINSessConn struct {
		Value int64 `json:"value"`
	} `json:"MAIN.sess_conn"`
	MAINSessFail struct {
		Value int64 `json:"value"`
	} `json:"MAIN.sess_fail"`
	MAINSessDropped struct {
		Value int64 `json:"value"`
	} `json:"MAIN.sess_dropped"`
	MAINNObject struct {
		Value int64 `json:"value"`
	} `json:"MAIN.n_object"`
	MAINNExpired struct {
		Value int64 `json:"value"`
	} `json:"MAIN.n_expired"`
	MAINNLruNuked struct {
		Value int64 `json:"value"`
	} `json:"MAIN.n_lru_nuked"`
	MAINNLruMoved struct {
		Value int64 `json:"value"`
	} `json:"MAIN.n_lru_moved"`
	MAINClientReq struct {
		Value int64 `json:"value"`
	} `json:"MAIN.client_req"`
	MAINBackendReq struct {
		Value int64 `json:"value"`
	} `json:"MAIN.backend_req"`
}

//  update convers Stats6_4 into a Stats struct, the varnish 6.5+ supported stats object.
func (s *Stats6_4) update() *Stats {
	return &Stats{
		Counters: struct {
			MAINBackendConn struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_conn\""
			MAINBackendUnhealthy struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_unhealthy\""
			MAINBackendBusy struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_busy\""
			MAINBackendFail struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_fail\""
			MAINBackendReuse struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_reuse\""
			MAINBackendRecycle struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_recycle\""
			MAINBackendRetry struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_retry\""
			MAINCacheHit struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.cache_hit\""
			MAINCacheHitpass struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.cache_hitpass\""
			MAINCacheMiss struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.cache_miss\""
			MAINThreadsCreated struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.threads_created\""
			MAINThreadsDestroyed struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.threads_destroyed\""
			MAINThreadsFailed struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.threads_failed\""
			MAINSessConn struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.sess_conn\""
			MAINSessFail struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.sess_fail\""
			MAINSessDropped struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.sess_dropped\""
			MAINNObject struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.n_object\""
			MAINNExpired struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.n_expired\""
			MAINNLruNuked struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.n_lru_nuked\""
			MAINNLruMoved struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.n_lru_moved\""
			MAINClientReq struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.client_req\""
			MAINBackendReq struct {
				Value int64 "json:\"value\""
			} "json:\"MAIN.backend_req\""
		}{
			MAINBackendConn:      s.MAINBackendConn,
			MAINBackendUnhealthy: s.MAINBackendUnhealthy,
			MAINBackendBusy:      s.MAINBackendBusy,
			MAINBackendFail:      s.MAINBackendFail,
			MAINBackendReuse:     s.MAINBackendReuse,
			MAINBackendRecycle:   s.MAINBackendRecycle,
			MAINBackendRetry:     s.MAINBackendRetry,
			MAINCacheHit:         s.MAINCacheHit,
			MAINCacheHitpass:     s.MAINCacheHitpass,
			MAINCacheMiss:        s.MAINCacheMiss,
			MAINThreadsCreated:   s.MAINThreadsCreated,
			MAINThreadsDestroyed: s.MAINThreadsDestroyed,
			MAINThreadsFailed:    s.MAINThreadsFailed,
			MAINSessConn:         s.MAINSessConn,
			MAINSessFail:         s.MAINSessFail,
			MAINSessDropped:      s.MAINSessDropped,
			MAINNObject:          s.MAINNObject,
			MAINNExpired:         s.MAINNExpired,
			MAINNLruNuked:        s.MAINNLruNuked,
			MAINNLruMoved:        s.MAINNLruMoved,
			MAINClientReq:        s.MAINClientReq,
			MAINBackendReq:       s.MAINBackendReq,
		},
	}
}

func (v *varnishScraper) recordVarnishBackendConnectionsCountDataPoint(now pdata.Timestamp, stats *Stats) {
	attributeMappings := map[string]int64{
		metadata.AttributeBackendConnectionType.Success:   stats.Counters.MAINBackendConn.Value,
		metadata.AttributeBackendConnectionType.Recycle:   stats.Counters.MAINBackendRecycle.Value,
		metadata.AttributeBackendConnectionType.Reuse:     stats.Counters.MAINBackendReuse.Value,
		metadata.AttributeBackendConnectionType.Fail:      stats.Counters.MAINBackendFail.Value,
		metadata.AttributeBackendConnectionType.Unhealthy: stats.Counters.MAINBackendUnhealthy.Value,
		metadata.AttributeBackendConnectionType.Busy:      stats.Counters.MAINBackendBusy.Value,
		metadata.AttributeBackendConnectionType.Retry:     stats.Counters.MAINBackendRetry.Value,
	}

	for attributeName, attributeValue := range attributeMappings {
		v.mb.RecordVarnishBackendConnectionsCountDataPoint(now, attributeValue, attributeName)
	}
}

func (v *varnishScraper) recordVarnishCacheOperationsCountDataPoint(now pdata.Timestamp, stats *Stats) {
	attributeMappings := map[string]int64{
		metadata.AttributeCacheOperations.Hit:     stats.Counters.MAINCacheHit.Value,
		metadata.AttributeCacheOperations.HitPass: stats.Counters.MAINCacheHitpass.Value,
		metadata.AttributeCacheOperations.Miss:    stats.Counters.MAINCacheMiss.Value,
	}

	for attributeName, attributeValue := range attributeMappings {
		v.mb.RecordVarnishCacheOperationsCountDataPoint(now, attributeValue, attributeName)
	}
}

func (v *varnishScraper) recordVarnishThreadOperationsCountDataPoint(now pdata.Timestamp, stats *Stats) {
	attributeMappings := map[string]int64{
		metadata.AttributeThreadOperations.Created:   stats.Counters.MAINThreadsCreated.Value,
		metadata.AttributeThreadOperations.Destroyed: stats.Counters.MAINThreadsDestroyed.Value,
		metadata.AttributeThreadOperations.Failed:    stats.Counters.MAINThreadsFailed.Value,
	}

	for attributeName, attributeValue := range attributeMappings {
		v.mb.RecordVarnishThreadOperationsCountDataPoint(now, attributeValue, attributeName)
	}
}

func (v *varnishScraper) recordVarnishSessionCountDataPoint(now pdata.Timestamp, stats *Stats) {
	attributeMappings := map[string]int64{
		metadata.AttributeSessionType.Accepted: stats.Counters.MAINSessConn.Value,
		metadata.AttributeSessionType.Dropped:  stats.Counters.MAINSessDropped.Value,
		metadata.AttributeSessionType.Failed:   stats.Counters.MAINSessFail.Value,
	}

	for attributeName, attributeValue := range attributeMappings {
		v.mb.RecordVarnishSessionCountDataPoint(now, attributeValue, attributeName)
	}
}
