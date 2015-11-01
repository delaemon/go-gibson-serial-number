package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	AccessTime time.Time = time.Now()
)

type RequestParams struct {
	SerialNumber string
}

type ResponseData struct {
	Year          string
	Month         string
	Day           string
	Factory       string
	Model         string
	Shape         string
	RankingNumber string
	SerialNumber  string
}

func handler(w http.ResponseWriter, req *http.Request) {
	logFile := fmt.Sprintf("./log/app/%d%d%d.log", AccessTime.Year(), AccessTime.Month(), AccessTime.Day())
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	log.SetOutput(f)

	var serialNumber string

	if req.Method == "POST" {
		decorder := json.NewDecoder(req.Body)
		var rp RequestParams
		err := decorder.Decode(&rp)
		if err != nil {
			log.Println(err)
		}
		serialNumber = rp.SerialNumber
	} else {
		req.ParseForm()
		serialNumber = req.Form.Get("serialNumber")
	}

	if !validSerialNumber(serialNumber) {
		fmt.Fprintf(w, "invalid serial number.")
		log.Printf("invalid serial number => %s\n", serialNumber)
		return
	}

	var out string
	if isCentennialYear(serialNumber) {
		// 1994 Gibson's Centennial year
		// 94RRRRRR
		p := convertRankingNumberToText(string(serialNumber[2:7]))
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Date: 1994 (In 1994, Gibson's Centennial year) \n"+
			"Instrument Rank: %d",
			serialNumber, p)
	} else if isCustomShopRegular(serialNumber) {
		// CSYRRRR
		p := convertRankingNumberToText(string(serialNumber[3:]))
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Year: 200%s\n"+
			"%s built that year.\n"+
			"CUSTOM SHOP Regular production.",
			serialNumber, string(serialNumber[2]), p)
	} else if isCustomShopReissues50s(serialNumber) {
		// 1952-1960 Les Paul, Explorer, Flying V, and Futura reissues (since late 1992):
		// M YRRR or MYRRRR
		// M is the model year being reissued
		y := serialNumber[1]
		k := 2
		if string(serialNumber[1]) == " " {
			k = 3
			y = serialNumber[2]
		}
		year := fmt.Sprintf("200%s", y)
		p := convertRankingNumberToText(string(serialNumber[k:]))
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Year: %s\n"+
			"%s built that year.\n"+
			"195%s reissue model\n"+
			"CUSTOM SHOP Reissues 50's.\n"+
			"Les Paul, Explorer, Flying V, and Futura reissues.",
			serialNumber, year, serialNumber[0], p)
	} else if isCustomShopReissues60s(serialNumber) {
		// 1961-1969 Firebird, Les Paul, and SG reissues (since 1997):
		// YYRRRM
		yy := string(serialNumber[0])
		y := string(serialNumber[1])
		yyyy := "19"
		if yy == "0" {
			yyyy = "20"
		}
		year := fmt.Sprintf("%s%s%s", yyyy, yy, y)
		p := convertRankingNumberToText(string(serialNumber[2:5]))
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Year: %s\n"+
			"%s built that year.\n"+
			"196%s reissue model\n"+
			"CUSTOM SHOP Reissues 60's.\n"+
			"Firebird, Les Paul, and SG reissues.",
			serialNumber, year, serialNumber[len(serialNumber)-1], p)
	} else if isCustomShopCarvedTop(serialNumber) {
		// YDDDYRRR
		yy := string(serialNumber[0])
		y := string(serialNumber[4])
		yyyy := "19"
		if yy == "0" {
			yyyy = "20"
		}
		year := fmt.Sprintf("%s%s%s", yyyy, yy, y)

		yi, err := strconv.Atoi(year)
		if err != nil {
			log.Println(err)
		}
		date := time.Date(yi, time.Month(1), 1, 0, 0, 0, 0, time.Local)
		ddd := fmt.Sprintf("%s%s%s",
			string(serialNumber[1]),
			string(serialNumber[2]),
			string(serialNumber[3]))
		di, err := strconv.Atoi(ddd)
		if err != nil {
			log.Println(err)
		}
		date = date.AddDate(0, 0, di-1)

		r := convertRankingNumberToText(string(serialNumber[5:8]))
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Date: %d.%d.%d\n"+
			"The %s instrument stamped that day.\n"+
			"CUSTOM SHOP Carved top %s\n",
			serialNumber, date.Year(), date.Month(), date.Day(), r)
	} else if isEsSeries(serialNumber) {
		/*
			ES (Electric Spanish)
			(A or B)-MYRRR
			M is the model year being reissued
			Y is the production year
			RRR indicates the guitar's place in the sequence of Historic ES production for that year.
			Reissue model codes:
			2= ES-295
			3= 1963 ES-335 (block inlays)
			4= ES-330
			5= ES-345
			9 with an "A" prefix = 1959 ES-335 (dot inlays)
			9 with a "B" prefix= ES-355
		*/
		year := fmt.Sprintf("200%", string(serialNumber[3]))
		r := convertRankingNumberToText(string(serialNumber[4:]))
		var model string
		m := string(serialNumber[2])
		if m == "2" {
			model = "ES-295"
		} else if m == "3" {
			model = "1963 ES-335 (block inlays)"
		} else if m == "4" {
			model = "ES-330"
		} else if m == "5" {
			model = "ES-345"
		} else if m == "9" {
			h := string(serialNumber[0])
			if h == "A" {
				model = "1959 ES-335 (dot inlays)"
			} else if h == "B" {
				model = "ES-355"
			}
		}
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Year: %s\n"+
			"%s built that year.\n"+
			"CUSTOM SHOP %s Reissues.\n",
			serialNumber, year, r, model)

	} else if isLesPaulClassic(serialNumber) {
		l := len(serialNumber)
		var y string
		if string(serialNumber[1]) == " " {
			y = string(serialNumber[0])
		} else {
			y = string(serialNumber[0:2])
		}

		var year string
		if l == 4 && y == "9" {
			year = "1989"
		} else if l == 5 && len(y) == 1 {
			year = "199" + y
		} else if l == 6 && len(y) == 2 {
			year = "20" + y
		}

		r := convertRankingNumberToText(string(serialNumber[2:]))
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Year: %s\n"+
			"%s built that year.\n"+
			"CUSTOM SHOP %s Reissues.\n",
			serialNumber, year, r)
	} else if isReguler(serialNumber) {
		yy := string(serialNumber[0])
		y := string(serialNumber[4])
		yyyy := "19"
		if yy == "0" {
			yyyy = "20"
		}
		year := fmt.Sprintf("%s%s%s", yyyy, yy, y)

		yi, err := strconv.Atoi(year)
		if err != nil {
			log.Println(err)
		}
		date := time.Date(yi, time.Month(1), 1, 0, 0, 0, 0, time.Local)
		ddd := fmt.Sprintf("%s%s%s",
			string(serialNumber[1]),
			string(serialNumber[2]),
			string(serialNumber[3]))
		di, err := strconv.Atoi(ddd)
		if err != nil {
			log.Println(err)
		}
		date = date.AddDate(0, 0, di-1)

		var ppp string
		if yi <= 2005 && date.Month() == 7 && len(serialNumber) == 9 {
			ppp = fmt.Sprintf("%s%s%s",
				string(serialNumber[6]),
				string(serialNumber[7]),
				string(serialNumber[8]))
			batchNumber := string(serialNumber[5])
			fmt.Fprintf(w, "batch %d, ", batchNumber)
		} else {
			ppp = fmt.Sprintf("%s%s%s",
				string(serialNumber[5]),
				string(serialNumber[6]),
				string(serialNumber[7]))
		}

		p := convertRankingNumberToText(ppp)
		pi, err := strconv.Atoi(ppp)
		if err != nil {
			log.Println(err)
		}

		var shapes string
		if pi >= 700 {
			shapes = "Flying V, T-Bird, Explorer, etc."
		} else if pi >= 300 {
			shapes = "Les Paul Style."
		} else {
			shapes = "Unknown."
		}

		// factory
		// supported electoric only.
		// accoustic@Bozeman(Montana) Since 1989
		// Mempis
		// 	2000 ES Series
		// 	2005 custom shop
		// 	2013 original
		var factory string
		if pi < 500 && yi <= 1984 {
			factory = "Kalamazoo"
		} else if 2000 <= yi &&
			string(serialNumber[0]) == "A" || string(serialNumber[0]) == "B" {
			factory = "Memphis"
		} else if yi < 2000 {
			factory = "Nashville or Memphis"
		} else {
			factory = "Nashville"
		}

		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Date: %d.%d.%d\n"+
			"Factory: %s\n"+
			"The %s instrument stamped that day.\n"+
			"Shapes: %s\n",
			serialNumber, date.Year(), date.Month(), date.Day(), factory, p, shapes)
	} else {
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(out))

	log.Printf(out)

	fmt.Printf("[AccessLog] %d-%02d-%02d %02d:%02d:%02d\n",
		AccessTime.Year(), AccessTime.Month(), AccessTime.Day(), AccessTime.Hour(), AccessTime.Minute(), AccessTime.Second())
	fmt.Println(out)
}

