package main

import "github.com/haibeey/doclite"

type Database struct {
	data *doclite.Collection
}

func InitDB() Database {

	col := doclite.Connect("dotfile-agent.doclite").Base()

	return Database{
		data: col,
	}

}

func (db *Database) Create(data GitPull) (err error) {
	_, err = db.data.Insert(data)
	return err
}
