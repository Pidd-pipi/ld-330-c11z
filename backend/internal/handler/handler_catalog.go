package handler

import (
	"github.com/blueship581/gbemr/internal/dto"
	"github.com/blueship581/gbemr/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type CatalogHandler struct{ s *service.CatalogService }

func NewCatalogHandler(s *service.CatalogService) *CatalogHandler { return &CatalogHandler{s} }
func (h *CatalogHandler) CreateDrug(c *gin.Context) {
	var in dto.CatalogInput
	if !bindJSON(c, &in) {
		return
	}
	v, e := h.s.Drug(in)
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}
func (h *CatalogHandler) Drugs(c *gin.Context) {
	v, e := h.s.Drugs()
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}
func (h *CatalogHandler) CreateDiagnosis(c *gin.Context) {
	var in dto.CatalogInput
	if !bindJSON(c, &in) {
		return
	}
	v, e := h.s.Diagnosis(in)
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}
func (h *CatalogHandler) Diagnoses(c *gin.Context) {
	v, e := h.s.Diagnoses()
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}
func (h *CatalogHandler) CreateTemplate(c *gin.Context) {
	var in dto.CatalogInput
	if !bindJSON(c, &in) {
		return
	}
	v, e := h.s.Template(in)
	if e != nil {
		Fail(c, e)
		return
	}
	uid, n, _ := currentUser(c)
	h.s.Audit(uid, n, "create", "record_template", v.Name+" v"+strconv.Itoa(v.Version))
	OK(c, v)
}

// UpdateTemplate 发布模板新版本，历史版本保持不变。
func (h *CatalogHandler) UpdateTemplate(c *gin.Context) {
	id, e := idParam(c)
	if e != nil {
		Fail(c, e)
		return
	}
	var in dto.CatalogInput
	if !bindJSON(c, &in) {
		return
	}
	v, e := h.s.UpdateTemplate(id, in)
	if e != nil {
		Fail(c, e)
		return
	}
	uid, n, _ := currentUser(c)
	h.s.Audit(uid, n, "update", "record_template", v.Name+" v"+strconv.Itoa(v.Version))
	OK(c, v)
}

// Templates 返回所有版本（含历史内容），供管理端查看。
func (h *CatalogHandler) Templates(c *gin.Context) {
	v, e := h.s.Templates()
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}

// LatestTemplates 每个模板只返回最新版，供医生书写病历时选择。
func (h *CatalogHandler) LatestTemplates(c *gin.Context) {
	v, e := h.s.LatestTemplates()
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}
func (h *CatalogHandler) Audits(c *gin.Context) {
	v, e := h.s.Audits()
	if e != nil {
		Fail(c, e)
		return
	}
	OK(c, v)
}
