package model

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	SPECTRUM_SCALE BackendType = iota
	SPECTRUM_SCALE_NFS
	SOFTLAYER_NFS
)

type BackendType int
type Volume struct {
	gorm.Model
	Name    string
	Backend BackendType
}
