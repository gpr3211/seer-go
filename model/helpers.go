package model

import "os"

type Project struct {
	Name  string
	Files map[string]*os.File // map of files for each package within a project.
	Hub   string
}

type Module struct {
	Items []*os.File
}
