// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cachedb

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPkg_full(t *testing.T) {
	want := []testentry{{A: true, B: "str", C: 42}, {A: false, B: "false", C: -1}}
	filename := "test"

	err := Write(filename, want)
	if err != nil {
		t.Fatal(err)
	}

	if !Exists(filename) {
		t.Fatal("cachedb does not exist")
	}

	got, err := Read[testentry](filename)
	if err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(want, got)
	if diff != "" {
		if diff != "" {
			t.Errorf("error mismatch (-want +got):\n%s", diff)
		}
	}
	err = os.Remove(Filename(filename))
	if err != nil {
		t.Fatal(err)
	}
}
