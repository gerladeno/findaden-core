package rest

import (
	"crypto/rsa"
	"encoding/json"
	"net/http"

	"github.com/gerladeno/homie-core/internal/models"
	"github.com/gerladeno/homie-core/pkg/common"

	"github.com/sirupsen/logrus"
)

type handler struct {
	log     *logrus.Entry
	service Service
	key     *rsa.PublicKey
}

func newHandler(log *logrus.Logger, service Service, key *rsa.PublicKey) *handler {
	return &handler{
		log:     log.WithField("module", "rest"),
		service: service,
		key:     key,
	}
}

func (h *handler) saveConfig(w http.ResponseWriter, r *http.Request) {
	uuid, ok := r.Context().Value(uuidKey).(string)
	if !ok {
		h.log.Warnf("err invalid context value")
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !common.IsValidUUID(uuid) {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	var config models.Config
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	config.UUID = uuid
	if err := h.service.SaveConfig(r.Context(), &config); err != nil {
		h.log.Warnf("err saving config: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, "Ok")
}

func (h *handler) getConfig(w http.ResponseWriter, r *http.Request) {
	uuid, ok := r.Context().Value(uuidKey).(string)
	if !ok {
		h.log.Warnf("err invalid context value")
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if !common.IsValidUUID(uuid) {
		writeErrResponse(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	config, err := h.service.GetConfig(r.Context(), uuid)
	if err != nil {
		h.log.Warnf("err getting config: %v", err)
		writeErrResponse(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	writeResponse(w, config)
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
