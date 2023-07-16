package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	for _, f := range os.Args[1:] {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var rockon Rockon
		err = json.Unmarshal(data, &rockon)
		if err != nil {
			fmt.Println(f, err)
			os.Exit(1)
		}

		var out strings.Builder
		enc := json.NewEncoder(&out)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")

		err = enc.Encode(rockon)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(out.String())
	}
}
