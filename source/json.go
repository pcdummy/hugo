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
	JsonSite struct {
		Path, Content string
	}

	JsonSites struct {
		cacheFiles []*File
		sites      []*JsonSite
	}
)

func (js *JsonSites) clear() {
	js.cacheFiles = nil
}

func (js *JsonSites) Files() []*File {
	if js.cacheFiles == nil {
		js.cacheFiles = make([]*File, len(js.sites))
		for i, s := range js.sites {
			js.cacheFiles[i] = NewFileWithContents(s.path(), s.reader())
		}
	}
	return js.cacheFiles
}

func (s *JsonSite) reader() io.Reader {
	return bytes.NewReader([]byte(s.Content))
}

func (s *JsonSite) path() string {
	return filepath.Clean(s.Path)
}

/*
   @todo implement polling and rebuild of the site via utils.CheckErr(commands.BuildSite(true))
*/
func GenerateSourceFromJson(hc *http.Client, fs afero.Fs) *JsonSites {

	if nil == hc {
		hc = http.DefaultClient
	}

	url := viper.GetString("SourceUrl")
	dec, err := streamContent(url, hc, fs)
	if err != nil {
		jww.ERROR.Printf("Failed to get json resource %s with error message %s", url, err)
		return nil
	}

	c := 0
	sources := make([]*JsonSite, 0, 10000)
	jww.INFO.Printf("Generating files from JSON %s", url)
	for {
		var s JsonSite
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

	return &JsonSites{
		sites: sources,
	}
}

func streamContent(url string, hc *http.Client, fs afero.Fs) (*json.Decoder, error) {
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

//func buildSite(site *Site) (err error) {
//	startTime := time.Now()
//
//	err = site.Build()
//	if err != nil {
//		return err
//	}
//	site.Stats()
//	jww.FEEDBACK.Printf("in %v ms\n", int(1000*time.Since(startTime).Seconds()))
//
//	return nil
//}
