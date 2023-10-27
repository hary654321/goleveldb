package iterator_test

import (
	"testing"

	"github.com/hary654321/goleveldb/leveldb/testutil"
)

func TestIterator(t *testing.T) {
	testutil.RunSuite(t, "Iterator Suite")
}
