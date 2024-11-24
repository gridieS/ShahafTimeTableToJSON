package main

/*
Finished json should have:
(List of dates)
Hour number, hour start + hour end
List of Hour lessons

Example:
days map:
Days: {
	1: {
		Day: 18,
		Month: 11
	},
	2: {
		Day: 19,
		Month: 11
	}
	...
}
Lessons: {
	1: {
		1: [
			{"Lesson": "Math","Teacher": "Svetlana","Location":"24-0-9"},
			{"Lesson": "Math","Teacher": "Berlin","Location":"24-0-10"},
			{"Lesson": "Math","Teacher": "Neor","Location":"24-0-11"}
		],
		2: [
			...
		]
		...
	}
	...
}
*/

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type Date struct {
	Day   int `json:"day"`
	Month int `json:"month"`
}

type HourEvent struct {
	HourStart   int `json:"hourStart"`
	MinuteStart int `json:"minuteStart"`
	HourEnd     int `json:"hourEnd"`
	MinuteEnd   int `json:"minuteEnd"`
}

type Lesson struct {
	Hour       int    `json:"hour"`
	LessonName string `json:"lessonName"`
	Teacher    string `json:"teacher"`
	Location   string `json:"location"`
}

const MIN_TEACHER_NAME_LENGTH = 5
const NUM_OF_LESSONS int = 14
const NUM_OF_DAYS int = 6

const DATE_STRING_LENGTH = 5
const DATE_STRING_DAY_END = 2
const DATE_STRING_MONTH_START = 3

const EVENT_STRING_HOUR_END = 2
const EVENT_STRING_MINUTE_START = 3

var dateMap map[int]Date = make(map[int]Date, NUM_OF_DAYS)
var hourMap map[int]HourEvent = make(map[int]HourEvent, NUM_OF_LESSONS)

var lessonMap map[int](map[int]([]Lesson)) = make(map[int](map[int]([]Lesson)), NUM_OF_DAYS)

var CLASS_TO_CODE map[string]string = map[string]string{ // const
	"7-1":   "9",
	"7-2":   "11",
	"7-3":   "12",
	"7-4":   "13",
	"7-6":   "15",
	"7-7":   "128",
	"7-8":   "222",
	"7-9":   "229",
	"8-1":   "18",
	"8-2":   "22",
	"8-3":   "151",
	"8-4":   "20",
	"8-6":   "103",
	"8-7":   "104",
	"8-8":   "134",
	"9-1":   "27",
	"9-2":   "28",
	"9-3":   "29",
	"9-4":   "30",
	"9-6":   "32",
	"9-7":   "108",
	"9-8":   "184",
	"9-9":   "227",
	"10-1":  "38",
	"10-2":  "40",
	"10-3":  "41",
	"10-4":  "43",
	"10-6":  "126",
	"10-7":  "44",
	"10-8":  "163",
	"10-9":  "164",
	"11-1":  "47",
	"11-2":  "49",
	"11-3":  "50",
	"11-4":  "51",
	"11-6":  "153",
	"11-7":  "137",
	"11-8":  "169",
	"11-9":  "170",
	"11-10": "228",
	"12-1":  "221",
	"12-2":  "63",
	"12-3":  "64",
	"12-4":  "65",
	"12-5":  "66",
	"12-6":  "125",
	"12-7":  "155",
	"12-8":  "181",
	"12-9":  "182",
}

