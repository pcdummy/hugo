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

package source

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"bytes"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/hugo/helpers"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

type (
	jsonPage struct {
		FilePath string `json:"Path"`
		Content  string `json:"Content"`
	}
)

func (p *jsonPage) Reader() io.Reader {
	return bytes.NewReader([]byte(p.Content))
}

func (p *jsonPage) Path() string {
	return filepath.Clean(p.FilePath)
}

// jsonStreamToFiles acts as the main function to be called in url.go
func loadJson(hc *http.Client, fs afero.Fs) []Pager {
	url := viper.GetString("SourceUrl")
	if url == "" {
		return nil
	}

	dec, err := jsonDecoder(url, hc, fs)
	if err != nil || dec == nil {
		jww.ERROR.Printf("Failed to get json resource \"%s\" with error message: %s", url, err)
		return nil
	}

	c := 0
	sources := make([]Pager, 0, 1000)
	jww.INFO.Printf("Generating files from JSON %s", url)
	for {
		var s jsonPage
		if err := dec.Decode(&s); err == io.EOF {
			jww.INFO.Printf("Generated %d file/s from JSON stream", c)
			break
		} else if err != nil {
			jww.WARN.Printf("Parser Error in JSON stream: %s", err.Error())
		} else {
			sources = append(sources, &s)
			c++
		}
	}

	return sources
}

func jsonDecoder(url string, hc *http.Client, fs afero.Fs) (*json.Decoder, error) {
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

	if e, err := helpers.Exists(url, fs); !e {
		return nil, err
	}

	f, err := fs.Open(url)
	if err != nil {
		return nil, err
	}
	return json.NewDecoder(f), nil
}
