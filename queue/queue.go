package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"io/ioutil"
	"time"
	"sync/atomic"
)

type Queue interface {
	Put(interface{}) error
	Get(interface{}) (string, error)
	Del(string) error
}

type Qobject struct {
	ID       string `json:"id"`
	JsonData string `json:"jsonData"`
}

type SQS struct {
	id           string
	sss          *s3.S3
	//s3Downloader *s3manager.Downloader
	//s3Uploader   *s3manager.Uploader
	client       sqsiface.SQSAPI
	url          string
	bucket       string
	waitTimeout  int64
	seq          int64
}

func NewSQS(id, url, bucket, region string) Queue {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	//s3Uploader := s3manager.NewUploader(sess)
	//s3Downloader := s3manager.NewDownloader(sess)
	return &SQS{
		id:           id,
		sss:          s3.New(sess),
		//s3Downloader: s3Downloader,
		//s3Uploader:   s3Uploader,
		client:       sqs.New(sess),
		url:          url,
		bucket:       bucket,
	}
}

func (q *SQS) Put(obj interface{}) (err error) {

	qo := Qobject{}
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	qo.JsonData = string(data)

	if len(data)+128 > 1024*256 {
		s := atomic.AddInt64(&q.seq, 1)
		qo.ID = fmt.Sprintf("%s:%d:%d", q.id, time.Now().UnixNano(), s)

		json, err := json.Marshal(qo)
		if err != nil {
			return err
		}

		params := &s3.PutObjectInput{
			Bucket:      aws.String(q.bucket), // Required
			Key:         aws.String(qo.ID),    // Required
			Body:        bytes.NewReader(json),
			ContentType: aws.String("application/json"),
		}

		//_, err = q.s3Uploader.Upload(params)
		_, err = q.sss.PutObject(params)
		if err != nil {
			return err
		}

		qo.JsonData = ""
	}

	data, err = json.Marshal(qo)
	if err != nil {
		return err
	}

	params := &sqs.SendMessageInput{
		QueueUrl:    aws.String(q.url),
		MessageBody: aws.String(string(data)), // Required
	}

	_, err = q.client.SendMessage(params)

	return
}

func (q *SQS) Get(obj interface{}) (id string, err error) {
	params := sqs.ReceiveMessageInput{
		QueueUrl: aws.String(q.url),
	}
	if q.waitTimeout > 0 {
		params.WaitTimeSeconds = aws.Int64(q.waitTimeout)
	}
	resp, err := q.client.ReceiveMessage(&params)
	if err != nil {
		return "", fmt.Errorf("failed to get messages, %v", err)
	}

	qo := Qobject{}
	for _, msg := range resp.Messages {
		//log.Println(msg)
		if err := json.Unmarshal([]byte(aws.StringValue(msg.Body)), &qo); err != nil {
			return "", fmt.Errorf("failed to unmarshal message, %v", err)
		}

		if qo.JsonData == "" {

			p := &s3.GetObjectInput{
				Bucket: aws.String(q.bucket), // Required
				Key:    aws.String(qo.ID),    // Required
			}
			resp, err := q.sss.GetObject(p)
			if err != nil {
				return "", fmt.Errorf("failed to get s3 object, %v", err)
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to get s3 object, %v", err)
			}

			if err := json.Unmarshal(data, &qo); err != nil {
				return "", fmt.Errorf("failed to unmarshal s3 object, %v", err)
			}

		}

		if err := json.Unmarshal([]byte(qo.JsonData), &obj); err != nil {
			return "", fmt.Errorf("failed to unmarshal message, %v", err)
		}

		id = *msg.ReceiptHandle
	}
	return
}

func (q *SQS) Del(id string) (err error) {
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(q.url),
		ReceiptHandle: aws.String(id),
	}

	_, err = q.client.DeleteMessage(params)

	return
}
