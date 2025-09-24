// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slog" // nee "log/slog"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
	"github.com/lmittmann/tint"

	"github.com/rockstor/rockon-validator/model"
)

const usage = `Usage:
    rockon-validator [--check] [--diff] [--write] [--root FILE] [--verbose|--debug] FILE...

Options:
    -c, --check    Check the FILE(s) for the correct syntax and return non-zero if invalid.
    -d, --diff     Check the FILE(s) for the correct syntax and output a diff if different.
    -w, --write    Check the FILE(s) and write any changes back to disk in-place.

    -r, --root     root.json file used to verify that the rockon is mentioned in said file.
                   Default: same directory as FILE

    -v, --verbose  Enable more logging
    --debug        Enable debug logging
`

var (
	checkFlag, diffFlag, writeFlag, verboseFlag, debugFlag bool
	rootFlag, rootFile                                     string
	logger                                                 *slog.Logger
)

func parseFlags() {
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	flag.BoolVar(&checkFlag, "c", false, "check the file")
	flag.BoolVar(&checkFlag, "check", false, "check the file")
	flag.BoolVar(&diffFlag, "d", false, "diff the file")
	flag.BoolVar(&diffFlag, "diff", false, "diff the file")
	flag.BoolVar(&writeFlag, "w", false, "write the file")
	flag.BoolVar(&writeFlag, "write", false, "write the file")
	flag.StringVar(&rootFlag, "r", "", "root.json file to check")
	flag.StringVar(&rootFlag, "root", "", "root.json file to check")
	flag.BoolVar(&verboseFlag, "v", false, "enable more logging")
	flag.BoolVar(&verboseFlag, "verbose", false, "enable more logging")
	flag.BoolVar(&debugFlag, "debug", false, "enable debug logging")

	flag.Parse()
}

func parseFileArgs() (filePaths []string) {
	logger.Debug("parseFileArgs()", slog.Any("called with", filePaths))
	for _, f := range flag.Args() {
		glob, _ := filepath.Glob(f)
		if glob == nil {
			logger.Error("No matching files found.")
			os.Exit(2)
		}
		filePaths = append(filePaths, glob...)
	}
	// recurse sub-directories
	//for i, f := range filePaths {
	//	files, err := os.ReadDir(f)
	//	if err != nil {
	//		continue // What we got was not a directory, so we can leave it be
	//	}
	//
	//	entries := []string{}
	//	for _, e := range files {
	//		if !e.IsDir() {
	//			entries = append(entries, filepath.Join(f, e.Name()))
	//		}
	//	}
	//	head := filePaths[:i]
	//	if i == 0 {
	//		head = []string{}
	//	}
	//	tail := filePaths[i+1:]
	//	filePaths = append(head, entries...)
	//	filePaths = append(filePaths, tail...)
	//}
	logger.Debug("paseFileArgs()", slog.Any("Return", filePaths))
	return filePaths
}

func setupLogger(logLevel *slog.LevelVar) *slog.Logger {
	logOpts := &tint.Options{
		Level: logLevel,
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return attr
		},
	}
	logHandler := tint.NewHandler(os.Stderr, logOpts)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)
	return logger
}

func checkRootMap(rootMap map[string]string, filename string, rockon model.RockOn) {
	var filenameFound bool
	var keyName string
	// Index file key expected to match lowercase Rockon name.
	for key := range rootMap {
		filenameFound = rootMap[key] == filename
		if filenameFound {
			keyName = key
			break
		}
	}
	for name := range rockon {
		if filenameFound {
			if name != keyName {
				slog.Warn("RockOn name does not match", slog.String("root.json", keyName), slog.String("rockon", name), slog.String("file", filepath.Base(filename)))
			}
		} else {
			rootMap[name] = filename
			logger.Debug("root.json map", slog.Any("rootMap", rootMap))
		}
	}
}

