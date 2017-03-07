package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
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

func NewQueue(s *sqs.SQS, name string, max, wait int64) (*Queue, error) {
	q := &Queue{SQS: s}
	u, err := q.fetchURL(name)
	if err != nil {
		return nil, err
	}
	q.URL = u
	q.ReceiveConfig = &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(u),
		AttributeNames:        []*string{aws.String(sqs.QueueAttributeNameAll)},
		MessageAttributeNames: []*string{aws.String(sqs.QueueAttributeNameAll)},
		MaxNumberOfMessages:   aws.Int64(max),
		WaitTimeSeconds:       aws.Int64(wait),
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

type message struct {
	Message   *sqs.Message
	Extracted map[string]interface{}
}

func (q *Queue) Receive() ([]message, error) {
	r, err := q.SQS.ReceiveMessage(q.ReceiveConfig)
	if err != nil {
		return nil, err
	}
	messages := make([]message, 0, len(r.Messages))
	for _, m := range r.Messages {
		mm := message{Message: m}
		for k, v := range mm.Message.MessageAttributes {
			// Extract custom typed messages.
			//
			// see also: http://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-message-attributes.html
			if strings.HasSuffix(*v.DataType, ".json") {
				// var attr map[string]interface{}
				attr := make(map[string]interface{})
				if err := json.Unmarshal(v.BinaryValue, &attr); err != nil {
					log.Printf("unmarshal MessageAttribute %s err: %s", k, err)
					continue // skip it
				}
				// if successfully unmarshaled, append it on response
				mm.Extracted = attr
			}
		}
		messages = append(messages, mm)
	}
	return messages, nil
}

func main() {
	var (
		queueName   = flag.String("queue", "your-queue", "your SQS queue name")
		region      = flag.String("region", "ap-northeast-1", "the AWS region where your SQS is.")
		maxMessages = flag.Int64("max-messages", 10, "maximum messages per request")
		wait        = flag.Int64("wait", 3, "wait seconds for long polling.")
		forever     = flag.Bool("f", false, "tailing sqs queue forever or not (like: tail -f)")
		interval    = flag.Duration("interval", time.Second*3, "seconds for waiting next receiveMessage request.")
	)

	flag.Parse()

	s := session.New(&aws.Config{Region: aws.String(*region)})
	q, err := NewQueue(sqs.New(s), *queueName, *maxMessages, *wait)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot initialize sqs-tail. please verify your queue is accecible.: %s", err)
		os.Exit(1)
	}

	for {
		messages, err := q.Receive()
		if err != nil {
			fmt.Fprintf(os.Stderr, "get messages failed: %s", err)
			os.Exit(1)
		}
		fmt.Printf("%#v\n", messages)
		if *forever == false {
			os.Exit(0)
		}
		time.Sleep(*interval)
	}
}
