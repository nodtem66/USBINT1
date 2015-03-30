package db

type SqliteHandle struct {
	Done chan struct{}
	DataTag
}

func (s *SqliteHandle) nullPipe() {
}
