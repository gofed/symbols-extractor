package util

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gofed/symbols-extractor/pkg/util/internal/load"
	"github.com/gofed/symbols-extractor/pkg/util/internal/work"
	"github.com/golang/glog"
)

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func findPackageLocation(packagePath string) (string, string, error) {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return "", "", fmt.Errorf("GOPATH not set")
	}

	var abspath, pathPrefix string

	// first, find the absolute path
	for _, gpath := range strings.Split(gopath, ":") {
		abspath = path.Join(gpath, "src", packagePath)

		if e, err := exists(abspath); err == nil && e {
			pathPrefix = path.Join(gpath, "src")
			break
		}
	}

	if pathPrefix == "" {
		return "", "", fmt.Errorf("Path %v not found on %v", packagePath, gopath)
	}

	return abspath, pathPrefix, nil
}

// Ignore specifies a set of resources to ignore
type Ignore struct {
	Dirs  []string
	Trees []string
	Regex *regexp.Regexp
}

func (ignore *Ignore) ignore(path string) bool {
	if ignore.Regex != nil && ignore.Regex.MatchString(path) {
		return true
	}

	for _, dir := range ignore.Trees {
		if strings.HasPrefix(path+"/", dir+"/") {
			return true
		}
	}

	for _, dir := range ignore.Dirs {
		if path == dir {
			return true
		}
	}
	return false
}

func collectPackageInfos(abspath, pathPrefix string, ignore *Ignore) (map[string]*load.PackagePublic, error) {
	pkgs := make(map[string]*load.PackagePublic)

	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		relPath := path[len(pathPrefix)+1:]
		if strings.HasSuffix(relPath, "/") {
			relPath = relPath[:len(relPath)-1]
		}
		// skip .git directory
		if strings.HasSuffix(relPath, ".git") {
			return filepath.SkipDir
		}

		// skip vendor directory
		if strings.HasSuffix(relPath, "/vendor") {
			return filepath.SkipDir
		}

		if ignore.ignore(relPath) {
			return nil
		}

		pkgInfo, err := ListPackage(relPath)
		if err != nil {
			if strings.Contains(err.Error(), "no Go files in") {
				return nil
			}
			// TODO(jchaloup): remove later
			if strings.Contains(err.Error(), "build constraints exclude all Go files in") {
				return nil
			}
			panic(err)
			return nil
		}
		if len(pkgInfo.GoFiles) > 0 {
			pkgs[relPath] = pkgInfo
		}

		return nil
	}

	err := filepath.Walk(abspath+"/", visit)
	if err != nil {
		return nil, err
	}

	return pkgs, nil
}

type OtherResources struct {
	ProtoFiles []string
	TmplFiles  []string
	MDFiles    []string
	Other      []string
}

func collectOtherResources(abspath, pathPrefix string, extensions []string, ignore *Ignore) (*OtherResources, error) {
	otherResources := &OtherResources{}

	visit := func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			for _, ext := range extensions {
				if strings.HasSuffix(path, ext) {
					switch ext {
					case ".proto":
						otherResources.ProtoFiles = append(otherResources.ProtoFiles, path)
					case ".md":
						otherResources.MDFiles = append(otherResources.MDFiles, path)
					default:
						otherResources.Other = append(otherResources.Other, path)
					}
				}
			}
			return nil
		}

		relPath := path[len(pathPrefix)+1:]
		if strings.HasSuffix(relPath, "/") {
			relPath = relPath[:len(relPath)-1]
		}
		// skip .git directory
		if strings.HasSuffix(relPath, ".git") {
			return filepath.SkipDir
		}

		// skip vendor directory
		if strings.HasSuffix(relPath, "/vendor") {
			return filepath.SkipDir
		}

		if ignore.ignore(relPath) {
			return nil
		}

		return nil
	}

	err := filepath.Walk(abspath+"/", visit)
	if err != nil {
		return nil, err
	}

	return otherResources, nil
}

func CollectInstalledResources(packagePath string, extensions []string, ignore *Ignore) ([]string, error) {
	abspath, pathPrefix, err := findPackageLocation(packagePath)
	if err != nil {
		return nil, err
	}

	pkgInfos, err := collectPackageInfos(abspath, pathPrefix, ignore)
	if err != nil {
		return nil, err
	}

	var resources []string

	for _, info := range pkgInfos {
		resources = append(resources, info.Dir)
		for _, file := range info.GoFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.CgoFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.CFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.CXXFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.MFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.HFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.FFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.SFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.SwigFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.SwigCXXFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.SysoFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.TestGoFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}
		for _, file := range info.XTestGoFiles {
			resources = append(resources, path.Join(info.Dir, file))
		}

	}

	otherResources, err := collectOtherResources(abspath, pathPrefix, extensions, ignore)
	if err != nil {
		return nil, err
	}

	if otherResources != nil {
		if otherResources.ProtoFiles != nil {
			resources = append(resources, otherResources.ProtoFiles...)
		}
		if otherResources.MDFiles != nil {
			resources = append(resources, otherResources.MDFiles...)
		}
		if otherResources.Other != nil {
			resources = append(resources, otherResources.Other...)
		}
	}

	return resources, nil
}

