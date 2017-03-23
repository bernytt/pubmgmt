package bolt

import (
	"reflect"
	"time"

	"github.com/boltdb/bolt"
	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/api/bolt/internal"
	"github.com/fengxsong/pubmgmt/helper"
)

type Store struct {
	Path          string // Path where is stored the BoltDB database
	UserService   *UserService
	HostService   *HostService
	MailerService *MailerService
	TaskService   *TaskService
	db            *bolt.DB
}

const (
	databaseFileName    = "pubmgmt.db"
	userBucketName      = "users"
	hostBucketName      = "hosts"
	hostgroupBucketName = "hostgroups"
	emailBucketName     = "emails"
	taskBucketName      = "tasks"
	cronBucketName      = "crons"
)

var bucketFuncMap = map[string]func() pub.Model{
	userBucketName:      func() pub.Model { return &pub.User{} },
	hostBucketName:      func() pub.Model { return &pub.Host{} },
	hostgroupBucketName: func() pub.Model { return &pub.Hostgroup{} },
	emailBucketName:     func() pub.Model { return &pub.Email{} },
	taskBucketName:      func() pub.Model { return &pub.Task{} },
	cronBucketName:      func() pub.Model { return &pub.Cron{} },
}

func NewStore(storePath string) (*Store, error) {
	store := &Store{
		Path:          storePath,
		UserService:   &UserService{},
		HostService:   &HostService{},
		MailerService: &MailerService{},
		TaskService:   &TaskService{},
	}
	store.UserService.store = store
	store.HostService.store = store
	store.MailerService.store = store
	store.TaskService.store = store
	return store, nil
}

func (store *Store) Open() error {
	path := store.Path + "/" + databaseFileName
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	store.db = db
	return db.Update(func(tx *bolt.Tx) error {
		for bucket, _ := range bucketFuncMap {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return err
			}
		}
		return nil
	})
}

func (store *Store) Close() error {
	if store.db != nil {
		return store.db.Close()
	}
	return nil
}

// container is a struct pointer
func (store *Store) getObjectByID(bucketName string, ID uint64, container interface{}) error {
	var data []byte
	err := store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		value := bucket.Get(internal.Itob(ID))
		if value == nil {
			return pub.ErrObjNotFound
		}
		data = make([]byte, len(value))
		copy(data, value)
		return nil
	})
	if err != nil {
		return err
	}
	if err = internal.Unmarshal(data, container); err != nil {
		return err
	}
	return nil
}

func (store *Store) deleteObject(bucketName string, ID uint64) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if err := bucket.Delete(internal.Itob(ID)); err != nil {
			return err
		}
		return nil
	})
}

// container is a struct pointer
func (store *Store) updateObjectByID(bucketName string, ID uint64, container interface{}) error {
	data, err := internal.Marshal(container)
	if err != nil {
		return err
	}
	return store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if err = bucket.Put(internal.Itob(ID), data); err != nil {
			return err
		}
		return nil
	})
}

func setObjectID(v interface{}, x uint64) error {
	s := reflect.Indirect(reflect.ValueOf(v))
	id := s.FieldByName("ID")
	if id.CanSet() {
		id.SetUint(x)
		return nil
	}
	return pub.ErrIDCanNotSet
}

func (store *Store) createObject(bucketName string, container interface{}) error {
	return store.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		id, _ := bucket.NextSequence()
		if err := setObjectID(container, id); err != nil {
			return err
		}

		data, err := internal.Marshal(container)
		if err != nil {
			return err
		}
		if err = bucket.Put(internal.Itob(id), data); err != nil {
			return err
		}
		return nil
	})
}

func getFieldVal(v interface{}, fieldName string) interface{} {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct && val.Kind() != reflect.Ptr {
		return nil
	}
	val = reflect.Indirect(val).FieldByName(fieldName)
	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.Uint64:
		return val.Uint()
	case reflect.String:
		return val.String()
	}
	return nil
}

func (store *Store) getObjectByFieldName(bucketName, fieldName string, fieldVal interface{}) ([]pub.Model, error) {
	function, ok := bucketFuncMap[bucketName]
	if !ok {
		return nil, pub.ErrModelNotFound
	}
	var (
		filter   bool
		modelSet []pub.Model
	)
	if fieldName != "" {
		filter = true
	}
	unique := helper.Contains(function().UniqueFields(), fieldName)
	err := store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var m = function()
			err := internal.Unmarshal(v, m)
			if err != nil {
				return err
			}
			if filter {
				if getFieldVal(m, fieldName) == fieldVal {
					modelSet = append(modelSet, m)
					if unique {
						break
					}
				}
			} else {
				modelSet = append(modelSet, m)
			}
		}
		if len(modelSet) == 0 {
			return pub.ErrModelSetEmpty
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return modelSet, nil
}
