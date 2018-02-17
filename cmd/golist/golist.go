package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/util"
)

// TODO(jchaloup):
// - add --version to print a Go version the code was built with

type flags struct {
	// Go package entry point
	packagePath *string
	// List all dependencies
	allDeps *bool
	// List provided packages
	provided *bool
	// List imported packages
	imported *bool
	// List of files to install
	toinstall  *bool
	extensions *string
	// Skip all packages prefixed with packagePath
	skipSelf *bool
	// search tests instead
	tests    *bool
	showMain *bool
	//
	ignoreTrees *string
	ignoreDirs  *string
	ignoreRegex *string
}

func buildFlags() *flags {
	return &flags{
		packagePath: flag.String("package-path", "", "Package entry point"),
		allDeps:     flag.Bool("all-deps", false, "List imported packages including stdlib"),
		provided:    flag.Bool("provided", false, "List provided packages"),
		imported:    flag.Bool("imported", false, "List imported packages"),
		skipSelf:    flag.Bool("skip-self", false, "Skip imported packages with the same --package-path"),
		tests:       flag.Bool("tests", false, "Apply the listing options over tests"),
		showMain:    flag.Bool("show-main", false, "Including main files in listings"),
		ignoreTrees: flag.String("ignore-trees", "", "Directory trees to ignore"),
		ignoreDirs:  flag.String("ignore-dirs", "", "Directories to ignore"),
		ignoreRegex: flag.String("ignore-regex", "", "Directories to ignore specified by a regex"),
		toinstall:   flag.Bool("to-install", false, "List all resources recognized as essential part of the Go project"),
		extensions:  flag.String("with-extensions", "", "Include all files with the extensions in the recognized resources, e.g. *.proto,*.tmpl"),
	}
}

func (f *flags) parse() error {
	flag.Parse()

	if *f.packagePath == "" {
		return fmt.Errorf("--package-path is not set")
	}

	if !*f.provided && !*f.imported && !*f.toinstall {
		return fmt.Errorf("At least one of --provided, --imported or --to-install must be set")
	}

	return nil
}

func main() {
	f := buildFlags()
	if err := f.parse(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	ignore := &util.Ignore{}

	if *f.ignoreTrees != "" {
		for _, dir := range strings.Split(*f.ignoreTrees, ",") {
			// skip all ignored dirs that are prefixes of the package-path
			if strings.HasPrefix(*f.packagePath, dir) {
				continue
			}
			ignore.Trees = append(ignore.Trees, dir)
		}
	}

	if *f.ignoreDirs != "" {
		for _, dir := range strings.Split(*f.ignoreDirs, ",") {
			// skip all ignored dirs that are prefixes of the package-path
			if strings.HasPrefix(*f.packagePath, dir) {
				continue
			}
			ignore.Dirs = append(ignore.Dirs, dir)
		}
	}

	if *f.ignoreRegex != "" {
		ignore.Regex = regexp.MustCompile(*f.ignoreRegex)
	}

	if *f.provided {
		pkgs, err := util.BuildPackageTree(*f.packagePath, *f.showMain, *f.tests, ignore)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		sort.Strings(pkgs)
		for _, item := range pkgs {
			fmt.Println(item)
		}
		return
	}

	if *f.imported {
		pkgs, err := util.CollectProjectDeps(*f.packagePath, *f.allDeps, *f.skipSelf, *f.tests, ignore)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		sort.Strings(pkgs)
		for _, item := range pkgs {
			fmt.Println(item)
		}
	}

	if *f.toinstall {
		var exts []string
		for _, item := range strings.Split(*f.extensions, ",") {
			if item != "" {
				exts = append(exts, item)
			}
		}
		pkgs, err := util.CollectInstalledResources(*f.packagePath, exts, ignore)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}
		sort.Strings(pkgs)
		for _, item := range pkgs {
			fmt.Println(item)
		}
	}

}
