package bolt

import (
	"github.com/boltdb/bolt"
	"github.com/fengxsong/pubmgmt/api"
	"github.com/fengxsong/pubmgmt/api/bolt/internal"
)

type UserService struct {
	store *Store
}

func (service *UserService) User(ID uint64) (*pub.User, error) {
	var user pub.User
	if err := service.store.getObjectByID(userBucketName, ID, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (service *UserService) UserByEmail(email string) (*pub.User, error) {
	users, err := service.store.getObjectByFieldName(userBucketName, "Email", email)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return users[0].(*pub.User), nil
}

func (service *UserService) UserByUsername(username string) (*pub.User, error) {
	users, err := service.store.getObjectByFieldName(userBucketName, "Username", username)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrUserNotFound
	} else if err != nil {
		return nil, err
	}
	return users[0].(*pub.User), nil
}

func (service *UserService) UsersByRole(role pub.UserRole) ([]pub.User, error) {
	users, err := service.store.getObjectByFieldName(userBucketName, "Role", uint64(role))
	if err != nil {
		return nil, err
	}
	return trUs(users), nil
}

func trUs(ms []pub.Model) []pub.User {
	var users []pub.User
	for _, m := range ms {
		users = append(users, *m.(*pub.User))
	}
	return users
}

// Users return an array containing all the users.
func (service *UserService) Users() ([]pub.User, error) {
	var users = make([]pub.User, 0)
	err := service.store.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(userBucketName))
		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var user pub.User
			err := internal.UnmarshalUser(v, &user)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (service *UserService) UpdateUser(ID uint64, user *pub.User) error {
	return service.store.updateObjectByID(userBucketName, ID, user)
}

func (service *UserService) CreateUser(user *pub.User) error {
	return service.store.createObject(userBucketName, user)
}

func (service *UserService) DeleteUser(ID uint64) error {
	return service.store.deleteObject(userBucketName, ID)
}
