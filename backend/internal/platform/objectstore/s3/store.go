package s3

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type s3Store struct {
	client     *s3.Client
	bucketName string
}

func New(client *s3.Client, buckeName string) *s3Store {
	return new(s3Store{client, buckeName})
}

func (s *s3Store) Put(ctx context.Context, name, contentType string, body []byte) error {
	_, err := s.client.PutObject(ctx, new(s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(name),
		Body:        bytes.NewReader(body),
		ContentType: aws.String(contentType),
	}))
	if err != nil {
		return err
	}

	return nil
}

func (s *s3Store) Get(ctx context.Context, name string) (io.ReadCloser, string, error) {
	result, err := s.client.GetObject(ctx, new(s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(name),
	}))
	if err != nil {
		return nil, "", err
	}

	return result.Body, aws.ToString(result.ContentType), nil
}

func (s *s3Store) Delete(ctx context.Context, name string) error {
	_, err := s.client.DeleteObject(ctx, new(s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(name),
	}))
	if err != nil {
		return err
	}

	return nil
}

func (s *s3Store) Rename(ctx context.Context, name, newName string) error {
	_, err := s.client.RenameObject(ctx, new(s3.RenameObjectInput{
		Bucket:       aws.String(s.bucketName),
		Key:          aws.String(name),
		RenameSource: aws.String(newName),
	}))
	if err != nil {
		return err
	}

	return nil
}
