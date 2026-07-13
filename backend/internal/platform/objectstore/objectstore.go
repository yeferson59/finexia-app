// Package objectstore owns the S3 client shared by every module. It is part
// of the platform layer and must stay free of business logic.
package objectstore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Connect builds the S3 client. A non-empty endpointURL switches to
// path-style addressing (MinIO and friends).
func Connect(ctx context.Context, accessKeyID, defaultRegion, endpointURL, secretAccessKey string) (*s3.Client, error) {
	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(defaultRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)
	if endpointURL != "" {
		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpointURL)
			o.UsePathStyle = true
		})
	}

	return s3Client, nil
}
