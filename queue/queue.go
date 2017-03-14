package queue

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"encoding/json"
	"errors"
	"fmt"
)

type Queue interface {
	Put(interface{}) error
	Get(interface{}) (string, error)
	Del(string) error
}

type SQS struct {
	client sqsiface.SQSAPI
	url    string
	WaitTimeout int64
}

func NewSQS(url string) Queue {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("ap-northeast-2")}))
	return SQS{
		client: sqs.New(sess),
		url:    url,
	}
}

func (q SQS) Put(obj interface{}) (err error) {

	data, err := json.Marshal(obj)
	if err != nil {
		return
	}

	if len(data) > 1024 * 256 {
		return errors.New("obj size limit, max : 256KB")
	}

	params := &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.url),
		MessageBody: aws.String(string(data)), // Required
	}

	_, err = q.client.SendMessage(params)
	if err != nil {
		return
	}

	return
}

func (q SQS) Get(obj interface{}) (id string, err error) {
	params := sqs.ReceiveMessageInput{
		QueueUrl: aws.String(q.url),
	}
	if q.WaitTimeout > 0 {
		params.WaitTimeSeconds = aws.Int64(q.WaitTimeout)
	}
	resp, err := q.client.ReceiveMessage(&params)
	if err != nil {
		return "", fmt.Errorf("failed to get messages, %v", err)
	}

	for _, msg := range resp.Messages {
		//log.Println(msg)
		if err := json.Unmarshal([]byte(aws.StringValue(msg.Body)), &obj); err != nil {
			return "", fmt.Errorf("failed to unmarshal message, %v", err)
		}
		id = *msg.ReceiptHandle
	}
	return
}

func (q SQS) Del(id string) (err error) {
	params := &sqs.DeleteMessageInput{
		QueueUrl:    aws.String(q.url),
		ReceiptHandle: aws.String(id),
	}

	_, err = q.client.DeleteMessage(params)

	return
}