func AddFormFields(writer *multipart.Writer, classCode string) {
	formField, err := writer.CreateFormField("__EVENTTARGET")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte("dnn$ctr30329$TimeTableView$btnTimeTable"))

	formField, err = writer.CreateFormField("__VIEWSTATE") //REQUIRED
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte("/wEPDwUIMjU3MTQzOTcPZBYGZg8WAh4EVGV4dAU+PCFET0NUWVBFIEhUTUwgUFVCTElDICItLy9XM0MvL0RURCBIVE1MIDQuMCBUcmFuc2l0aW9uYWwvL0VOIj5kAgEPZBYMAgEPFgIeB1Zpc2libGVoZAICDxYCHgdjb250ZW50BRjXqNeV16TXmdefINei157XpyDXl9ek16hkAgMPFgIfAgUn16jXldek15nXnyDXotee16cg15fXpNeoLERvdE5ldE51a2UsRE5OZAIEDxYCHwIFINeb15wg15TXlteb15XXmdeV16og16nXnteV16jXldeqZAIFDxYCHwIFC0RvdE5ldE51a2UgZAIGDxYCHwIFGNeo15XXpNeZ158g16LXntenINeX16TXqGQCAg9kFgJmD2QWAgIED2QWAmYPZBYGAgIPZBYCZg8PFgYeCENzc0NsYXNzBQtza2luY29sdHJvbB4EXyFTQgICHwFoZGQCAw9kFgJmDw8WBh8DBQtza2luY29sdHJvbB8ABQVMb2dpbh8EAgJkZAIGD2QWAgICD2QWCAIBDw8WAh8BaGRkAgMPDxYCHwFoZGQCBQ9kFgICAg8WAh8BaGQCBw9kFgICAQ9kFgICAQ9kFggCBg9kFgJmD2QWDAICDxYCHgVjbGFzcwUKSGVhZGVyQ2VsbGQCBA8WAh8FBQpIZWFkZXJDZWxsZAIGDxYCHwUFCkhlYWRlckNlbGxkAggPFgIfBQUKSGVhZGVyQ2VsbGQCCg8WAh8FBQpIZWFkZXJDZWxsZAIMDxYCHwUFEEhlYWRlckNlbGxCdXR0b25kAgcPEGQQFQAVABQrAwBkZAIMD2QWAmYPZBYaZg9kFgICAQ8QZBAVMQPXljED15YyA9eWMwPXljQD15Y2A9eWNwPXljgD15Y5A9eXMQPXlzID15czA9eXNAPXlzYD15c3A9eXOAPXmDED15gyA9eYMwPXmDQD15g2A9eYNwPXmDgD15g5A9eZMQPXmTID15kzA9eZNAPXmTYD15k3A9eZOAPXmTkF15nXkDEF15nXkDIF15nXkDMF15nXkDQF15nXkDYF15nXkDcF15nXkDgF15nXkDkG15nXkDEwBdeZ15ExBdeZ15EyBdeZ15EzBdeZ15E0BdeZ15E1BdeZ15E2BdeZ15E3BdeZ15E4BdeZ15E5FTEBOQIxMQIxMgIxMwIxNQMxMjgDMjIyAzIyOQIxOAIyMgMxNTECMjADMTAzAzEwNAMxMzQCMjcCMjgCMjkCMzACMzIDMTA4AzE4NAMyMjcCMzgCNDACNDECNDMDMTI2AjQ0AzE2MwMxNjQCNDcCNDkCNTACNTEDMTUzAzEzNwMxNjkDMTcwAzIyOAMyMjECNjMCNjQCNjUCNjYDMTI1AzE1NQMxODEDMTgyFCsDMWdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cWAQIUZAICDxYEHwUFCkhlYWRlckNlbGwfAWhkAgMPFgIfAWhkAgQPFgIfBQUKSGVhZGVyQ2VsbGQCBg8WAh8FBRJIZWFkZXJDZWxsU2VsZWN0ZWRkAggPFgIfBQUKSGVhZGVyQ2VsbGQCCg8WAh8FBQpIZWFkZXJDZWxsZAIMDxYCHwUFCkhlYWRlckNlbGxkAg4PFgIfBQUKSGVhZGVyQ2VsbGQCEA8WAh8FBQpIZWFkZXJDZWxsZAISDxYEHwUFCkhlYWRlckNlbGwfAWhkAhMPFgIfAWhkAhQPFgIfBQUQSGVhZGVyQ2VsbEJ1dHRvbmQCDw8PFgIfAAU7157XoteV15PXm9efINecOiAxOC4xMS4yMDI0LCDXqdei15Q6IDIwOjQ0LCDXnteh15o6IEExMzAzMjlkZGR1CcRP+gmN0gm8+oYknIC/sTy65Q=="))

	formField, err = writer.CreateFormField("dnn$ctr30329$TimeTableView$ClassesList")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte(classCode))

	formField, err = writer.CreateFormField("dnn$ctr30329$TimeTableView$MainControl$WeekShift")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte("0"))

	formField, err = writer.CreateFormField("dnn$ctr30329$TimeTableView$ControlId")
	if err != nil {
		log.Fatal(err)
	}
	_, err = formField.Write([]byte("8"))

	writer.Close()

}

func createRequest(className string, shahafURL string) *http.Response {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)

	AddFormFields(writer, CLASS_TO_CODE[className])

	client := &http.Client{} // Create a client to send the http requests
	req, err := http.NewRequest("POST", shahafURL, form)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	return resp
}

var curDay int