func convertRankingNumberToText(rrr string) string {
	ri, err := strconv.Atoi(rrr)
	if err != nil {
		log.Println(err)
	}
	var r string
	if ri == 1 {
		r = "1st"
	} else if ri == 2 {
		r = "2nd"
	} else if ri == 3 {
		r = "3rd"
	} else {
		r = fmt.Sprintf("%dth", ri)
	}
	return r
}

func isReguler(serialNumber string) bool {
	// 1977 ~ 2005.07
	// YDDDYPPP
	//   YY is the production year
	//   DDD is the day of the year
	//   PPP is the plant designation and/or instrument rank
	// 2005.07 ~ now
	// YDDDYBPPP
	//   B is the batch number
	var re *regexp.Regexp
	re = regexp.MustCompile("^\\d{8,9}$")
	if !re.MatchString(serialNumber) {
		return false
	}

	yy := string(serialNumber[0])
	y := string(serialNumber[4])
	yyyy := "19"
	if yy == "0" {
		yyyy = "20"
	}
	year := fmt.Sprintf("%s%s%s", yyyy, yy, y)
	yi, err := strconv.Atoi(year)
	if err != nil {
		log.Println(err)
	}
	if 1977 <= yi && yi < 2005 && len(serialNumber) == 8 {
		return true
	}
	if 2005 <= yi {
		date := time.Date(yi, time.Month(1), 1, 0, 0, 0, 0, time.Local)
		ddd := fmt.Sprintf("%s%s%s",
			string(serialNumber[1]),
			string(serialNumber[2]),
			string(serialNumber[3]))
		di, err := strconv.Atoi(ddd)
		if err != nil {
			log.Println(err)
		}
		date = date.AddDate(0, 0, di-1)
		if date.Month() <= 6 && len(serialNumber) == 8 {
			return true
		} else if 7 <= date.Month() && len(serialNumber) == 9 {
			return true
		}
	}
	return false
}

