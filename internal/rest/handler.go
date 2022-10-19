package rest

import (
	"crypto/rsa"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/sirupsen/logrus"
)

//go:embed public.pub
var publicSigningKey []byte

type handler struct {
	log     *logrus.Logger
	balance Balance
	pubKey  *rsa.PublicKey
}

func newHandler(log *logrus.Logger, balance Balance) *handler {
	return &handler{
		log:     log,
		balance: balance,
		pubKey:  musGetPublicKey(publicSigningKey),
	}
}

func (h *handler) Test(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	accountID := sessionInfo.AccountID
	h.log.Info(accountID)
	h.writeJSONResponse(w, accountID)
}

func (h *handler) GetBalance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	balance, err := h.balance.GetBalance(ctx, sessionInfo.AccountID)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	default:
		h.log.Errorf("Error get balance: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	result := models.Balance{
		Currency: "RUB",
		Amount:   balance,
	}
	h.writeJSONResponse(w, result)
}

func (h *handler) DepositMoneyToWallet(w http.ResponseWriter, r *http.Request) {
	transaction := models.Transaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	err := h.balance.AddDeposit(ctx, sessionInfo.AccountID, transaction)
	switch {
	case err == nil:
	default:
		h.log.Errorf("Error deposit money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) WithdrawMoneyFromWallet(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	transaction := models.Transaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	err := h.balance.WithdrawMoney(ctx, sessionInfo.AccountID, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughMoney.Error())
		return
	default:
		h.log.Errorf("Error withdraw money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) TransferMoney(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	transaction := models.TransferTransaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	err := h.balance.TransferMoney(ctx, sessionInfo.AccountID, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughMoney.Error())
		return
	default:
		h.log.Errorf("Error transfer money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}
