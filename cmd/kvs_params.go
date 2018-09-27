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

func parseKVConnectionString(s string) (*kvsParams, error) {
	// s must be in url format( provider://host:port/path.format)
	u, err := url.Parse(s)
	if err != nil {
		return nil, errors.Wrapf(err, "can't parse string %s into components", s)
	}

	if u.Scheme != "etcd" && u.Scheme != "consul" {
		return nil, errors.Wrapf(err, "%s is not correct key value store provider", u.Scheme)
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
	if len(pathParts) < 2 {
		return nil, errors.New("key value store format is not provided")
	}
	path := strings.Join(pathParts[:len(pathParts)-1], ".")

	format := pathParts[len(pathParts)]
	switch format {
	case "json", "toml", "yaml", "yml", "properties", "props", "prop", "hcl":
		// the format is ok
	default:
		return nil, errors.Errorf("%s is not correct key value store format", format)
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
	s := fmt.Sprintf("%s:%s", p.host, p.port)
	if p.provider == "etcd" {
		s = "http://" + s
	}
	return s
}
