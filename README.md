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
./rockon-validator --diff ./temp/OpenSpeedTest.json 
time=2026-01-05T13:42:05.137Z level=WARN msg="Name mismatch:" index="SpeedTest by OpenSpeedTest" expected="speedtest by openspeedtest" file=OpenSpeedTest.json
time=2026-01-05T13:42:05.137Z level=WARN msg="(if --write) Removing and adding expected entry."
--- a/./temp/OpenSpeedTest.json
+++ b/./temp/OpenSpeedTest.json
@@ -1,25 +1,21 @@
 {
   "SpeedTest by OpenSpeedTest": {
+    "description": "<p>SpeedTest by OpenSpeedTest™ is a free and open-source HTML5 network performance estimation tool. Written in vanilla Javascript it only uses built-in Web APIs like XMLHttpRequest (XHR), HTML, CSS, JS, &amp; SVG. No third-party frameworks or libraries are required.<p>Based on the official docker image: <a href='https://hub.docker.com/r/openspeedtest/latest' target='_blank'> https://hub.docker.com/r/openspeedtest/latest/</a>, available for amd64 and arm64 architecture.</p>",
+    "version": "1.0",
+    "website": "https://openspeedtest.com/",
     "containers": {
       "speedtest": {
         "image": "openspeedtest/latest",
-               "launch_order": 1,
-               "ports": {
-                       "3000": {
-                               "description": "SpeedTest http WebUI port. Suggested Default: 3000",
-                               "host_default": 3000,
-                               "label": "http WebUI port",
-                               "ui": true
-                       }
-               }
+        "launch_order": 1,
+        "ports": {
+          "3000": {
+            "description": "SpeedTest http WebUI port. Suggested Default: 3000",
+            "label": "http WebUI port",
+            "host_default": 3000,
+            "ui": true
+          }
+        }
       }
-    },
-    "description": "<p>SpeedTest by OpenSpeedTest™ is a free and open-source HTML5 network performance estimation tool. Written in vanilla Javascript it only uses built-in Web APIs like XMLHttpRequest (XHR), HTML, CSS, JS, &amp; SVG. No third-party frameworks or libraries are required.<p>Based on the official docker image: <a href='https://hub.docker.com/r/openspeedtest/latest' target='_blank'> https://hub.docker.com/r/openspeedtest/latest/</a>, available for amd64 and arm64 architecture.</p>",
-    "ui": {
-      "https": false,
-      "slug": ""
-    },
-    "website": "https://openspeedtest.com/",
-    "version": "1.0"
+    }
   }
 }
# the following is abridged:
--- a/temp/root.json
+++ b/temp/root.json
@@ -1,89 +1,89 @@
 {
...
-    "SpeedTest by OpenSpeedTest": "OpenSpeedTest.json",
...
+  "speedtest by openspeedtest": "OpenSpeedTest.json"
...
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
Requires GO version 1.24 or later.
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
