package main
import (
	"net/http"
	"fmt"
	"encoding/json"
	"strconv"
	"time"
	"regexp"
)

func main () {
	Server()
}

type RequestParams struct {
	SerialNumber string
}

func handler(w http.ResponseWriter, req *http.Request) {
	var serialNumber string

	if req.Method == "POST" {
		decorder := json.NewDecoder(req.Body)
		var rp RequestParams
		err := decorder.Decode(&rp)
		if err != nil {
			fmt.Println(err)
		}
		serialNumber = rp.SerialNumber
	} else {
		req.ParseForm()
		serialNumber = req.Form.Get("serialNumber")
	}

	if !validSerialNumber(serialNumber) {
		fmt.Fprintf(w, "invalid serial number.")
		return
	}

	var out string
	if isCentennialYear(serialNumber) {
		// 1994 Gibson's Centennial year
		ppp := fmt.Sprintf("%s%s%s",
				string(serialNumber[2]),
				string(serialNumber[3]),
				string(serialNumber[4]),
				string(serialNumber[5]),
				string(serialNumber[6]),
				string(serialNumber[7]))
		pi, err := strconv.Atoi(ppp)
		if err != nil {
			fmt.Println(err)
		}
		var p string
		if pi == 1 {
			p = "1st"
		} else if pi == 2 {
			p = "2nd"
		} else if pi == 3 {
			p = "3rd"
		} else {
			p = fmt.Sprintf("%dth", pi)
		}
		out = fmt.Sprintf("SerialNumber: %s\n"+
			"Date: 1994 (In 1994, Gibson's Centennial year) \n"+
			"Instrument Rank: %d",
			serialNumber, p)
	} else if isCustomShopRegular(serialNumber) {
		// custom shop
	} else if isCustomShopReissues50s(serialNumber) {
		// custom shop reissues50s
	} else if isCustomShopReissues60s(serialNumber) {
		// custom shop reissues60s
	} else if isEsSeries(serialNumber) {
		// ES (Electric Spanish)
	} else if isLesPaulClassic(serialNumber) {
		// Les Paul Classic
	} else {
		// regular
		// 1977 ~ 2005.07 = YDDDYPPP
		//   YY is the production year
		//   DDD is the day of the year
		//   PPP is the plant designation and/or instrument rank
		// 2005.07 ~ now = YDDDYBPPP
		//   B is the batch number

		// YY
		yy := string(serialNumber[0])
		y := string(serialNumber[4])
		yyyy := "19"
		if yy == "0" {
			yyyy = "20"
		}
		year := fmt.Sprintf("%s%s%s", yyyy, yy, y)

		// DDD
		yi, err := strconv.Atoi(year)
		if err != nil {
			fmt.Println(err)
		}
		date := time.Date(yi, time.Month(1), 1, 0, 0, 0, 0, time.Local)
		ddd := fmt.Sprintf("%s%s%s",
			string(serialNumber[1]),
			string(serialNumber[2]),
			string(serialNumber[3]))
		di, err := strconv.Atoi(ddd)
		if err != nil {
			fmt.Println(err)
		}
		date = date.AddDate(0, 0, di - 1)

		// PPP
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
		pi, err := strconv.Atoi(ppp)
		if err != nil {
			fmt.Println(err)
		}
		var p string
		if pi == 1 {
			p = "1st"
		} else if pi == 2 {
			p = "2nd"
		} else if pi == 3 {
			p = "3rd"
		} else {
			p = fmt.Sprintf("%dth", pi)
		}

		// shapes
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
	}

	n := time.Now()
	fmt.Printf("[AccessLog] %d-%02d-%02d %02d:%02d:%02d\n",
		n.Year(), n.Month(), n.Day(), n.Hour(), n.Minute(), n.Second())
	fmt.Println(out)
	fmt.Fprintf(w, "%s", out)
}

func isReguler(serialNumber string) bool {
	// length 8, 1977 - 2005.06
	// or
	// length 9, 2005.07 ~
	yy := string(serialNumber[0])
	y := string(serialNumber[4])
	yyyy := "19"
	if yy == "0" {
		yyyy = "20"
	}
	year := fmt.Sprintf("%s%s%s", yyyy, yy, y)
	yi, err := strconv.Atoi(year)
	if err != nil {
		fmt.Println(err)
	}
	if 1977 <= yi && yi < 2005 && len(serialNumber) > 8 {

	}
	if 2005 <= yi {
		date := time.Date(yi, time.Month(1), 1, 0, 0, 0, 0, time.Local)
		ddd := fmt.Sprintf("%s%s%s",
			string(serialNumber[1]),
			string(serialNumber[2]),
			string(serialNumber[3]))
		di, err := strconv.Atoi(ddd)
		if err != nil {
			fmt.Println(err)
		}
		date = date.AddDate(0, 0, di - 1)
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
	yy := string(serialNumber[0])
	y := string(serialNumber[1])
	yyyy := "19"
	if yy == "0" {
		yyyy = "20"
	}
	year := fmt.Sprintf("%s%s%s", yyyy, yy, y)
	yi, err := strconv.Atoi(year)
	if err != nil {
		fmt.Println(err)
	}
	if yi < 1992 {
		return false
	}
	var re *regexp.Regexp
	re = regexp.MustCompile("^\\d{1}\\s{1}\\d{4}")
	if len(re.FindString(serialNumber)) > 0 {
		return true
	}
	re = regexp.MustCompile("^\\d{6}")
	if len(re.FindString(serialNumber)) > 0 {
		return true
	}
	return false
}

func isCustomShopReissues60s(serialNumber string) bool {
	// 1961-1969 Firebird, Les Paul, and SG reissues (since 1997):
	// YYRRRM
	yy := string(serialNumber[0])
	y := string(serialNumber[1])
	yyyy := "19"
	if yy == "0" {
		yyyy = "20"
	}
	year := fmt.Sprintf("%s%s%s", yyyy, yy, y)
	yi, err := strconv.Atoi(year)
	if err != nil {
		fmt.Println(err)
	}
	if yi < 1997 {
		return false
	}
	re := regexp.MustCompile("^\\d{6}")
	if len(re.FindString(serialNumber)) > 0 {
		return true
	}
	return false
}

func isCustomShopRegular(serialNumber string) bool {
	// CSYRRRR
	head := fmt.Sprintf("%s%s", string(serialNumber[0]), string(serialNumber[1]))
	if head == "CS" && len(serialNumber) == 7 {
		return true
	}
	return false
}

func isEsSeries(serialNumber string) bool {
	// length 6, head A or B
	head := fmt.Sprintf("%s%s", string(serialNumber[0]), string(serialNumber[1]))
	if (head == "A" || head == "B") && len(serialNumber) == 6 {
		return true
	}
	return false
}

func isLesPaulClassic(serialNumber string) bool {
	var match string

	upTo1999 := regexp.MustCompile("^\\d{1}\\s{1}\\d{3,4}")
	match = upTo1999.FindString(serialNumber)
	if len(match) > 0 {
		return true
	}

	since2000 := regexp.MustCompile("^\\d{6}")
	match = since2000.FindString(serialNumber)
	if len(match) > 0 {
		return true
	}

	return false
}

func validSerialNumber(serialNumber string) bool{
	if isLesPaulClassic (serialNumber) {
		return true
	} else if isCustomShopReissues50s (serialNumber) {
		return true
	} else if isCustomShopReissues60s (serialNumber) {
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

func Server() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}


