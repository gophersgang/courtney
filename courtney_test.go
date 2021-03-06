package main

import (
	"testing"

	"io/ioutil"
	"path/filepath"

	"bytes"
	"strings"

	"github.com/dave/courtney/shared"
	"github.com/dave/patsy"
	"github.com/dave/patsy/builder"
	"github.com/dave/patsy/vos"
)

func TestRun(t *testing.T) {
	name := "run"
	env := vos.Mock()
	b, err := builder.New(env, "ns")
	if err != nil {
		t.Fatalf("Error creating builder in %s: %s", name, err)
	}
	defer b.Cleanup()

	ppath, pdir, err := b.Package("a", map[string]string{
		"a.go": `package a
		
			func Foo(i int) int {
				i++ // 1
				return i
			}
			
			func Bar(i int) int {
				i++ // 0
				return i
			}
		`,
		"a_test.go": `package a
					
			import "testing"
			
			func TestFoo(t *testing.T) {
				i := Foo(1)
				if i != 2 {
					t.Fail()
				}
			}
		`,
	})
	if err != nil {
		t.Fatalf("Error creating builder in %s: %s", name, err)
	}

	if err := env.Setwd(pdir); err != nil {
		t.Fatalf("Error in Setwd in %s: %s", name, err)
	}

	sout := &bytes.Buffer{}
	serr := &bytes.Buffer{}
	env.Setstdout(sout)
	env.Setstderr(serr)

	setup := &shared.Setup{
		Env:      env,
		Paths:    patsy.NewCache(env),
		Enforce:  true,
		Verbose:  true,
		Output:   "",
		TestArgs: []string{ppath},
	}
	if err := Run(setup); err != nil {
		if !strings.Contains(err.Error(), "Error: untested code") {
			t.Fatalf("Error running program in %s: %s", name, err)
		}
	}

	coverage, err := ioutil.ReadFile(filepath.Join(pdir, "coverage.out"))
	if err != nil {
		t.Fatalf("Error reading coverage file in %s: %s", name, err)
	}
	expected := `mode: set
ns/a/a.go:3.24,6.5 2 1
ns/a/a.go:8.24,11.5 2 0
`
	if string(coverage) != expected {
		t.Fatalf("Error in %s coverage. Got: \n%s\nExpected: \n%s\n", name, string(coverage), expected)
	}

	expected = `Untested code:
ns/a/a.go:8-11:
	func Bar(i int) int {
		i++ // 0
		return i
	}`

	if !strings.Contains(sout.String(), expected) {
		t.Fatalf("Error in %s stdout. Got: \n%s\nExpected to contain: \n%s\n", name, sout.String(), expected)
	}

	setup = &shared.Setup{
		Env:      env,
		Paths:    patsy.NewCache(env),
		Enforce:  false,
		Verbose:  false,
		Output:   "",
		TestArgs: []string{ppath},
	}
	if err := Run(setup); err != nil {
		t.Fatalf("Error running program (second try) in %s: %s", name, err)
	}

}
