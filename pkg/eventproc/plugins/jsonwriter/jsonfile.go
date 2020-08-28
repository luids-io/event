// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package jsonwriter

import (
	"encoding/json"
	"io"
	"os"
)

const dataBuffSize = 100

type jsonfile struct {
	path   string
	file   *os.File
	data   chan interface{}
	closed bool
}

func (j *jsonfile) open(bsize int) error {
	fh, err := os.OpenFile(j.path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	j.file = fh
	j.data = make(chan interface{}, bsize)
	go writeData(fh, j.data)
	return nil
}

func (j *jsonfile) write(e interface{}) {
	if !j.closed {
		j.data <- e
	}
}

func (j *jsonfile) close() {
	if !j.closed {
		j.closed = true
		close(j.data)
		j.file.Sync()
		j.file.Close()
		j.file = nil
	}
}

func writeData(w io.Writer, data <-chan interface{}) {
	for e := range data {
		bytes, err := json.Marshal(e)
		if err != nil {
			continue
		}
		bytes = append(bytes, byte('\n'))
		w.Write(bytes)
	}
}

var mapFiles map[string]*jsonfile

func getJSONFile(fpath string) *jsonfile {
	file, ok := mapFiles[fpath]
	if ok {
		return file
	}
	file = &jsonfile{path: fpath}
	mapFiles[fpath] = file
	return file
}

func init() {
	mapFiles = make(map[string]*jsonfile)
}
