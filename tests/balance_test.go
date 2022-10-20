package tests

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/DANDA322/balance-service/internal/models"
	"github.com/stretchr/testify/require"
)

const (
	token1 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50X2lkIjo1NTUsInJvbGUiOiJhZG1pbiJ9.tD-jH7f6HzdnWMhyxuLzwomXDc4di3sAe9G2xldZ2lPYWAc4gcGifZyxdunBsNbwZk9VH5OBOV7MuozPFAuGhi9ZwTCt0F27kRMfSt70P5G8EzaqOR2pxxX8rgcui3ZUpE7AXbPaGd49sY94flV_oxFE9-ikuQrH018-qhMAwQ-dKS3lBwwDFtM9rF37iMJX7Omw52TcwpELL2ovQZOQVqNuqs6CZYzLZiTMXR3cBLSCymT7PDs0Rjdtkc5grmBdZVYUwOjzH5-Yjf8ctGBagu5aOTFd2tOAxkmc64xPU-VnmfoG7EkwXLYE9dmlsvQTqRabviWSUoin7Y-XsLSofQ" //nolint:lll,gosec
)

var transaction1 = &models.Transaction{
	Amount:  100.50,
	Comment: "Пополнение баланса",
}

var balance1 = &models.Balance{
	Currency: "RUB",
	Amount:   100.5,
}

func (s *IntegrationTestSuite) TestAddDeposit() {
	resp, code, err := s.processRequest("POST", "/wallet/addDeposit", token1, transaction1)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	require.Equal(s.T(), "{\"response\":\"OK\"}\n", string(resp))
	checkBalance(s.T(), s, token1, balance1)
}

func checkBalance(t *testing.T, s *IntegrationTestSuite, token string, balance *models.Balance) {
	t.Helper()
	resp, code, err := s.processRequest("GET", "/wallet/getBalance", token, nil)
	require.NoError(s.T(), err)
	require.Equal(s.T(), http.StatusOK, code)
	respStruct := models.Balance{}
	err = json.Unmarshal(resp, &respStruct)
	require.NoError(s.T(), err)
	require.Equal(s.T(), balance.Amount, respStruct.Amount)
	require.Equal(s.T(), balance.Currency, respStruct.Currency)
}
