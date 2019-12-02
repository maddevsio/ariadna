package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	c, err := Get()
	require.NoError(t, err)
	assert.Equal(t, []string{"http://localhost:9200"}, c.ElasticURLs)
	assert.Equal(t, "addresses", c.ElasticIndex)
	os.Clearenv()
	os.Setenv("ELASTIC_INDEX", "override")
	c, err = Get()
	require.NoError(t, err)
	assert.Equal(t, "override", c.ElasticIndex)
}
