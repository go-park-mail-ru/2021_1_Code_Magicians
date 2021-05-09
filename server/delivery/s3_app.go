package delivery

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3App struct {
	sess       *session.Session
	bucketName string
}

func NewS3App(sess *session.Session, bucketName string) *S3App {
	return &S3App{sess, bucketName}
}

type S3AppInterface interface {
	UploadFile(io.Reader, string) error // Upload file to s3 bucket
	DeleteFile(string) error            // Delete file from s3 bucket
}

func (s3App *S3App) UploadFile(file io.Reader, filename string) error {
	uploader := s3manager.NewUploader(s3App.sess)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3App.bucketName),
		ACL:    aws.String("public-read"),
		Key:    aws.String(filename),
		Body:   file,
	})

	return handleS3Error(err)
}

func (s3App *S3App) DeleteFile(filename string) error {
	deleter := s3.New(s3App.sess)
	_, err := deleter.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s3App.bucketName),
		Key:    aws.String(filename),
	})

	return handleS3Error(err)
}

func handleS3Error(err error) error {
	if err == nil {
		return nil
	}

	aerr, ok := err.(awserr.Error)
	if ok {
		switch aerr.Code() {
		case s3.ErrCodeNoSuchBucket:
			return fmt.Errorf("Specified bucket does not exist")
		case s3.ErrCodeNoSuchKey:
			return fmt.Errorf("No file found with such filename")
		case s3.ErrCodeObjectAlreadyInActiveTierError:
			return fmt.Errorf("S3 bucket denied access to you")
		default:
			return fmt.Errorf("Unknown S3 error")
		}
	}

	return fmt.Errorf("Not an S3 error")
}
