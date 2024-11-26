package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
)

func main() {
	var classNum int
	var listClasses bool
	var shahafURL string
	var outputToFile string
	flag.BoolVar(&listClasses, "list", false, "Lists all classes in JSON form")
	flag.IntVar(&classNum, "class", 0, "Class Name")
	flag.StringVar(&shahafURL, "url", "", "Shahaf Time Table Website URL (REQUIRED)")
	flag.StringVar(&outputToFile, "output", "", "Output the JSON to your specified file. By default: Prints the JSON in stdout")
	flag.Parse()

	_, err := url.ParseRequestURI(shahafURL)
	if err != nil {
		flag.PrintDefaults()
		panic(err)
	}

	var jsonOut string
	if listClasses == true {
		jsonOut = getClassJSON(shahafURL)
	} else {
		jsonOut = getTimeTableJSON(classNum, shahafURL)
	}
	if outputToFile == "" {
		fmt.Printf(jsonOut)
	} else {
		err = os.WriteFile(outputToFile, []byte(jsonOut), 0644)
		if err != nil {
			panic(err)
		}
	}
}