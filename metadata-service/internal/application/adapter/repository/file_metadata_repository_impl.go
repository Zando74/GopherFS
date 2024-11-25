package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Zando74/GopherFS/metadata-service/config"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/metadata-service/logger"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type FileMetadataRepositoryImpl struct {
	client *clientv3.Client
	ttl    int
}

func NewEtcdFileMetadataRepository() *FileMetadataRepositoryImpl {

	cfg := config.ConfigSingleton.GetInstance()
	log := logger.LoggerSingleton.GetInstance()

	cfgEtcd := clientv3.Config{
		Endpoints:   cfg.Etcd.Endpoints,
		DialTimeout: time.Duration(cfg.Etcd.DialTimeout) * time.Second,
	}
	client, err := clientv3.New(cfgEtcd)
	if err != nil {
		log.Fatal("failed to create etcd client %v", err)
	}
	return &FileMetadataRepositoryImpl{
		client: client,
		ttl:    cfg.Etcd.TTL,
	}
}

func (repo *FileMetadataRepositoryImpl) SaveFileMetadata(metadata *entity.FileMetadata, onSuccess func()) error {
	key := "/file_metadata/" + metadata.FileID
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(repo.ttl)*time.Second)
	defer cancel()

	session, err := concurrency.NewSession(repo.client)
	if err != nil {
		return err
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, key)
	if err := mutex.Lock(ctx); err != nil {
		return err
	}
	defer mutex.Unlock(ctx)

	data, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	resp, err := repo.client.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, string(data))).
		Commit()
	if err != nil {
		return err
	}

	if resp.Succeeded {
		onSuccess()
		return nil
	}
	return errors.New("file metadata already exists")
}

func (repo *FileMetadataRepositoryImpl) DeleteFileMetadata(fileID string) error {
	key := "/file_metadata/" + fileID
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(repo.ttl)*time.Second)
	defer cancel()

	session, err := concurrency.NewSession(repo.client)
	if err != nil {
		return err
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, key)
	if err := mutex.Lock(ctx); err != nil {
		return err
	}
	defer mutex.Unlock(ctx)

	resp, err := repo.client.Delete(ctx, key)
	if err != nil {
		return err
	}

	if resp.Deleted == 0 {
		return errors.New("no file metadata found to delete")
	}
	return nil
}

func (repo *FileMetadataRepositoryImpl) GetFileMetadata(fileID string) (*entity.FileMetadata, error) {
	key := "/file_metadata/" + fileID
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(repo.ttl)*time.Second)
	defer cancel()

	session, err := concurrency.NewSession(repo.client)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	mutex := concurrency.NewMutex(session, key)
	if err := mutex.Lock(ctx); err != nil {
		return nil, err
	}
	defer mutex.Unlock(ctx)

	resp, err := repo.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	var metadata entity.FileMetadata
	if err := json.Unmarshal([]byte(resp.Kvs[0].Value), &metadata); err != nil {
		return nil, err
	}

	return &metadata, nil
}
