package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	// "log"
)

const filePath string = "./data/users.txt"

func SlowSearch(out io.Writer) {
	file, err := os.Open(filePath)					//1.06kb
	if err != nil {
		panic(err)
	}

	fileContents, err := ioutil.ReadAll(file)		//15,98Mb
	if err != nil {
		panic(err)
	}


	r := regexp.MustCompile("@")				//5.96kb
	seenBrowsers := []string{}
	uniqueBrowsers := 0
	foundUsers := ""

	users := make([]map[string]interface{}, 0)
	byteLines := bytes.Split(fileContents, []byte("\n"))
	for _, byteLine := range byteLines{
		user := make(map[string]interface{})

		err := json.Unmarshal(byteLine, &user)

		if err != nil {
			panic(err)
		}
		users = append(users, user)

	}

	for i, user := range users {

		isAndroid := false
		isMSIE := false

		browsers, ok := user["browsers"].([]interface{})
		if !ok {
			continue
		}

		for _, browserRaw := range browsers {
			browser, ok := browserRaw.(string)
			if !ok {
				continue
			}
			if strings.Contains( browser, "Android") {
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)									//20.38kb 20.38kb
					uniqueBrowsers++
				}
			} else if strings.Contains(browser, "MSIE" ){
				isMSIE = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)					// 10ms
					uniqueBrowsers++
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := r.ReplaceAllString(user["email"].(string), " [at] ")				//72.20kb
		foundUsers += fmt.Sprintf("[%d] %s <%s>\n", i, user["name"], email)				//1.45mb 1.58mb
	}

	fmt.Fprintln(out, "found users:\n"+foundUsers)										//38.12kb 85.62kb
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))					//80b 18.58kb
}
