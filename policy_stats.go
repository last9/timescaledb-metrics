package main

import (
	"log"
	"time"

	"github.com/jackc/pgx"
)

const query = `SELECT
COALESCE(CA.view_name, CAS.view_name, PS.hypertable) AS entity_name,
job_type,
EXTRACT(EPOCH FROM PS.last_start)::int as last_start_on,
EXTRACT(EPOCH FROM PS.last_successful_finish)::int as last_success_on,
PS.total_failures
FROM timescaledb_information.policy_stats PS
LEFT JOIN timescaledb_information.continuous_aggregates CA
ON PS.hypertable = CA.materialization_hypertable
LEFT JOIN timescaledb_information.continuous_aggregate_stats CAS
ON PS.job_id = CAS.job_id;`

type record struct {
	Entity      string `json:"entity"`
	JobType     string `json:"job_type"`
	LastStart   int    `json:"last_start"`
	LastSuccess int    `json:"last_success"`
	TotalFail   int64  `json:"total_fail"`
}

func recordTags(r *record) map[string]string {
	return map[string]string{
		"entity":   r.Entity,
		"job_type": r.JobType,
	}
}

const (
	lastSuccessMetric = "last_success_on"
	lastStartMetric   = "last_start_on"
	totalFailMetric   = "total_fail"
)

func policyMetrics(conn *pgx.Conn, tm Telemeter) error {
	rows, err := conn.Query(query)
	if err != nil {
		return err
	}

	defer rows.Close()

	var rec record

	for rows.Next() {
		if err := rows.Scan(
			&rec.Entity, &rec.JobType, &rec.LastStart, &rec.LastSuccess, &rec.TotalFail,
		); err != nil {
			log.Println(err)
			break
		}

		recordToMetrics(&rec, recordTags(&rec), tm)
	}

	return nil
}

func recordToMetrics(rec *record, tags map[string]string, tm Telemeter) {
	ts := time.Now()

	tm.Emit(&Metric{
		Name:      lastStartMetric,
		Type:      Gauge,
		Timestamp: ts,
		Tags:      tags,
		Value:     float64(rec.LastStart),
	})

	tm.Emit(&Metric{
		Name:      lastSuccessMetric,
		Type:      Gauge,
		Timestamp: ts,
		Tags:      tags,
		Value:     float64(rec.LastSuccess),
	})

	tm.Emit(&Metric{
		Name:      totalFailMetric,
		Type:      Counter,
		Timestamp: ts,
		Tags:      tags,
		Value:     float64(rec.TotalFail),
	})
}