func isCentennialYear(serialNumber string) bool {
	// 94RRRRRR
	re := regexp.MustCompile("^94\\d{6}$")
	if !re.MatchString(serialNumber) {
		return false
	}

	head := fmt.Sprintf("%s%s", string(serialNumber[0]), string(serialNumber[1]))
	if head == "94" && len(serialNumber) == 8 {
		return true
	}
	return false
}

func isCustomShopReissues50s(serialNumber string) bool {
	// 1952-1960 Les Paul, Explorer, Flying V, and Futura reissues (since late 1992):
	// M YRRR or MYRRRR
	// M is the model year being reissued
	re1 := regexp.MustCompile("^\\d{1}\\s{1}\\d{4}$")
	re2 := regexp.MustCompile("^\\d{6}$")
	if !re1.MatchString(serialNumber) && !re2.MatchString(serialNumber) {
		return false
	}

	y := string(serialNumber[1])
	if y == " " {
		y = string(serialNumber[2])
	}
	year := fmt.Sprintf("200%s", y) // todo
	yi, err := strconv.Atoi(year)
	if err != nil {
		log.Println(err)
	}
	if 1992 <= yi {
		return true
	}
	return false
}

func isCustomShopReissues60s(serialNumber string) bool {
	// 1961-1969 Firebird, Les Paul, and SG reissues (since 1997):
	// YYRRRM
	re := regexp.MustCompile("^\\d{6}$")
	if !re.MatchString(serialNumber) {
		return false
	}

	yy := string(serialNumber[0])
	y := string(serialNumber[1])
	yyyy := "19"
	if yy == "0" {
		yyyy = "20"
	}
	year := fmt.Sprintf("%s%s%s", yyyy, yy, y)
	yi, err := strconv.Atoi(year)
	if err != nil {
		log.Println(err)
	}
	if 1997 <= yi {
		return true
	}
	return false
}

