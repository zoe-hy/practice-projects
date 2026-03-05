package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.KafkaBroker == "" || cfg.KafkaTopic == "" || cfg.HTTPPort == "" || cfg.MySQLDSN == "" {
		t.Fatalf("expected defaults to be set, got %+v", cfg)
	}
}
