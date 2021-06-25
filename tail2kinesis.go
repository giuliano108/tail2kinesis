package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/giuliano108/tail2kinesis/lib"
	"github.com/nxadm/tail"
	"github.com/qntfy/frinesis"
	"github.com/qntfy/frinesis/batchproducer"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func makeKinesisClient(args map[string]interface{}) *kinesis.Kinesis {
	var sess *session.Session
	var creds *credentials.Credentials
	sess = session.Must(session.NewSession())
	switch args["--auth"].(string) {
	case "env":
		creds = credentials.NewEnvCredentials()
	case "metadata":
		creds = ec2rolecreds.NewCredentials(sess)
	case "assumerole":
		creds = stscreds.NewCredentials(sess, args["--role-arn"].(string))
	}
	_, err := creds.Get()
	if err != nil {
		log.Fatalf("Error getting credentials: %v", err)
	}

	var region string
	if args["--region"] != nil {
		region = args["--region"].(string)
	} else {
		// ParseArgs ensures this is not empty
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

	return kinesis.New(sess, &aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
}

func makeKinesisProducer(args map[string]interface{}) (batchproducer.Producer, error) {
	var err error
	var ksis *kinesis.Kinesis
	if args["--endpoint"] != nil {
		ksis = frinesis.NewClientWithEndpoint("dummyregion", args["--endpoint"].(string))
	} else {
		ksis = makeKinesisClient(args)
	}

	// frinesis uses zap, which is ever so slightly incompatible with the standard logger :(
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerLevel := zap.NewAtomicLevel()
	loggerLevel.SetLevel(lib.LogLevelsZap[args["--log-level"].(string)])
	loggerConfig.Level = loggerLevel
	loggerConfig.EncoderConfig.TimeKey = "time"
	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}

	config := batchproducer.Config{
		AddBlocksWhenBufferFull: false,
		BufferSize:              10000,
		FlushInterval:           1 * time.Second,
		BatchSize:               100,
		MaxAttemptsPerRecord:    2,
		StatInterval:            1 * time.Second,
		Logger:                  logger,
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
