package host_test

import (
	"testing"

	"xiaosong372089396/learning-Restful-API-HTTP-Demo/apps/host"

	"github.com/stretchr/testify/assert"
)

func TestHostUpdate(t *testing.T) {
	should := assert.New(t)

	h := host.NewDefaultHost()
	patch := host.NewDefaultHost()
	patch.Name = "patch01"

	err := h.Patch(patch.Resource, patch.Describe)
	if should.NoError(err) {
		should.Equal(patch.Name, h.Name)
	}
}
