package main

import (
	"encoding/json"
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
	// Display as JSON artefact
	json *bool
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
		json:        flag.Bool("json", false, "Output as JSON artefact"),
	}
}

func (f *flags) parse() error {
	flag.Parse()

	if *f.packagePath == "" {
		return fmt.Errorf("--package-path is not set")
	}

	if !*f.provided && !*f.imported && !*f.toinstall && !*f.json {
		return fmt.Errorf("At least one of --provided, --imported, --to-install or --json must be set")
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
			if strings.HasPrefix(*f.packagePath, dir) && *f.packagePath != dir {
				continue
			}
			ignore.Dirs = append(ignore.Dirs, dir)
		}
	}

	if *f.ignoreRegex != "" {
		ignore.Regex = regexp.MustCompile(*f.ignoreRegex)
	}

	var exts []string
	for _, item := range strings.Split(*f.extensions, ",") {
		if item != "" {
			exts = append(exts, item)
		}
	}

	c := util.NewPackageInfoCollector(ignore, exts)
	if err := c.CollectPackageInfos(*f.packagePath); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if *f.json {
		// Collect everything
		artifact, _ := c.BuildArtifact()
		str, _ := json.Marshal(artifact)
		fmt.Printf("%v\n", string(str))
		return
	}

	if *f.provided {
		pkgs, err := c.BuildPackageTree(*f.showMain, *f.tests)
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
		pkgs, err := c.CollectProjectDeps(*f.allDeps, *f.skipSelf, *f.tests)
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
		pkgs, err := c.CollectInstalledResources()
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
