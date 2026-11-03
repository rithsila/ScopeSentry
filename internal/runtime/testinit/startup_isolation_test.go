//go:build scopesentry_unit

package testinit_test

import (
	"testing"

	"github.com/Autumn-27/ScopeSentry/internal/config"
	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/database/redis"
	"github.com/Autumn-27/ScopeSentry/internal/runtime/testinit"
)

func TestUnitStartupIsolation(t *testing.T) {
	if !testinit.Enabled {
		t.Fatal("scopesentry_unit bootstrap is not enabled")
	}
	want := config.Config{System: config.SystemConfig{Timezone: "UTC"}, Log: config.LogConfig{Level: "info"}}
	if config.GlobalConfig != want {
		t.Fatal("unit config must use deterministic safe defaults, ignoring sentinel environment")
	}
	if mongodb.Client != nil || mongodb.DB != nil || mongodb.IsConnected || redis.Client != nil {
		t.Fatal("unit startup created a database client or connected")
	}
}
