package usrbin

import (
	"github.com/usrbinapp/usrbin-go/pkg/pkgmgr"
	"github.com/usrbinapp/usrbin-go/pkg/updatechecker"
)

type Logger interface {
	Printf(format string, v ...interface{})
}

type SDK struct {
	version                 string
	updateChecker           updatechecker.UpdateChecker
	externalPackageManagers []pkgmgr.ExternalPackageManager
	logger                  Logger
}
