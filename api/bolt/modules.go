package bolt

import (
	"github.com/fengxsong/pubmgmt/api"
)

type ModuleService struct {
	store *Store
}

func (s *ModuleService) SvnByID(id uint64) (*pub.SubversionInfo, error) {
	var svnInfo pub.SubversionInfo
	if err := s.store.getObjectByID(svnInfoBucketName, id, &svnInfo); err != nil {
		return nil, err
	}
	return &svnInfo, nil
}

func (s *ModuleService) SvnInfos() ([]pub.SubversionInfo, error) {
	modelSet, err := s.store.getObjectByFieldName(svnInfoBucketName, "", nil)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrSvnInfoSetEmpty
	} else if err != nil {
		return nil, err
	}
	return trSvnInfos(modelSet), nil
}

func trSvnInfos(ms []pub.Model) []pub.SubversionInfo {
	var s []pub.SubversionInfo
	for _, m := range ms {
		s = append(s, *m.(*pub.SubversionInfo))
	}
	return s
}

func (s *ModuleService) CreateSvnInfo(svnInfo *pub.SubversionInfo) error {
	return s.store.createObject(svnInfoBucketName, svnInfo)
}

func (s *ModuleService) UpdateSvnInfo(id uint64, svnInfo *pub.SubversionInfo) error {
	return s.store.updateObjectByID(svnInfoBucketName, id, svnInfo)
}

func (s *ModuleService) DeleteSvnInfo(id uint64) error {
	return s.store.deleteObject(svnInfoBucketName, id)
}
