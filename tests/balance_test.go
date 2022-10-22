package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/stretchr/testify/require"
)

const (
	token1      = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ" //nolint:lll,gosec
	token2      = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjozMzMsInJvbGUiOiJhZG1pbiJ9.SxZaQcVFDVSb72MPMLbGPE5-23s-FZO9Lgip2oS13vwKy9f5Qe0L_xrCtWQbrAodlFphwmF-dCTd59hAaahcoNzN1Jgj0b15NJBKDcQgZDhQN8jehXrDrFdfj2UUi9y3KpHfRtepBDPiMXNCUd5zaY_3eW5ilbBtUj8GDchN0SiRyg9d3v4THvk21o3CDWRwLe8exKTdP7KCfuGeqLG8315aMSIuOUCNw25m-JKzQUYlgeaxQDK0d6DDitogBy0WYI77KZXVK5M5r-tYWj9aIcy7pCk2jCZ-NkuL5ekLbYfI5NHzNbF3sJUdxE4GkIx2x4LrX38lJvZe80bH0aQIMQ" //nolint:lll,gosec
	dateTimeFmt = "2006-01-02T15:04:05Z"
)

var transaction1 = &models.Transaction{
	Amount:  100.50,
	Comment: "Пополнение баланса",
}

var transaction2 = &models.Transaction{
	Amount:  100.50,
	Comment: "Снятие средств",
}

var transaction3 = &models.Transaction{
	Amount:  10000.0,
	Comment: "Снятие средств",
}

var transaction4 = &models.Transaction{
	Amount:  50,
	Comment: "Пополнение баланса",
}

var transaction5 = &models.Transaction{
	Amount:  1000.50,
	Comment: "Пополнение баланса",
}

var balance0 = &models.Balance{
	Currency: "RUB",
	Amount:   0,
}

var balance1 = &models.Balance{
	Currency: "RUB",
	Amount:   100.5,
}

var balance2 = &models.Balance{
	Currency: "RUB",
	Amount:   201,
}

var transferTransaction = &models.TransferTransaction{
	Target:  333,
	Amount:  100.5,
	Comment: "Перевод",
}

var transactions = []models.TransactionFullInfo{
	{
		ID:             3,
		WalletID:       1,
		Amount:         transaction2.Amount,
		TargetWalletID: nil,
		Comment:        transaction2.Comment,
		Timestamp:      time.Now().UTC(),
	},
	{
		ID:             2,
		WalletID:       1,
		Amount:         transaction5.Amount,
		TargetWalletID: nil,
		Comment:        transaction5.Comment,
		Timestamp:      time.Now().UTC(),
	},
	{
		ID:             1,
		WalletID:       1,
		Amount:         transaction1.Amount,
		TargetWalletID: nil,
		Comment:        transaction1.Comment,
		Timestamp:      time.Now().UTC(),
	},
}

var reserveTransaction = &models.ReserveTransaction{
	AccountID: 555,
	ServiceID: 1,
	OrderID:   111,
	Amount:    100.5,
}

var reserveTransaction2 = &models.ReserveTransaction{
	AccountID: 555,
	ServiceID: 1,
	OrderID:   111,
	Amount:    5000,
}

var reserveTransaction3 = &models.ReserveTransaction{
	AccountID: 555,
	ServiceID: 1,
	OrderID:   222,
	Amount:    100.5,
}

var csvText = "ServiceTitle;Amount\nУслуга;201\n"

func (s *IntegrationTestSuite) TestNotFound() {
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/blabla", token1, "")
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "", string(resp))
}

func (s *IntegrationTestSuite) TestAddDeposit() {
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/addDeposit", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, balance1)
}

