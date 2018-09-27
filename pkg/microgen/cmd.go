package microgen

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"

	"github.com/vetcher/go-astra"
	"github.com/vetcher/go-astra/types"

	"github.com/devimteam/microgen/generator"
	lg "github.com/devimteam/microgen/logger"
)

var (
	flagSource    = flag.String("src", ".", "Path to input package with interfaces.")
	flagDstDir    = flag.String("dst", ".", "Destiny path.")
	flagVerbose   = flag.Int("v", 1, "Sets microgen verbose level.")
	flagDebug     = flag.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagInterface = flag.String("interface", "", "Name of the target interface. If package contains one interface with microgen tags, arg may be omitted.")
)

func init() {
	if !flag.Parsed() {
		flag.Parse()
	}
}

func Exec() {
	begin := time.Now()
	defer func() { lg.Logger.Logln(1, "Full:", time.Since(begin)) }()
	if *flagVerbose < 0 {
		*flagVerbose = 0
	}
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = 100
	}
	lg.Logger.Logln(1, "microgen", "1.0.0")

	processConfig()

	lg.Logger.Logln(4, "Source package:", *flagSource)
	info, err := astra.GetPackage(*flagSource, astra.AllowAnyImportAliases,
		astra.IgnoreStructs, astra.IgnoreFunctions, astra.IgnoreConstants,
		astra.IgnoreMethods, astra.IgnoreTypes, astra.IgnoreVariables,
	)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		os.Exit(1)
	}
	ii := findInterfaces(info)
	iface, err := selectInterface(ii)
	if err != nil {
		lg.Logger.Logln(0, "fatal:", err)
		lg.Logger.Logln(4, "All founded interfaces:")
		lg.Logger.Logln(4, listInterfaces(info.Interfaces))
		os.Exit(1)
	}

}

func findInterfaces(file *types.File) []*types.Interface {
	var ifaces []*types.Interface
	for i := range file.Interfaces {
		if docsContainMicrogenTag(file.Interfaces[i].Docs) {
			ifaces = append(ifaces, &file.Interfaces[i])
		}
	}
	return ifaces
}

func listInterfaces(ii []types.Interface) string {
	var s string
	for _, i := range ii {
		s = s + fmt.Sprintf("\t%s(%d methods, %d embedded interfaces)\n", i.Name, len(i.Methods), len(i.Interfaces))
	}
	return s
}

func selectInterface(ii []*types.Interface) (*types.Interface, error) {
	if len(ii) == 0 {
		return ii[0], nil
	}
	if *flagInterface == "" {
		return nil, fmt.Errorf("%d interfaces founded, but -interface is empty. Please, provide an interface name", len(ii))
	}
	for i := range ii {
		if ii[i].Name == *flagInterface {
			return ii[i], nil
		}
	}
	return nil, fmt.Errorf("%s interface not found, but %d others are available", *flagInterface, len(ii))
}

func docsContainMicrogenTag(strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(str, generator.TagMark+generator.MicrogenMainTag) {
			return true
		}
	}
	return false
}

func processConfig() ([]byte, error) {
	file, err := os.Open("microgen.toml")
	if err != nil {
		return nil, errors.WithMessage(err, "open file")
	}
	var rawToml bytes.Buffer
	_, err = rawToml.ReadFrom(file)
	if err != nil {
		return nil, errors.WithMessage(err, "read from config")
	}
	var cfg config
	err = toml.Unmarshal(rawToml.Bytes(), &cfg)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}

}
