// +build dev

package core

import "net/http"

var Client http.FileSystem = http.Dir("../../client/build")
