package dcrawl

import (
	"library/go-logging"
	"os"
)

var Log = logging.MustGetLogger("spider")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} - %{shortfunc} - %{level:.4s}%{color:reset} - %{message}`,
)

func init() {
	backend := logging.NewLogBackend(os.Stderr, "", 1)
	bf := logging.AddModuleLevel(logging.NewBackendFormatter(backend, format))
	bf.SetLevel(logging.INFO, "")
	logging.SetBackend(bf)
}