func main() {
	logLevel := &slog.LevelVar{}
	logLevel.Set(slog.LevelWarn)
	logger = setupLogger(logLevel)

	flag.Usage = func() {
		_, err := fmt.Fprint(os.Stderr, usage)
		if err != nil {
			logger.Error("Options print failure")
			os.Exit(6)
		}
	}

	parseFlags()

	if verboseFlag {
		logLevel.Set(slog.LevelInfo)
	}

	if debugFlag {
		logLevel.Set(slog.LevelDebug)
	}

	logger.Debug("Operation flags", slog.Bool("checkFlag", checkFlag), slog.Bool("diffFlag", diffFlag), slog.Bool("writeFlag", writeFlag))
	logger.Debug("Verbosity flags", slog.Bool("verboseFlag", verboseFlag), slog.Bool("debugFlag", debugFlag))
	logger.Debug("root.json flags", slog.String("rootFlag", rootFlag), slog.String("rootFile", rootFile))

	var numDiffFiles int
	for _, fileName := range parseFileArgs() {
		logger.Info("Checking", slog.String("file", fileName))
		data, err := os.ReadFile(fileName)
		if err != nil {
			logger.Error("Reading file", slog.String("file", fileName), slog.Any("err", err))
			os.Exit(1) // We should be able to read all the files
		}
		dataString := string(data)

		if !json.Valid(data) {
			logger.Error("Invalid JSON format", slog.String("file", fileName))
			os.Exit(3) // All files should at least parse as JSON.
		}

		rootFile = rootFlag
		if rootFlag == "" {
			rootFile = filepath.Join(filepath.Dir(fileName), "root.json")
			logger.Info("Using same-path index", slog.String("file", rootFile))
		} else {
			logger.Info("Using passed index", slog.String("file", rootFile))
		}

		rootData, err := os.ReadFile(rootFile)
		if err != nil {
			logger.Error("Reading index", slog.String("file", rootFile), slog.Any("err", err))
			os.Exit(4)
		}
		if !json.Valid(rootData) {
			logger.Error("Invalid JSON format in index", slog.String("file", rootFile))
			os.Exit(5) // All files should at least parse as JSON.
		}

		// We re-validate our rootData on every parseFileArgs() entry
		// Likely associated with multiple embedded rockon repos, each with possibly their own index file.
		rootMap := map[string]string{}
		err = json.Unmarshal(rootData, &rootMap)
		logger.Debug("root.json flags", slog.String("rootFlag", rootFlag), slog.String("rootFile", rootFile))
		if err != nil {
			logger.Error("Index validation failed for", slog.String("file", rootFile))
			os.Exit(1)
		}

		// Validate Rockon file data against RockOn model, confirming matching index entry (root.json).
		// But only:
		if fileName != rootFile {
			var rockon model.RockOn
			err = json.Unmarshal(data, &rockon)
			if err != nil {
				if filepath.Ext(fileName) == ".json" {
					logger.Error("Unmarshaling json data", slog.String("file", fileName), slog.Any("err", err))
					os.Exit(1) // File was named `*.json`, but couldn't be marshalled as expected, so we need to exit.
				}
				logger.Warn("Non *.json filename passed as input, skipping", slog.String("file", fileName))
				continue // File was not named `*.json`, so we shouldn't worry about it.
			}

			checkRootMap(rootMap, filepath.Base(fileName), rockon)

			result, err := rockon.ToJSON()
			if err != nil {
				logger.Error("Marshaling to JSON", slog.Any("err", err))
				os.Exit(1) // This should basically never happen
			}

			if dataString != result {
				numDiffFiles++
			}

			if diffFlag {
				aPath := "a/" + strings.TrimPrefix(fileName, "/")
				bPath := "b/" + strings.TrimPrefix(fileName, "/")
				edits := myers.ComputeEdits(span.URIFromPath(aPath), dataString, result)
				fmt.Println(gotextdiff.ToUnified(aPath, bPath, dataString, edits))
			}

			if writeFlag {
				stat, _ := os.Stat(fileName)
				logger.Debug("Writing rockon", slog.String("file", fileName))
				err = os.WriteFile(fileName, []byte(result), stat.Mode())
				if err != nil {
					logger.Error("Writing rockon", slog.String("file", fileName), slog.Any("err", err))
				}
				rootStat, err := os.Stat(rootFile)
				if os.IsNotExist(err) {
					rootStat = stat
				}
				rootJson, _ := json.MarshalIndent(rootMap, "", "  ")
				logger.Debug("Writing root", slog.String("file", rootFile))
				err = os.WriteFile(rootFile, rootJson, rootStat.Mode())
				if err != nil {
					logger.Error("Writing root", slog.String("file", rootFile), slog.Any("err", err))
				}
			}
		} else {
			logger.Warn("Skipped RockOn validation for index", slog.String("file", rootFile))
		}
	}

	if checkFlag {
		os.Exit(numDiffFiles)
	}
}
