package view

type QueryResultView interface {
	Success(data []byte) error
	Error(err error) error
}