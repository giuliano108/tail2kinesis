package lib

import (
	"fmt"
	"github.com/docopt/docopt-go"
	log "github.com/sirupsen/logrus"
	"go.uber.org/zap/zapcore"
	"os"
	"strings"
)

var LogLevels map[string]log.Level = map[string]log.Level{
	"trace": log.TraceLevel,
	"debug": log.DebugLevel,
	"info":  log.InfoLevel,
	"warn":  log.WarnLevel,
	"error": log.ErrorLevel,
	"fatal": log.FatalLevel,
	"panic": log.PanicLevel,
}

// This sucks
var LogLevelsZap map[string]zapcore.Level = map[string]zapcore.Level{
	"trace": zapcore.DebugLevel,
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
	"fatal": zapcore.FatalLevel,
	"panic": zapcore.PanicLevel,
}

func ParseArgs(argv []string) map[string]interface{} {
	usage := `tail2kinesis - Like 'tail -f', outputs to Amazon Kinesis, optionally transforms the data

Usage:
    tail2kinesis -h | --help
    tail2kinesis -v | --version
    tail2kinesis --stream-name=STREAM_NAME [--auth=AUTH_TYPE] [--role-arn=ROLE_ARN] [--region=REGION] [--transform=TRANSFORM] [--endpoint=ENDPOINT] [--log-level=LOG_LEVEL] <filename>

Options:
    -h --help
    -v --version
    --auth=AUTH_TYPE
                 Get the credentials from the environment (AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY/AWS_SESSION_TOKEN),
                 the EC2 metadata service or STS AssumeRole. For 'metadata', the calling instance to be attached to
                 an instance profile. For 'assumerole', you need to pass '--role-arn'.
                 Allowed: {env,metadata,assumerole} - [default: env].
    --role-arn=ROLE_ARN
                 The Amazon Resource Name (ARN) of the role to assume.
    --region=REGION
                 AWS region. If not specified, the AWS_DEFAULT_REGION environment variable is used.
                 No region is required when using the --endpoint option.
    --stream-name=STREAM_NAME
                 Name of the Kinesis stream that will receive the data.
                 It must exist.
    --transform=TRANSFORM
                 Every input line is passed to the given TRANSFORM function.
                 Allowed: {identity,accesslog-query} - [default: identity].
    --endpoint=ENDPOINT
                 Kinesis API endpoint, mostly useful in testing.
    --log-level=LOG_LEVEL
                 Log level.
                 Allowed: {trace,debug,info,warn,error,fatal,panic} - [default: error].

Arguments:
    <filename>   Name of the file being tailed, it does not need to exist.

Notes:
    * Lines are batched. Batches are sent to Kinesis no more than once per second, or when the maximum
      batch size is reached (default: 100 lines).
    * There's no "-n" option, files are always tailed from EOF.
`
	arguments, err := docopt.ParseArgs(usage, argv, VERSION)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if arguments["--version"].(bool) {
		fmt.Println(VERSION)
		os.Exit(1)
	}

	validAuths := []string{"env", "metadata", "assumerole"}
	validAuth := false
	for _, t := range validAuths {
		if t == arguments["--auth"].(string) {
			validAuth = true
			break
		}
	}
	if !validAuth {
		fmt.Printf("\"%s\" is not a valid --auth . Choose from: {%s} .\n", arguments["--auth"].(string), strings.Join(validAuths, ","))
		os.Exit(1)
	}

	if arguments["--auth"].(string) == "assumerole" && arguments["--role-arn"] == nil {
		fmt.Printf("\"--auth=metadata\" requires \"--role-arn\" to be passed.\n")
		os.Exit(1)
	}

	if arguments["--region"] == nil && arguments["--endpoint"] == nil && os.Getenv("AWS_DEFAULT_REGION") == "" {
		fmt.Printf("Unless you're specifying an --endpoint, you need to supply a --region or set the AWS_DEFAULT_REGION environment variable.\n")
		os.Exit(1)
	}

	validTransforms := []string{"identity", "accesslog-query"}
	validTransform := false
	for _, t := range validTransforms {
		if t == arguments["--transform"].(string) {
			validTransform = true
			break
		}
	}
	if !validTransform {
		fmt.Printf("\"%s\" is not a valid --transform . Choose from: {%s} .\n", arguments["--transform"].(string), strings.Join(validTransforms, ","))
		os.Exit(1)
	}

	validLogLevel := false
	for k, _ := range LogLevels {
		if k == arguments["--log-level"].(string) {
			validLogLevel = true
			break
		}
	}
	if !validLogLevel {
		fmt.Printf("\"%s\" is not a valid --log-level . Check the --help .\n", arguments["--log-level"].(string))
		os.Exit(1)
	}

	return arguments
}
