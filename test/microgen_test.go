package test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type TestCase struct {
	Name string
	Got  string
	Want string
}

const (
	assestsPath = "./assets"
	wantSubPath = "./want"
	gotSubPath  = "./got"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestAssets(t *testing.T) {
	cases, err := ioutil.ReadDir(assestsPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if !c.IsDir() {
			continue
		}
		t.Run(c.Name(), func(t *testing.T) {
			path := filepath.Join(c.Name(), wantSubPath)
			wantFiles := make(map[string][]byte)
			err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				wantFiles[path] = data
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}

			path = filepath.Join(c.Name(), gotSubPath)
			gotFiles := make(map[string][]byte)
			err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				data, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				gotFiles[path] = data
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			for k, v := range wantFiles {
				if v2, ok := gotFiles[k]; !ok {
					t.Fatal(k, "not found")
				} else if string(v) != string(v2) {
					t.Fatal("not same")
				}
			}
		})
	}
}
