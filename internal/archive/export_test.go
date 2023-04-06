package archive

import (
	"net/http"
)

func FakeDo(do func(req *http.Request) (*http.Response, error)) (restore func()) {
	_httpDo := httpDo
	httpDo = do
	return func() {
		httpDo = _httpDo
	}
}

const UbuntuURL = ubuntuURL
const UbuntuPortsURL = ubuntuPortsURL
const UbuntuProURL = ubuntuProURL
