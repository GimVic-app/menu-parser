package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type Kosilo struct {
	Navadno        []string
	Vegetarijansko []string
}

type Malica struct {
	Navadna         []string
	VegSPerutnino   []string
	Vegetarijanska  []string
	SadnoZelenjavna []string
}

func main() {

	//check for file argument
	if len(os.Args) < 2 {
		fmt.Println("Specifiy the file!")
	} else {
		fileArg := os.Args[1]

		if isJedilnikValid(fileArg) {
			fmt.Println("Parsing file " + fileArg)

			csvfile, err := os.Open(fileArg)
			defer csvfile.Close()
			check(err)

			reader := csv.NewReader(csvfile)
			reader.Comma = ';'
			reader.FieldsPerRecord = -1
			rawCSVdata, err := reader.ReadAll()
			check(err)

			fmt.Println("Parser read file as csv")

			secNumbers, malica := getSectionNumbers(rawCSVdata)

			if malica {
				fmt.Println("Recognized as malica")
			} else {
				fmt.Println("Recognized as kosilo")
			}
			fmt.Println("Parser generated section numbers list: ", secNumbers, "\n\n")

			if malica {
				processMalica(rawCSVdata, secNumbers)
			} else {
				processKosilo(rawCSVdata, secNumbers)
			}

		} else {
			panic("Invalid jedilnik!")
		}
	}
}

//provides table sections and type (malica = true, kosilo = false)
func getSectionNumbers(csv [][]string) ([]int, bool) {
	var result []int
	var malica bool

	for i, line := range csv {
		if strings.Contains(strings.ToLower(line[1]), "navadna") || strings.Contains(strings.ToLower(line[1]), "kosilo") {
			result = append(result, i)
			if strings.Contains(line[1], "navadna") {
				malica = true
			}
		}
	}

	return result, malica
}

func processMalica(table [][]string, selNumbers []int) {

	for i, num := range selNumbers {
		if i+1 == len(selNumbers) {
			processMalicaSel(table[num:len(table)])
		} else {
			processMalicaSel(table[num:selNumbers[i+1]])
		}
	}

	fmt.Println("Parser done processing malica.\n")
}

func processMalicaSel(sel [][]string) {
	date := findDate(sel)

	temp := new(Malica)
	for i := 1; i < len(sel); i++ {
		if sel[i][1] != "" {
			temp.Navadna = append(temp.Navadna, sel[i][1])
		}
		if sel[i][3] != "" {
			temp.VegSPerutnino = append(temp.VegSPerutnino, sel[i][3])
		}
		if sel[i][5] != "" {
			temp.Vegetarijanska = append(temp.Vegetarijanska, sel[i][5])
		}
		if sel[i][6] != "" {
			temp.SadnoZelenjavna = append(temp.SadnoZelenjavna, sel[i][5])
		}
	}

	jsonData, _ := json.Marshal(temp)
	fmt.Println("Generated json for ", date, ": ", string(jsonData))

	con, err := sql.Open("mysql", "app:urnikZAvse@/app")
	check(err)
	defer con.Close()

	_, err = con.Exec("insert into malica (date, json) values (?, ?)", date, jsonData)
	check(err)

}

func processKosiloSel(sel [][]string) {
	date := findDate(sel)

	temp := new(Kosilo)
	for i := 1; i < len(sel); i++ {
		if sel[i][1] != "" {
			temp.Navadno = append(temp.Navadno, sel[i][1])
		}
		if sel[i][3] != "" {
			temp.Vegetarijansko = append(temp.Vegetarijansko, sel[i][3])
		}
	}

	jsonData, _ := json.Marshal(temp)
	fmt.Println("Generated json for ", date, ": ", string(jsonData))

	con, err := sql.Open("mysql", "app:urnikZAvse@/app")
	check(err)
	defer con.Close()

	_, err = con.Exec("insert into kosilo (date, json) values (?, ?)", date, jsonData)
	check(err)

}

func processKosilo(table [][]string, selNumbers []int) {
	for i, num := range selNumbers {
		if i+1 == len(selNumbers) {
			processKosiloSel(table[num:len(table)])
		} else {
			processKosiloSel(table[num:selNumbers[i+1]])
		}
	}

	fmt.Println("Parser done processing kosilo.\n")
}

func findDate(sel [][]string) time.Time {
	var index int = 0
	for i, line := range sel {
		cell := strings.ToLower(line[0])
		if strings.Contains(cell, "pon") || strings.Contains(cell, "tor") || strings.Contains(cell, "sre") || strings.Contains(cell, "rtek") || strings.Contains(cell, "pet") {
			index = i + 1
		}
	}

	date, err := time.Parse("2.1.2006", sel[index][0])
	check(err)
	return date
}

func isJedilnikValid(fileName string) bool {
	csv, err := ioutil.ReadFile(fileName)
	check(err)

	if strings.Count(string(csv), ";") == 240 && (strings.Contains(strings.ToLower(string(csv)), "navadna") || strings.Contains(strings.ToLower(string(csv)), "kosilo")) {
		return true
	}
	return false
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
