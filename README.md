# Rockon-validator

A GO script to validate [Rockstor](https://rockstor.com/)'s [Rock-on](https://github.com/rockstor/rockon-registry) definitions.
See also: [Rock-ons (Docker Plugins)](https://rockstor.com/docs/interface/overview.html).

## GO download and install

Requires GO version 1.20 or later.
- [Upstream install instructions](https://go.dev/doc/install)

Alternatively a docker image definition is included.
See the: [Docker](#docker) subsection below for details.


## Script install/run
```
go install github.com/rockstor/rockon-validator@latest
```

### Run options

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

To Check `rockon.json` meets formating guidelines:

```
rockon-validator --check rockon.json
```

Returns `0` (success), or `1` (fail).

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

## Docker

This repo contains a docker image using the "FROM golang as builder" directive.

### Build container

```
docker build -t rockon-validator:latest .
```

### Container use

Volume mount the local rock-on file(s) directory under `/files` within the container.
In the following example we use the current/working directory,
and specify a single file to validate. 

```
docker run -v $(pwd):/files rockon-validator --diff rockon.json
```
