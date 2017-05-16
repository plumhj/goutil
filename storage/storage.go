package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io/ioutil"
)

type Storage interface {
	Save(string, interface{}) error
	Read(string, interface{}) error
	Delete(string) error
}

type S3 struct {
	sss    *s3.S3
	bucket string
}

func NewS3(bucket, region string) Storage {
	sess := session.Must(session.NewSession(&aws.Config{Region: aws.String(region)}))
	return &S3{
		sss:    s3.New(sess),
		bucket: bucket,
	}
}

func (s *S3) Save(key string, obj interface{}) (err error) {

	json, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	params := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket), // Required
		Key:    aws.String(key),      // Required
		Body:   bytes.NewReader(json),
		//ContentType: aws.String("application/json"),
	}

	_, err = s.sss.PutObject(params)
	if err != nil {
		return err
	}

	return
}

func (s *S3) Read(key string, obj interface{}) (err error) {
	p := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket), // Required
		Key:    aws.String(key),      // Required
	}
	resp, err := s.sss.GetObject(p)
	if err != nil {
		return fmt.Errorf("failed to get s3 object, %v", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to get s3 object, %v", err)
	}

	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("failed to unmarshal s3 object, %v", err)
	}

	return
}

func (s *S3) Delete(key string) (err error) {
	params := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket), // Required
		Key:    aws.String(key),      // Required
	}

	_, err = s.sss.DeleteObject(params)
	return
}