func (s *IntegrationTestSuite) TestAddDepositWrongJSON() {
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/addDeposit", token1, transactions)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, code)
	require.Equal(s.T(), "{\"error\":\"Can't decode json\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalance() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest(http.MethodGet, "/wallet/getBalance", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance1.Amount, respStruct.Amount)
	require.Equal(s.T(), balance1.Currency, respStruct.Currency)
}

func (s *IntegrationTestSuite) TestGetBalanceNotFound() {
	resp, code, err := s.processRequest(http.MethodGet, "/wallet/getBalance", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetBalanceNotAuth() {
	resp, code, err := s.processRequest(http.MethodGet, "/wallet/getBalance", "", nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusUnauthorized, code)
	require.Equal(s.T(), "{\"error\":\"Unauthorized\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/withdrawMoney", token1, transaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, balance0)
}

func (s *IntegrationTestSuite) TestWithdrawMoneyWrongJSON() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/withdrawMoney", token1, transactions)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusBadRequest, code)
	require.Equal(s.T(), "{\"error\":\"Can't decode json\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoneyNotFound() {
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/withdrawMoney", token1, transaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestWithdrawMoneyNotEnoughMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/withdrawMoney", token1, transaction3)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough money on the balance\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestTransferMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	depositMoney(s.T(), s, token2, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/transferMoney", token1, transferTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, balance0)
	checkBalance(s.T(), s, token2, balance2)
}

func (s *IntegrationTestSuite) TestTransferMoneyWalletNotFound() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/transferMoney", token1, transferTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestTransferMoneyNotEnoughMoney() {
	depositMoney(s.T(), s, token1, transaction4)
	depositMoney(s.T(), s, token2, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/transferMoney", token1, transferTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough money on the balance\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetWalletTransactionsSortByDateDesc() {
	depositMoney(s.T(), s, token1, transaction1)
	depositMoney(s.T(), s, token1, transaction5)
	withdrawMoney(s.T(), s, token1, transaction2)
	from := time.Now()
	to := from.Add(time.Hour * 24)
	resp, code, err := s.processRequest(http.MethodGet, "/wallet/getTransactions?from="+from.Format(dateTimeFmt)+
		"&to="+to.Format(dateTimeFmt)+"&limit=10&offset=0&descending=true&sorting=date", token1, nil)
	var respStruct []models.TransactionFullInfo
	require.NoError(s.T(), err)
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	compareTransactions(s.T(), transactions, respStruct)
}

func (s *IntegrationTestSuite) TestGetWalletTransactionsWalletNotFound() {
	from := time.Now()
	to := from.Add(time.Hour * 24)
	resp, code, err := s.processRequest(http.MethodGet, "/wallet/getTransactions?from="+from.Format(dateTimeFmt)+
		"&to="+to.Format(dateTimeFmt)+"&limit=10&offset=0&descending=true&sorting=amount", token1, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestReserveMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/reserveMoney", token1, reserveTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, balance0)
}

func (s *IntegrationTestSuite) TestReserveMoneyWalletNotFound() {
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/reserveMoney", token1, reserveTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestReserveMoneyNotEnoughMoney() {
	depositMoney(s.T(), s, token1, transaction4)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/reserveMoney", token1, reserveTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough money on the balance\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestApplyReserveMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	reserveMoney(s.T(), s, token1, reserveTransaction)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/applyReserve", token1, reserveTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestApplyReserveMoneyWalletNotFound() {
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/applyReserve", token1, reserveTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"wallet not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestApplyReserveNotEnoughMoney() {
	depositMoney(s.T(), s, token1, transaction1)
	reserveMoney(s.T(), s, token1, reserveTransaction)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/applyReserve", token1, reserveTransaction2)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusConflict, code)
	require.Equal(s.T(), "{\"error\":\"not enough reserved money on the balance\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestApplyReserveOrderNotFound() {
	depositMoney(s.T(), s, token1, transaction1)
	reserveMoney(s.T(), s, token1, reserveTransaction)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/applyReserve", token1, reserveTransaction3)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusNotFound, code)
	require.Equal(s.T(), "{\"error\":\"reserved order not found\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestCancelReserve() {
	depositMoney(s.T(), s, token1, transaction1)
	reserveMoney(s.T(), s, token1, reserveTransaction)
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/cancelReserve", token1, reserveTransaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func (s *IntegrationTestSuite) TestGetReport() {
	depositMoney(s.T(), s, token1, transaction1)
	depositMoney(s.T(), s, token1, transaction1)
	reserveMoney(s.T(), s, token1, reserveTransaction)
	reserveMoney(s.T(), s, token1, reserveTransaction3)
	applyMoney(s.T(), s, token1, reserveTransaction)
	applyMoney(s.T(), s, token1, reserveTransaction3)
	path := fmt.Sprintf("http://localhost%s%s", addr, "/wallet/getReport?month="+time.Now().Format("2006-01"))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, path, nil)
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+token1)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, resp.StatusCode)
	require.Equal(s.T(), "text/csv", resp.Header.Get("Content-Type"))
	require.Equal(s.T(), "attachment; filename=\"report.csv\"", resp.Header.Get("Content-Disposition"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	require.Equal(s.T(), csvText, string(body))
	err = resp.Body.Close()
	require.NoError(s.T(), err)
}

func checkBalance(t *testing.T, s *IntegrationTestSuite, token string, balance *models.Balance) {
	t.Helper()
	resp, code, err := s.processRequest(http.MethodGet, "/wallet/getBalance", token, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance.Amount, respStruct.Amount)
	require.Equal(s.T(), balance.Currency, respStruct.Currency)
}

func depositMoney(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.Transaction) {
	t.Helper()
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/addDeposit", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func withdrawMoney(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.Transaction) {
	t.Helper()
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/withdrawMoney", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func reserveMoney(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.ReserveTransaction) {
	t.Helper()
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/reserveMoney", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func applyMoney(t *testing.T, s *IntegrationTestSuite, token string, transaction *models.ReserveTransaction) {
	t.Helper()
	resp, code, err := s.processRequest(http.MethodPost, "/wallet/applyReserve", token, transaction)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
}

func compareTransactions(t *testing.T, expected, actual []models.TransactionFullInfo) {
	t.Helper()
	for index, element := range actual {
		require.Equal(t, element.ID, expected[index].ID)
		require.Equal(t, element.WalletID, expected[index].WalletID)
		require.Equal(t, element.Amount, expected[index].Amount)
		require.Equal(t, element.ServiceID, expected[index].ServiceID)
		require.Equal(t, element.TargetWalletID, expected[index].TargetWalletID)
		require.Equal(t, element.Timestamp.Truncate(time.Second), expected[index].Timestamp.Truncate(time.Second))
	}
}
