package main

import (
	"fmt"
	"github.com/giuliano108/tail2kinesis/lib"
	"github.com/hpcloud/tail"
	"github.com/sendgridlabs/go-kinesis"
	"github.com/sendgridlabs/go-kinesis/batchproducer"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func makeKinesisProducer(args map[string]interface{}) (batchproducer.Producer, error) {
	var err error
	var ksis *kinesis.Kinesis
	var auth kinesis.Auth
	if args["--auth"].(string) == "env" {
		auth, err = kinesis.NewAuthFromEnv()
	} else {
		auth, err = kinesis.NewAuthFromMetadata()
	}
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve authentication credentials from the environment: %v", err)
	}
	if args["--endpoint"] != nil {
		ksis = kinesis.NewWithEndpoint(auth, "dummyregion", args["--endpoint"].(string))
	} else {
		var region string
		if args["--region"] != nil {
			region = args["--region"].(string)
		} else {
			// ParseArgs ensures this is not empty
			region = os.Getenv("AWS_DEFAULT_REGION")
		}
		ksis = kinesis.New(auth, region)
	}

	config := batchproducer.Config{
		AddBlocksWhenBufferFull: false,
		BufferSize:              10000,
		FlushInterval:           1 * time.Second,
		BatchSize:               100,
		MaxAttemptsPerRecord:    2,
		StatInterval:            1 * time.Second,
		Logger:                  log.StandardLogger(),
	}
	producer, err := batchproducer.New(ksis, args["--stream-name"].(string), config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)

	var err error
	args := lib.ParseArgs(os.Args[1:])
	log.SetLevel(lib.LogLevels[args["--log-level"].(string)])
	fileName := args["<filename>"].(string)
	var xform lib.Transform
	switch args["--transform"].(string) {
	case "accesslog-query":
		xform, err = lib.NewAccessLog2JSON("query")
	default:
		xform, err = lib.NewIdentity()
	}
	if err != nil {
		log.Fatal(err)
	}

	producer, err := makeKinesisProducer(args)
	if err != nil {
		log.Fatal(err)
	}
	producer.Start()

	tailConfig := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: false,
		Location:  &tail.SeekInfo{0, os.SEEK_END},
		Logger:    log.StandardLogger(),
	}
	t, err := tail.TailFile(fileName, tailConfig)
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT)
	go func() {
		sig := <-sigs
		log.WithFields(log.Fields{"signal": sig}).Info("Signal received")
		done <- true
	}()

	go func() {
		for line := range t.Lines {
			transformed, err := xform.Do(line.Text)
			if err == nil {
				err = producer.Add([]byte(transformed), "pk")
				if err != nil {
					log.Errorf("Cannot add record to the Kinesis producer: %v", err)
				}
			} else {
				log.Warnf("Could not transform line: %v", err)
				log.Debugf("Could not transform line: %v\nLine that failed: %s", err, line.Text)
			}
		}
		err = t.Wait()
		if err != nil {
			log.Errorf("Error while tailing file: %v", err)
		}
		done <- true
	}()

	<-done
	sent, remaining, err := producer.Flush(3*time.Second, false)
	if err == nil {
		log.Infof("Producer stopped: %d sent, %d remaining", sent, remaining)
	} else {
		log.Errorf("Error stopping producer: %v, %d sent, %d remaining", err, sent, remaining)
	}
}
