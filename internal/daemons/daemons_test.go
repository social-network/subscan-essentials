package daemons_test

import (
	"github.com/social-network/netscan/internal/daemons"
	"github.com/sevlyar/go-daemon"
	"testing"
)

func TestRun(t *testing.T) {
	daemons.Run("demo", "start")
	if len(daemon.Flags()) != 1 {
		t.Errorf("Run Daemon failed")
	}
}
