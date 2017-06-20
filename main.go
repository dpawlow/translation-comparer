package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"
	"path/filepath"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getFilesDiffAndWrite(filename string, path string) error {
	var path1 string = fmt.Sprintf("%s/crowdin/%s", path, filename)
	var path2 string = fmt.Sprintf("%s/babel/%s", path, filename)
	var diff map[string][2]string = compareFiles(path1, path2)

	outputDir := filepath.Join(".", "output")
	check(os.MkdirAll(outputDir, os.ModePerm))

	var output string = fmt.Sprintf("%s/output/%s", path, filename)
	file, err := os.Create(output)
	check(err)
	defer file.Close()

	w := bufio.NewWriter(file)
	for key, value := range diff {
		fmt.Fprintln(w, key)
		fmt.Fprintln(w, value[0])
		fmt.Fprintln(w, value[1])
		fmt.Fprintln(w, "")
	}

	fmt.Printf("Wrote to %s \n", output)
	return w.Flush()
}

func compareFiles(filepath1 string, filepath2 string) map[string][2]string {
	diff := make(map[string][2]string)
	var m1 map[string]string = loadFileToMemory(filepath1)
	var m2 map[string]string = loadFileToMemory(filepath2)

	for key, value1 := range m1 {
		value2, ok := m2[key]
		if ok && value1 != value2 {
			diff[key] = [2]string{value1, value2}
		} else if !ok && !strings.Contains(value1, "\"\"") {
			diff[key] = [2]string{value1, "No hay key en babel"}
		}
	}
	return diff
}

func loadFileToMemory(filepath string) map[string]string {
	translations := make(map[string]string)
	file, err := os.Open(filepath)
	check(err)
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var key string
	var value string
	var estadoCarga [2]bool = [2]bool{false, false}
	for scanner.Scan() {
		var line []string = strings.Split(scanner.Text(), "\"")
		if strings.Contains(line[0], "msgid") {
			key = line[1]
			estadoCarga[0] = true
		} else if strings.Contains(line[0], "msgstr") {
			value = line[1]
			estadoCarga[1] = true
		}
		if estadoCarga[0] && estadoCarga[1] {
			translations[key] = value
			key = ""
			value = ""
			estadoCarga[0], estadoCarga[1] = false, false
		}
	}
	return translations
}

func generateTests(filename string, filePath string) error {
	var path string = fmt.Sprintf("%s/crowdin/%s", filePath, filename)
	m := loadFileToMemory(path)

	outputDir := filepath.Join(".", "test")
	check(os.MkdirAll(outputDir, os.ModePerm))

	var output string = fmt.Sprintf("%s/test/%s", filePath, filename)
	file, err := os.Create(output)
	check(err)
	defer file.Close()

	w := bufio.NewWriter(file)
	for key, value := range m {
		if value != "" {
			test := fmt.Sprintf("Assert.assertEquals(\"%s\".encodeAsHTML() ,\tt9nService.tr(s:\"%s\", locale:\"${lang}\"))", value, key)
			fmt.Fprintln(w, test)
		}
	}

	fmt.Printf("Wrote to %s \n", output)
	return w.Flush()
}

func main() {
	path := flag.String("path", ".", "Path to files.")
	test := flag.String("test", "false", "Generate tests?")
	flag.Parse()
	fileNames := flag.Args()

	var wg sync.WaitGroup

	for _, f := range fileNames {
		wg.Add(1)
		go func(file string, pathtofile string) {
			if *test == "true" {
				err := generateTests(file, pathtofile)
				check(err)
			} else {
				err := getFilesDiffAndWrite(file, pathtofile)
				check(err)
			}
			wg.Done()
		}(f, *path)
	}

	wg.Wait()
}
