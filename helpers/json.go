// Copyright © 2014 Steve Francia <spf@spf13.com>.
//
// Licensed under the Simple Public License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://opensource.org/licenses/Simple-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helpers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spf13/afero"
	jww "github.com/spf13/jwalterweatherman"
)

func JsonDecoder(url string, hc *http.Client, fs afero.Fs) (*json.Decoder, error) {
	if url == "" {
		return nil, nil
	}
	if strings.Contains(url, "://") {
		jww.INFO.Printf("Downloading content JSON: %s ...", url)
		res, err := hc.Get(url)
		if err != nil {
			return nil, err
		}
		return json.NewDecoder(res.Body), nil
	}

	if e, err := Exists(url, fs); !e {
		return nil, err
	}

	f, err := fs.Open(url)
	if err != nil {
		return nil, err
	}
	return json.NewDecoder(f), nil
}
