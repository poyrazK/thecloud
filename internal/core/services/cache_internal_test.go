package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheService_InternalParsers(t *testing.T) {
	t.Run("parseAllocatedPort", func(t *testing.T) {
		s := &CacheService{}
		port, err := s.parseAllocatedPort([]string{"30001:6379"}, "6379")
		assert.NoError(t, err)
		assert.Equal(t, 30001, port)

		port, _ = s.parseAllocatedPort([]string{"80:80"}, "6379")
		assert.Equal(t, 0, port)
	})

	t.Run("parseRedisClients", func(t *testing.T) {
		info := "# Clients\r\nconnected_clients:10\r\nclient_longest_output_list:0\r\n"
		clients := parseRedisClients(info)
		assert.Equal(t, 10, clients)
	})

	t.Run("parseRedisKeys", func(t *testing.T) {
		info := "# Keyspace\r\ndb0:keys=100,expires=0,avg_ttl=0\r\ndb1:keys=50,expires=5,avg_ttl=100\r\n"
		keys := parseRedisKeys(info)
		assert.Equal(t, int64(150), keys)
	})
}
