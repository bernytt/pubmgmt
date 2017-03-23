package bolt

import (
	"github.com/fengxsong/pubmgmt/api"
)

type MailerService struct {
	store *Store
}

func (service *MailerService) CreateEmail(email *pub.Email) error {
	return service.store.createObject(emailBucketName, email)
}

func (service *MailerService) EmailByUser(userId uint64) ([]pub.Email, error) {
	modelSets, err := service.store.getObjectByFieldName(emailBucketName, "FromUserID", userId)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrEmailOutboxEmpty
	} else if err != nil {
		return nil, err
	}
	return trMs(modelSets), nil
}

func trMs(ms []pub.Model) []pub.Email {
	var emails []pub.Email
	for _, m := range ms {
		emails = append(emails, *m.(*pub.Email))
	}
	return emails
}

func (service *MailerService) EmailByUUID(uuid string) (*pub.Email, error) {
	modelSets, err := service.store.getObjectByFieldName(emailBucketName, "UUID", uuid)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrEmailNotFound
	} else if err != nil {
		return nil, err
	}
	return modelSets[0].(*pub.Email), nil
}
