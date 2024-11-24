package fsm

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/hashicorp/raft"
)

type badgerFSM struct {
	db *badger.DB
}

func (b badgerFSM) get(key string) (interface{}, error) {
	var keyByte = []byte(key)
	var data interface{}

	txn := b.db.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(keyByte)
	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	var value = make([]byte, 0)
	err = item.Value(func(val []byte) error {
		value = val
		return nil
	})

	if err != nil {
		data = map[string]interface{}{}
		return data, err
	}

	if len(value) > 0 {
		err = json.Unmarshal(value, &data)
	}

	if err != nil {
		data = map[string]interface{}{}
	}

	return data, err
}

func (b badgerFSM) set(key string, value interface{}) error {
	var data = make([]byte, 0)
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if data == nil || len(data) <= 0 {
		return nil
	}

	txn := b.db.NewTransaction(true)
	err = txn.Set([]byte(key), data)
	if err != nil {
		txn.Discard()
		return err
	}

	err = txn.Commit()

	return err
}

func (b badgerFSM) delete(key string) error {
	var keyByte = []byte(key)

	txn := b.db.NewTransaction(true)
	defer txn.Discard()

	err := txn.Delete(keyByte)
	if err != nil {
		return err
	}

	return txn.Commit()
}

func (b badgerFSM) Apply(log *raft.Log) interface{} {
	if log.Type != raft.LogCommand {
		_, _ = fmt.Fprintf(os.Stderr, "not raft log command type\n")
		return nil
	}

	var payload CommandPayload
	if err := json.Unmarshal(log.Data, &payload); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error unmarshalling store payload: %s\n", err.Error())
		return nil
	}

	op := strings.ToUpper(strings.TrimSpace(payload.Operation))
	switch op {
	case "SET":
		err := b.set(payload.Key, payload.Value)
		return &ApplyResponse{Error: err, Data: payload.Value}
	case "GET":
		data, err := b.get(payload.Key)
		return &ApplyResponse{Error: err, Data: data}
	case "DELETE":
		err := b.delete(payload.Key)
		return &ApplyResponse{Error: err}
	default:
		_, _ = fmt.Fprintf(os.Stderr, "unsupported operation: %s\n", op)
		return nil
	}
}


func (b badgerFSM) Snapshot() (raft.FSMSnapshot, error) {
	return newSnapshotNoop()
}


func (b badgerFSM) Restore(rClose io.ReadCloser) error {
	defer func() {
		if err := rClose.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[FINALLY RESTORE] close error %s\n", err.Error())
		}
	}()

	_, _ = fmt.Fprintf(os.Stdout, "[START RESTORE] read all message from snapshot\n")
	var totalRestored int

	decoder := json.NewDecoder(rClose)
	for decoder.More() {
		var data = &CommandPayload{}
		err := decoder.Decode(data)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error decode data %s\n", err.Error())
			return err
		}

		if err := b.set(data.Key, data.Value); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error persist data %s\n", err.Error())
			return err
		}

		totalRestored++
	}

	_, err := decoder.Token()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] error %s\n", err.Error())
		return err
	}

	_, _ = fmt.Fprintf(os.Stdout, "[END RESTORE] success restore %d messages in snapshot\n", totalRestored)
	return nil
}


func NewBadger(badgerDB *badger.DB) raft.FSM {
	return &badgerFSM{
		db: badgerDB,
	}
}
