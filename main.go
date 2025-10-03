// SPDX-License-Identifier: GPL-3.0-or-later
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
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
	indexOrigContent                                       string
	fileInfo                                               os.FileInfo
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
	for _, f := range flag.Args() {
		glob, _ := filepath.Glob(f)
		if glob == nil {
			logger.Error("No matching files found.")
			os.Exit(2)
		}
		filePaths = append(filePaths, glob...)
	}
	// recurse subdirectories
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
	logger = slog.New(logHandler)
	slog.SetDefault(logger)
	return logger
}

func checkRootMap(rootMap map[string]string, filename string, rockon model.RockOn) {
	filenameFound, keyName := false, ""
	// Index file key expected to match lowercase Rockon name.
	for key, value := range maps.All(rootMap) {
		filenameFound = value == filename
		if filenameFound {
			keyName = key
			break
		}
	}

	// maps.keys(rockon) returns an iterator over our single entry Rockon map.
	// slices.Collect enables retrieval of key by index on slice.
	// https://pkg.go.dev/iter#hdr-Standard_Library_Usage
	var rockonTitle = slices.Collect(maps.Keys(rockon))[0]
	var lowerCaseName = strings.ToLower(rockonTitle)
	if filenameFound {
		if lowerCaseName != keyName {
			slog.Info("Found match in index for", slog.String("filename", filename))
			slog.Warn("Name mismatch:", slog.String("index", keyName), slog.String("expected", lowerCaseName), slog.String("file", filepath.Base(filename)))
			slog.Warn("(if --write) Removing and adding expected entry.")
			delete(rootMap, keyName)
		}
	} else {
		slog.Warn("No match in index for", slog.String("filename", filename))
		slog.Info("(if --write) Adding entry", slog.String("index", lowerCaseName), slog.String("filename", filename))
	}
	rootMap[lowerCaseName] = filename
	logger.Debug("root.json map", slog.Any("rootMap", rootMap))

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

	rootFile = rootFlag

	diffToValid := false
	indexValidated := false
	var indexValidatedJSONBArray []byte
	// Working map for index file entries.
	var rootMap = make(map[string]string)

	for _, fileName := range parseFileArgs() {
		logger.Info("Checking", slog.String("file", fileName))
		fileData, err := os.ReadFile(fileName)
		if err != nil {
			logger.Error("Reading file", slog.String("file", fileName), slog.Any("err", err))
			os.Exit(1) // We should be able to read all the files
		}
		// Loop local re-declared variable
		origFileContent := string(fileData)

		if !json.Valid(fileData) {
			logger.Error("Invalid JSON format", slog.String("file", fileName))
			os.Exit(3) // All files should at least parse as JSON.
		}

		// Enables same-dir index default via: filepath.Dir(fileName)
		// Avoid reprocessing index on every Rockon definition validation
		// Optimise: we may already have just read, and JSON Validated our index file.
		if indexValidated == false {
			if rootFlag == "" {
				rootFile = filepath.Join(filepath.Dir(fileName), "root.json")
				logger.Info("Using same-path index", slog.String("file", rootFile))
			} else {
				logger.Info("Using passed index", slog.String("file", rootFile))
			}

			rootData, rootReadErr := os.ReadFile(rootFile)
			// TODO: Warn on no index when using '--check' as this can create an index file from the passed definitions.
			//  Set flag on --check and no index file found to avoid further references.
			if rootReadErr != nil {
				logger.Error("Reading index", slog.String("file", rootFile), slog.Any("rootReadErr", rootReadErr))
				os.Exit(4)
			}
			if !json.Valid(rootData) {
				logger.Error("Invalid JSON format in index", slog.String("file", rootFile))
				os.Exit(5) // All files should at least parse as JSON.
			}

			// Stash Original index file content.
			indexOrigContent = string(rootData)

			rootValidErr := json.Unmarshal(rootData, &rootMap)
			logger.Debug("root.json flags", slog.String("rootFlag", rootFlag), slog.String("rootFile", rootFile))
			if rootValidErr != nil {
				logger.Error("Index validation failed for", slog.String("file", rootFile))
				os.Exit(1)
			}
			indexValidated = true
		}

		// Skip Rockon validation for index file: validated above.
		if filepath.Clean(fileName) == filepath.Clean(rootFile) {
			logger.Warn("Skipped RockOn validation for index", slog.String("file", rootFile))
			continue
		}

		// Validate Rockon file fileData against RockOn model, confirming matching index entry (root.json).
		var rockon model.RockOn
		rockonValidErr := json.Unmarshal(fileData, &rockon)
		if rockonValidErr != nil {
			if filepath.Ext(fileName) == ".json" {
				logger.Error("Unmarshalling json fileData", slog.String("file", fileName), slog.Any("err", rockonValidErr))
				os.Exit(1) // File was named `*.json`, but couldn't be marshalled as expected, so we need to exit.
			}
			logger.Warn("Non *.json filename passed as input, skipping", slog.String("file", fileName))
			continue // File was not named `*.json`, so we shouldn't worry about it.
		}

		// Check and update rootMap from index file against this Rock-on's filename and title.
		checkRootMap(rootMap, filepath.Base(fileName), rockon)

		rockonValidatedJSON, rockonToJsonErr := rockon.ToJSON()
		if rockonToJsonErr != nil {
			logger.Error("Marshaling to JSON", slog.Any("err", rockonToJsonErr))
			os.Exit(1) // This should basically never happen
		}

		if origFileContent != rockonValidatedJSON {
			diffToValid = true
		}

		// Print diff for this Rockon.
		if diffFlag {
			aPath := "a/" + strings.TrimPrefix(fileName, "/")
			bPath := "b/" + strings.TrimPrefix(fileName, "/")
			edits := myers.ComputeEdits(span.URIFromPath(aPath), origFileContent, rockonValidatedJSON)
			fmt.Println(gotextdiff.ToUnified(aPath, bPath, origFileContent, edits))
		}

		// Get existing FileInfo from local variable to reuse in os.WriteFile overwrite.
		fileInfo, _ = os.Stat(fileName)

		if writeFlag { // this rockon
			logger.Debug("Overwriting rockon", slog.String("file", fileName))
			err = os.WriteFile(fileName, []byte(rockonValidatedJSON), fileInfo.Mode())
			if err != nil {
				logger.Error("Overwriting rockon", slog.String("file", fileName), slog.Any("err", err))
				os.Exit(6)
			}

		}
	} // fileName in parseFileArgs()

	// Remaining index file treatment/feedback:

	// Slice.sorted of index file names GO 1.23 onwards
	// https://www.dolthub.com/blog/2024-12-20-collection-functions-in-go-1-23/#sorting-map-elements
	// Strings in GO are read-only slices of bytes.
	// sortedKeys := slices.Sorted(maps.Keys(rootMap))
	// logger.Info("Sorted index", slog.Any("Keys", sortedKeys))

	// From: https://go.dev/src/encoding/json/encode.go
	// "The map keys are sorted and used as JSON object keys ..."
	// Works when arbitrary index file elements (Rockon Titles) are all lower-case.
	indexValidatedJSONBArray, _ = json.MarshalIndent(rootMap, "", "  ")

	if writeFlag { // index file
		rootStat, rootStatErr := os.Stat(rootFile)
		// if no index file for fileInfo, use last Rockon FileInfo
		if os.IsNotExist(rootStatErr) {
			rootStat = fileInfo
		}
		logger.Debug("Overwriting index", slog.String("file", rootFile))
		indexOverwriteErr := os.WriteFile(rootFile, indexValidatedJSONBArray, rootStat.Mode())
		if indexOverwriteErr != nil {
			logger.Error("Overwriting index", slog.String("file", rootFile), slog.Any("err", indexOverwriteErr))
			os.Exit(7)
		}
	}

	// Print diff for the index file.
	if diffFlag {
		aPath := "a/" + strings.TrimPrefix(rootFile, "/")
		bPath := "b/" + strings.TrimPrefix(rootFile, "/")
		edits := myers.ComputeEdits(span.URIFromPath(aPath), indexOrigContent, string(indexValidatedJSONBArray))
		fmt.Println(gotextdiff.ToUnified(aPath, bPath, indexOrigContent, edits))
	}

	// Return 0 when --diff and diff to valid successfully generated.
	if diffToValid {
		if diffFlag {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}
