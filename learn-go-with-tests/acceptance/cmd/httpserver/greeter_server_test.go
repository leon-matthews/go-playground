package main_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	go_specs_greet "local.dev/go-specs-greet"
	"local.dev/go-specs-greet/specifications"
)

func TestGreeterServer(t *testing.T) {
	// Start/build Docker container
	ctx := context.Background()
	containerRequest := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "../../.",
			Dockerfile: "./cmd/httpserver/Dockerfile",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("8080"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerRequest,
		Started:          true,
	})
	testcontainers.CleanupContainer(t, container)
	assert.NoError(t, err)

	// Determine URL
	baseURL, err := container.PortEndpoint(ctx, "8080/tcp", "http")
	assert.NoError(t, err)

	// Actual test
	driver := go_specs_greet.Driver{BaseURL: baseURL}
	specifications.GreetSpecification(t, driver)
}
