package main

import (
	"log"
	"sync"
	"time"

	"github.com/jackc/pgx"
)

// MetricType enumerator
type MetricType int

const (
	Gauge MetricType = iota
	Counter
)

func (mt MetricType) String() string {
	return [...]string{"Gauge", "Counter"}[mt]
}

type Metric struct {
	Name      string
	Timestamp time.Time
	Tags      map[string]string
	Value     float64
	Type      MetricType
}

type Telemeter interface {
	Emit(m *Metric)
}

type TelemetrySender interface {
	Telemeter
	Flush() error
}

type Emitter interface {
	Query(*pgx.Conn, Telemeter) error
}

type emitterFunc func(*pgx.Conn, Telemeter) error

func (e emitterFunc) Query(conn *pgx.Conn, tm Telemeter) error {
	return e(conn, tm)
}

func emitAll(conn *pgx.Conn, all []Emitter, tm Telemeter) {
	var wg sync.WaitGroup

	for _, e := range all {
		wg.Add(1)

		go func() {
			defer wg.Done()
			if err := e.Query(conn, tm); err != nil {
				log.Println("Error in emitter", err)
			}
		}()
	}

	wg.Wait()
}
