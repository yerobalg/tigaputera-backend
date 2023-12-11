package storage

import (
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
	"tigaputera-backend/sdk/file"

	"context"
	"encoding/json"
	"io"
	"net/url"
)

type GCPServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

type storageLib struct {
	ServiceAccount GCPServiceAccount
	BucketName     string
	client         *storage.Client
}

type Interface interface {
	Upload(ctx context.Context, file *file.File, path string, filename string) (string, error)
	Delete(ctx context.Context, path string, filename string) error
	getObjectPlace(objectPath string) *storage.ObjectHandle
}

func Init(serviceAccount GCPServiceAccount, bucketName string) Interface {
	serviceAccountJson, err := json.Marshal(serviceAccount)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(serviceAccountJson))
	if err != nil {
		panic(err)
	}

	return &storageLib{
		BucketName:     bucketName,
		ServiceAccount: serviceAccount,
		client:         client,
	}
}

func (s *storageLib) getObjectPlace(objectPath string) *storage.ObjectHandle {
	return s.client.Bucket(s.BucketName).Object(objectPath)
}

func (s *storageLib) Upload(
	ctx context.Context,
	file *file.File,
	filename string,
	path string,
) (string, error) {
	var imageURL string
	writer := s.getObjectPlace(path + "/" + filename).NewWriter(ctx)

	if _, err := io.Copy(writer, file.Content); err != nil {
		return imageURL, err
	}

	if err := writer.Close(); err != nil {
		return imageURL, err
	}

	parsedURL, err := url.Parse(writer.Attrs().MediaLink)
	if err != nil {
		return imageURL, err
	}

	imageURL = parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path

	return imageURL, nil
}

func (s *storageLib) Delete(
	ctx context.Context,
	filename string,
	path string,
) error {
	return s.getObjectPlace(path + "/" + filename).Delete(ctx)
}
