package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseKVSConnectionString(t *testing.T) {
	for _, s := range []string{"wrongstring", "scheme://wrongstring"} {
		_, err := parseKVSConnectionString(s)
		assert.EqualError(t, err, fmt.Sprintf("can't parse string %s into components", s))
	}

	s := "wrongscheme://somehost:1144/configs/myconf.json"
	_, err := parseKVSConnectionString(s)
	assert.EqualError(t, err, "wrongscheme is not correct key value store provider")

	s = "etcd://somehost:abc/configs/myconf.json"
	_, err = parseKVSConnectionString(s)
	assert.EqualError(t, err, "abc is not correct key value store port")

	s = "etcd://somehost:1144/configs/myconf"
	_, err = parseKVSConnectionString(s)
	assert.EqualError(t, err, "key value store format is not provided")

	s = "etcd://somehost:1144/configs/myconf.conf"
	_, err = parseKVSConnectionString(s)
	assert.EqualError(t, err, "conf is not correct config format")

	s = "etcd://somehost:1144/configs/my.conf.yml"
	params, err := parseKVSConnectionString(s)
	require.NoError(t, err)
	assert.Equal(t, &kvsParams{"etcd", "somehost", 1144, "/configs/my.conf", "yml"}, params)

	s = "etcd://somehost/configs/myconf.yml"
	params, err = parseKVSConnectionString(s)
	require.NoError(t, err)
	assert.Equal(t, &kvsParams{"etcd", "somehost", 2379, "/configs/myconf", "yml"}, params)

	s = "consul://somehost/configs/myconf.yml"
	params, err = parseKVSConnectionString(s)
	require.NoError(t, err)
	assert.Equal(t, &kvsParams{"consul", "somehost", 8500, "/configs/myconf", "yml"}, params)
}

func Test_kvsParams_kvsParams(t *testing.T) {
	params := &kvsParams{
		provider: "etcd",
		host:     "somehost",
		port:     1144,
	}
	assert.Equal(t, "http://somehost:1144", params.formatEndpoint())

	params.provider = "consul"
	assert.Equal(t, "somehost:1144", params.formatEndpoint())
}
