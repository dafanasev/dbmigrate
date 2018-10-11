package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// kvsParams holds data used to connect to and get data configuration from key value store
type kvsParams struct {
	// provider is the key value store, consul or etcd
	provider string
	host     string
	port     int
	path     string
	// format is the forat supported by Viper
	// json, toml, yaml, properties or hcl
	format string
}

// parseKVSConnectionString parses key vaue store connection string into kvsParams struct
func parseKVSConnectionString(s string) (*kvsParams, error) {
	// s must be in url format (provider://host(:port)/path.format)
	u, err := url.Parse(s)
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse string %s into components", s)
	}
	if u.Scheme == "" || u.Host == "" || u.Path == "" {
		return nil, errors.Errorf("can't parse string %s into components", s)
	}

	if u.Scheme != "etcd" && u.Scheme != "consul" {
		return nil, errors.Errorf("%s is not correct key value store provider", u.Scheme)
	}

	if u.Port() == "" {
		switch u.Scheme {
		case "etcd":
			u.Host += ":2379"
		case "consul":
			u.Host += ":8500"
		}
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil {
		return nil, errors.Errorf("%s is not correct key value store port", u.Port())
	}

	pathParts := strings.Split(u.Path, ".")
	l := len(pathParts)
	if l < 2 {
		return nil, errors.New("key value store format is not provided")
	}
	path := strings.Join(pathParts[:l-1], ".")

	format := pathParts[l-1]
	switch format {
	case "json", "toml", "yaml", "yml", "properties", "props", "prop", "hcl":
		// format is ok
	default:
		return nil, errors.Errorf("%s is not correct config format", format)
	}

	return &kvsParams{
		provider: u.Scheme,
		host:     u.Hostname(),
		port:     port,
		path:     path,
		format:   format,
	}, nil
}

// formatEndpoint formats endpoint so it can be passed to key value store
func (p *kvsParams) formatEndpoint() string {
	s := fmt.Sprintf("%s:%d", p.host, p.port)
	if p.provider == "etcd" {
		s = "http://" + s
	}
	return s
}
