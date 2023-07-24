# rockon-validator

This is a small script to validate and format RockOn json files. To install, simply run:

```
go install github.com/rockstor/rockon-validator@latest
# NOTE: Requires go version 1.20 or greater
```

and then to run, use any of the three options `--check`, `--diff`, or `--write` to validate your file:

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

For example, to Check that your file meets the correct format:

```
rockon-validator -c rockon.json
```

will exit with `0` (success), or non-zero (`1` in this case) if the file does not meet the correct format.

Similarly, `-d` will output a diff between the existing and expected format,

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

and `-w` will re-write the file, assuming it meets the correct syntax, but not the right formatting.

## Multiple files

Multiple files (or glob patterns) can be passed to validate several files simultaneously.

## Root.json

In addition, the script will check for a `root.json` file in the same directory as the given file (or files)
and ensure that an entry exists for said file in the `root.json`, and that the name referenced matches, warning
if they differ. If the `--root` flag is passed with a path to a `root.json` file, that file will be used instead.

No change to the name is made if they are different, but an entry is added if it is missing.

## Docker

If you do not have or want go 1.20+ on your machine, you can use the Docker container provided instead.

To build, run:

```
docker build -t validator:latest .
```

And then to run, mount the directory containing your rockon file(s) to `/files` in the container:

```
docker run -v $(pwd):/files validator -w rockon.json
```
