package main

/*
Finished json should have:
(List of dates)
Hour number, hour start + hour end
List of Hour lessons

Example:
days map:
Days: { 1: {
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
	"errors"
	"mime/multipart"
	"net/http"
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

type Class struct {
	ClassName string `json:"className"`
	ClassNum  int    `json:"classNum"`
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

var classCodeSlice []Class = []Class{}

func AddFormFields(writer *multipart.Writer, classNum int) {
	formField, err := writer.CreateFormField("__EVENTTARGET")
	if err != nil {
		panic(err)
	}
	_, err = formField.Write([]byte("dnn$ctr30329$TimeTableView$btnTimeTable"))

	formField, err = writer.CreateFormField("__VIEWSTATE") //REQUIRED
	if err != nil {
		panic(err)
	}
	_, err = formField.Write([]byte("/wEPDwUIMjU3MTQzOTcPZBYGZg8WAh4EVGV4dAU+PCFET0NUWVBFIEhUTUwgUFVCTElDICItLy9XM0MvL0RURCBIVE1MIDQuMCBUcmFuc2l0aW9uYWwvL0VOIj5kAgEPZBYMAgEPFgIeB1Zpc2libGVoZAICDxYCHgdjb250ZW50BRjXqNeV16TXmdefINei157XpyDXl9ek16hkAgMPFgIfAgUn16jXldek15nXnyDXotee16cg15fXpNeoLERvdE5ldE51a2UsRE5OZAIEDxYCHwIFINeb15wg15TXlteb15XXmdeV16og16nXnteV16jXldeqZAIFDxYCHwIFC0RvdE5ldE51a2UgZAIGDxYCHwIFGNeo15XXpNeZ158g16LXntenINeX16TXqGQCAg9kFgJmD2QWAgIED2QWAmYPZBYGAgIPZBYCZg8PFgYeCENzc0NsYXNzBQtza2luY29sdHJvbB4EXyFTQgICHwFoZGQCAw9kFgJmDw8WBh8DBQtza2luY29sdHJvbB8ABQVMb2dpbh8EAgJkZAIGD2QWAgICD2QWCAIBDw8WAh8BaGRkAgMPDxYCHwFoZGQCBQ9kFgICAg8WAh8BaGQCBw9kFgICAQ9kFgICAQ9kFggCBg9kFgJmD2QWDAICDxYCHgVjbGFzcwUKSGVhZGVyQ2VsbGQCBA8WAh8FBQpIZWFkZXJDZWxsZAIGDxYCHwUFCkhlYWRlckNlbGxkAggPFgIfBQUKSGVhZGVyQ2VsbGQCCg8WAh8FBQpIZWFkZXJDZWxsZAIMDxYCHwUFEEhlYWRlckNlbGxCdXR0b25kAgcPEGQQFQAVABQrAwBkZAIMD2QWAmYPZBYaZg9kFgICAQ8QZBAVMQPXljED15YyA9eWMwPXljQD15Y2A9eWNwPXljgD15Y5A9eXMQPXlzID15czA9eXNAPXlzYD15c3A9eXOAPXmDED15gyA9eYMwPXmDQD15g2A9eYNwPXmDgD15g5A9eZMQPXmTID15kzA9eZNAPXmTYD15k3A9eZOAPXmTkF15nXkDEF15nXkDIF15nXkDMF15nXkDQF15nXkDYF15nXkDcF15nXkDgF15nXkDkG15nXkDEwBdeZ15ExBdeZ15EyBdeZ15EzBdeZ15E0BdeZ15E1BdeZ15E2BdeZ15E3BdeZ15E4BdeZ15E5FTEBOQIxMQIxMgIxMwIxNQMxMjgDMjIyAzIyOQIxOAIyMgMxNTECMjADMTAzAzEwNAMxMzQCMjcCMjgCMjkCMzACMzIDMTA4AzE4NAMyMjcCMzgCNDACNDECNDMDMTI2AjQ0AzE2MwMxNjQCNDcCNDkCNTACNTEDMTUzAzEzNwMxNjkDMTcwAzIyOAMyMjECNjMCNjQCNjUCNjYDMTI1AzE1NQMxODEDMTgyFCsDMWdnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2dnZ2cWAQIUZAICDxYEHwUFCkhlYWRlckNlbGwfAWhkAgMPFgIfAWhkAgQPFgIfBQUKSGVhZGVyQ2VsbGQCBg8WAh8FBRJIZWFkZXJDZWxsU2VsZWN0ZWRkAggPFgIfBQUKSGVhZGVyQ2VsbGQCCg8WAh8FBQpIZWFkZXJDZWxsZAIMDxYCHwUFCkhlYWRlckNlbGxkAg4PFgIfBQUKSGVhZGVyQ2VsbGQCEA8WAh8FBQpIZWFkZXJDZWxsZAISDxYEHwUFCkhlYWRlckNlbGwfAWhkAhMPFgIfAWhkAhQPFgIfBQUQSGVhZGVyQ2VsbEJ1dHRvbmQCDw8PFgIfAAU7157XoteV15PXm9efINecOiAxOC4xMS4yMDI0LCDXqdei15Q6IDIwOjQ0LCDXnteh15o6IEExMzAzMjlkZGR1CcRP+gmN0gm8+oYknIC/sTy65Q=="))

	formField, err = writer.CreateFormField("dnn$ctr30329$TimeTableView$ClassesList")
	if err != nil {
		panic(err)
	}

	_, err = formField.Write([]byte(strconv.Itoa(classNum)))

	formField, err = writer.CreateFormField("dnn$ctr30329$TimeTableView$MainControl$WeekShift")
	if err != nil {
		panic(err)
	}
	_, err = formField.Write([]byte("0"))

	formField, err = writer.CreateFormField("dnn$ctr30329$TimeTableView$ControlId")
	if err != nil {
		panic(err)
	}
	_, err = formField.Write([]byte("8"))

	writer.Close()

}

func createRequest(classNum int, shahafURL string) *http.Response {
	form := new(bytes.Buffer)
	writer := multipart.NewWriter(form)

	AddFormFields(writer, classNum)

	client := &http.Client{} // Create a client to send the http requests
	req, err := http.NewRequest("POST", shahafURL, form)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}

var curDay int

func processTD(n *html.Node) {
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

func processSelect(n *html.Node) {
	for optionChild := n.FirstChild; optionChild != nil; optionChild = optionChild.NextSibling {
		if optionChild.FirstChild != nil {
			var classNum int
			for _, attr := range optionChild.Attr {
				if attr.Key == "value" {
					classNum, _ = strconv.Atoi(attr.Val)
					break // Only one attribute
				}
			}
			className := optionChild.FirstChild.Data

			classCodeSlice = append(classCodeSlice, Class{className, classNum})
		}
	}
}

func processNode(n *html.Node) {
	if n.Data == "td" {
		processTD(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNode(c)
	}
}

func computeResponse(resp *http.Response) {
	var processAllProduct func(*html.Node)
	processAllProduct = func(n *html.Node) {
		for _, attr := range n.Attr {
			if attr.Key == "class" && attr.Val == "TTTable" {
				processNode(n)
				return //Only need the first tbody tag
			}
		}
		if n.Data == "select" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "HeaderClasses" {
					processSelect(n)
					break
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			processAllProduct(c)
		}
	}

	// Make a recursive call to the function
	htmlParsedPage, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}
	processAllProduct(htmlParsedPage)
}

func addJSONs(jsonNameList []string, jsonSlice ...string) string {
	if len(jsonNameList) != len(jsonSlice) {
		panic(errors.New("Slices have different lengths"))
	}
	finalStr := ""
	for i, jsonName := range jsonNameList {
		stringToAdd := "\"" + jsonName + "\"" + ": " + jsonSlice[i] + ","
		if i == len(jsonNameList)-1 {
			stringToAdd = stringToAdd[:len(stringToAdd)-1]
		}
		finalStr += stringToAdd
	}
	return "{" + finalStr + "}"
}

func getClassJSON(shahafURL string) string {
	resp := createRequest(-1, shahafURL)

	defer resp.Body.Close()

	computeResponse(resp)

	jsonOut, err := json.Marshal(classCodeSlice)
	if err != nil {
		panic(err)
	}

	return string(jsonOut)
}

func getTimeTableJSON(classNum int, shahafURL string) string {
	resp := createRequest(classNum, shahafURL)

	defer resp.Body.Close()

	computeResponse(resp)

	jsonOut, err := json.Marshal(lessonMap)
	if err != nil {
		panic(err)
	}
	lessonMapJson := string(jsonOut)

	jsonOut, err = json.Marshal(dateMap)
	if err != nil {
		panic(err)
	}

	dateMapJson := string(jsonOut)

	jsonOut, err = json.Marshal(hourMap)
	if err != nil {
		panic(err)
	}

	hourMapJson := string(jsonOut)

	jsonNameList := []string{"Lessons", "Dates", "Hours"}

	finalJson := addJSONs(jsonNameList, lessonMapJson, dateMapJson, hourMapJson)

	return finalJson
}
