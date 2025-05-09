package daraja_test

import (
	"testing"
	"time"

	"github.com/SirWaithaka/payments-api/service/daraja"
)

func TestCache_Get(t *testing.T) {

	t.Run("test that empty string is returned when cache is empty", func(t *testing.T) {
		cache := daraja.NewCache[string]()

		if cache.Get() != "" {
			t.Errorf("expected empty string, got %s", cache.Get())
		}
	})

	t.Run("test that empty string is returned when cache is expired", func(t *testing.T) {
		cache := daraja.NewCache[string]()
		// set cache expiry to 10 seconds ago
		cache.Set("fake_value", time.Now().Add(-time.Second*10))

		if cache.Get() != "" {
			t.Errorf("expected empty string, got %s", cache.Get())
		}
	})

	t.Run("test that correct value is returned when cache is not expired or empty", func(t *testing.T) {
		cache := daraja.NewCache[string]()
		// set cache expiry to 10 seconds from now
		cache.Set("fake_value", time.Now().Add(time.Second*10))

		if cache.Get() != "fake_value" {
			t.Errorf("expected fake_value, got %s", cache.Get())
		}
	})
}
