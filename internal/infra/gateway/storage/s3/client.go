package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Client cliente S3 compatível com LocalStack (endpoint customizado, path-style).
type Client struct {
	client     *s3.Client
	bucket     string
	region     string
	endpoint   string
	baseURL    string // URL base para montar URLs públicas (ex: http://localhost:4566/bucket)
}

// Config configuração do cliente S3.
type Config struct {
	Bucket          string
	Region          string
	Endpoint        string // ex: http://localhost:4566
	AccessKeyID     string
	SecretAccessKey string
}

// NewClient cria um cliente S3. Para LocalStack use Endpoint e UsePathStyle.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{URL: cfg.Endpoint}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("carregar config aws: %w", err)
	}
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
	}
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	baseURL := cfg.Endpoint
	if baseURL != "" && cfg.Bucket != "" {
		baseURL = strings.TrimSuffix(cfg.Endpoint, "/") + "/" + cfg.Bucket
	}
	c := &Client{
		client:   client,
		bucket:   cfg.Bucket,
		region:   cfg.Region,
		endpoint: cfg.Endpoint,
		baseURL:  baseURL,
	}
	if err := c.EnsureBucket(ctx); err != nil {
		return nil, fmt.Errorf("garantir bucket: %w", err)
	}
	return c, nil
}

// EnsureBucket cria o bucket se não existir.
func (c *Client) EnsureBucket(ctx context.Context) error {
	_, err := c.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &c.bucket})
	if err != nil {
		// LocalStack/S3 pode retornar erro se já existir; ignorar
		if !strings.Contains(err.Error(), "BucketAlreadyExists") && !strings.Contains(err.Error(), "AlreadyExists") {
			return err
		}
	}
	return nil
}

// UploadFile envia um arquivo local para a chave key no bucket.
func (c *Client) UploadFile(ctx context.Context, localPath, key string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    &key,
		Body:   f,
	})
	return err
}

// UploadHLSDir envia todos os arquivos do diretório localDir para o prefixo S3 (prefix/videoID/).
// Faz upload em paralelo (workers) e retorna a URL do manifest (primeiro .m3u8 encontrado).
func (c *Client) UploadHLSDir(ctx context.Context, localDir, prefix string, workers int) (manifestURL string, err error) {
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return "", err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		files = append(files, filepath.Join(localDir, e.Name()))
	}
	if workers <= 0 {
		workers = 5
	}
	type upload struct {
		localPath string
		key       string
	}
	var jobs []upload
	for _, f := range files {
		key := strings.TrimPrefix(strings.TrimPrefix(f, localDir), string(filepath.Separator))
		key = filepath.Join(prefix, key)
		key = filepath.ToSlash(key)
		jobs = append(jobs, upload{localPath: f, key: key})
	}
	var wg sync.WaitGroup
	errCh := make(chan error, len(jobs))
	sem := make(chan struct{}, workers)
	for _, j := range jobs {
		wg.Add(1)
		go func(localPath, key string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if e := c.UploadFile(ctx, localPath, key); e != nil {
				errCh <- e
			}
		}(j.localPath, j.key)
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		if err == nil {
			err = e
		}
	}
	if err != nil {
		return "", err
	}
	// URL do manifest: baseURL/prefix/index.m3u8 (ou o nome do arquivo .m3u8)
	for _, j := range jobs {
		if strings.HasSuffix(j.key, ".m3u8") {
			manifestURL = c.baseURL + "/" + j.key
			break
		}
	}
	return manifestURL, nil
}

// PutObject envia um body para a chave key (útil para testes).
func (c *Client) PutObject(ctx context.Context, key string, body io.Reader) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &c.bucket,
		Key:    &key,
		Body:   body,
	})
	return err
}
