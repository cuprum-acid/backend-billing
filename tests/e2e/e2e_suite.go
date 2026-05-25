// Package e2e provides end-to-end tests for the billing API.
// These tests verify the complete workflow from a user perspective.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Suite is the main test suite for E2E testing.
type Suite struct {
	suite.Suite
	apiProcess         *exec.Cmd
	apiURL             string
	pgContainer        testcontainers.Container
	pgConnectionString string
	ctx                context.Context
	cancel             context.CancelFunc
}

// SetupSuite runs once before all tests.
func (s *Suite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	// Skip if running in short mode
	if testing.Short() {
		s.T().Skip("Skipping E2E tests in short mode")
	}

	// Check if E2E flag is set
	if !isE2EEnabled() {
		s.T().Skip("E2E tests disabled. Run with -e2e flag")
	}

	// Start PostgreSQL container
	s.startPostgreSQL()

	// Start API process
	s.startAPI()

	// Wait for API to be ready
	s.waitForAPI()
}

// TearDownSuite runs once after all tests.
func (s *Suite) TearDownSuite() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.apiProcess != nil && s.apiProcess.Process != nil {
		_ = s.apiProcess.Process.Kill()
	}

	if s.pgContainer != nil {
		_ = s.pgContainer.Terminate(context.Background())
	}
}

func isE2EEnabled() bool {
	for _, arg := range os.Args {
		if arg == "-e2e" || arg == "--e2e" {
			return true
		}
	}
	return false
}

func (s *Suite) startPostgreSQL() {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "password",
			"POSTGRES_DB":       "billing_e2e",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(s.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err, "Failed to start PostgreSQL container")
	s.pgContainer = container

	// Get connection string
	host, err := container.Host(s.ctx)
	s.Require().NoError(err)
	port, err := container.MappedPort(s.ctx, "5432")
	s.Require().NoError(err)

	s.pgConnectionString = fmt.Sprintf(
		"host=%s user=postgres password=password dbname=billing_e2e port=%s sslmode=disable",
		host, port.Port(),
	)
}

func (s *Suite) startAPI() {
	// Build the API first
	s.T().Log("Building API binary...")
	buildCmd := exec.Command("go", "build", "-o", "billing-api-e2e", ".")
	buildCmd.Dir = "../.."
	buildOutput, err := buildCmd.CombinedOutput()
	s.Require().NoError(err, "Failed to build API: %s", string(buildOutput))

	// Start API process
	s.apiProcess = exec.Command("../../billing-api-e2e")
	s.apiProcess.Env = append(os.Environ(),
		fmt.Sprintf("DATABASE_URL=%s", s.pgConnectionString),
		"PORT=8181",
	)

	var out bytes.Buffer
	s.apiProcess.Stdout = &out
	s.apiProcess.Stderr = &out

	err = s.apiProcess.Start()
	s.Require().NoError(err, "Failed to start API process")

	s.apiURL = "http://localhost:8181"
	s.T().Log("API started on", s.apiURL)
}

func (s *Suite) waitForAPI() {
	s.T().Log("Waiting for API to be ready...")

	for i := 0; i < 30; i++ {
		resp, err := http.Get(s.apiURL + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			_ = resp.Body.Close()
			s.T().Log("API is ready")
			return
		}
		time.Sleep(1 * time.Second)
	}

	s.Fail("API did not become ready in time")
}

// TestE2E runs the E2E test suite.
func TestE2E(t *testing.T) {
	suite.Run(t, new(Suite))
}

// Helper methods

func (s *Suite) makeRequest(method, path string, body interface{}) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, s.apiURL+path, reqBody)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	s.Require().NoError(err)

	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	return resp, respBody
}

func (s *Suite) decodeJSON(data []byte, target interface{}) {
	err := json.Unmarshal(data, target)
	s.Require().NoError(err)
}
