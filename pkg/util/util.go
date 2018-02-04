package util

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/golang/glog"
)

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
