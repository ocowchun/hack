package main

import (
	"fmt"
	"hack/vm/translator"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the vm file")
	}
	inputStr := os.Args[1]
	f, err := os.Open(inputStr)
	if err != nil {
		log.Fatal(err)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	inputFilePaths := make([]string, 0)
	if fileInfo.IsDir() {
		files, err := os.ReadDir(inputStr)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".vm") {
				inputFilePaths = append(inputFilePaths, inputStr+"/"+file.Name())
			}
		}
		sort.SliceStable(inputFilePaths, func(i, j int) bool {
			return strings.HasSuffix(inputFilePaths[i], "/Sys.vm")
		})
	} else {
		inputFilePaths = append(inputFilePaths, inputStr)
	}
	// TODO: improve it
	fmt.Println("@256")
	fmt.Println("D=A")
	fmt.Println("@SP")
	fmt.Println("M=D")
	for _, inputFilePath := range inputFilePaths {
		inputFile, err := os.Open(inputFilePath)
		if err != nil {
			log.Fatal(err)
		}

		parser := translator.NewParser(inputFile)
		tokens := strings.Split(inputFilePath, "/")
		writer := translator.NewWriter(tokens[len(tokens)-1])

		for parser.HasMoreCommands() {
			err = parser.Advance()
			if err != nil {
				fmt.Println("parser.Advance")
				log.Fatal(err)
			}
			cmd := parser.CurrentCommand()
			asms, err := writer.Write(cmd)
			if err != nil {
				fmt.Println("writer.Write")
				log.Fatal(err)
			}
			fmt.Printf("// %s\n", cmd.String())
			for _, a := range asms {
				fmt.Println(a)
			}
		}
	}

}
