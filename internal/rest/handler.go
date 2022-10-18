package rest

import (
	"crypto/rsa"
	_ "embed"
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
