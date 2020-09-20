package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/jackc/pgx"
	"github.com/jinzhu/configor"
)

// VERSION no. of the binary. To be set using ldflags during build time.
var VERSION = "0.0.3"

// Config is set using an external YAML, JSON, Toml file
var Config = struct {
	DB struct {
		URL string `required:"true" env:"DATABASE_URL"`
	}

	Cloudwatch struct {
		Namespace string `required:"true" env:"CLOUDWATCH_NAMESPACE"`
		Region    string `required:"true" env:"AWS_DEFAULT_REGION"`
		AccessKey string `env:"AWS_ACCESS_KEY_ID"`
		SecretKey string `env:"AWS_SECRET_ACCESS_KEY"`
	}
}{}

var (
	config  = flag.String("config", "", "config file")
	version = flag.Bool("v", false, "prints current roxy version")
)

func printVersion(v string) {
	fmt.Println("Version: ", v)
	os.Exit(0)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()

	if *version {
		printVersion(VERSION)
	}

	if *config == "" {
		configor.Load(&Config)
	} else {
		configor.Load(&Config, *config)
	}

	conn, err := dbConn(Config.DB.URL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	defer conn.Close()

	cw := newCloudwatchClient(&cloudwatchCfg{
		Namespace: Config.Cloudwatch.Namespace,
		Region:    Config.Cloudwatch.Region,
		AccessKey: Config.Cloudwatch.AccessKey,
		SecretKey: Config.Cloudwatch.SecretKey,
	})

	registerEmitter(policyMetrics)
	process(conn, cw)
}

var emitters []Emitter
var eMux sync.RWMutex

func registerEmitter(e Emitter) {
	eMux.Lock()
	defer eMux.Unlock()

	emitters = append(emitters, e)
}

func process(conn *pgx.Conn, cw TelemetrySender) {
	eMux.RLock()
	defer eMux.RUnlock()

	emitAll(conn, emitters, cw)
	cw.Flush()
}
