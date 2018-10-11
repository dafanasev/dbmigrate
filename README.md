[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg?style=flat-square)](/LICENSE)
[![Build Status](https://travis-ci.org/dafanasev/lu.svg?branch=master)](https://travis-ci.org/dafanasev/dbmigrate)
[![GoDoc](https://godoc.org/github.com/dafanasev/go-yandex-translate?status.svg)](https://godoc.org/github.com/dafanasev/dbmigrate)
[![Go Report Card](https://goreportcard.com/badge/github.com/dafanasev/lu)](https://goreportcard.com/report/github.com/dafanasev/dbmigrate)
[![Coverage Status](https://coveralls.io/repos/github/dafanasev/lu/badge.svg)](https://coveralls.io/github/dafanasev/dbmigrate)

# dbmigrate

## Overview
__dbmigrate is a sql database migration tool.__   

dbmigrate can be used both as a CLI application or Go package, it does not use any DSL for migrations, just plain old SQL we all know and love
so it can be used with any framework and programming language. 

This readme covers CLI tool, for the Go package documentation please look at https://godoc.org/github.com/dafanasev/dbmigrate

## Features
* Can be used as a CLI app or as a Go package
* Support for PostgreSQL, MySQL and SQLite
* Migrations generator
* Up and down migrations in different files
* Database specific migrations (e.g. ones that executed only on Postgres)
* Uses timestamps as migration version
* Migrates all the way up or by specified number of steps
* Applies migrations in batches, that can be rolled back/reapplied at once   
* View migrations status and other information such as if database is up to date or not, last applied migration, etc
* Gets database connection settings from command line flags, environment variables, file in JSON, TOML, YAML, HCL, or Java properties format, consul or etcd.
* Support for different environments, e.g. for tests

## Install
* using homebrew: `brew install dbmigrate`
* using go get: `go get -u github.com/dafanasev/dbmigrate`
* download binaries from https://github.com/dafanasev/dbmigrate/releases

## Usage
dbmigrate command could be called from any subdirectory of a directory containing dbmigrations directory, there migrations should be stored.  

### Database settings
In order to use dbmigrate you should provide database settings.
Mandatory settings are database engine (postgres, mysql or sqlite), database name and user (except for sqlite).
Host defaults to localhost and port has a specific default value for each database engine.
 
Settings can be read from the following sources, sorted by precedence: 
* Command line flags
* Environment variables
* Configuration files (in JSON, TOML, YAML, HCL, or Java properties format)
* Key value store (consul or etcd)

#### Environments
When using environment variables, configuration files or key value store the --environment (-e) command line flag can be provided 
to specify alternative database settings if your project uses more than one database, e.g. for tests. 

#### Command line flags
Database settings related command line flags are:
* -n, --engine: database engine (postgres, mysql or sqlite)
* -d, --database: database name
* -u, --user: database user
* -p, --password: database password
* -b, --host: database host, default is localhost
* -o, --port: database port, default is specific for each database engine
* -t, --table: migrations table, default is migrations

The full list of command line flags can be obtained by running dbmigrate --help.

#### Environment variables
Environment variables names should be in the following format:
uppercased project dir name (the one holding dbmigrations dir) followed by correspnding flag name, joined by underscore and uppercased.
For example, the environment variable name for the database engine for the project that located in the directory named theservice would be `THESERVICE_ENGINE`.

If the --prefix (-x) flag is provided, it would be used instead of project dir as environment variables prefix.

if the --env (-e) flag is provided, it would be used as a second part of the variable name, 
e.g. when the --env flag set to 'test' and the project is in the directory theservice, 
variable name for engine would be `THESERVICE_TEST_ENGINE`

#### Configuration files
Configuration files could be in in JSON, TOML, YAML, HCL, or Java properties format. 
Default file name is dbmigrate, so dbmigrate looks for dbmigrate.yml, dbmigrate.yaml, dbmigrate.json, etc.
The alternative configuration file name (without extension) can be set using the --config (-c) flag.

Example configuration file in yaml format:

```yaml
engine: postgres
database: blog
user: author
password: topsecret

test:
  engine: sqlite
  database: blog.db
``` 

#### Key-value stores

dbmigrate can use consul or etcd to store configuration.
In order to use it the --kvsparams (-k) flag should be provided, with a value in the following format:
provider://host(:port)/path.type where provider is consul or etcd, host is key-value store host without scheme part,
port can be omitted and defaults to 8500 for consul and 2379 for etcd, 
path is the key path and type is the file format used to encode the config. 
File formats are the same ones as used for configuration files.

[crypt](https://github.com/xordataexchange/crypt) can be used to put configurations in the key-value store, e.g.:
`crypt set -endpoint http://localhost:2379 -plaintext /configs/weebservice weebservice.json`

if the --secretkeyring (-r) flag is provided, which should point to the path of a secret key ring path, 
the configuration will be stored encrypted and will be automatically decrypted when retrieved.

### Commands
dbmigrate has the following commands: generate, migrate (the root, default command), rollback, reapply and status.

#### Generate
The generate command generates up and down migrations. It uses command line arguments to build migration name,
e.g. `dbmigrate generate Posts table` will create TIMESTAMP_posts_table.up.sql and TIMESTAMP_posts_table.down.sql migrations.

If the --engines (-g) flag is set then migrations will be created only for the specified engines, 
e.g. `dbmigrate -g=postgres,sqlite generate Posts table` will generate TIMESTAMP_posts_table.up.postgres.sql, TIMESTAMP_posts_table.up.sqlite.sql
and corresponding down migrations.

If the --engines flag is set without value, the database engine specified in connection settings will be used,
e.g. `dbmigrate -n=sqlite -d=test.db generate Posts table` will generate TIMESTAMP_posts_table.up.sqlite.sql and TIMESTAMP_posts_table.down.sqlite.sql files.


#### Migrate
The migrate command applies all unapplied migrations or, if the --steps (-s) flag is set, only -s migrations.

#### Rollback
The rollback command rolls back the latest migration operation, e.g. if 3 migrations have been applied during the last operation, exactly these 3 migrations will be rolled back.
If the --steps (-s) flag is set, exactly -s migrations will be rolled back.

#### Reapply
The reapply command rolls back and applies again migrations applied during the latest migration operation.
If the --steps (-s) flag is set, exactly -s migrations will be reapplied.

#### Status
The status command shows migrations list with names, versions and applied at times, if the migration was applied.
It also shows the latest version migration, the last applied migrations (they are not necessarily the same ones), 
number of applied migrations and if the database schema is up to date or not. 

### Other settings
The --missingdowns (-m) boolean flag specifies if it is ok to have missing or empty down migrations.
 
## Todo
- [ ] Embed migrations into binary or get them from zip/tar archives, http, ssh, s3 or github

## License
dbmigrate is distributed under the MIT license.