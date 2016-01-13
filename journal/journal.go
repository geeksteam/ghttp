package journal

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/geeksteam/GoTools/boltdb"
)

const (
	// TimeLayout is a time layout for time module. Should be used to be consistent with
	// functions which require time strings.
	TimeLayout = time.RFC3339

	dbMode       = os.FileMode(0600)
	keyDelimiter = "|"
)

var (
	cfg Journal
)

func SetConfig(c Journal) {
	cfg = c
}

// Operation is a journal operations struct representation.
type Operation struct {
	SessionID string
	Date      string
	Username  string // имя юзера
	Operation string // Название операции
	Content   string // Содержание операции
	Extra     string // Дополнительная информация
}

// GetAll fetches all operation from BoltDB storage.
func GetAll() (result []Operation) {
	view(func(k, v []byte) {
		op := &Operation{}
		boltdb.DecodeValue(v, op, cfg.DataEncoding)
		result = append(result, *op)
	})
	return
}

// FetchByDate attempts to fetch all operations, which are operated by user with
// username (or all users if username = "") and in time bounds beetween dateFrom and dateTo dates.
func FetchByDate(dateFrom, dateTo, username string) ([]Operation, error) {
	result := []Operation{}

	db, err := bolt.Open(cfg.BoltDB, dbMode, nil)
	if err != nil {
		return result, err
	}
	defer db.Close()

	//halper func for parsing date and handling errors
	parseTime := func(source string) (time.Time, error) {
		t, err := time.Parse(TimeLayout, source)
		if err != nil {
			return t, err
		}
		return t, nil
	}

	from, err := parseTime(dateFrom)
	if err != nil {
		return result, err
	}

	to, err := parseTime(dateTo)
	if err != nil {
		return result, err
	}

	now, err := parseTime(time.Now().Format(TimeLayout))
	if err != nil {
		return result, err
	}

	if to.Equal(now) {
		to = time.Now()
	} else {
		to = to.Add(24 * time.Hour)
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cfg.BucketForOperations))
		c := b.Cursor()
		for k, v := c.Seek([]byte(from.String())); k != nil && bytes.Compare(k, []byte(to.String())) <= 0; k, v = c.Next() {
			//check username
			if username != "" && !bytes.HasSuffix(k, []byte(username)) {
				continue
			}

			op := &Operation{}
			boltdb.DecodeValue(v, op, cfg.DataEncoding)
			result = append(result, *op)
		}
		return nil
	})

	return result, nil
}

// Add attempts to add given operation into BoltDB storage.
func Add(operation Operation) error {
	db, err := bolt.Open(cfg.BoltDB, dbMode, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(cfg.BucketForOperations))
		if err != nil {
			return err
		}
		operation.Date = getCurrentDateString()
		key := createKey(operation.Date, operation.Username)
		value, err := boltdb.EncodeValue(operation, cfg.DataEncoding)
		if errPut := bucket.Put([]byte(key), value); errPut != nil {
			err = errPut
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// CleanOld Delete entries which out of date
func CleanOld() error {
	db, err := bolt.Open(cfg.BoltDB, dbMode, nil)
	defer db.Close()
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cfg.BucketForOperations))
		if bucket == nil {
			return errors.New("Jornal bucket for operation not found")
		}

		cursor := bucket.Cursor()

		//Calculate latest date
		latest := time.Now().Add(time.Duration(-cfg.Capacity) * 24 * time.Hour).Format(TimeLayout)

		for k, _ := cursor.Seek([]byte("0")); k != nil && bytes.Compare(k, []byte(latest)) <= 0; k, _ = cursor.Next() {
			cursor.Delete()
		}
		return nil
	})
}

func getCurrentDateString() string {
	t := time.Now()
	return t.UTC().Format(TimeLayout)
}

func createKey(date, username string) string {
	return date + keyDelimiter + username
}

// view is a generic walk function. Maps given function to all elements of
// BucketForOperations bucket.
func view(f func(k, v []byte)) error {
	db, err := bolt.Open(cfg.BoltDB, dbMode, nil)
	if err != nil {
		return err
	}
	defer db.Close()
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(cfg.BucketForOperations))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			f(k, v)
		}
		return nil
	})
}
