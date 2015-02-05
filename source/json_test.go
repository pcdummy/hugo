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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func getTestServer(handler func(w http.ResponseWriter, r *http.Request)) (*httptest.Server, *http.Client) {
	testServer := httptest.NewServer(http.HandlerFunc(handler))
	client := &http.Client{
		Transport: &http.Transport{Proxy: func(*http.Request) (*url.URL, error) { return url.Parse(testServer.URL) }},
	}
	return testServer, client
}

func TestJsonPage(t *testing.T) {
	tests := []struct {
		pathActual   string
		pathExpected string
		content      string
	}{
		{
			"./path/./to/file.md",
			"path/to/file.md",
			"A SQL query walks into a bar ...",
		},
		{
			"/../path//to/file.md",
			"/path/to/file.md",
			"A SQL query walks again into a bar ...",
		},
		{
			"./../path//to/file.md",
			"../path/to/file.md",
			"",
		},
	}
	for _, test := range tests {
		jp := &JsonPage{
			FilePath: test.pathActual,
			Content:  test.content,
		}
		if jp.Path() != test.pathExpected {
			t.Errorf("Expected: %s but got: %s", test.pathExpected, jp.Path())
		}
		data, err := ioutil.ReadAll(jp.Reader())
		if err != nil {
			t.Error(err)
		}
		if string(data) != test.content {
			t.Errorf("Expected: %s but got: %s", test.content, string(data))
		}

	}
}

func TestJsonStreamToFiles(t *testing.T) {

	jsonStream, err := ioutil.ReadFile("./streamOfPages.json")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		sourceUrl     string
		content       []byte
		expectedCount int
	}{
		{"http://jay-son/file/stream.json", jsonStream, 93},
		{"i/am/a/local/file/stream.json", jsonStream, 93},
		{"", jsonStream, 0},
		{"", nil, 0},
	}

	for _, test := range tests {
		fs := new(afero.MemMapFs)
		inMemFile, err := fs.Create(test.sourceUrl)
		if err != nil {
			t.Fatal(err)
		}
		inMemFile.Write(test.content)

		srv, cl := getTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.Write(test.content)
		})
		defer func() { srv.Close() }()
		viper.Set("SourceUrl", test.sourceUrl)
		pages := jsonStreamToFiles(cl, fs)

		if len(pages) != test.expectedCount {
			t.Errorf("URL %s Expected: %d but got: %d", test.sourceUrl, test.expectedCount, len(pages))
		}
		for _, page := range pages {
			if len(page.Path()) < 1 {
				t.Errorf("Path in demo data has length of 0: %#v", page)
			}
		}
	}
}
