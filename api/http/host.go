package http

import (
	"net/http"
	"strconv"

	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/gin-gonic/gin.v1"
)

type HostHandler struct {
	Logger      logger
	HostService pub.HostService
}

// url: /hostgroups  method: PUT  body: pub.Hostgroup
func (h *HostHandler) createHostgroup(ctx *gin.Context) {
	var req pub.Hostgroup
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	hostgroup, err := h.HostService.HostgroupByName(req.Name)
	if err != nil && err != pub.ErrHostgroupNotFound {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	if hostgroup != nil {
		Error(ctx, pub.ErrHostgroupAlreadyExists, http.StatusConflict, nil)
		return
	}
	hostgroup = &pub.Hostgroup{
		Name:    req.Name,
		Comment: req.Comment,
	}
	err = h.HostService.CreateHostgroup(hostgroup)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: "Put hostgroup success"})
}

// url: /hostgroups  method: GET
func (h *HostHandler) getHostgroups(ctx *gin.Context) {
	hostgroups, err := h.HostService.Hostgroups()
	if err == pub.ErrHostgroupSetEmpty {
		Error(ctx, err, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
	} else {
		ctx.IndentedJSON(http.StatusOK, hostgroups)
	}
}

// url: /hostgroups/pk/:id  method: GET,
func (h *HostHandler) getHostgroupByID(ctx *gin.Context) {
	hostgroupId, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}

	hostgroup, err := h.HostService.Hostgroup(hostgroupId)

	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrHostgroupNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, hostgroup)
}

// url: /hostgroups/pk/:id  method: DELETE
func (h *HostHandler) deleteHostgroupByID(ctx *gin.Context) {
	ID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}

	_, err = h.HostService.Hostgroup(ID)

	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrHostgroupNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	err = h.HostService.DeleteHostgroup(ID)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Delete hostgroup success"})
}

// url: /hosts  method: GET  body: putHostRequest
func (h *HostHandler) createHost(ctx *gin.Context) {
	var req putHostRequest
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	reqHost := h.HostService.NewHost(req.Format)
	host, err := h.HostService.HostByName(reqHost.Hostname)
	if err != nil && err != pub.ErrHostNotFound {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	if host != nil {
		Error(ctx, pub.ErrHostAlreadyExists, http.StatusConflict, nil)
		return
	}
	reqHost.HostgroupID = req.HostgroupID
	reqHost.Password = req.Password
	reqHost.IdentityFile = req.IdentityFile
	reqHost.Comment = req.Comment
	reqHost.IsActive = req.IsActive
	err = h.HostService.CreateHost(reqHost)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: "Put host success"})
}

// for creating a host from specical json format
type putHostRequest struct {
	Format       string `json:"format" binding:"required"` // "root@localhost:22"
	Password     string `json:"password"`
	HostgroupID  uint64 `json:"hostgroup_id"`
	IdentityFile string `json:"identity_file"`
	Comment      string `json:"comment"`
	IsActive     bool   `json:"is_active"`
}

// url: /hosts  method: GET
func (h *HostHandler) getHosts(ctx *gin.Context) {
	hosts, err := h.HostService.Hosts()
	if err == pub.ErrHostSetEmpty {
		Error(ctx, err, http.StatusNotFound, nil)
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
	} else {
		ctx.IndentedJSON(http.StatusOK, hosts)
	}
}

// url: /hosts/pk/:id  method: GET
func (h *HostHandler) getHostByID(ctx *gin.Context) {
	ID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}

	host, err := h.HostService.Host(ID)

	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrHostNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, host)
}

// url: /hosts/pk/:id method: POST
func (h *HostHandler) updateHostByID(ctx *gin.Context) {
	host := h._getHostByID(ctx)
	if host == nil {
		return
	}
	var req postHostRequest
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if req.Hostname != "" {
		host.Hostname = req.Hostname
	}
	if req.Password != "" {
		host.Password = req.Password
	}
	if req.HostgroupID != 0 {
		host.HostgroupID = req.HostgroupID
	}
	if req.IdentityFile != "" {
		host.IdentityFile = req.IdentityFile
	}
	if req.Comment != "" {
		host.Comment = req.Comment
	}
	host.IsActive = req.IsActive
	err := h.HostService.UpdateHost(host.ID, host)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, nil)
	} else {
		ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Update host success"})
	}
}

type postHostRequest struct {
	ID           uint64
	Hostname     string `json:"hostname"`
	Password     string `json:"password,omitempty"`
	HostgroupID  uint64 `json:"hostgroup_id"`
	IdentityFile string `json:"identity_file,omitempty"`
	Comment      string `json:"comment"`
	IsActive     bool   `json:"is_active"`
}

func (h *HostHandler) _getHostByID(ctx *gin.Context) *pub.Host {
	ID := getID(ctx)
	if ID == 0 {
		return nil
	}
	host, err := h.HostService.Host(ID)
	if err == nil {
		return host
	} else if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrHostNotFound, http.StatusNotFound, nil)
	} else {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
	}
	return nil
}

// url: /hosts/pk/:id method: DELETE
func (h *HostHandler) deleteHostByID(ctx *gin.Context) {
	host := h._getHostByID(ctx)
	if host == nil {
		return
	}
	err := h.HostService.DeleteHost(host.ID)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, h.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "Delete host success"})
}
