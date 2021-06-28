package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

// easyjson:json

// User struct contains info about user
type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"-"`
	Country  string   `json:"-"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Job      string   `json:"-"`
	Phone    string   `json:"-"`
}

// Users struct contains users
type Users struct {
	user []User
}

// FastSearch is the func search through the file
func FastSearch(out io.Writer) {
	var seenBrowsers []string

	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	uniqueBrowsers := 0
	foundUsers := ""

	byteLines := bytes.Split(fileContents, []byte("\n"))

	users := Users{}

	for i, line := range byteLines {
		isAndroid := false
		isMsie := false
		d := &User{}
		err = easyjson.Unmarshal(line, d)
		for _, browser := range d.Browsers {
			switch {
			case strings.Contains(browser, "Android"):
				isAndroid = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			case strings.Contains(browser, "MSIE"):
				isMsie = true
				notSeenBefore := true
				for _, item := range seenBrowsers {
					if item == browser {
						notSeenBefore = false
					}
				}
				if notSeenBefore {
					seenBrowsers = append(seenBrowsers, browser)
					uniqueBrowsers++
				}
			default:
				continue
			}
			if isAndroid && isMsie {
				foundUsers += "[" + strconv.Itoa(i) + "] " + d.Name + " <" + strings.ReplaceAll(d.Email, "@", " [at] ") + ">" + "\n"
				break
			}
		}
		if err != nil {
			panic(err)
		}
		users.user = append(users.user, *d)
	}
	fmt.Fprintln(out, "found users:\n"+foundUsers)
	fmt.Fprintln(out, "Total unique browsers", uniqueBrowsers)
}

// Autogenerated functions (See EasyJson from Mail Group)
func easyJsonE0340b5dDecodeHw3BenchJsonPackage(in *jlexer.Lexer, out *User) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "browsers":
			if in.IsNull() {
				in.Skip()
				out.Browsers = nil
			} else {
				in.Delim('[')
				if out.Browsers == nil {
					if !in.IsDelim(']') {
						out.Browsers = make([]string, 0, 4)
					} else {
						out.Browsers = []string{}
					}
				} else {
					out.Browsers = (out.Browsers)[:0]
				}
				for !in.IsDelim(']') {
					var v1 = in.String()
					out.Browsers = append(out.Browsers, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		case "email":
			out.Email = in.String()
		case "name":
			out.Name = in.String()
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyJsonE0340b5dEncodeHw3BenchJsonPackage(out *jwriter.Writer, in User) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"browsers\":"
		out.RawString(prefix[1:])
		if in.Browsers == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Browsers {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(v3)
			}
			out.RawByte(']')
		}
	}
	{
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		out.String(in.Email)
	}
	{
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		out.String(in.Name)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v User) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyJsonE0340b5dEncodeHw3BenchJsonPackage(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v User) MarshalEasyJSON(w *jwriter.Writer) {
	easyJsonE0340b5dEncodeHw3BenchJsonPackage(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *User) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyJsonE0340b5dDecodeHw3BenchJsonPackage(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *User) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyJsonE0340b5dDecodeHw3BenchJsonPackage(l, v)
}
