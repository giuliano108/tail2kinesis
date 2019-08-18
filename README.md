```
tail2kinesis - Like 'tail -f', outputs to Amazon Kinesis, optionally transforms the data

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
```

### Development

```
make docker-build
make docker-shell

kinesalite &
. /src/kinesis.sh
kCreateStream test

find /src -name '*.go' | entr -d /src/dockerdevsetup.sh
```

### TODO

- [ ] Configurable partition key
- [ ] Configurable flush interval
