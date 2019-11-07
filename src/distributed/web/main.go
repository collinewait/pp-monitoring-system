package main

import (
	"net/http"

	"github.com/collinewait/pp-monitoring-system/src/distributed/web/controller"
)

func main() {
	controller.Initialize()

	http.ListenAndServe(":3000", nil)
}
