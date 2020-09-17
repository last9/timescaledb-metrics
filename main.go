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

const VERSION = "0.0.2"

var Config = struct {
	DB struct {
		Url string `required:"true" env:"DATABASE_URL"`
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

	conn, err := dbConn(Config.DB.Url)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	defer conn.Close()

	cw := NewCloudwatchClient(&CloudwatchCfg{
		Namespace: Config.Cloudwatch.Namespace,
		Region:    Config.Cloudwatch.Region,
		AccessKey: Config.Cloudwatch.AccessKey,
		SecretKey: Config.Cloudwatch.SecretKey,
	})

	registerEmitter(emitterFunc(policyMetrics))
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
