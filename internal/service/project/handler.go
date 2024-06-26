package project

import (
	"net/http"
	"strconv"

	"github.com/arthurlee945/Docrilla/internal/errors"
	"github.com/arthurlee945/Docrilla/internal/model"
	"github.com/arthurlee945/Docrilla/internal/util/json"
)

func RegisterHandler(router *http.ServeMux, service Service) {
	handler := NewHandler(service)
	router.HandleFunc("GET /projects", handler.GetAll)
	router.HandleFunc("GET /projects/{id}/overview", handler.GetOverviewById)
	router.HandleFunc("GET /projects/{id}/detail", handler.GetDetailById)
	router.HandleFunc("POST /projects", handler.Create)
	router.HandleFunc("PUT /projects", handler.Update)
	router.HandleFunc("DELETE /projects/{id}", handler.Delete)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{
		service: s,
	}
}

type GetAllResponse struct {
	Projects []model.Project `json:"projects"`
	Cursor   string          `json:"cursor"`
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	limit, cursor := r.FormValue("limit"), r.FormValue("cursor")
	var limitToPass uint8
	if intLimit, err := strconv.Atoi(limit); err == nil && intLimit > 0 {
		limitToPass = uint8(intLimit)
	} else {
		limitToPass = 10
	}
	projs, cursor, err := h.service.GetAll(r.Context(), GetAllRequest{limitToPass, cursor})

	if err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	if err := json.Encode(w, http.StatusOK, GetAllResponse{
		Projects: projs,
		Cursor:   cursor,
	}); err != nil {
		errors.ServerError(r.Context(), w, err)
	}
}

func (h *Handler) GetOverviewById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	proj, err := h.service.GetOverviewById(r.Context(), id)
	if err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	if err := json.Encode(w, http.StatusOK, proj); err != nil {
		errors.ServerError(r.Context(), w, err)
	}
}

func (h *Handler) GetDetailById(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	proj, err := h.service.GetDetailById(r.Context(), id)
	if err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	if err := json.Encode(w, http.StatusOK, proj); err != nil {
		errors.ServerError(r.Context(), w, err)
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	reqObj := &CreateRequest{}
	if err := json.Decode(r, reqObj); err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	proj, err := h.service.Create(r.Context(), *reqObj)
	if err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	if err := json.Encode(w, http.StatusCreated, proj); err != nil {
		errors.ServerError(r.Context(), w, err)
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	reqObj := &UpdateRequest{}
	if err := json.Decode(r, reqObj); err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	proj, err := h.service.Update(r.Context(), *reqObj)
	if err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	if err := json.Encode(w, http.StatusAccepted, proj); err != nil {
		errors.ServerError(r.Context(), w, err)
	}
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := h.service.Delete(r.Context(), id)
	if err != nil {
		errors.ServerError(r.Context(), w, err)
		return
	}
	if err := json.Encode(w, http.StatusAccepted, struct{}{}); err != nil {
		errors.ServerError(r.Context(), w, err)
	}
}
