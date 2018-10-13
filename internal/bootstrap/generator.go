package bootstrap

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func Run(plugins, interfaces []string, pkg string) error {
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

	mainContent, err := mainFunc(plugins, interfaces, pkg)
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

func mainFunc(plugins, interfaces []string, pkg string) ([]byte, error) {
	var b lnBuilder
	b.L(0, "import (")
	if len(plugins) > 0 {
		b.L(1, `// List of imported plugins`)
	}
	for i := range plugins {
		b.L(1, fmt.Sprintf(`_ "%s"`, plugins[i]))
	}
	if len(plugins) > 0 {
		b.L(0)
	}
	b.L(1, `pkg `, strconv.Quote(pkg))
	b.L(1, `microgen "github.com/devimteam/microgen/pkg/microgen"`)
	b.L(0, ")")
	b.L(0, "func main() {")
	for i := range interfaces {
		b.L(1, fmt.Sprintf(`microgen.RegisterInterface("%s", pkg.%s(nil))`, interfaces[i], interfaces[i]))
	}
	b.L(1, "microgen.Run()")
	b.L(0, "}")
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

func (b *lnBuilder) L(tabs int, ss ...string) {
	_, _ = b.WriteString(strings.Repeat("\t", tabs))
	for i := range ss {
		_, _ = b.WriteString(ss[i])
	}
	_ = b.WriteByte('\n')
}

func (b *lnBuilder) Bytes() []byte {
	return []byte(b.String())
}
