package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/devimteam/microgen/pkg/microgen"
	"github.com/pkg/errors"
)

func Run(plugins []string, iface microgen.Interface, currentPkg string) error {
	f, err := ioutil.TempFile(".", "microgen-bootstrap-*.go")
	if err != nil {
		return errors.Wrap(err, "can't create bootstrap file")
	}
	//defer os.Remove(f.Name())

	if n, err := f.Write(prefix); err != nil {
		return errors.Wrap(err, "writing error")
	} else if n != len(prefix) {
		return errors.New("prefix content was loosed")
	}

	mainContent, err := mainFunc(plugins, iface, currentPkg)
	if err != nil {
		return err
	}

	if n, err := f.Write(mainContent); err != nil {
		return errors.Wrap(err, "writing error")
	} else if n != len(mainContent) {
		return errors.New("main content was loosed")
	}
	genName := f.Name()
	if err := f.Close(); err != nil {
		return errors.Wrap(err, "close error")
	}
	if err := runFile(genName); err != nil {
		return errors.Wrap(err, "run new generator")
	}
	return nil
}

var prefix = []byte(`// +build microgen-ignore

// TEMPORARY microgen FILE. DO NOT EDIT.

package main
`)

func mainFunc(plugins []string, iface microgen.Interface, currentPkg string) ([]byte, error) {
	var b lnBuilder
	b.L("import (")
	b.L(`// List of imported plugins`)
	b.L(`_ "github.com/devimteam/microgen/pkg/plugins"`)
	for i := range plugins {
		b.L(fmt.Sprintf(`_ "%s"`, plugins[i]))
	}
	if len(plugins) > 0 {
		b.L()
	}
	b.L(strconv.Quote("reflect"))
	b.L(`pkg `, strconv.Quote(currentPkg))
	b.L(`microgen "github.com/devimteam/microgen/pkg/microgen"`)
	b.L(")")
	b.L("func main() {")
	b.L("microgen.RegisterPackage(", strconv.Quote(currentPkg), ")")
	b.L("targetInterface := ", iface.String())
	//b.L("targetInterface.Value=reflect.ValueOf((*pkg.", iface.Name, ")(nil)).Elem()")
	b.L("targetInterface.Value=reflect.ValueOf(new(pkg.", iface.Name, ")).Elem()")
	b.L("microgen.RegisterInterface(targetInterface)") //
	b.L("microgen.Exec(", strings.Join(stringpipe(strconv.Quote)(os.Args[1:]), ","), ")")
	b.L("}")
	return b.Bytes(), nil
}

func runFile(name string) error {
	cmd := exec.Command("go", "run", name)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

type lnBuilder struct {
	strings.Builder
}

func (b *lnBuilder) L(ss ...string) {
	for i := range ss {
		_, _ = b.WriteString(ss[i])
	}
	_ = b.WriteByte('\n')
}

func (b *lnBuilder) Bytes() []byte {
	return []byte(b.String())
}

func stringpipe(ff ...func(string) string) func([]string) []string {
	return func(ww []string) []string {
		ss := make([]string, len(ww))
		copy(ss, ww)
		for i := range ff {
			for j := range ss {
				ss[j] = ff[i](ss[j])
			}
		}
		return ss
	}
}
