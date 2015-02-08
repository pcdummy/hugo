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
	"io"
	"net/http"

	"github.com/spf13/afero"
)

type (
	SourceFiles struct {
		files      Input
		cacheFiles []*File
		pages      []Pager
	}

	Pager interface {
		Reader() io.Reader
		Path() string
	}
)

// merge merges a slice of pages into pages variable
func (sf *SourceFiles) merge(ps []Pager) {
	if ps != nil {
        sf.pages = append(sf.pages, ps...)
	}
}

// Files returns all available files. In this case it merges the virtual files from the URL
// source into input files (mostly the ones from the HDD)
func (sf *SourceFiles) Files() []*File {
	if sf.cacheFiles == nil {
		l := len(sf.pages)
		if sf.files != nil {
			l = l + len(sf.files.Files())
		}
		sf.cacheFiles = make([]*File, l)
		for i, s := range sf.pages {
			sf.cacheFiles[i] = NewFileWithContents(s.Path(), s.Reader())
		}
		if sf.files != nil {
			l := len(sf.pages)
			for i, f := range sf.files.Files() {
				sf.cacheFiles[l+i] = f
			}
		}
	}
	return sf.cacheFiles
}

// MergeUrl merges any remote or local URL into the files variable
// MergeUrl is of course also triggere when watch mode is active but only when
// a local file changes
func MergeUrl(files Input, fs afero.Fs) *SourceFiles {

	sf := &SourceFiles{
		pages: make([]Pager, 0, 10000), // 10k pages should be on average enough
	}
	sf.merge(loadJson(http.DefaultClient, fs))
	// pages can now be merged with more sources like XML or freaky CSV

	sf.files = files
	return sf
}
