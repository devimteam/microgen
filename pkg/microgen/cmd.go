package microgen

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/devimteam/microgen/internal/pkgpath"

	"github.com/devimteam/microgen/logger"
	lg "github.com/devimteam/microgen/logger"
	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/vetcher/go-astra/types"
)

var (
	fset = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	flagVerbose = fset.Int("v", logger.Common, "Sets microgen verbose level.")
	flagDebug   = fset.Bool("debug", false, "Print all microgen messages. Equivalent to -v=100.")
	flagConfig  = fset.String("config", "microgen.toml", "")
	flagDry     = fset.Bool("dry", false, "Do everything except writing files.")
	flagKeep    = fset.Bool("keep", false, "")
	flagForce   = fset.Bool("force", false, "Forcing microgen to overwrite files, that was marked as 'edited manually'")
)

const (
	Microgen          = "microgen"
	MicrogenVersion   = "1.0"
	DefaultFileHeader = `Code generated by microgen. DO NOT EDIT.`
)

const (
	VariablesSection  = "vars"
	GenerationSection = "generate"
	ImportsSection    = "import"
)

func Exec(args ...string) {
	err := fset.Parse(args)
	defer func() {
		if err != nil {
			lg.Logger.Logln(logger.Critical, "fatal:", err)
			os.Exit(1)
		}
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			lg.Logger.Logln(logger.Critical, "panic:", err, "\n", string(buf))
			os.Exit(1)
		}
	}()
	begin := time.Now()
	defer func() {
		lg.Logger.Logln(logger.Info, "Duration:", time.Since(begin))
	}()
	if *flagVerbose < logger.Critical {
		*flagVerbose = logger.Critical
	}
	lg.Logger.Level = *flagVerbose
	if *flagDebug {
		lg.Logger.Level = logger.Debug
		//lg.Logger.Logln(logger.Debug, "Debug logs mode in on")
	}
	lg.Logger.Logln(logger.Common, Microgen, MicrogenVersion)

	lg.Logger.Logln(logger.Detail, "Config:", *flagConfig)
	cfg, err := processConfig(*flagConfig)
	if err != nil {
		return
	}

	lg.Logger.Logln(logger.Debug, "Validate interface")
	if err = ValidateInterface(&targetInterface); err != nil {
		return
	}

	currentPkg, err := pkgpath.GetPkgPath(".", true)
	if err != nil {
		return
	}

	source, err := os.Getwd()
	if err != nil {
		return
	}
	lg.Logger.Logln(logger.Debug, "Start generation")
	ctx := Context{
		Interface:           &targetInterface,
		Source:              source,
		SourcePackageName:   sourcePackage,
		SourcePackageImport: currentPkg,
		FileHeader:          DefaultFileHeader,
		Files:               nil,
		Variables:           prepareVariables(cfg.Get(VariablesSection)),
		AllowedMethods:      makeAllowedMethods(&targetInterface),
	}
	lg.Logger.Logln(logger.Debug, "Exec plugins")
	for i, pcfg := range cfg.Get(GenerationSection).([]*toml.Tree) {
		fnErr := func() error {
			defer func() {
				if e := recover(); e != nil {
					panic(errors.Errorf("recover panic from plugin. Message: %v", e))
				}
			}()
			plugin := pcfg.GetDefault("plugin", "").(string)
			p, ok := pluginsRepository[plugin]
			if !ok {
				return errors.Errorf("plugin '%s' not registered", plugin)
			}
			var input string
			switch params := pcfg.Get("params").(type) {
			case *toml.Tree:
				input = params.String()
			case nil:
				input = ""
			default:
				return errors.Errorf("params should be tree, but got '%T'", pcfg.Get("params"))
			}
			lg.Logger.Logln(logger.Debug, "\t", i+1, "\texec plugin", "'"+plugin+"'", "with args:\n", input)
			ctx, err = p.Generate(ctx, []byte(input))
			if err != nil {
				return errors.Wrapf(err, "plugin '%s' returns an error", plugin)
			}
			return nil
		}()
		if fnErr != nil {
			err = fnErr
			return
		}
	}
	if *flagDry {
		lg.Logger.Logln(logger.Debug, "dry execution: do not create files")
		return
	}
	lg.Logger.Logln(logger.Debug, "Write files")
	for i, f := range ctx.Files {
		lg.Logger.Logln(logger.Debug, "\t", i+1, "\tcreate\t", f.Path)
		tgtFile, e := makeDirsAndCreateFile(f.Path)
		if e != nil {
			err = errors.Wrapf(e, "plugin-file '%s': during creating '%s' file", f.Name, f.Path)
			return
		}
		if tgtFile == nil {
			continue
		}
		defer tgtFile.Close()
		lg.Logger.Logln(logger.Debug, "\t", i+1, "\twrite\t", f.Path)
		_, e = tgtFile.Write(f.Content)
		if e != nil {
			err = errors.Wrapf(e, "plugin-file '%s': during writing '%s' file", f.Name, f.Path)
			return
		}
	}
	lg.Logger.Logln(logger.Info, "Done")
}

