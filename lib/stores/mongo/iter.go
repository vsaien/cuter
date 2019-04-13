package mongo

import "gopkg.in/mgo.v2"

type ClosableIter struct {
	*mgo.Iter
	Cleanup func()
}

func (it *ClosableIter) Close() error {
	err := it.Iter.Close()
	it.Cleanup()
	return err
}
