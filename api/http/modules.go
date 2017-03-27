package http

import (
	"net/http"
	"strconv"

	"github.com/fengxsong/pubmgmt/api"
	"gopkg.in/gin-gonic/gin.v1"
)

type ModuleHandler struct {
	Logger        logger
	ModuleService pub.ModuleService
}

func (m *ModuleHandler) createSvnInfo(ctx *gin.Context) {
	var req pub.SubversionInfo
	if err := ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	err := m.ModuleService.CreateSvnInfo(&req)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusCreated, &msgResponse{Msg: "create subversion info success"})
}

func (m *ModuleHandler) getSvnByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	svnInfo, err := m.ModuleService.SvnByID(id)
	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrSvnInfoNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, svnInfo)
}

func (m *ModuleHandler) getSvnInfos(ctx *gin.Context) {
	svnInfos, err := m.ModuleService.SvnInfos()
	if err == pub.ErrModelSetEmpty {
		Error(ctx, pub.ErrSvnInfoSetEmpty, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &svnInfos)
}

func (m *ModuleHandler) updateSvnByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	var req pub.SubversionInfo
	if err = ctx.BindJSON(&req); err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if req.ID == 0 || req.ID != id {
		Error(ctx, errIDField, http.StatusBadRequest, nil)
		return
	}
	_, err = m.ModuleService.SvnByID(req.ID)
	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrSvnInfoNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	err = m.ModuleService.UpdateSvnInfo(req.ID, &req)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "update subversion info success"})
}

func (m *ModuleHandler) deleteSvnByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		Error(ctx, err, http.StatusBadRequest, nil)
		return
	}
	_, err = m.ModuleService.SvnByID(id)
	if err == pub.ErrObjNotFound {
		Error(ctx, pub.ErrSvnInfoNotFound, http.StatusNotFound, nil)
		return
	} else if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	err = m.ModuleService.DeleteSvnInfo(id)
	if err != nil {
		Error(ctx, err, http.StatusInternalServerError, m.Logger)
		return
	}
	ctx.IndentedJSON(http.StatusOK, &msgResponse{Msg: "delete subversion info success"})
}
