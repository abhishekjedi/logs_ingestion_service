package config

import "github.com/knadh/koanf/v2"

type S3Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
	UseSSL    bool
}

func NewS3Config(k *koanf.Koanf) S3Config {
	region := k.String("s3.region")
	if region == "" {
		region = "us-east-1"
	}
	return S3Config{
		Endpoint:  k.String("s3.endpoint"),
		AccessKey: k.String("s3.access_key"),
		SecretKey: k.String("s3.secret_key"),
		Bucket:    k.String("s3.bucket"),
		Region:    region,
		UseSSL:    k.Bool("s3.use_ssl"),
	}
}
