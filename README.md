# Rockon-validator

A GO script to validate [Rockstor](https://rockstor.com/)'s [Rock-on](https://github.com/rockstor/rockon-registry) definitions.
See also: [Rock-ons (Docker Plugins)](https://rockstor.com/docs/interface/overview.html).

## Function

Validation is performed via the GO [encoding/json](https://pkg.go.dev/encoding/json) standard library package.
Initially via the built-in `Valid` function, to establish basic JSON formatting.
There-after the JSON string is [Unmarshal](https://pkg.go.dev/encoding/json#Unmarshal)'ed as follows:
- index file: `map[string]string{}` - see main.go.
- Rock-on definition file: `map[string]RockonDetails` - see model/rockon.go.

Where `RockonDetails` is a GO embedded [struct](https://go.dev/tour/moretypes/2) defining the expected field types.
If there is a failure to transition the JSON string (file content) to the RockonDetails struct,
a failed validation (rc=1) error is returned.

### Omitted entries

Many json Rock-on definition elements are optional: e.g. `icon`, `more_info`, `devices` etc.
Where-as in a GO `struct`, our backing validation, all possible fields are defined.
To handle this miss-match the [omitempty](https://www.sohamkamani.com/golang/omitempty/) json tag is used.
This effectively ignores/removes empty, default, or missing json elements during Un/marshalling from/to JSON string format.

For the more advanced `--diff` and `--write` options, this can have surprising consequences.
I.e. if we want to maintain an explicit `"uid:" 0` element, which would otherwise be removed as a default int32 value,
we can instead use a pointer to int32, default is undefined: ergo no json element no json marshalling (struct to json).
Embedded struct pointers, and custom variable types can also approach this same problem re `omitemtpy` compliance.
All of the above approaches are used within this project.

## Docker run

Volume mount the definitions directory under `/files` within the container.
Here we use the current/working directory, asking for any differences from the recommended format/content. 
See [Run options](#run-options).

```shell
docker run -v $(pwd):/files ghcr.io/rockstor/rockon-validator:main --diff forgejo-runner.json
```

See the: [Development](#development) subsection below for running without docker,
and for locally building both the GO binary and Docker container.   

## Run options

```
rockon-validator [--check] [--diff] [--write] [--root FILE] [--verbose|--debug] FILE...

Options:
    -c, --check    Check the FILE(s) for the correct syntax and return non-zero if invalid.
    -d, --diff     Check the FILE(s) for the correct syntax and output a diff if different.
    -w, --write    Check the FILE(s) and write any changes back to disk in-place.

    -r, --root     root.json file used to verify that the rockon is mentioned in said file.
                   Default: same directory as FILE

    -v, --verbose  Enable more logging
    --debug        Enable debug logging
```

## Example

To Check `forgejo-runner.json` meets formating guidelines.
N.B. uses a `go install` instantiated binary; see [Development](#development) below.

```shell
~/go/bin/rockon-validator --check forgejo-runner.json
```

Returns:
- `0` Success: file is valid; or `--diff` to valid status was achieved.   
- `1` Failed validation.
- `2` No matching files found.
- `3` Invalid JSON format file=./path/not-json.json.
- `4` No matching index file: defaults to root.json (same path as tested definition).
- `5` Invalid JSON format in index file=./path/root.json.
- `6` Overwrite Rockon file error.
- `7` Overwire/write index file error.

Similarly, `--diff` produces a `diffutils` formated output re: existing and proposed file format:

```diff
--- a/files/bitcoind.json
+++ b/files/bitcoind.json
@@ -1,21 +1,25 @@
 {
     "Bitcoin": {
+        "description": "Bitcoin full node. <p>Based on a custom docker image: <a href='https://hub.docker.com/r/kylemanna/bitcoind' target='_blank'>https://hub.docker.com/r/kylemanna/bitcoind</a>, available for amd64 architecture only.</p>",
+        "version": "1.1",
+        "website": "https://bitcoin.org/en/full-node",
+        "volume_add_support": true,
         "containers": {
             "bitcoind": {
                 "image": "kylemanna/bitcoind",
                 "launch_order": 1,
                 "ports": {
-                    "8333": {
-                        "description": "Listening port",
-                        "host_default": 28333,
-                        "label": "Port for incoming connections",
-                        "protocol": "tcp"
-                    },
                     "8332": {
                         "description": "JSONRPC port",
-                        "host_default": 28332,
                         "label": "The JSONRPC server allows to query and control the server remotely",
+                        "host_default": 28332,
                         "protocol": "tcp"
+                    },
+                    "8333": {
+                        "description": "Listening port",
+                        "label": "Port for incoming connections",
+                        "host_default": 28333,
+                        "protocol": "tcp"
                     }
                 },
                 "volumes": {
@@ -26,10 +21,6 @@
                     }
                 }
             }
-        },
-        "description": "Bitcoin full node. <p>Based on a custom docker image: <a href='https://hub.docker.com/r/kylemanna/bitcoind' target='_blank'>https://hub.docker.com/r/kylemanna/bitcoind</a>, available for amd64 architecture only.</p>",
-        "volume_add_support": true,
-        "website": "https://bitcoin.org/en/full-node",
-        "version": "1.1"
+        }
     }
 }
```

N.B. `--write` **USE WITH CAUTION** re-writes the file in-place; assuming correct syntax but incorrect formatting.

## Multiple files

Multiple files (or glob patterns) can be passed to validate several files simultaneously.

## The root.json index

The script checks for a `root.json` index file and ensures a matching entry exists for the processed definitions.
A warning will result for differing names, but an entry is added if non is found.
An alternative target index file can be passed via the `--root` flag.

## Development

Details more associated with development.

### GO download and install

Alternatives to the `docker run` approach, and required for development purposes.
Requires GO version 1.23 or later.
- [Upstream install instructions](https://go.dev/doc/install)

### Go build

Creates a `rockon-validator` binary directly in/from the source root:

```shell
go build
```

Run via:

```shell
./rockon-validator
````

### Script install/run (from repo)

Builds and installs the `rockon-validator` binary directly from GitHub repo.

```shell
go install github.com/rockstor/rockon-validator@latest
```

Run via:

```shell
~/go/bin/rockon-validator
```

### Script run (from local source)

Builds and runs `rockon-validator` from within a local copy of the source.

```shell
go run . --diff ./path/to/OpenSpeedTest.json
```

### Build container

This repo contains a docker image using the "FROM golang as builder" directive.
See also [Script docker run](#docker-run) for pre-built container use.

```shell
docker build -t rockon-validator-local:latest .
```

Run via:

```shell
docker run -v $(pwd):/files rockon-validator-local --diff rockon.json
```
