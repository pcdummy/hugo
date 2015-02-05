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
	"bytes"
	"io"
	"testing"
)

type (
	TestPage struct {
		FilePath string
		Content  string
	}
)

func (p *TestPage) Reader() io.Reader {
	return bytes.NewReader([]byte(p.Content))
}

func (p *TestPage) Path() string {
	return p.FilePath
}

func TestSourceFiles(t *testing.T) {
	sf := &SourceFiles{
		pages: []Pager{
			&TestPage{
				FilePath: "f1.md",
				Content:  "We go together",
			},
			&TestPage{
				FilePath: "f2.md",
				Content:  "We go together, sometimes",
			},
		},
	}
	sf.merge([]Pager{
		&TestPage{
			FilePath: "f3.md",
			Content:  "We go together, today",
		},
	})
	if len(sf.pages) != 3 {
		t.Errorf("Expected 3 Pages but got: %d", len(sf.pages))
	}

	if len(sf.Files()) != 3 {
		t.Errorf("Expected 3 Files but got: %d", len(sf.Files()))
	}

	ts4 := `PHP vs Go is like a Humvee vs a Tesla Model S P85D`
	fs := &Filesystem{}
	fs.add("fs1.md", bytes.NewReader([]byte(ts4)))
	sf.files = fs

	if len(sf.Files()) != 3 {
		t.Errorf("Expected 3 Files from cache but got: %d", len(sf.Files()))
	}
	sf.cacheFiles = nil

	f := sf.Files()
	if len(f) != 4 {
		t.Errorf("Expected 4 Files but got: %d", len(f))
	}

	at := sf.cacheFiles[3].String()
	if at != ts4 {
		t.Errorf("Expected %s but got: %s", ts4, at)
	}

}
