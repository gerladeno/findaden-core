package rest

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gerladeno/homie-core/internal"

	"github.com/go-chi/chi/v5"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/pkg/common"

	"github.com/sirupsen/logrus"
)

type handler struct {
	log     *logrus.Entry
	service Service
	key     *rsa.PublicKey
}

const defaultLimit = 10

func newHandler(log *logrus.Logger, service Service, key *rsa.PublicKey) *handler {
	return &handler{
		log:     log.WithField("module", "rest"),
		service: service,
		key:     key,
	}
}

func (h *handler) saveConfig(w http.ResponseWriter, r *http.Request) {
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	var config models.Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil || config.Personal.Gender == models.Any {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	config.SetUUID(uuid)
	if err := h.service.SaveConfig(r.Context(), &config); err != nil {
		h.log.Warnf("err saving config: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, "Ok")
}

func (h *handler) getConfig(w http.ResponseWriter, r *http.Request) {
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	config, err := h.service.GetConfig(r.Context(), uuid)
	switch {
	case err == nil:
	case errors.Is(err, internal.ErrConfigNotFound):
		writeErrResponse(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
		return
	}
	if err != nil {
		h.log.Warnf("err getting config: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, config)
}

func (h *handler) getMatches(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query().Get("count")
	count, _ := strconv.ParseInt(val, 10, 64)
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	result, err := h.service.GetMatches(r.Context(), uuid, count)
	if err != nil {
		h.log.Warnf("err getting config: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, result)
}

func (h *handler) like(w http.ResponseWriter, r *http.Request) {
	val := r.URL.Query().Get("super")
	super, err := strconv.ParseBool(val)
	if err != nil {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	targetUUID := chi.URLParam(r, "uuid")
	if !common.IsValidUUID(targetUUID) {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	if err = h.service.Like(r.Context(), uuid, targetUUID, super); err != nil {
		h.log.Warnf("err liking: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, "Ok")
}

func (h *handler) dislike(w http.ResponseWriter, r *http.Request) {
	targetUUID := chi.URLParam(r, "uuid")
	if !common.IsValidUUID(targetUUID) {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	if err := h.service.Dislike(r.Context(), uuid, targetUUID); err != nil {
		h.log.Warnf("err disliking: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, "Ok")
}

func (h *handler) listLiked(w http.ResponseWriter, r *http.Request) {
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	limit, offset := h.limitOffset(w, r)
	result, err := h.service.ListLikedProfiles(r.Context(), uuid, limit, offset)
	if err != nil {
		h.log.Warnf("err listing liked: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, result)
}

func (h *handler) listDisliked(w http.ResponseWriter, r *http.Request) {
	uuid, ok := h.getUUID(w, r)
	if !ok {
		return
	}
	limit, offset := h.limitOffset(w, r)
	result, err := h.service.ListDislikedProfiles(r.Context(), uuid, limit, offset)
	if err != nil {
		h.log.Warnf("err listing disliked: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, result)
}

func (h *handler) getUUID(w http.ResponseWriter, r *http.Request) (string, bool) {
	uuid, ok := r.Context().Value(uuidKey).(string)
	if !ok {
		h.log.Warnf("err invalid context value")
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return "", false
	}
	if !common.IsValidUUID(uuid) {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return "", false
	}
	return uuid, ok
}

func (h *handler) limitOffset(w http.ResponseWriter, r *http.Request) (int64, int64) {
	val := r.URL.Query().Get("limit")
	limit, _ := strconv.ParseInt(val, 10, 64)
	if limit == 0 {
		limit = defaultLimit
	}
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	return limit, offset
}

func (h *handler) getRegions(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.GetRegions(r.Context())
	if err != nil {
		h.log.Warnf("err getting regions: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, result)
}