func CollectProjectDeps(packagePath string, standard bool, skipSelf bool, tests bool, ignore *Ignore) ([]string, error) {
	abspath, pathPrefix, err := findPackageLocation(packagePath)
	if err != nil {
		return nil, err
	}

	pkgInfos, err := collectPackageInfos(abspath, pathPrefix, ignore)
	if err != nil {
		return nil, err
	}

	imports := make(map[string]struct{})

	if tests {
		for _, info := range pkgInfos {
			for _, item := range info.TestImports {
				if item == "C" {
					continue
				}
				if pos := strings.LastIndex(item, "/vendor/"); pos != -1 {
					item = item[pos+8:]
				}
				if _, ok := imports[item]; !ok {
					imports[item] = struct{}{}
				}
			}
		}
	} else {
		for _, info := range pkgInfos {
			for _, item := range info.Imports {
				if item == "C" {
					continue
				}
				if pos := strings.LastIndex(item, "/vendor/"); pos != -1 {
					item = item[pos+8:]
				}
				if _, ok := imports[item]; !ok {
					imports[item] = struct{}{}
				}
			}
		}
	}

	var pkgs []string

	for relPath := range imports {
		if skipSelf && strings.HasPrefix(relPath, packagePath) {
			continue
		}

		if ignore.ignore(relPath) {
			continue
		}

		pkgInfo, err := ListPackage(relPath)
		// assuming the stdlib is always processed properly
		if !standard && err == nil && pkgInfo.Standard {
			continue
		}

		pkgs = append(pkgs, relPath)
	}

	return pkgs, nil
}

func BuildPackageTree(packagePath string, includeMain bool, tests bool, ignore *Ignore) ([]string, error) {
	// TODO(jchaloup): strip all main package unless explicitely requested
	abspath, pathPrefix, err := findPackageLocation(packagePath)
	if err != nil {
		return nil, err
	}

	pkgInfos, err := collectPackageInfos(abspath, pathPrefix, ignore)
	if err != nil {
		return nil, err
	}

	var entryPoints []string
	if tests {
		for p, pkgInfo := range pkgInfos {
			if len(pkgInfo.TestGoFiles) > 0 {
				entryPoints = append(entryPoints, p)
			}
		}
	} else {
		for p, pkgInfo := range pkgInfos {
			// check package name of each file
			var nonMainFiles []string
			files := pkgInfo.GoFiles
			files = append(files, pkgInfo.CgoFiles...)
			for _, file := range files {
				f, err := parser.ParseFile(token.NewFileSet(), path.Join(pkgInfo.Dir, file), nil, 0)
				if err != nil {
					return nil, err
				}
				if !includeMain && f.Name.Name == "main" {
					continue
				}
				nonMainFiles = append(nonMainFiles, file)
			}
			if len(nonMainFiles) > 0 {
				entryPoints = append(entryPoints, p)
			}
		}
	}

	return entryPoints, nil
}

func ListPackage(path string) (*load.PackagePublic, error) {
	// TODO(jchaloup): more things need to be init most likely
	work.BuildModeInit()

	d := load.PackagesAndErrors([]string{path})
	if d == nil {
		return nil, fmt.Errorf("No package listing found for %v", path)
	}

	pkg := d[0]
	if pkg.Error != nil {
		return nil, pkg.Error
	}
	// Show vendor-expanded paths in listing
	pkg.TestImports = pkg.Vendored(pkg.TestImports)
	pkg.XTestImports = pkg.Vendored(pkg.XTestImports)
	return &pkg.PackagePublic, nil
}

func ListGoFiles(packagePath string, cgo bool) ([]string, error) {

	collectFiles := func(output string) []string {
		line := strings.Split(string(output), "\n")[0]
		line = line[1 : len(line)-1]
		if line == "" {
			return nil
		}
		return strings.Split(line, " ")
	}
	// check GOPATH/packagePath
	filter := "{{.GoFiles}}"
	if cgo {
		filter = "{{.CgoFiles}}"
	}
	cmd := exec.Command("go", "list", "-f", filter, packagePath)
	output, e := cmd.CombinedOutput()
	if e == nil {
		return collectFiles(string(output)), nil
	}

	if strings.Contains(string(output), "no buildable Go source files in") {
		return nil, nil
	}

	// if strings.Contains(string(output), "no Go files in") {
	// 	return nil, nil
	// }

	return nil, fmt.Errorf("%v: %v, %v", strings.Join(cmd.Args, " "), e, string(output))
}

func GetPackageFiles(packageRoot, packagePath string) (files []string, packageLocation string, err error) {

	files, ppath, e := func() ([]string, string, error) {
		var searched []string
		// First searched the vendor directories
		pathParts := strings.Split(packageRoot, string(os.PathSeparator))
		for i := len(pathParts); i >= 0; i-- {
			vendorpath := path.Join(path.Join(pathParts[:i]...), "vendor", packagePath)
			glog.V(1).Infof("Checking %v directory", vendorpath)
			if l, e := ListGoFiles(vendorpath, false); e == nil {
				return l, vendorpath, e
			}
			searched = append(searched, vendorpath)
		}

		glog.V(1).Infof("Checking %v directory", packagePath)
		if l, e := ListGoFiles(packagePath, false); e == nil {
			return l, packagePath, e
		}
		searched = append(searched, packagePath)

		return nil, "", fmt.Errorf("Unable to find %q in any of:\n\t\t%v\n", packagePath, strings.Join(searched, "\n\t\t"))
	}()

	if e != nil {
		return nil, "", e
	}

	// cgo files enabled?
	cgoFiles, e := ListGoFiles(ppath, true)
	if e != nil {
		return nil, "", e
	}

	if len(cgoFiles) > 0 {
		files = append(files, cgoFiles...)
	}

	{
		cmd := exec.Command("go", "list", "-f", "{{.Dir}}", ppath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return nil, "", fmt.Errorf("go list -f {{.Dir}} %v failed: %v", ppath, err)
		}
		lines := strings.Split(string(output), "\n")
		packageLocation = string(lines[0])
	}

	return files, packageLocation, nil
}
