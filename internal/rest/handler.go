package rest

import (
	"crypto/rsa"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/sirupsen/logrus"
)

const (
	dateTimeFmt1 = "2006-01-02T15:04:05Z"
	dateTimeFmt2 = "2006-01"
	roleAdmin    = "admin"
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

func (h *handler) ReserveMoney(w http.ResponseWriter, r *http.Request) {
	transaction := models.ReserveTransaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	if sessionInfo.Role != roleAdmin && transaction.AccountID != sessionInfo.AccountID {
		h.writeErrResponse(w, http.StatusForbidden, "")
		return
	}
	err := h.balance.ReserveMoney(ctx, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughMoney.Error())
		return
	default:
		h.log.Errorf("Error reserve money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) ApplyReservedMoney(w http.ResponseWriter, r *http.Request) {
	transaction := models.ReserveTransaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	if sessionInfo.Role != roleAdmin && transaction.AccountID != sessionInfo.AccountID {
		h.writeErrResponse(w, http.StatusForbidden, "")
		return
	}
	err := h.balance.ApplyReservedMoney(ctx, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughReservedMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughReservedMoney.Error())
		return
	case errors.Is(err, models.ErrOrderNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrOrderNotFound.Error())
		return
	case errors.Is(err, models.ErrServiceNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrServiceNotFound.Error())
		return
	default:
		h.log.Errorf("Error recognize money: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) GetWalletTransactions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	from, err := h.parseTime(r.URL.Query().Get("from"), dateTimeFmt1)
	if err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't parse time")
		h.log.Info(err)
		return
	}
	to, err := h.parseTime(r.URL.Query().Get("to"), dateTimeFmt1)
	if err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't parse time")
		h.log.Info(err)
		return
	}
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		h.writeErrResponse(w, http.StatusInternalServerError, err)
		return
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, err)
		return
	}
	sorting := r.URL.Query().Get("sorting")
	descending := r.URL.Query().Get("descending")
	var transactions []models.TransactionFullInfo
	queryParams := models.TransactionsQueryParams{
		From:       from,
		To:         to,
		Limit:      limit,
		Offset:     offset,
		Sorting:    sorting,
		Descending: descending,
	}
	h.log.Info(queryParams)
	transactions, err = h.balance.GetWalletTransaction(ctx, sessionInfo.AccountID, &queryParams)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	default:
		h.log.Errorf("Error get wallet transactions: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, transactions)
}

func (h *handler) CancelReserve(w http.ResponseWriter, r *http.Request) {
	transaction := models.ReserveTransaction{}
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't decode json")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	sessionInfo := ctx.Value(SessionKey).(models.SessionInfo)
	if sessionInfo.Role != roleAdmin && transaction.AccountID != sessionInfo.AccountID {
		h.writeErrResponse(w, http.StatusForbidden, "")
		return
	}
	err := h.balance.CancelReserve(ctx, transaction)
	switch {
	case err == nil:
	case errors.Is(err, models.ErrWalletNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrWalletNotFound.Error())
		return
	case errors.Is(err, models.ErrNotEnoughReservedMoney):
		h.writeErrResponse(w, http.StatusConflict, models.ErrNotEnoughReservedMoney.Error())
		return
	case errors.Is(err, models.ErrOrderNotFound):
		h.writeErrResponse(w, http.StatusNotFound, models.ErrOrderNotFound.Error())
		return
	default:
		h.log.Errorf("Error cancel reserve: %v", err)
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
		return
	}
	h.writeJSONResponse(w, map[string]interface{}{"response": "OK"})
}

func (h *handler) GetReport(w http.ResponseWriter, r *http.Request) {
	month, err := h.parseTime(r.URL.Query().Get("month"), dateTimeFmt2)
	if err != nil {
		h.writeErrResponse(w, http.StatusBadRequest, "Can't parse time")
		h.log.Info(err)
		return
	}
	ctx := r.Context()
	file, err := h.balance.GetReport(ctx, month)
	if err != nil {
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal server error: %v", err))
	}
	h.writeCSVResponse(w, file)
}

func (h *handler) parseTime(s, layout string) (time.Time, error) {
	t, err := time.Parse(layout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("could not parse time: %w", err)
	}
	return t, err
}
