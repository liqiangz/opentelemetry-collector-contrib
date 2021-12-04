// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package processesscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper"

import (
	"context"
	"time"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/process"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/model/pdata"
	"go.opentelemetry.io/collector/receiver/scrapererror"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/processesscraper/internal/metadata"
)

var metricsLength = func() int {
	n := 0
	if enableProcessesCount {
		n++
	}
	if enableProcessesCreated {
		n++
	}
	return n
}()

// scraper for Processes Metrics
type scraper struct {
	config    *Config
	startTime pdata.Timestamp

	// for mocking gopsutil
	getMiscStats func() (*load.MiscStat, error)
	getProcesses func() ([]proc, error)
}

// for mocking out gopsutil process.Process
type proc interface {
	Status() ([]string, error)
}

type processesMetadata struct {
	countByStatus    map[string]int64 // ignored if enableProcessesCount is false
	processesCreated *int64           // ignored if enableProcessesCreated is false
}

// newProcessesScraper creates a set of Processes related metrics
func newProcessesScraper(_ context.Context, cfg *Config) *scraper {
	return &scraper{
		config:       cfg,
		getMiscStats: load.Misc,
		getProcesses: func() ([]proc, error) {
			ps, err := process.Processes()
			ret := make([]proc, len(ps))
			for i := range ps {
				ret[i] = ps[i]
			}
			return ret, err
		},
	}
}

func (s *scraper) start(context.Context, component.Host) error {
	bootTime, err := host.BootTime()
	if err != nil {
		return err
	}
	// bootTime is seconds since 1970, timestamps are in nanoseconds.
	s.startTime = pdata.Timestamp(bootTime * 1e9)
	return nil
}

func (s *scraper) scrape(_ context.Context) (pdata.Metrics, error) {
	now := pdata.NewTimestampFromTime(time.Now())

	md := pdata.NewMetrics()
	metrics := md.ResourceMetrics().AppendEmpty().InstrumentationLibraryMetrics().AppendEmpty().Metrics()
	metrics.EnsureCapacity(metricsLength)

	processMetadata, err := s.getProcessesMetadata()
	if err != nil {
		return md, scrapererror.NewPartialScrapeError(err, metricsLength)
	}

	if enableProcessesCount && processMetadata.countByStatus != nil {
		setProcessesCountMetric(metrics.AppendEmpty(), s.startTime, now, processMetadata.countByStatus)
	}

	if enableProcessesCreated && processMetadata.processesCreated != nil {
		setProcessesCreatedMetric(metrics.AppendEmpty(), s.startTime, now, *processMetadata.processesCreated)
	}

	return md, err
}

func setProcessesCountMetric(metric pdata.Metric, startTime, now pdata.Timestamp, countByStatus map[string]int64) {
	metadata.Metrics.SystemProcessesCount.Init(metric)
	ddps := metric.Sum().DataPoints()

	for status, count := range countByStatus {
		setProcessesCountDataPoint(ddps.AppendEmpty(), startTime, now, status, count)
	}
}

func setProcessesCountDataPoint(dataPoint pdata.NumberDataPoint, startTime, now pdata.Timestamp, statusLabel string, value int64) {
	dataPoint.Attributes().InsertString(metadata.Attributes.Status, statusLabel)
	dataPoint.SetStartTimestamp(startTime)
	dataPoint.SetTimestamp(now)
	dataPoint.SetIntVal(value)
}

func setProcessesCreatedMetric(metric pdata.Metric, startTime, now pdata.Timestamp, value int64) {
	metadata.Metrics.SystemProcessesCreated.Init(metric)
	ddp := metric.Sum().DataPoints().AppendEmpty()
	ddp.SetStartTimestamp(startTime)
	ddp.SetTimestamp(now)
	ddp.SetIntVal(value)
}
