package main

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx"
)

type mockTelemetry struct {
	cloudwatchClient
	flushed []*Metric
}

func (c *mockTelemetry) Flush() error {
	c.Lock()
	defer c.Unlock()

	c.flushed = append(c.flushed, c.metrics...)
	return nil
}

func (c *mockTelemetry) FlushedMetrics() []*Metric {
	c.Lock()
	defer c.Unlock()
	return c.flushed
}

func mockPolicyEmitter(conn *pgx.Conn, tm Telemeter) error {
	f, err := os.Open("testdata/records.log")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	var rec record
	for scanner.Scan() {
		t := scanner.Text()
		if err := json.Unmarshal([]byte(t), &rec); err != nil {
			break
		}

		recordToMetrics(&rec, recordTags(&rec), tm)
	}

	return nil
}

func TestEmitter(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	registerEmitter(mockPolicyEmitter)

	mt := &mockTelemetry{}
	process(nil, mt)

	t.Run("Length should be the same", func(t *testing.T) {
		if len(mt.FlushedMetrics()) != 27 {
			t.Fatal("Length of metrics emitted does not match")
		}
	})
}
