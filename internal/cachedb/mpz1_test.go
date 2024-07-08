// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package cachedb

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testentry struct {
	A bool
	B string
	C int
}

func TestMpz1_write(t *testing.T) {
	want := []testentry{{A: true, B: "str", C: 42}, {A: false, B: "false", C: -1}}
	filename := "test"

	err := mpz1write(filename, want)
	if err != nil {
		t.Fatal(err)
	}

	got, err := mpz1read[testentry](filename)
	if err != nil {
		t.Fatal(err)
	}
	diff := cmp.Diff(want, got)
	if diff != "" {
		if diff != "" {
			t.Errorf("error mismatch (-want +got):\n%s", diff)
		}
	}

	err = os.Remove(fname(filename, mpz1))
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Printf("%+v\n", db1)
}
