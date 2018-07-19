package migrate

import "time"

// migration holds info about migration
type migration struct {
	name string
	// timestamp when the migration was created
	timestamp time.Time
}

type byTimestamp []*migration

func (bts byTimestamp) Len() int           { return len(bts) }
func (bts byTimestamp) Swap(i, j int)      { bts[i], bts[j] = bts[j], bts[i] }
func (bts byTimestamp) Less(i, j int) bool { return bts[i].timestamp.Unix() < bts[j].timestamp.Unix() }

// run executes single migration
func (m *migration) run() error {
	return nil
}
