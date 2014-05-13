package ar

/*
   Copyright 2013 Am Laher

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/


import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"os"
)
//TODO: this is commented out because it's an unfinished proof-of-concept.

func TestAr(t *testing.T) {
	err := os.MkdirAll(filepath.Join("_test","usr","bin"), 0777)
	if err != nil {
		t.Errorf("Error making dir %v", err)
	}
	err = ioutil.WriteFile(filepath.Join("_test","debian-binary"), []byte("2.0\n"), 0644)
	if err != nil {
		t.Errorf("Error making file %v", err)
	}
	ioutil.WriteFile(filepath.Join("_test","control"), []byte(`Package: testbin
Priority: extra
Maintainer: Am Laher
Architecture: i386
Version: 0.5.2
Depends: golang
Provides: testbin
Description: utility for Go
`), 0644)
	if err != nil {
		t.Errorf("Error making file %v", err)
	}
	err = ioutil.WriteFile(filepath.Join("_test","usr","bin", "testbin"), []byte("#!/bin/bash\necho \"testbin\"\n"), 0644)
	if err != nil {
		t.Errorf("Error making file %v", err)
	}

/*
	TarGz(filepath.Join("_test","control.tar.gz"), [][]string{[]string{ filepath.Join("_test","control"), "control"} })
	TarGz(filepath.Join("_test","data.tar.gz"), [][]string{[]string{ filepath.Join("_test","goxc"), "/usr/bin/goxc"} })
	*/
	targetFile := filepath.Join("_test","testbin_0.5.2_i386.deb")

	inputs := []Archivable{
	 &FilePathArchivable{filepath.Join("_test","debian-binary"),"debian-binary",nil},
	 &FilePathArchivable{filepath.Join("_test","control.tar.gz"),"control.tar.gz",nil},
	 &FilePathArchivable{filepath.Join("_test","data.tar.gz"),"data.tar.gz",nil}}
	err = Ar(targetFile, inputs)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

