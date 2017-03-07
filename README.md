# tail for SQS

## features

* tail SQS queue.
* extract attributes with custom type
  * especially, when the suffix is `.json`, read as json.
  * if not, read as text.

## Installation

	go get github.com/suzuken/sqs-tail

## Usage

```
$ sqs-tail -help
Usage of ./sqs-tail:
  -f    tailing sqs queue forever or not (like: tail -f) (default true)
  -interval duration
        seconds for waiting next receiveMessage request. (default 3s)
  -max-messages int
        maximum messages per request (default 10)
  -queue string
        your SQS queue name (default "your-queue")
  -region string
        the AWS region where your SQS is. (default "ap-northeast-1")

```

## LICENSE

MIT
