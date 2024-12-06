// If you want to use the api in any other language supported by c, build this main file
package main

import "C"
import (
	"net/url"
)

//export mainC
func mainC(shahafURLPtr *C.char, listClassesC C.int, classNumC C.int) *C.char {
	shahafURL := C.GoString(shahafURLPtr)
	listClasses := int(listClassesC)
	classNum := int(classNumC)
	_, err := url.ParseRequestURI(shahafURL)
	if err != nil {
		panic(err)
	}

	var jsonOut string
	if listClasses == 1 {
		jsonOut = getClassJSON(shahafURL)
	} else {
		jsonOut = getTimeTableJSON(classNum, shahafURL)
	}
	return C.CString(jsonOut)
}

func main() {

}

// // Deprecated
// func main() {
// 	var classNum int
// 	var listClasses bool
// 	var shahafURL string
// 	var outputToFile string
// 	flag.BoolVar(&listClasses, "list", false, "Lists all classes in JSON form")
// 	flag.IntVar(&classNum, "class", 0, "Class Name")
// 	flag.StringVar(&shahafURL, "url", "", "Shahaf Time Table Website URL (REQUIRED)")
// 	flag.StringVar(&outputToFile, "output", "", "Output the JSON to your specified file. By default: Prints the JSON in stdout")
// 	flag.Parse()

// 	_, err := url.ParseRequestURI(shahafURL)
// 	if err != nil {
// 		flag.PrintDefaults()
// 		panic(err)
// 	}

// 	var jsonOut string
// 	if listClasses == true {
// 		jsonOut = getClassJSON(shahafURL)
// 	} else {
// 		jsonOut = getTimeTableJSON(classNum, shahafURL)
// 	}
// 	if outputToFile == "" {
// 		println(jsonOut)
// 	} else {
// 		err = os.WriteFile(outputToFile, []byte(jsonOut), 0644)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}
// }
