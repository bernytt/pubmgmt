package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/api/bolt/internal"
)

type TaskService struct {
	store *Store
}

func (service *TaskService) Task(ID uint64) (*pub.Task, error) {
	var task pub.Task
	if err := service.store.getObjectByID(taskBucketName, ID, &task); err != nil {
		return nil, err
	}
	return &task, nil
}

func (service *TaskService) Tasks(reqApproval, unfinished bool) ([]pub.Task, error) {
	var tasks []pub.Task
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(taskBucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var task pub.Task
			if err := internal.Unmarshal(v, &task); err != nil {
				return err
			}
			if reqApproval && task.RequiredApproval == reqApproval {
				if unfinished && task.Done.IsZero() {
					tasks = append(tasks, task)
				} else {
					tasks = append(tasks, task)
				}
			} else {
				tasks = append(tasks, task)
			}
		}
		if len(tasks) == 0 {
			return pub.ErrTaskSetEmpty
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (service *TaskService) TasksSchedule() ([]pub.Task, error) {
	var tasks []pub.Task
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(taskBucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var task pub.Task
			if err := internal.Unmarshal(v, &task); err != nil {
				return err
			}
			if task.Spec != "" && !task.Suspended {
				tasks = append(tasks, task)
			}
		}
		if len(tasks) == 0 {
			return pub.ErrTaskSetEmpty
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (service *TaskService) UpdateTask(ID uint64, task *pub.Task) error {
	return service.store.updateObjectByID(taskBucketName, ID, task)
}

func (service *TaskService) CreateTask(task *pub.Task) error {
	return service.store.createObject(taskBucketName, task)
}

func (service *TaskService) DeleteTask(ID uint64) error {
	return service.store.deleteObject(taskBucketName, ID)
}

func (service *TaskService) Cron(ID uint64) (*pub.Cron, error) {
	var cron pub.Cron
	if err := service.store.getObjectByID(cronBucketName, ID, &cron); err != nil {
		return nil, err
	}
	return &cron, nil
}

func (service *TaskService) CronByName(name string) (*pub.Cron, error) {
	crons, err := service.store.getObjectByFieldName(cronBucketName, "Name", name)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrCronNotFound
	} else if err != nil {
		return nil, err
	}
	return crons[0].(*pub.Cron), nil
}

func (service *TaskService) Crons() ([]pub.Cron, error) {
	var crons []pub.Cron
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(cronBucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var cron pub.Cron
			if err := internal.Unmarshal(v, &cron); err != nil {
				return err
			}
			crons = append(crons, cron)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return crons, nil
}

func (service *TaskService) UpdateCron(ID uint64, cron *pub.Cron) error {
	return service.store.updateObjectByID(cronBucketName, ID, cron)
}

func (service *TaskService) CreateCron(cron *pub.Cron) error {
	return service.store.createObject(cronBucketName, cron)
}

func (service *TaskService) DeleteCron(ID uint64) error {
	return service.store.deleteObject(cronBucketName, ID)
}
