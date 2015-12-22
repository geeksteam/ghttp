package api

import (
	"fmt"
	"testing"

	"github.com/geekbros/SHM-Backend/core/api"
)

func TestRead(t *testing.T) {
	t.SkipNow()
	api.Trigger.Read()
	fmt.Printf("%+v", api.Trigger.Triggers)
}
