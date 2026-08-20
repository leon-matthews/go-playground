package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"
)

const (
	containerDBName         = "snippetbox_test"
	containerDBPassword     = "pass"
	containerDBRootPassword = "root"
	containerDBUser         = "test_web"
	containerImage          = "docker.io/library/mariadb:12"
	containerName           = "snippetbox-test-mariadb"
	containerPort           = 3307
)

// newTestDBContainer wraps newTestDB, starting container if required.
func newTestDBContainer(t *testing.T) *sql.DB {
	// Is everything okay? Let's go home early!
	db, err := newTestDB(t, 1*time.Second)
	if err == nil {
		return db
	}

	// Don't try starting container unless we timed-out waiting
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("connecting to database: %v", err)
	}

	// Connecting to database has timed out. Check if container exists.
	// If container exists, but isn't accepting DB connection we can't help much.
	if containerExists(t, containerName) {
		t.Fatalf("Container %q exists but connection to the database failed: %v", containerName, err)
	}

	// Run command to start container
	startDBContainer(t)

	// Wait for database to start accepting connections
	db, err = newTestDB(t, 10*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

// newTestDB connects to a test-only database, returning a connection pool.
// Setup and teardown SQL scripts are used to create and drop tables
// Returns only database connection errors, calls t.Fatal() for anything else.
func newTestDB(t *testing.T, timeout time.Duration) (*sql.DB, error) {
	testDSN := fmt.Sprintf(
		"%s:%s@tcp(127.0.0.1:%d)/%s?parseTime=true&multiStatements=true",
		containerDBUser,
		containerDBPassword,
		containerPort,
		containerDBName,
	)

	ctx, cancel := context.WithTimeout(t.Context(), timeout)
	defer cancel()
	db, err := OpenDB(ctx, testDSN)
	if err != nil {
		return nil, err
	}

	// Execute statements in our setup SQL script
	script, err := os.ReadFile("./testdata/setup.sql")
	if err != nil {
		db.Close()
		t.Fatalf("Error reading setup script: %v", err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		db.Close()
		t.Fatalf("Error running setup script: %v", err)
	}

	// Run teardown SQL script when test has finished
	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("./testdata/teardown.sql")
		if err != nil {
			t.Errorf("Error reading teardown script: %v", err)
			return
		}
		_, err = db.Exec(string(script))
		if err != nil {
			t.Errorf("Error running teardown script: %v", err)
			return
		}
	})

	return db, nil
}

// containerExists returns true if we can verify that a container by that name exists
func containerExists(t *testing.T, name string) bool {
	ctx, cancel := context.WithTimeout(t.Context(), 1*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "podman", "container", "exists", name)
	t.Logf("Running command: %q", cmd.String())
	out, err := cmd.CombinedOutput()

	// Container found?
	if err == nil {
		return true
	}

	// Podman not installed?
	if errors.Is(err, exec.ErrNotFound) {
		t.Fatal("Podman command not found, please install to run integration tests.")
	}

	// Exit code 125 is for local storage error
	if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
		if exitErr.ExitCode() == 125 {
			t.Fatalf("Podman 'container exists' command reports local storage error: %s", string(out))
		}
	}

	return false
}

// startDBContainer executes podman run for our MariaDB container
func startDBContainer(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(
		ctx,
		"podman", "run", "--rm", "--detach",
		"--name", containerName,
		"--publish", fmt.Sprintf("%d:3306", containerPort),
		"--tmpfs", "/var/lib/mysql",
		"--env", "MARIADB_ROOT_PASSWORD="+containerDBRootPassword,
		"--env", "MARIADB_DATABASE="+containerDBName,
		"--env", "MARIADB_USER="+containerDBUser,
		"--env", "MARIADB_PASSWORD="+containerDBPassword,
		containerImage,
	)
	t.Logf("Running command: %q", cmd.String())

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Starting DB container: %s\n%s", err, string(out))
	}
}
