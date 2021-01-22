/*
 * Copyright 2019 Dgraph Labs, Inc. and Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package alpha

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMetricTxnCommits(t *testing.T) {
	name := "dgraph_txn_commits_total"
	mt := `
    {
	  set {
		<0x71>  <name> "Bob" .
	  }
	}
	`

	// Create initial 'dgraph_txn_aborts_total' metric
	mr1, err := mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	mr2, err := mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	require.NoError(t, commitWithTs(mr1.keys, mr1.preds, mr1.startTs))
	require.Error(t, commitWithTs(mr2.keys, mr2.preds, mr2.startTs))

	// Fetch Metrics
	txnMetric := fetchMetric(t, name)

	// Create second 'dgraph_txn_aborts_total' metric
	mr1, err = mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	mr2, err = mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	require.NoError(t, commitWithTs(mr1.keys, mr1.preds, mr1.startTs))
	require.Error(t, commitWithTs(mr2.keys, mr2.preds, mr2.startTs))

	// Fetch and check updated metrics
	require.NoError(t, retryableFetchMetrics(t, name, txnMetric+1))
}

/*
func TestMetricTxnDiscards(t *testing.T) {
	name := "dgraph_txn_discards_total"
	mt := `
    {
	  set {
		<0x71>  <name> "Bob" .
	  }
	}
	`

	// Create initial 'dgraph_txn_aborts_total' metric
	mr1, err := mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	mr2, err := mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	require.NoError(t, commitWithTs(mr1.keys, mr1.preds, mr1.startTs))
	require.Error(t, commitWithTs(mr2.keys, mr2.preds, mr2.startTs))

	// Fetch Metrics
	txnMetric := fetchMetric(t, name)

	// Create second 'dgraph_txn_aborts_total' metric
	mr1, err = mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	mr2, err = mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	require.NoError(t, commitWithTs(mr1.keys, mr1.preds, mr1.startTs))
	require.Error(t, commitWithTs(mr2.keys, mr2.preds, mr2.startTs))

	// Fetch and check updated metrics
	require.NoError(t, retryableFetchMetrics(t, name, txnMetric+1))
}
*/

func TestMetricTxnAborts(t *testing.T) {
	name := "dgraph_txn_aborts_total"
	mt := `
    {
	  set {
		<0x71>  <name> "Bob" .
	  }
	}
	`

	// Create initial 'dgraph_txn_aborts_total' metric
	mr1, err := mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	mr2, err := mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	require.NoError(t, commitWithTs(mr1.keys, mr1.preds, mr1.startTs))
	require.Error(t, commitWithTs(mr2.keys, mr2.preds, mr2.startTs))

	// Fetch Metrics
	txnMetric := fetchMetric(t, name)

	// Create second 'dgraph_txn_aborts_total' metric
	mr1, err = mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	mr2, err = mutationWithTs(mt, "application/rdf", false, false, 0)
	require.NoError(t, err)
	require.NoError(t, commitWithTs(mr1.keys, mr1.preds, mr1.startTs))
	require.Error(t, commitWithTs(mr2.keys, mr2.preds, mr2.startTs))

	// Fetch and check updated metrics
	require.NoError(t, retryableFetchMetrics(t, name, txnMetric+1))
}

func retryableFetchMetrics(t *testing.T, name string, expected int) error {
	var txnMetric int
	for i := 0; i < 10; i++ {
		txnMetric = fetchMetric(t, name)
		if expected == txnMetric {
			return nil
		}
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("metric '%s' was not incremented. wanted %d, got %d",
		name, expected, txnMetric)
}

func fetchMetric(t *testing.T, name string) int {
	req, err := http.NewRequest("GET", addr+"/debug/prometheus_metrics", nil)
	require.NoError(t, err)
	_, body, _, err := runRequest(req)
	require.NoError(t, err)
	metricsMap, err := extractMetrics(string(body))
	require.NoError(t, err)

	txnMetric, ok := metricsMap[name]
	require.True(t, ok, "the required metric '%s' is not found", name)
	m, _ := strconv.Atoi(txnMetric.(string))
	return m
}

func TestMetrics(t *testing.T) {
	req, err := http.NewRequest("GET", addr+"/debug/prometheus_metrics", nil)
	require.NoError(t, err)

	_, body, _, err := runRequest(req)
	require.NoError(t, err)
	metricsMap, err := extractMetrics(string(body))
	require.NoError(t, err, "Unable to get the metrics map: %v", err)

	requiredMetrics := []string{
		// Go Runtime Metrics
		"go_goroutines", "go_memstats_gc_cpu_fraction", "go_memstats_heap_alloc_bytes",
		"go_memstats_heap_idle_bytes", "go_memstats_heap_inuse_bytes", "dgraph_latency_bucket",

		// Badger Metrics
		"badger_v3_disk_reads_total", "badger_v3_disk_writes_total", "badger_v3_gets_total",
		"badger_v3_memtable_gets_total", "badger_v3_puts_total", "badger_v3_read_bytes",
		"badger_v3_written_bytes",

		// Transaction Metrics
		"dgraph_txn_aborts_total", "dgraph_txn_commits_total", "dgraph_txn_discards_total",

		// Dgraph Memory Metrics
		"dgraph_memory_idle_bytes", "dgraph_memory_inuse_bytes", "dgraph_memory_proc_bytes",
		"dgraph_memory_alloc_bytes",
		// Dgraph Activity Metrics
		"dgraph_active_mutations_total", "dgraph_pending_proposals_total",
		"dgraph_pending_queries_total",
		"dgraph_num_queries_total", "dgraph_alpha_health_status",
	}
	for _, requiredM := range requiredMetrics {
		_, ok := metricsMap[requiredM]
		require.True(t, ok, "the required metric %s is not found", requiredM)
	}
}

func extractMetrics(metrics string) (map[string]interface{}, error) {
	lines := strings.Split(metrics, "\n")
	metricRegex, err := regexp.Compile("(^\\w+|\\d+$)")

	if err != nil {
		return nil, err
	}
	metricsMap := make(map[string]interface{})
	for _, line := range lines {
		matches := metricRegex.FindAllString(line, -1)
		if len(matches) > 0 {
			metricsMap[matches[0]] = matches[1]
		}
	}
	return metricsMap, nil
}
