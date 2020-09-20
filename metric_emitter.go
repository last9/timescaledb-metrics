package main

import (
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx"
)

// MetricType enumerator
type MetricType int

// Gauge and Counter are the only two supported metric types
const (
	Gauge MetricType = iota
	Counter
)

func (mt MetricType) String() string {
	return [...]string{"Gauge", "Counter"}[mt]
}

// Metric represents a cross-sink compatible struct
// It has a fixed timestamp that should be set.
type Metric struct {
	Name      string
	Timestamp time.Time
	Tags      map[string]string
	Value     float64
	Type      MetricType
}

// Telemeter is anything that can accept a Metric.
type Telemeter interface {
	Emit(m *Metric)
}

// TelemetrySender is Telemeter + Flush capability.
// Using a separate flush method allows caching and batching to keep
// downstream costs and pressure in check, at some risk of loss in case
// of a crash in this program.
type TelemetrySender interface {
	Telemeter
	Flush() error
}

// Emitter can technically be anything that accepts a PG conn.
// Primary motivation of making this an interface was to allow testing
type Emitter func(*pgx.Conn, Telemeter) error

func emitAll(conn *pgx.Conn, all []Emitter, tm Telemeter) {
	var wg sync.WaitGroup

	for _, e := range all {
		wg.Add(1)

		go func(ef Emitter) {
			defer wg.Done()
			if err := ef(conn, tm); err != nil {
				log.Println("Error in emitter", err)
			}
		}(e)
	}

	wg.Wait()
}
