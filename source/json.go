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
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/hugo/helpers"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

var (
	ps = string(os.PathSeparator)
	cd string
)

type Site struct {
	Path, Content string
}

func (s *Site) reader() io.Reader {
	cb := []byte(s.Content)
	return bytes.NewReader(cb)
}

func (s *Site) path() string {
	return cd + filepath.Clean(s.Path)
}

func GenerateSourceFromJson(fs afero.Fs) {

	cd = helpers.AbsPathify(viper.GetString("ContentDir")) + ps + "FromJSON" + ps
	if err := fs.RemoveAll(cd); err != nil {
		jww.ERROR.Printf("Failed remove %s with error message %s", cd, err.Error())
	}

	url := viper.GetString("SourceUrl")
	dec, err := streamContent(url, http.DefaultClient, fs)
	if err != nil {
		jww.ERROR.Printf("Failed to get json resource %s with error message %s", url, err)
		return
	}

	c := 0
	jww.INFO.Printf("Generating files from JSON %s in: %s", url, cd)
	for {
		var s Site
		if err := dec.Decode(&s); err == io.EOF {
			jww.INFO.Printf("Generated %d file/s from JSON stream", c)
			break
		} else if err != nil {
			jww.WARN.Printf("Parser Error in JSON stream: %s", err.Error())
		} else {
			if err := helpers.SafeWriteToDisk(s.path(), s.reader(), fs); err != nil {
				jww.FATAL.Fatalf("Failed to write to disc: %s\n%#v", err, s)
			}
			c++
		}
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
