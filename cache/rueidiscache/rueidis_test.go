package rueidiscache

import (
	"testing"

	"github.com/redis/rueidis"
	"github.com/stretchr/testify/assert"
)

func TestNewRueidisClientReturnsConfigError(t *testing.T) {
	client, err := NewRueidisClient(&rueidis.ClientOption{})
	assert.Nil(t, client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no alive address")
}

func TestNewRueidisClientRejectsNilOption(t *testing.T) {
	client, err := NewRueidisClient(nil)
	assert.Nil(t, client)
	assert.Error(t, err)
}

func TestNewRueidisAsideClientRejectsNilOption(t *testing.T) {
	client, err := NewRueidisAsideClient(nil)
	assert.Nil(t, client)
	assert.Error(t, err)
}
