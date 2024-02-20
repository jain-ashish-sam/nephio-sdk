package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"github.com/sirupsen/logrus"
)

/*
Reads the input.txt file and returns the Comma-Separated Rows in a 2D-Slice
*/
func readInputFile(filename string) (out [][]string) {
	file, err := os.Open(filename)
	if err != nil {
		logrus.Fatal("Error Occured While Opening Input-File| Error : ", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 1
	for scanner.Scan() {
		line := scanner.Text()
		lineContent := strings.Split(line, ",")
		if len(lineContent) != 2 {
			logrus.Warn("Expected only Single Comma , | Found ", len(lineContent)-1, " at Line Number ", lineCount, " | Skipping")
			continue
		}

		if lineContent[0] == "Module" {
			continue
		}

		out = append(out, lineContent)

		lineCount += 1
	}

	if err := scanner.Err(); err != nil {
		logrus.Fatal("Error Occured While Reading Input-File| Error : ", err)
	}
	return
}

/*
Fetchs the FileContent from url provided
*/
func getFileFromUrl(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

/*
It finds and list all type-name matching the regex "type <struct-name> typeToFind"
Eg: "type <struct-name> struct" will list all the struct-types
*/
func findAndList(fileContent string, typeToFind string) (out []string) {
	regExpression := fmt.Sprintf(`type([ ]+)([A-Za-z]+)([ ]+)%s`, typeToFind)
	re := regexp.MustCompile(regExpression)
	matchedLines := re.FindAllString(fileContent, -1)
	structNameRegex := regexp.MustCompile(`([ ]+)([A-Za-z]+)([ ]+)`)
	for _, lines := range matchedLines {
		structName := structNameRegex.FindAllString(lines, 1)[0]
		structName = strings.Trim(structName, " ")
		out = append(out, structName)
	}
	return
}

func writeToJson(data map[string][]string, outfilename string) {
	jsonString, _ := json.MarshalIndent(data, "", "    ")
	err := os.WriteFile(outfilename, jsonString, 0777)
	if err != nil {
		logrus.Fatal("Failed to Write Json File ", outfilename, " : Error | ", err)
	}

}

func printSummary(summary map[string]map[string]int) {
	logrus.Info("---------------- Run Successful ---------------\n Check the Summary")
	// Printing the summary
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"MOduleName", "Struct", "String-Enum", "Map-Struct", "Int-Enum", "Float-Enum"})
	for filename, val := range summary {
		cur_row := table.Row{filename}
		keys := []string{"struct", "string", "map", "int", "float"}
		for _, k := range keys {
			cur_row = append(cur_row, val[k])
		}
		t.AppendRow(cur_row)
	}
	t.Render()
}

func main() {
	fileContents := readInputFile("input.txt")
	var summary = make(map[string]map[string]int)
	var structModuleMap = make(map[string][]string)
	var enumModuleMap = make(map[string][]string)

	for _, content := range fileContents {
		moduleName, url := content[0], content[1]
		var interSummary = make(map[string]int)

		fileContent, err := getFileFromUrl(strings.TrimSpace(url))
		if err != nil {
			logrus.Error("Error While Downloading File from URL ", url, " | Error : ", err)
			continue
		}

		// Listing all the structs i.e. `type <structname> struct`
		struct_names_found := findAndList(fileContent, "struct")
		structModuleMap[moduleName] = struct_names_found
		interSummary["struct"] = len(struct_names_found)

		// Listing enums
		emptySlice := []string{}
		enumModuleMap[moduleName] = emptySlice
		enumTypes := []string{"string", "map", "int", "float"}
		for _, enumType := range enumTypes {
			enumNamesFound := findAndList(fileContent, enumType)
			structModuleMap[moduleName] = append(structModuleMap[moduleName], enumNamesFound...)
			if enumType != "map" {
				enumModuleMap[moduleName] = append(enumModuleMap[moduleName], enumNamesFound...)
			}

			interSummary[enumType] = len(enumNamesFound)
		}
		summary[moduleName] = interSummary
	}

	writeToJson(structModuleMap, "output/struct_module_mapping.json")
	writeToJson(enumModuleMap, "output/enum_module_mapping.json")
	printSummary(summary)
	logrus.Info("Check the output folder for generated-json files")

}
