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
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/hugo/helpers"
	"github.com/spf13/hugo/hugofs"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

type site struct {
	Directory, Filename, Content string
}

func GenerateSourceFromJson() {

	j := getJson()
	jww.INFO.Printf("JSON: %#v", j)

	jww.INFO.Printf("Generating files in: %s", helpers.AbsPathify(viper.GetString("ContentDir")))

}

// getRemote loads the content of a remote file.
func getRemote(url string, fs afero.Fs, hc *http.Client) ([]byte, error) {

	jww.INFO.Printf("Downloading content JSON: %s ...", url)
	res, err := hc.Get(url)
	if err != nil {
		return nil, err
	}
	c, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return c, nil
}

// getLocal loads the content of a local file
func getLocal(filePath string, fs afero.Fs) ([]byte, error) {

	if e, err := helpers.Exists(filePath, fs); !e {
		return nil, err
	}

	f, err := fs.Open(filePath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

// resGetResource loads the content of a local or remote file
func downloadContent(url string) ([]byte, error) {
	if url == "" {
		return nil, nil
	}
	if strings.Contains(url, "://") {
		return getRemote(url, hugofs.SourceFs, http.DefaultClient)
	}
	return getLocal(url, hugofs.SourceFs)
}

// GetJson expects the url to a resource which can either be a local or a remote one.
// GetJson returns nil or parsed JSON to use in a short code.
func getJson() interface{} {
	url := viper.GetString("SourceUrl")
	c, err := downloadContent(url)
	if err != nil {
		jww.ERROR.Printf("Failed to get json resource %s with error message %s", url, err)
		return nil
	}
	// implement all readers via streams stream file and stream from URL
	dec := json.NewDecoder(strings.NewReader(jsonStream))
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", m.Name, m.Text)
	}
}
