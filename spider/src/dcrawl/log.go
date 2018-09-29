package dcrawl

import (
	"github.com/op/go-logging"
	"os"
)

var Log = logging.MustGetLogger("spider")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} - %{shortfunc} - %{level:.4s}%{color:reset} - %{message}`,
)

func init() {
	backend := logging.NewLogBackend(os.Stderr, "", 1)
	bf := logging.NewBackendFormatter(backend, format)
	logging.SetLevel(logging.INFO, "")
	logging.SetBackend(bf)
}
