package bolt

import (
	"os/user"
	"strings"

	"github.com/fengxsong/pubmgmt/api"
)

type HostService struct {
	store *Store
}

func (service *HostService) Hostgroup(ID uint64) (*pub.Hostgroup, error) {
	var hostgroup pub.Hostgroup
	if err := service.store.getObjectByID(hostgroupBucketName, ID, &hostgroup); err != nil {
		return nil, err
	}
	return &hostgroup, nil
}

func (service *HostService) HostgroupByName(name string) (*pub.Hostgroup, error) {
	modelSet, err := service.store.getObjectByFieldName(hostgroupBucketName, "Name", name)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrHostgroupNotFound
	} else if err != nil {
		return nil, err
	}
	return modelSet[0].(*pub.Hostgroup), nil
}

// Hostgroups return an array containing all the hostgroups.
func (service *HostService) Hostgroups() ([]pub.Hostgroup, error) {
	modelSet, err := service.store.getObjectByFieldName(hostgroupBucketName, "", nil)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrHostgroupSetEmpty
	} else if err != nil {
		return nil, err
	}
	return trHgs(modelSet), nil
}

func trHgs(ms []pub.Model) []pub.Hostgroup {
	var hostgroups []pub.Hostgroup
	for _, m := range ms {
		hostgroups = append(hostgroups, *m.(*pub.Hostgroup))
	}
	return hostgroups
}

func (service *HostService) UpdateHostgroup(ID uint64, hostgroup *pub.Hostgroup) error {
	return service.store.updateObjectByID(hostgroupBucketName, ID, hostgroup)
}

func (service *HostService) CreateHostgroup(hostgroup *pub.Hostgroup) error {
	return service.store.createObject(hostgroupBucketName, hostgroup)
}

func (service *HostService) DeleteHostgroup(ID uint64) error {
	return service.store.deleteObject(hostgroupBucketName, ID)
}

func (service *HostService) Host(ID uint64) (*pub.Host, error) {
	var host pub.Host
	if err := service.store.getObjectByID(hostBucketName, ID, &host); err != nil {
		return nil, err
	}
	return &host, nil
}

func (service *HostService) Hosts() ([]pub.Host, error) {
	modelSet, err := service.store.getObjectByFieldName(hostBucketName, "", nil)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrHostSetEmpty
	} else if err != nil {
		return nil, err
	}
	return trHosts(modelSet), nil
}

func trHosts(ms []pub.Model) []pub.Host {
	var hosts []pub.Host
	for _, m := range ms {
		hosts = append(hosts, *m.(*pub.Host))
	}
	return hosts
}

func (service *HostService) HostByName(hostname string) (*pub.Host, error) {
	modelSet, err := service.store.getObjectByFieldName(hostBucketName, "Hostname", hostname)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrHostNotFound
	} else if err != nil {
		return nil, err
	}
	return modelSet[0].(*pub.Host), nil
}

func (service *HostService) HostsByStatus(status bool) ([]pub.Host, error) {
	modelSet, err := service.store.getObjectByFieldName(hostBucketName, "IsActive", status)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrHostSetEmpty
	} else if err != nil {
		return nil, err
	}
	return trHosts(modelSet), nil
}

func (service *HostService) HostsByHostgroupID(ID uint64) ([]pub.Host, error) {
	modelSet, err := service.store.getObjectByFieldName(hostBucketName, "HostgroupID", ID)
	if err == pub.ErrModelSetEmpty {
		return nil, pub.ErrHostSetEmpty
	} else if err != nil {
		return nil, err
	}
	return trHosts(modelSet), nil
}

func (service *HostService) UpdateHost(ID uint64, host *pub.Host) error {
	return service.store.updateObjectByID(hostBucketName, ID, host)
}

func (service *HostService) CreateHost(host *pub.Host) error {
	return service.store.createObject(hostBucketName, host)
}

func (service *HostService) DeleteHost(ID uint64) error {
	return service.store.deleteObject(hostBucketName, ID)
}

func (service *HostService) NewHost(str string) *pub.Host {
	host := new(pub.Host)
	if at := strings.Index(str, "@"); at != -1 {
		host.Username = str[:at]
		host.Hostname = str[at+1:]
	} else {
		u, err := user.Current()
		if err != nil {
			host.Username = "root"
		} else {
			host.Username = u.Username
		}
		host.Hostname = str
	}
	if colon := strings.Index(host.Hostname, ":"); colon != -1 {
		host.Port = host.Hostname[colon+1:]
		host.Hostname = host.Hostname[:colon]
	} else {
		host.Port = "22"
	}
	return host
}
