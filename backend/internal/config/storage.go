package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AWS_ACCESS_KEY_ID
// AWS_DEFAULT_REGION
// AWS_ENDPOINT_URL
// AWS_S3_BUCKET_NAME
// AWS_SECRET_ACCESS_KEY

func (Config) Storage(ctx context.Context, awsAccessKeyID, awsDefaultRegion, awsEndpointURL, awsSecretAccessKey string) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(awsDefaultRegion), awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(awsAccessKeyID, awsSecretAccessKey, "")))
	if err != nil {
		panic("")
	}

	s3Client := s3.NewFromConfig(cfg)
	if awsEndpointURL != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(awsEndpointURL)
			o.UsePathStyle = true
		})
	}

	return s3Client, nil
}
