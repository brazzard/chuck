package storage

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gavrilaf/chuck/utils"
	"github.com/spf13/afero"
)

type seekerImpl struct {
	root  *afero.Afero
	index Index
}

func NewSeekerWithFs(fs afero.Fs, folder string) (Seeker, error) {
	folder = strings.Trim(folder, " \\/")
	logDirExists, _ := afero.DirExists(fs, folder)
	if !logDirExists {
		return nil, fmt.Errorf("Folder %s doesn't exists", folder)
	}

	root := &afero.Afero{Fs: afero.NewBasePathFs(fs, folder)}

	index, err := LoadIndex(root, "index.txt", true)
	if err != nil {
		return nil, err
	}

	seeker := &seekerImpl{
		index: index,
		root:  root,
	}

	return seeker, nil
}

func (seeker *seekerImpl) Look(method string, url string) (*http.Response, error) {
	item := seeker.index.Find(method, url, SEARCH_SUBSTR)
	if item == nil {
		return nil, nil
	}

	header, err := seeker.readHeader(item.Folder + "/resp_header.json")
	if err != nil {
		return nil, fmt.Errorf("Read header error for %s: %v", item.Folder, err)
	}

	body, err := seeker.readBody(item.Folder + "/resp_body.json")
	if err != nil {
		return nil, fmt.Errorf("Read header body for %s: %v", item.Folder, err)
	}

	return utils.MakeResponse(item.Code, header, body, 0), nil
}

/*
 * Private
 */

func (seeker *seekerImpl) readHeader(fname string) (http.Header, error) {
	exists, err := seeker.root.Exists(fname)
	if err != nil {
		return nil, err
	}

	if !exists {
		return make(http.Header), nil
	}

	fp, err := seeker.root.Open(fname)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	header, err := utils.DecodeHeaders(buf)
	return header, err
}

func (seeker *seekerImpl) readBody(fname string) (io.ReadCloser, error) {
	exists, err := seeker.root.Exists(fname)
	if err != nil {
		return nil, err
	}

	if !exists {
		return ioutil.NopCloser(strings.NewReader("")), nil
	}

	fp, err := seeker.root.Open(fname)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	buf, err := ioutil.ReadAll(fp)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewReader(buf)), nil
}
