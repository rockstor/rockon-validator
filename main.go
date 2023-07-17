package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	for _, f := range os.Args[1:] {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var rockon RockOn
		err = json.Unmarshal(data, &rockon)
		if err != nil {
			// It may be the root json, so skip it
			root := map[string]string{}
			err1 := json.Unmarshal(data, &root)
			if err1 == nil {
				continue
			}
			fmt.Println(f, err)
			os.Exit(1)
		}

		fmt.Println(rockon.ToJSON())
	}
}
