package storage

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gavrilaf/chuck/utils"
	"github.com/spf13/afero"
)

type scRecorderImpl struct {
	root     *afero.Afero
	name     string
	recorder Recorder
	log      utils.Logger
}

func NewScenarioRecorder(fs afero.Fs, log utils.Logger, folder string, createNewFolder bool) (ScenarioRecorder, error) {
	// TODO: move to the shared code
	folder = strings.Trim(folder, " \\/")
	logDirExists, err := afero.DirExists(fs, folder)
	if err != nil {
		return nil, err
	}

	if !logDirExists {
		err := fs.Mkdir(folder, 0777)
		if err != nil {
			return nil, err
		}
	}

	name := ""
	path := folder
	if createNewFolder {
		tm := time.Now()
		name = fmt.Sprintf("%d_%d_%d_%d_%d_%d", tm.Year(), tm.Month(), tm.Day(), tm.Hour(), tm.Minute(), tm.Second())
		path = folder + "/" + name

		err = fs.Mkdir(path, 0777)
		if err != nil {
			return nil, err
		}
	}

	root := &afero.Afero{Fs: afero.NewBasePathFs(fs, path)}

	return &scRecorderImpl{
		root: root,
		name: name,
		log:  log,
	}, nil
}

func (p *scRecorderImpl) Close() error {
	return nil
}

func (p *scRecorderImpl) Name() string {
	return p.name
}

func (self *scRecorderImpl) PendingCount() int {
	return 0
}

func (p *scRecorderImpl) ActivateScenario(name string) error {
	recorder, err := NewRecorder(p.root, p.log, name, false, true)
	if err != nil {
		return err
	}

	if p.recorder != nil {
		p.recorder.Close()
	}
	p.recorder = recorder

	return nil
}

func (p *scRecorderImpl) RecordRequest(req *http.Request, session int64) (*PendingRequest, error) {
	if p.recorder == nil {
		return nil, ErrScenarioNotFound
	}

	return p.recorder.RecordRequest(req, session)
}

func (p *scRecorderImpl) RecordResponse(resp *http.Response, session int64) (*PendingRequest, error) {
	if p.recorder == nil {
		return nil, ErrScenarioNotFound
	}

	return p.recorder.RecordResponse(resp, session)
}