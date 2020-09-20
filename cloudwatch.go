package main

import (
	"log"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

var batchLen = 20

type cloudwatchCfg struct {
	Namespace string
	Region    string
	AccessKey string
	SecretKey string
}

type cloudwatchClient struct {
	sync.Mutex
	cfg     *cloudwatchCfg
	metrics []*Metric
}

func newCloudwatchClient(cfg *cloudwatchCfg) TelemetrySender {
	client := &cloudwatchClient{}
	client.cfg = cfg
	client.metrics = []*Metric{}
	return client
}

func (c *cloudwatchClient) Emit(m *Metric) {
	c.Lock()
	defer c.Unlock()

	c.metrics = append(c.metrics, m)
}

func (c *cloudwatchClient) Flush() error {

	awsCfg := &aws.Config{Region: aws.String(c.cfg.Region)}

	// If static credentials are not found, rely on other authentication mechanism like default profile or IAM Role.
	if c.cfg.AccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentials(c.cfg.AccessKey, c.cfg.SecretKey, "")
	}

	sess, err := session.NewSession(awsCfg)

	if err != nil {
		return err
	}

	client := cloudwatch.New(sess)

	c.Lock()
	batches := datumBatches(c.metrics)
	c.metrics = c.metrics[:0]
	c.Unlock()

	for _, batch := range batches {
		input := cloudwatch.PutMetricDataInput{}
		input.SetNamespace(c.cfg.Namespace)
		input.SetMetricData(batch)

		// TODO: Try this in highbrow loop
		if _, cerr := client.PutMetricData(&input); cerr != nil {
			err = cerr
			log.Printf("error: %v", cerr)
		}
	}

	return err
}

func datumBatches(in []*Metric) [][]*cloudwatch.MetricDatum {
	var batches [][]*cloudwatch.MetricDatum
	batch := []*cloudwatch.MetricDatum{}

	for _, m := range in {
		if len(batch) == batchLen {
			batches = append(batches, batch)
			batch = []*cloudwatch.MetricDatum{}
		}

		md := cloudwatch.MetricDatum{}
		md.SetMetricName(m.Name)

		var dim []*cloudwatch.Dimension

		for k, v := range m.Tags {
			dim = append(dim, &cloudwatch.Dimension{Name: aws.String(k), Value: aws.String(v)})
		}

		md.SetDimensions(dim)
		md.SetTimestamp(m.Timestamp.UTC())
		md.SetValue(m.Value)

		batch = append(batch, &md)
	}

	if len(batch) > 0 {
		batches = append(batches, batch)
	}

	return batches
}
