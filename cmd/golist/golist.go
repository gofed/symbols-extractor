package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/util"
	"github.com/urfave/cli"
)

// TODO(jchaloup):
// - add --version to print a Go version the code was built with

func main() {
	app := cli.NewApp()
	app.Name = "golist"
	app.Usage = "List Go project resources (built with Go ???)"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{"ignore-dir, d", nil, "Directory to ignore", "", false},
		cli.StringSliceFlag{"ignore-tree, t", nil, "Directory tree to ignore", "", false},
		cli.StringSliceFlag{"ignore-regex, r", nil, "Regex specified files/dirs to ignore", "", false},
		cli.StringFlag{"package-path", "", "Package entry point", "", nil, false},
		cli.BoolFlag{"all-deps", "List imported packages including stdlib", "", nil, false},
		cli.BoolFlag{"provided", "List provided packages", "", nil, false},
		cli.BoolFlag{"imported", "List imported packages", "", nil, false},
		cli.BoolFlag{"skip-self", "Skip imported packages with the same --package-path", "", nil, false},
		cli.BoolFlag{"tests", "Apply the listing options over tests", "", nil, false},
		cli.BoolFlag{"show-main", "Including main files in listings", "", nil, false},
		cli.BoolFlag{"to-install", "List all resources recognized as essential part of the Go project", "", nil, false},
		cli.StringFlag{"with-extensions", "", "Include all files with the extensions in the recognized resources, e.g. *.proto,*.tmpl", "", nil, false},
		cli.BoolFlag{"json", "Output as JSON artefact", "", nil, false},
	}

	app.Action = func(c *cli.Context) error {

		if c.String("package-path") == "" {
			return fmt.Errorf("--package-path is not set")
		}

		if !c.Bool("provided") && !c.Bool("imported") && !c.Bool("to-install") && !c.Bool("json") {
			return fmt.Errorf("At least one of --provided, --imported, --to-install or --json must be set")
		}

		ignore := &util.Ignore{}

		for _, dir := range c.StringSlice("ignore-tree") {
			// skip all ignored dirs that are prefixes of the package-path
			if strings.HasPrefix(c.String("package-path"), dir) {
				continue
			}
			ignore.Trees = append(ignore.Trees, dir)
		}

		for _, dir := range c.StringSlice("ignore-dir") {
			// skip all ignored dirs that are prefixes of the package-path
			if strings.HasPrefix(c.String("package-path"), dir) && c.String("package-path") != dir {
				continue
			}
			ignore.Dirs = append(ignore.Dirs, dir)
		}

		for _, dir := range c.StringSlice("ignore-regex") {
			ignore.Regexes = append(ignore.Regexes, regexp.MustCompile(dir))
		}

		var exts []string
		for _, item := range strings.Split(c.String("extensions"), ",") {
			if item != "" {
				exts = append(exts, item)
			}
		}

		collector := util.NewPackageInfoCollector(ignore, exts)
		if err := collector.CollectPackageInfos(c.String("package-path")); err != nil {
			return err

		}

		if c.Bool("json") {
			// Collect everything
			artifact, _ := collector.BuildArtifact()
			str, _ := json.Marshal(artifact)
			fmt.Printf("%v\n", string(str))
			return nil
		}

		if c.Bool("provided") {
			pkgs, err := collector.BuildPackageTree(c.Bool("show-main"), c.Bool("tests"))
			if err != nil {
				return err
			}
			sort.Strings(pkgs)
			for _, item := range pkgs {
				fmt.Println(item)
			}
			return nil
		}

		if c.Bool("imported") {
			pkgs, err := collector.CollectProjectDeps(c.Bool("all-deps"), c.Bool("skip-self"), c.Bool("tests"))
			if err != nil {
				return err
			}
			sort.Strings(pkgs)
			for _, item := range pkgs {
				fmt.Println(item)
			}
			return nil
		}

		if c.Bool("to-install") {
			pkgs, err := collector.CollectInstalledResources()
			if err != nil {
				return err
			}
			sort.Strings(pkgs)
			for _, item := range pkgs {
				fmt.Println(item)
			}
			return nil
		}

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
