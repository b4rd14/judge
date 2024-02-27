package replier

import (
	model "GO/Judge/Model"
	"context"
	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinIoClient() (*minio.Client, error) {
	env := NewEnv()
	cfg := model.MinioConfig{
		Endpoint:  env.MinioEndpoint,
		AccessKey: env.MinioAccessKey,
		SecretKey: env.MinioSecretKey,
		UseSSL:    true,
	}

	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})

	return client, err

}

func NewBucket(client *minio.Client, bucketName string) error {
	err := client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{
		ObjectLocking: false,
		Region:        "us-east-1",
	})
	return err
}

func ListBuckets(client *minio.Client) ([]minio.BucketInfo, error) {
	buckets, err := client.ListBuckets(context.Background())
	return buckets, err
}

func Download(ctx context.Context, client *minio.Client, bucketName string, Prefix string, dirName string) error {
	objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    Prefix,
		Recursive: true,
	})
	for object := range objectCh {
		if object.Err != nil {
			return object.Err
		}
		err := client.FGetObject(ctx, bucketName, object.Key, dirName+"/"+object.Key, minio.GetObjectOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
