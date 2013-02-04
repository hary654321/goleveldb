// Copyright (c) 2012, Suryandaru Triandana <syndtr@gmail.com>
// All rights reserved.
//
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This LevelDB Go implementation is based on LevelDB C++ implementation.
// Which contains the following header:
//   Copyright (c) 2011 The LevelDB Authors. All rights reserved.
//   Use of this source code is governed by a BSD-style license that can be
//   found in the LEVELDBCPP_LICENSE file. See the LEVELDBCPP_AUTHORS file
//   for names of contributors.

package leveldb

import (
	"encoding/binary"
	"io"
	"sort"

	"github.com/syndtr/goleveldb/leveldb/descriptor"
	"github.com/syndtr/goleveldb/leveldb/journal"
)

type readByteReader interface {
	io.Reader
	io.ByteReader
}

func sliceBytes(b []byte) []byte {
	z, n := binary.Uvarint(b)
	return b[n : n+int(z)]
}

func sliceBytesTest(b []byte) (valid bool, v, rest []byte) {
	z, n := binary.Uvarint(b)
	m := n + int(z)
	if n <= 0 || m > len(b) {
		return
	}
	valid = true
	v = b[n:m]
	rest = b[m:]
	return
}

func readBytes(r readByteReader) (b []byte, err error) {
	var n uint64
	n, err = binary.ReadUvarint(r)
	if err != nil {
		return
	}
	b = make([]byte, n)
	_, err = io.ReadFull(r, b)
	if err != nil {
		b = nil
	}
	return
}

func shorten(str string) string {
	if len(str) <= 13 {
		return str
	}
	return str[:5] + "..." + str[len(str)-5:]
}

type files []descriptor.File

func (p files) Len() int {
	return len(p)
}

func (p files) Less(i, j int) bool {
	return p[i].Num() < p[j].Num()
}

func (p files) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p files) sort() {
	sort.Sort(p)
}

type journalReader struct {
	file    descriptor.File
	reader  descriptor.Reader
	journal *journal.Reader
}

func newJournalReader(file descriptor.File, checksum bool, dropf journal.DropFunc) (p *journalReader, err error) {
	r := new(journalReader)
	r.file = file
	r.reader, err = file.Open()
	if err != nil {
		return
	}
	r.journal = journal.NewReader(r.reader, checksum, dropf)
	return r, nil
}

func (r *journalReader) closed() bool {
	return r.reader == nil
}

func (r *journalReader) close() {
	if r.closed() {
		return
	}
	r.reader.Close()
	r.reader = nil
	r.journal = nil
}

func (r *journalReader) remove() error {
	r.close()
	return r.file.Remove()
}

type journalWriter struct {
	file    descriptor.File
	writer  descriptor.Writer
	journal *journal.Writer
}

func newJournalWriter(file descriptor.File) (p *journalWriter, err error) {
	w := new(journalWriter)
	w.file = file
	w.writer, err = file.Create()
	if err != nil {
		return
	}
	w.journal = journal.NewWriter(w.writer)
	return w, nil
}

func (w *journalWriter) closed() bool {
	return w.writer == nil
}

func (w *journalWriter) close() {
	if w.closed() {
		return
	}
	w.writer.Close()
	w.writer = nil
	w.journal = nil
}

func (w *journalWriter) remove() error {
	w.close()
	return w.file.Remove()
}