package main

import (
	"fmt"
	"hack/compiler"
	"log"
	"os"
	"strings"
)

func buildVmFileName(input string) string {
	removedSpace := strings.Replace(input, " ", "_", -1)
	tokens := strings.Split(removedSpace, "/")
	return tokens[len(tokens)-1] + ".vm"
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the jack file")
	}
	inputFilePath := os.Args[1]
	f, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	inputFilePaths := make([]string, 0)
	if fileInfo.IsDir() {
		files, err := os.ReadDir(inputFilePath)
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".jack") {
				fmt.Println("yo", inputFilePath+"/"+file.Name())
				inputFilePaths = append(inputFilePaths, inputFilePath+"/"+file.Name())
			}
		}
	} else {
		inputFilePaths = append(inputFilePaths, inputFilePath)
	}

	for _, filePath := range inputFilePaths {
		inputFile, err := os.Open(filePath)

		if err != nil {
			log.Fatal(err)
		}
		engine := compiler.NewEngine(inputFile)
		class, err := engine.CompileClass()

		if err != nil {
			log.Fatal(err)
		}

		outputFileName := strings.Replace(filePath, ".jack", ".vm", -1)
		output, err := os.Create(outputFileName)
		writer := compiler.NewVmWriter(output, class)
		err = writer.Write()
		if err != nil {
			log.Fatal(err)
		}
	}
	//outputFileName := ""
	//
	//inputFile, err := os.Open(inputFilePath)
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//engine := compiler.NewEngine(inputFile)
	//class, err := engine.CompileClass()
	//
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//writer := compiler.NewVmWriter(os.Stdout, class)
	//err = writer.Write()
	//if err != nil {
	//	log.Fatal(err)
	//}
}
