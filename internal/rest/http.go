package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

type Balance interface {
	AddDeposit(ctx context.Context, accountID int, transaction models.Transaction) error
	GetBalance(ctx context.Context, accountID int) (float64, error)
	WithdrawMoney(ctx context.Context, accountID int, transaction models.Transaction) error
	TransferMoney(ctx context.Context, accountID int, transaction models.TransferTransaction) error
	ReserveMoney(ctx context.Context, transaction models.ReserveTransaction) error
	ApplyReservedMoney(ctx context.Context, transaction models.ReserveTransaction) error
	GetWalletTransaction(ctx context.Context, accountID int,
		queryParams *models.TransactionsQueryParams) ([]models.TransactionFullInfo, error)
	CancelReserve(ctx context.Context, transaction models.ReserveTransaction) error
	GetReport(ctx context.Context, month time.Time) (*os.File, error)
}

func NewRouter(log *logrus.Logger, balance Balance) chi.Router {
	handler := newHandler(log, balance)
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(cors.AllowAll().Handler)
	r.NotFound(notFoundHandler)
	r.Route("/wallet", func(r chi.Router) {
		r.Use(handler.auth)
		r.Get("/getBalance", handler.GetBalance)
		r.Get("/getTransactions", handler.GetWalletTransactions)
		r.Get("/getReport", handler.GetReport)
		r.Post("/addDeposit", handler.DepositMoneyToWallet)
		r.Post("/withdrawMoney", handler.WithdrawMoneyFromWallet)
		r.Post("/transferMoney", handler.TransferMoney)
		r.Post("/reserveMoney", handler.ReserveMoney)
		r.Post("/applyReserve", handler.ApplyReservedMoney)
		r.Post("/cancelReserve", handler.CancelReserve)
	})

	return r
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func (h *handler) writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.log.Errorf("unable to encode data %v", err)
	}
}

func (h *handler) writeCSVResponse(w http.ResponseWriter, file *os.File) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report.csv\"")
	fileResponse, err := ioutil.ReadFile(file.Name())
	if err != nil {
		h.writeErrResponse(w, http.StatusInternalServerError, fmt.Sprintf("Cannot open file %v", err))
	}
	_, err = w.Write(fileResponse)
	if err != nil {
		h.log.Errorf("err to write report: %v", err)
	}
}

func (h *handler) writeErrResponse(w http.ResponseWriter, code int, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if newErr := json.NewEncoder(w).Encode(map[string]interface{}{"error": err}); newErr != nil {
		h.log.Errorf("unable to encode %v", newErr)
	}
}
