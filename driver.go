package migrate

import (
	"database/sql"
	"time"
	
	"github.com/pkg/errors"
)

var (
	errDBNameNotProvided = errors.New("db name is not provided")
	errUserNotProvided   = errors.New("user is not provided")
)

type driver struct {
	credentials *Credentials
	db          *sql.DB
	provider
	placeholdersProvider
}

func NewDriver(cr *Credentials, provider provider) *driver {
	d := &driver{
		credentials: cr,
		provider:    provider,
	}
	if pp, ok := d.provider.(placeholdersProvider); ok {
		d.placeholdersProvider = pp
	}
	return d
}

func (d *driver) open() error {
	dsn, err := d.provider.dsn(d.credentials)
	if err != nil {
		return err
	}
	
	d.db, err = sql.Open(d.credentials.DriverName, dsn)
	if err != nil {
		return errors.Wrap(err, "can't open database")
	}
	
	return nil
}

func (d *driver) close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "Error shutting down migrator")
	}
	return nil
}

func (d *driver) appliedMigrations() ([]time.Time, error) {
	rows, err := d.db.Query("SELECT version FROM migrations ASC")
	if err != nil {
		return nil, errors.Wrap(err, "can't get executed migrations versions")
	}
	defer rows.Close()
	
	vs := []time.Time{}
	var v time.Time
	for rows.Next() {
		err = rows.Scan(&v)
		if err != nil {
			return nil, errors.Wrap(err, "can't scan migration version row")
		}
		vs = append(vs, v)
	}
	return vs, nil
}

func (d *driver) setPlaceholders(s string) string {
	if d.placeholdersProvider == nil {
		return s
	}
	return d.placeholdersProvider.setPlaceholders(s)
}

func (d *driver) createTable() error {
	_, err := d.db.Query(d.setPlaceholders("CREATE TABLE migrations (version timestamp NOT NULL, PRIMARY KEY(version));"))
	if err != nil {
		return errors.Wrapf(err, "Can't create migrations table")
	}
	return nil
}