func isCustomShopRegular(serialNumber string) bool {
	// CSYRRRR
	re := regexp.MustCompile("^CS\\d{5}$")
	if re.MatchString(serialNumber) {
		return true
	}
	return false
}

func isCustomShopCarvedTop(serialNumber string) bool {
	// YDDDYRRR
	re := regexp.MustCompile("^\\d{8}$")
	if re.MatchString(serialNumber) {
		return true
	}
	return false
}

func isEsSeries(serialNumber string) bool {
	// (A or B)-MYRRR
	re := regexp.MustCompile("^[A|B]-\\d{5}$")
	if re.MatchString(serialNumber) {
		return true
	}
	return false
}

func isLesPaulClassic(serialNumber string) bool {
	upTo1999 := regexp.MustCompile("^\\d{1}\\s{1}\\d{3,4}")
	if upTo1999.MatchString(serialNumber) {
		return true
	}

	since2000 := regexp.MustCompile("^\\d{6}")
	if since2000.MatchString(serialNumber) {
		return true
	}

	return false
}

func validSerialNumber(serialNumber string) bool {
	if isLesPaulClassic(serialNumber) {
		return true
	} else if isCustomShopReissues50s(serialNumber) {
		return true
	} else if isCustomShopReissues60s(serialNumber) {
		return true
	} else if isCustomShopCarvedTop(serialNumber) {
		return true
	} else if isCustomShopRegular(serialNumber) {
		return true
	} else if isEsSeries(serialNumber) {
		return true
	} else if isReguler(serialNumber) {
		return true
	}
	return false
}

type LogLine struct {
	RemoteAddr  string
	ContentType string
	Path        string
	Query       string
	Method      string
	Body        string
	UserAgent   string
}

var LogTemplate = `{{.RemoteAddr}} {{.ContentType}} {{.Method}} {{.Path}} {{.Query}} {{.Body}} {{.UserAgent}}`

func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bufbody := new(bytes.Buffer)
		bufbody.ReadFrom(req.Body)
		body := bufbody.String()
		line := LogLine{
			req.RemoteAddr,
			req.Header.Get("Content-Type"),
			req.URL.Path,
			req.URL.RawQuery,
			req.Method, body, req.UserAgent(),
		}
		tmpl, err := template.New("line").Parse(LogTemplate)
		if err != nil {
			panic(err)
		}
		bufline := new(bytes.Buffer)
		err = tmpl.Execute(bufline, line)
		if err != nil {
			panic(err)
		}

		logFile := fmt.Sprintf("./log/access/%d%d%d.log", AccessTime.Year(), AccessTime.Month(), AccessTime.Day())
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
		log.Printf(bufline.String())

		handler.ServeHTTP(w, req)
	})
}

func Server() {
	http.HandleFunc("/", handler)
	defaultPort := 80
	port := defaultPort
	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.IntVar(&port, "p", defaultPort, "listen port")
	f.IntVar(&port, "port", defaultPort, "listen port")
	f.Parse(os.Args[1:])
	for 0 < f.NArg() {
		f.Parse(f.Args()[1:])
	}
	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, Log(http.DefaultServeMux))
}

func main() {
	Server()
}
