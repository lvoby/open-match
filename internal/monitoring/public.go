// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitoring

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"open-match.dev/open-match/internal/config"
)

var (
	publicLogger = logrus.WithFields(logrus.Fields{
		"app":       "openmatch",
		"component": "monitoring",
	})
)

// Setup configures the monitoring for the server.
func Setup(mux *http.ServeMux, cfg config.View) {
	periodString := cfg.GetString("monitoring.reportingPeriod")
	reportingPeriod, err := time.ParseDuration(periodString)
	if err != nil {
		publicLogger.WithFields(logrus.Fields{
			"error":           err,
			"reportingPeriod": periodString,
		}).Info("Failed to parse monitoring.reportingPeriod, defaulting to 10s")
		reportingPeriod = time.Second * 10
	}

	bindJaeger(cfg)
	bindPrometheus(mux, cfg)
	bindStackDriver(cfg)
	bindZipkin(cfg)
	bindZpages(mux, cfg)

	// Change the frequency of updates to the metrics endpoint
	view.SetReportingPeriod(reportingPeriod)

	publicLogger.WithFields(logrus.Fields{
		"reportingPeriod": reportingPeriod,
	}).Info("Monitoring has been configured.")
}

// Counter creates a counter metric.
func Counter(name string, description string) *stats.Int64Measure {
	s := stats.Int64(name, "Count of "+description+".", "1")
	counterView(s)
	return s
}

// IncrementCounter +1's the counter metric.
func IncrementCounter(ctx context.Context, s *stats.Int64Measure) {
	IncrementCounterN(ctx, s, 1)
}

// IncrementCounterN increases a metric by n.
func IncrementCounterN(ctx context.Context, s *stats.Int64Measure, n int) {
	stats.Record(ctx, s.M(int64(n)))
}

// CounterView converts the measurement into a view for a counter.
func counterView(s *stats.Int64Measure) *view.View {
	v := &view.View{
		Name:        s.Name(),
		Measure:     s,
		Description: s.Description(),
		Aggregation: view.Count(),
	}
	err := view.Register(v)
	if err != nil {
		publicLogger.WithError(err).Infof("cannot register view for metric: %s, it will not be reported", s.Name())
	}
	return v
}
