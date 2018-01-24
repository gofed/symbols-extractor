# Symbols extractor

![tTravis CI](https://api.travis-ci.org/gofed/symbols-extractor.svg?branch=master)

STILL A DRAFT!!! STILL IN EXPERIMENTAL STATE!!!

Extracted artefacts generated in the following directory hierarchy:

```sh
generated/golang
└── 1.9.2
    ├── archive
    │   ├── tar
    │   └── zip
    ├── bufio
    │   ├── allocated.json
    │   ├── api.json
    │   └── contracts.json
    ├── builtin
    │   ├── allocated.json
    │   ├── api.json
    │   └── contracts.json
    ├── bytes
    │   ├── allocated.json
    │   ├── api.json
    │   └── contracts.json
    ...
```

#### Extraction

The golang standard library is extracted separately and must be available
before any project is processed. The standard library (e.g. of version `1.9.2`) can be extracted by running:

```bash
./extract --stdlib --symbol-table-dir generated --cgo-symbols-path cgo/cgo.yml
```

All other project packages must be available under `GOPATH` environment variable.

Assuming the standard Go library is processed, you can extract artefacts from a project (e.g. `github.com/coreos/etcd`, version `3.2.15`) by running:

```bash
./extract \
    --symbol-table-dir generated \
    --cgo-symbols-path cgo/cgo.yml \
    --package-path github.com/coreos/etcd/cmd/etcd \
    --package-prefix github.com/coreos/etcd:1b3ac99e8a431b381e633802cc42fe70e663baf5 \
    --glidefile src/github.com/coreos/etcd/glide.lock
```

.
The extractor goes through every dependency and collects various information from the source code:

```
Package "gopkg.in/yaml.v2" processed
Package "github.com/ghodss/yaml" processed
Package "github.com/coreos/go-semver/semver" processed
Package "github.com/coreos/etcd/version" processed
Package "github.com/coreos/pkg/capnslog" processed
Package "github.com/coreos/etcd/pkg/types" processed
...
```

The `1b3ac99e8a431b381e633802cc42fe70e663baf5` corresponds to etcd `3.2.15`.
All the artefacts are stored under `generated/github.com/coreos/etcd` directory:

```
generated/github.com/coreos/etcd/
├── alarm
│   └── 1b3ac99e8a431b381e633802cc42fe70e663baf5
│       ├── allocated.json
│       ├── api.json
│       └── contracts.json
├── auth
│   ├── 1b3ac99e8a431b381e633802cc42fe70e663baf5
│   │   ├── allocated.json
│   │   ├── api.json
│   │   └── contracts.json
│   └── authpb
│       └── 1b3ac99e8a431b381e633802cc42fe70e663baf5
├── client
│   ├── 1b3ac99e8a431b381e633802cc42fe70e663baf5
│   │   ├── allocated.json
│   │   ├── api.json
│   │   └── contracts.json
├── clientv3
│   ├── 1b3ac99e8a431b381e633802cc42fe70e663baf5
│   │   ├── allocated.json
│   │   ├── api.json
│   │   └── contracts.json
│   ├── concurrency
│   │   └── 1b3ac99e8a431b381e633802cc42fe70e663baf5
│   ├── namespace
│   │   └── 1b3ac99e8a431b381e633802cc42fe70e663baf5
│   └── naming
│       └── 1b3ac99e8a431b381e633802cc42fe70e663baf5
├── cmd
│   └── etcd
│       └── 1b3ac99e8a431b381e633802cc42fe70e663baf5
```

The extractor is not capable of detecting which version of the project is
processed. So it is up to a invoker to set the proper commit under which
the extracted information are stored.

If the `--stdlib` option is not set, the latest available processed version is used.

#### Allocation of symbols

As the main purpose of the extractor is to collect a list of symbols imported (a.k.a allocated) in a Go source code,
you can generate various statistics and derive other results based on the data. E.g.
- how much a Go project/package depends on another (how many symbols a package `A` imports/allocates from a package `B`)
- how many times is a given symbol imported in a project (does it make sense to wrap the symbol into a function call and call the function instead?)
- how often a symbol imported from another project changes?

To get a list of imported symbols for a given Go package you can run the `extract` command with the `allocated` flag:

```sh
./extract \
    --symbol-table-dir generated \
    --cgo-symbols-path cgo/cgo.yml \
    --package-path github.com/coreos/etcd/cmd/etcdctl \
    --package-prefix github.com/coreos/etcd:1b3ac99e8a431b381e633802cc42fe70e663baf5 \
    --glidefile src/github.com/coreos/etcd/glide.lock \
    --allocated
Package "github.com/coreos/etcd/cmd/etcdctl" already processed
file: github.com/coreos/etcd/cmd/etcdctl
==========================================================================
	F: fmt.Fprintln:                              	1
	F: github.com/coreos/etcd/etcdctl/ctlv2.Start:	1
	F: github.com/coreos/etcd/etcdctl/ctlv3.Start:	1
	F: os.Exit:                                   	1
	F: os.Getenv:                                 	1
	F: os.Unsetenv:                               	1
	V: os.Stderr:                                 	1
==========================================================================
```



#### API Compatibility detection

All projects (and its dependencies) must be processed before the API check is performed.

To get a list of detected backward incompatibilities (e.g. checking kube-apiserver `v1.9.0` with respect to etcd `v2.0.9`) you can run:

```bash
./checkapi \
    --allocated apiserver.json \
    --package-prefix github.com/coreos/etcd/client \
    --allocated-godepsfile src/k8s.io/kubernetes/Godeps/Godeps.json \
    --package-commit 02697ca725e5c790cc1f9d0918ff22fad84cb4c5 \
    --godepsfile src/github.com/coreos/etcd/Godeps/Godeps.json \
    --symbol-table-dir generated \
    --go-version 1.9.2
```

The `--allocated` flag points to a file with a list of all symbols imported by a given package (including all its dependencies). In this case it corresponds to `cmd/kube-apiserver`. The `apiserver.json` was generated by running:

```sh
./extract \
    --symbol-table-dir generated \
    --cgo-symbols-path cgo/cgo.yml \
    --package-path k8s.io/kubernetes/cmd/kube-apiserver \
    --package-prefix k8s.io/kubernetes:925c127ec6b946659ad0fd596fa959be43f0cc05 \
    --godepsfile src/k8s.io/kubernetes/Godeps/Godeps.json \
    --allocated \
    -recursive-from k8s.io/apiserver \
    --json > apiserver.json
```

The `925c127ec6b946659ad0fd596fa959be43f0cc05` of `github.com/kubernetes/kubernetes` projects corresponds to `v1.9.0`.

Once the `checkapi` command is run with the generated `apiserver.json`, the following (shorten) list of reports can be generated:

```sh
Comparing github.com/coreos/etcd/client:0520cb9304cb2385f7e72b8bc02d6e4d3257158a with github.com/coreos/etcd/client:02697ca725e5c790cc1f9d0918ff22fad84cb4c5
-type "github.com/coreos/etcd/client.Client" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:2439
	used at k8s.io/apiserver/pkg/storage/storagebackend/factory/etcd2.go:1386
-type "github.com/coreos/etcd/client.WatcherOptions" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_watcher.go:6769
-type "github.com/coreos/etcd/clientv3.GetResponse" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:10930
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:21482
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:7819
-type "github.com/coreos/etcd/clientv3.Cmp" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:26018
...
-function "github.com/coreos/etcd/client.ErrorCodeNodeExist" missing
	used at k8s.io/apiserver/pkg/storage/etcd/util/etcd_util.go:1015
-function "github.com/coreos/etcd/client.ErrorCodeTestFailed" missing
	used at k8s.io/apiserver/pkg/storage/etcd/util/etcd_util.go:1190
-function "github.com/coreos/etcd/clientv3.EventTypeDelete" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/event.go:1275
-function "github.com/coreos/etcd/client.PrevNoExist" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:18772
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:5144
-function "github.com/coreos/etcd/client.ErrClusterUnavailable" missing
	used at k8s.io/apiserver/pkg/storage/etcd/util/etcd_util.go:1757
-function "github.com/coreos/etcd/client.ErrorCodeEventIndexCleared" missing
	used at k8s.io/apiserver/pkg/storage/etcd/util/etcd_util.go:1499
-function "github.com/coreos/etcd/client.ErrorCodeKeyNotFound" missing
	used at k8s.io/apiserver/pkg/storage/etcd/util/etcd_util.go:830
-function "github.com/coreos/etcd/clientv3.ModRevision" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:10659
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:26058
	used at k8s.io/apiserver/pkg/storage/etcd3/store.go:7610
...
-field "github.com/coreos/etcd/clientv3.WatchResponse.Events" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/watcher.go:6422
-field "github.com/coreos/etcd/clientv3.Config.TLS" missing
	used at k8s.io/apiserver/pkg/storage/storagebackend/factory/etcd3.go:1385
-field "github.com/coreos/etcd/client.SetOptions.PrevIndex" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:19589
-field "github.com/coreos/etcd/client.Node.Expiration" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:17735
-field "github.com/coreos/etcd/client.GetOptions.Quorum" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:12424
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:16033
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:10238
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_watcher.go:7454
-field "github.com/coreos/etcd/client.Error.Index" missing
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_helper.go:16210
	used at k8s.io/apiserver/pkg/storage/etcd/etcd_watcher.go:7834
...
-method "github.com/coreos/etcd/clientv3.WatchResponse.Err" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/watcher.go:6184
	used at k8s.io/apiserver/pkg/storage/etcd3/watcher.go:6214
-method "github.com/coreos/etcd/clientv3.Client.Endpoints" missing
	used at k8s.io/apiserver/pkg/storage/etcd3/compact.go:4636
	used at k8s.io/apiserver/pkg/storage/etcd3/compact.go:5681
	used at k8s.io/apiserver/pkg/storage/etcd3/compact.go:1587
	used at k8s.io/apiserver/pkg/storage/etcd3/compact.go:1709
	used at k8s.io/apiserver/pkg/storage/etcd3/compact.go:1766
...
```
