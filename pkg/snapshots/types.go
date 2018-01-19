package snapshots

type Snapshot interface {
	Commit(pkg string) (string, error)
}
