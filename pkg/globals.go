package pkg

import (
	"net/http"
	"time"
)

var CLIENT_ID = "zi1hlaxywl566lk3r00gy6b61qknuf"

var httpClient = http.Client{
	Timeout: time.Second * 10,
}
