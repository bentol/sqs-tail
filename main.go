package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type Queue struct {
	*sqs.SQS
	maxItemSize   int
	URL           string
	ReceiveConfig *sqs.ReceiveMessageInput
}

func NewQueue(s *sqs.SQS, name string, max int64) (*Queue, error) {
	q := &Queue{SQS: s}
	u, err := q.fetchURL(name)
	if err != nil {
		return nil, err
	}
	q.URL = u
	q.ReceiveConfig = &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(u),
		AttributeNames:        []*string{aws.String("All")},
		MessageAttributeNames: []*string{aws.String("All")},
		MaxNumberOfMessages:   aws.Int64(max),
	}
	return q, nil
}

func (q *Queue) fetchURL(name string) (string, error) {
	resp, err := q.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	})
	if err != nil {
		return "", err
	}
	return *resp.QueueUrl, nil
}

func (q *Queue) Receive() (*sqs.ReceiveMessageOutput, error) {
	return q.SQS.ReceiveMessage(q.ReceiveConfig)
}

func main() {
	var (
		queueName   = flag.String("queue", "your-queue", "your SQS queue name")
		region      = flag.String("region", "ap-northeast-1", "the AWS region where your SQS is.")
		maxMessages = flag.Int64("max-messages", 10, "maximum messages per request")
		forever     = flag.Bool("f", true, "tailing sqs queue forever or not (like: tail -f)")
		interval    = flag.Duration("interval", time.Second*3, "seconds for waiting next receiveMessage request.")
	)

	flag.Parse()

	s := session.New(&aws.Config{Region: aws.String(*region)})
	q, err := NewQueue(sqs.New(s), *queueName, *maxMessages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot initialize sqs-tail. please verify your queue is accecible.: %s", err)
		os.Exit(1)
	}

	for {
		if *forever == false {
			fmt.Fprint(os.Stdout, "all messages in this queue is fetched.")
			os.Exit(0)
		}
		resp, err := q.Receive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "get messages failed: %s", err)
			os.Exit(1)
		}
		time.Sleep(*interval)
		fmt.Printf("messages = %+v\n", resp.Messages)
	}
}
