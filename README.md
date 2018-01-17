# symbols-extractor

![tTravis CI](https://api.travis-ci.org/gofed/symbols-extractor.svg?branch=master)

STILL A DRAFT!!!

Extracted artefacts generated in the following directory hierarchy:

```bash
generated
├── golang
│   └── 1.9.1
│       └── fmt
│       └── net
│           └── http
├── github.com
│   └── BurntSushi
│       └── toml
│           └──a368813c5e648fee92e5f6c30e3944ff9d5e8895
│              └── api.json
│              └── contracts.json
│              └── allocated.json
├── golang.org
```

#### Extraction

The golang standard library is extracted separately and must be available
before any project is processed. The standard library (e.g. of version `1.9.1`) can be extracted by running:

```bash
./extract --stdlib=1.9.1
```

To extract artefacts from a project (assuming the standard Go library is processed) you can run:

```bash
./extract \
    --project=github.com/coreos/etcd:7a8c192c8f24cbc77e505cabc5153108f59f789c
    --stdlib=1.9.1
```

If the `--stdlib` option is not set, the latest available processed version is used.

#### API Compatibility detection

To get a list of detected backward incompatibilities you can run:

```bash
./check-api \
    --project=github.com/coreos/etcd:7a8c192c8f24cbc77e505cabc5153108f59f789c
    --wrt=github.com/urfave/cli:cfb38830724cc34fedffe9a2a29fb54fa9169cd1
```

If any of the projects is not yet pre-processed, all artefacts are extracted and stored before the API check is performed.

In case a local project needs to be checked, you can run:

```bash
./check-api \
    --project-from-dir=<PATH>
    --wrt-from-dir=<PATH>
```

Artefacts of anonymous project (with unknown fully qualified identifier in a form ``PACKAGE:COMMIT``) are dropped once a command finishes.
