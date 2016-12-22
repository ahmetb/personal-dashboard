package main

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
)

type googleStorage struct {
	bucket string
	object string
}

func googleStorageFromParameters(config map[string]string) (uploader, error) {
	v := &googleStorage{
		bucket: config["bucket"],
		object: config["object"]}
	if v.bucket == "" {
		return nil, errors.New("googlestorage: 'bucket' not specified")
	}
	if v.object == "" {
		return nil, errors.New("googlestorage: 'object' not specified")
	}
	return v, nil
}

func (s *googleStorage) save(b []byte) error {
	client, err := storage.NewClient(context.TODO())
	if err != nil {
		return errors.Wrap(err, "failed to create storage client")
	}
	w := client.Bucket(s.bucket).Object(s.object).
		NewWriter(context.TODO())
	w.ContentType = "application/javascript"
	w.CacheControl = "no-cache"
	w.ACL = []storage.ACLRule{
		{Entity: storage.AllUsers, Role: storage.RoleReader}}
	if _, err = w.Write(b); err != nil {
		return errors.Wrap(err, "failed to write to object")
	}
	return errors.Wrap(w.Close(), "failed to close object writer")
}
