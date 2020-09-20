package main

import (
	"log"
	"strings"
	"testing"
)

func TestCloudWatch(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	c := newCloudwatchClient(&cloudwatchCfg{
		Namespace: "ok",
		Region:    "test",
		AccessKey: "hello",
	})

	mockPolicyEmitter(nil, c)
	m := c.(*cloudwatchClient).metrics
	t.Run("Length should be the same", func(t *testing.T) {
		if len(m) != 27 {
			t.Fatal("Length of metrics emitted does not match")
		}
	})

	t.Run("Batch length should match", func(t *testing.T) {
		batchLen = 10
		b := datumBatches(m)
		if len(b) != 3 {
			t.Fatal("Length of metric batch does not match")
		}
	})

	err := c.Flush()
	if err == nil || !strings.Contains(err.Error(), "static credentials are empty") {
		t.Fatal("Dial should fail")
	}
}