func prepareVariables(varsTree interface{}) map[string]string {
	m := make(map[string]string)
	if varsTree == nil {
		return m
	}
	switch v := varsTree.(type) {
	case *toml.Tree:
		for _, k := range v.Keys() {
			value := v.Get(k)
			if value == nil {
				continue
			}
			switch v := value.(type) {
			case string:
				m[k] = os.ExpandEnv(v)
			case int64:
				m[k] = strconv.FormatInt(v, 64)
			case uint64:
				m[k] = strconv.FormatUint(v, 64)
			case float64:
				m[k] = strconv.FormatFloat(v, 'f', -1, 64)
			case bool:
				m[k] = strconv.FormatBool(v)
			}
		}
	}
	return m
}

const mkdirPermissions = 0777

func makeDirsAndCreateFile(p string) (*os.File, error) {
	outpath, err := filepath.Abs(p)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to resolve path")
	}
	dir := filepath.Dir(outpath)

	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(dir, mkdirPermissions)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create directory '%s'", outpath)
		}
	} else if err != nil {
		return nil, errors.WithMessage(err, "could not stat file")
	}
	if !*flagForce {
		if marked, err := isMarkedManual(outpath); err != nil {
			return nil, err
		} else if marked {
			lg.Logger.Logln(lg.Critical, "WARN skip write:", p, "is marked as 'edited manually'")
			return nil, nil
		}
	}

	tgtFile, err := os.Create(p)
	if err != nil {
		return nil, errors.WithMessage(err, "create file")
	}
	return tgtFile, nil
}

var manualMark = []byte("microgen:manual")

func isMarkedManual(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	bts, err := bufio.NewReader(f).ReadBytes('\n')
	if err != nil {
		return false, errors.Wrap(err, "read bytes")
	}
	return bytes.Index(bts, manualMark) != -1, nil
}

func makeAllowedMethods(iface *Interface) map[string]bool {
	m := make(map[string]bool)
	for _, fn := range iface.Methods {
		m[fn.Name] = !FetchTags(fn.Docs, "//"+Microgen).Has("-")
	}
	return m
}

func nonTestFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
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

func docsContainMicrogenTag(strs []string) bool {
	for _, str := range strs {
		if strings.HasPrefix(str, "//"+Microgen) {
			return true
		}
	}
	return false
}

func processConfig(pathToConfig string) (*toml.Tree, error) {
	file, err := os.Open(pathToConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "open file")
	}
	var rawToml bytes.Buffer
	_, err = rawToml.ReadFrom(file)
	if err != nil {
		return nil, errors.WithMessage(err, "read from config")
	}
	tree, err := toml.LoadReader(&rawToml)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}
	return tree, nil
}

func composeErrors(errs ...error) error {
	if len(errs) > 0 {
		var strs []string
		for _, err := range errs {
			if err != nil {
				strs = append(strs, err.Error())
			}
		}
		if len(strs) == 1 {
			return fmt.Errorf(strs[0])
		}
		if len(strs) > 0 {
			return fmt.Errorf("many errors:\n%v", strings.Join(strs, "\n"))
		}
	}
	return nil
}

func Run(name string, iface interface{}) {
	for k, v := range pluginsRepository {
		fmt.Println(k, v)
	}
	fmt.Printf("%s %T", name, iface)
}
