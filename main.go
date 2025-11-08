package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		outputFile = flag.String("o", "emitter_callsites.go", "output file name")
		varName    = flag.String("var", "emitterCallSiteDetails", "name of the generated variable")
		pkgName    = flag.String("package", "", "package name (if empty, derived from directory)")
    version    = flag.Bool("version", false, "print version information and exit")

	)
	flag.Parse()

  if *version {
    fmt.Println("go-emitter version 1.0.0")
    os.Exit(0)
  }

	if flag.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Usage: go-emitter [flags] <directory>\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	dir := flag.Arg(0)

	config := GeneratorConfig{
		Directory:  dir,
		OutputFile: *outputFile,
		VarName:    *varName,
		PkgName:    *pkgName,
	}

	if err := Generate(config); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
