package generator

import (
	"context"
	"testing"
)

func Test_AzAppConf_succeeds(t *testing.T) {
	c, _ := NewAzAppConf(context.TODO(), "AZAPPCONF://configmanager-app-config/queue_name", *(NewConfig().WithTokenSeparator("://")))
	if c == nil {
		t.Errorf("got %v, wanted: %v", c, AzAppConf{})
	}
	// got, err := c.tokenVal()
}
