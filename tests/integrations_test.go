package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/DANDA322/balance-service/internal"
	"github.com/DANDA322/balance-service/internal/pgstore"
	"github.com/DANDA322/balance-service/internal/rest"
	"github.com/DANDA322/balance-service/pkg/logging"
	_ "github.com/jackc/pgx/v5/stdlib"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	addr = ":3333"
)

type IntegrationTestSuite struct {
	suite.Suite
	log     *logrus.Logger
	store   *pgstore.DB
	service *internal.App
	server  *http.Server
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.log = logging.GetLogger("true")
	ctx := context.Background()
	var err error
	s.store, err = pgstore.GetPGStore(ctx, s.log, "postgres://postgres:secret@localhost:5433/postgres")
	require.NoError(s.T(), err)
	err = s.store.Migrate(migrate.Down)
	require.NoError(s.T(), err)
	err = s.store.Migrate(migrate.Up)
	require.NoError(s.T(), err)
	s.service = internal.NewApp(s.log, s.store)
	router := rest.NewRouter(s.log, s.service)
	s.server = &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Second * 30,
	}
	go func() {
		_ = s.server.ListenAndServe()
	}()
	time.Sleep(100 * time.Millisecond)
}

func (s *IntegrationTestSuite) SetupTest() {
	err := s.store.Migrate(migrate.Down)
	require.NoError(s.T(), err)
	err = s.store.Migrate(migrate.Up)
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	_ = s.server.Shutdown(context.Background())
	err := s.store.Migrate(migrate.Down)
	require.NoError(s.T(), err)
	err = s.store.Migrate(migrate.Up)
	require.NoError(s.T(), err)
}

func (s *IntegrationTestSuite) processRequest(method, path, token string, body interface{}) ([]byte, int, error) {
	requestBody, err := json.Marshal(body)
	require.NoError(s.T(), err)
	path = fmt.Sprintf("http://localhost%s%s", addr, path)
	req, err := http.NewRequestWithContext(context.Background(), method, path, bytes.NewReader(requestBody))
	require.NoError(s.T(), err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(s.T(), err)
	return responseBody, resp.StatusCode, err
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
