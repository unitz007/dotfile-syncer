package main

import (
	"errors"
	"github.com/haibeey/doclite"
)

type Database interface {
	Create(data *Commit) error
	Get(id int) (*Commit, error)
}

type docliteImpl struct {
	data *doclite.Doclite
}

func (db *docliteImpl) Get(id int) (*Commit, error) {
	defer db.close()
	g := &Commit{}

	var i int64 = 0
	var err error

	for i = 0; i <= db.data.Base().GetCol().NumDocuments; i++ {

		if id == int(i) {
			err = db.data.Base().FindOne(i, g)
			break
		}
	}

	if g.Id == "" {
		err = errors.New("not found")
	}

	return g, err
}

func InitDB() Database {

	col := doclite.Connect("dotfile-agent.doclite")

	return &docliteImpl{
		data: col,
	}

}

func (db *docliteImpl) Create(data *Commit) error {
	defer db.close()

	var err error
	localCommitId := 1
	isExists := func() bool {
		_, err := db.Get(localCommitId)
		if err == nil {
			return true
		}

		return false
	}()

	if isExists {
		err = db.data.Base().UpdateOneDoc(int64(localCommitId), data)
	} else {
		_, err = db.data.Base().Insert(data)
	}

	//err := db.data.Base().UpdateOneDoc(1, data)

	if err != nil {
		return err
	}

	return nil
}

func (db *docliteImpl) close() {
	func(data *doclite.Doclite) {
		err := data.Close()
		if err != nil {
			Error("failed to close doclite")
		}
	}(db.data)
}
