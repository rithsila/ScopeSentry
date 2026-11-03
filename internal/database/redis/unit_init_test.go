//go:build scopesentry_unit

package redis

import "testing"

func TestUnitRedisBootstrap(t *testing.T) {
	if Client != nil {
		t.Fatal("unit init must return before redis.NewClient")
	}
}
