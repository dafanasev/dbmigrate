package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type kvsParams struct {
	provider string
	host     string
	port     int
	path     string
	format   string
}

func parseKVSConnectionString(s string) (*kvsParams, error) {
	// s must be in url format (provider://host:port/path.format)
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

func (p *kvsParams) formatEndpoint() string {
	s := fmt.Sprintf("%s:%d", p.host, p.port)
	if p.provider == "etcd" {
		s = "http://" + s
	}
	return s
}
