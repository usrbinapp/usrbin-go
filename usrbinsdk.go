package usrbin

type Logger interface {
	Printf(format string, v ...interface{})
}

type SDK struct {
	version                 string
	updateChecker           UpdateChecker
	externalPackageManagers []ExternalPackageManager
	logger                  Logger
}