func processNode(n *html.Node) {
	switch n.Data {
	case "tr":
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {

		}
	case "td":
		for _, attr := range n.Attr {
			if attr.Key == "class" {
				if strings.Contains(attr.Val, "CTitle") { // Signals date
					dateInString := n.FirstChild.Data[len(n.FirstChild.Data)-DATE_STRING_LENGTH : len(n.FirstChild.Data)]
					dateDay, _ := strconv.Atoi(dateInString[:DATE_STRING_DAY_END])
					dateMonth, _ := strconv.Atoi(dateInString[DATE_STRING_MONTH_START:])
					dateMap[len(dateMap)+1] = Date{dateDay, dateMonth}
				} else if strings.Contains(attr.Val, "CName") { // Signals Hour element
					var curHour int
					var hourToHour []string = make([]string, 2)
					if n.FirstChild == nil {
						continue
					}
					boldTag := n.FirstChild
					for timeChild := boldTag.FirstChild; timeChild != nil; timeChild = timeChild.NextSibling {
						if timeChild.Type == html.TextNode {
							convertedString, _ := strconv.Atoi(timeChild.Data)
							curHour = convertedString
						} else if timeChild.Data == "span" {
							if hourToHour[0] == "" {
								hourToHour[0] = timeChild.FirstChild.Data
							} else {
								hourToHour[1] = timeChild.FirstChild.Data
							}
						}
					}
					if len(hourToHour[0]) > 1 {
						startHour, _ := strconv.Atoi(hourToHour[0][:EVENT_STRING_HOUR_END])
						startMinutes, _ := strconv.Atoi(hourToHour[0][EVENT_STRING_MINUTE_START:])
						endHour, _ := strconv.Atoi(hourToHour[1][:EVENT_STRING_HOUR_END])
						endMinutes, _ := strconv.Atoi(hourToHour[1][EVENT_STRING_MINUTE_START:])
						hourMap[curHour] = HourEvent{startHour, startMinutes, endHour, endMinutes}
					}
				} else if strings.Contains(attr.Val, "TTCell") { // Signals hour's subject
					curHour := len(hourMap) - 1
					curDay = curDay%NUM_OF_DAYS + 1
					for hourChild := n.FirstChild; hourChild != nil; hourChild = hourChild.NextSibling {
						if hourChild.Data != "div" {
							continue
						}
						var location string
						var lessonName string
						var teacher string
						for lessonChild := hourChild.FirstChild; lessonChild != nil; lessonChild = lessonChild.NextSibling {
							if lessonChild.Data == "b" {
								lessonName = lessonChild.FirstChild.Data
							} else if lessonChild.Type == html.TextNode {
								if strings.Contains(lessonChild.Data, "(") { // location has to have paranthesis
									location = lessonChild.Data
								} else {
									if len(lessonChild.Data) >= MIN_TEACHER_NAME_LENGTH {
										teacher = lessonChild.Data
									}
								}
							}
						}
						curLesson := Lesson{curHour, lessonName, teacher, location}
						if lessonMap[curDay] == nil {
							lessonMap[curDay] = make(map[int][]Lesson)
						}
						lessonMap[curDay][curHour] = append(lessonMap[curDay][curHour], curLesson)
					}
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(c)
	}
}

func main() {
	var className string
	var shahafURL string
	var printToStdout bool
	flag.StringVar(&className, "class", "7-1", "Class Name")
	flag.StringVar(&shahafURL, "url", "", "Shahaf Time Table Website URL (REQUIRED)")
	flag.BoolVar(&printToStdout, "stdout", false, "Prints the json to stdout. By default: creates a new file 'output.json'")
	flag.Parse()

	_, err := url.ParseRequestURI(shahafURL)
	if err != nil {
		flag.PrintDefaults()
		panic(err)
	}

	if CLASS_TO_CODE[className] == "" {
		panic("Invalid class name")
	}

	resp := createRequest(className, shahafURL)

	defer resp.Body.Close()

	var processAllProduct func(*html.Node)
	processAllProduct = func(n *html.Node) {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "TTTable" {
				processNode(n)
				return //Only need the first tbody tag
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processAllProduct(c)
		}
	}

	// Make a recursive call to the function
	htmlParsedPage, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	processAllProduct(htmlParsedPage)

	lessonMapJson, err := json.Marshal(lessonMap)
	if err != nil {
		panic(err)
	}
	lessonMapJson = []byte("\"Lessons\":" + string(lessonMapJson) + ",")

	dateMapJson, err := json.Marshal(dateMap)
	if err != nil {
		panic(err)
	}

	dateMapJson = []byte("\"Dates\":" + string(dateMapJson) + ",")

	hourMapJson, err := json.Marshal(hourMap)
	if err != nil {
		panic(err)
	}

	hourMapJson = []byte("\"Hours\":" + string(hourMapJson))

	finalJson := []byte("{" + string(lessonMapJson) + string(dateMapJson) + string(hourMapJson) + "}")
	if printToStdout {
		fmt.Printf(string(finalJson))
	} else {
		err = os.WriteFile("output.json", finalJson, 0644)
		if err != nil {
			panic(err)
		}
	}
}
