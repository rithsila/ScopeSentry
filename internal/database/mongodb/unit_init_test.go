//go:build scopesentry_unit

package mongodb

import "testing"

func TestUnitMongoBootstrap(t *testing.T) {
	if Client != nil || DB != nil || IsConnected {
		t.Fatalf("unit init must return before mongo.Connect: Client=%v DB=%v IsConnected=%v", Client != nil, DB != nil, IsConnected)
	}
	if EnsureConnected() {
		t.Fatal("uninitialized unit client must not connect")
	}
}
