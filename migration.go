package migrate

import "time"

// migration holds info about migration
type migration struct {
	name string
	// timestamp when the migration was created
	timestamp time.Time
}

// run executes single migration
func (m *migration) run() error {
	return nil
}
