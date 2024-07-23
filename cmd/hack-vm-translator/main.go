package main

import (
	"bufio"
	"flag"
	"fmt"
	"hack/vm/translator"
	"log"
	"os"
	"sort"
	"strings"
)

func buildAsmFileName(input string) string {
	removedSpace := strings.Replace(input, " ", "_", -1)
	tokens := strings.Split(removedSpace, "/")
	return tokens[len(tokens)-1] + ".asm"
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the vm file")
	}
	inputStr := os.Args[len(os.Args)-1]
	f, err := os.Open(inputStr)
	if err != nil {
		log.Fatal(err)
	}
	fileInfo, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	inputFilePaths := make([]string, 0)
	outputFileName := ""
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
		outputFileName = buildAsmFileName(inputStr)
	} else {
		inputFilePaths = append(inputFilePaths, inputStr)
		outputFileName = buildAsmFileName(inputStr)
	}

	output, err := os.Create(outputFileName)
	w := bufio.NewWriter(output)
	shouldBootstrap := flag.Bool("bootstrap", false, "whether to bootstrap or not")
	flag.Parse()
	if shouldBootstrap == nil {
		fmt.Println("bootstrap is nil")
	} else {
		fmt.Println("bootstrap is ", *shouldBootstrap)
	}
	if shouldBootstrap != nil && *shouldBootstrap {
		// TODO: improve it
		boostrapCmds := []string{
			"@256",
			"D=A",
			"@SP",
			"M=D",
		}
		for _, cmd := range boostrapCmds {
			_, err = w.WriteString(fmt.Sprintln(cmd))
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	counter := int64(0)
	for _, inputFilePath := range inputFilePaths {
		inputFile, err := os.Open(inputFilePath)
		if err != nil {
			log.Fatal(err)
		}

		parser := translator.NewParser(inputFile)
		tokens := strings.Split(inputFilePath, "/")
		writer := translator.NewWriter(tokens[len(tokens)-1], counter)

		for parser.HasMoreCommands() {
			err = parser.Advance()
			if err != nil {
				log.Fatal(err)
			}
			cmd := parser.CurrentCommand()
			asms, err := writer.Write(cmd)
			if err != nil {
				log.Fatal(err)
			}

			// write vm command
			_, err = w.WriteString(fmt.Sprintf("// %s\n", cmd.String()))
			if err != nil {
				log.Fatal(err)
			}

			// write assembly
			for _, a := range asms {
				_, err = w.WriteString(fmt.Sprintln(a))
				if err != nil {
					log.Fatal(err)
				}
			}
		}
		counter = writer.Counter()
	}

	err = w.Flush()
	if err != nil {
		log.Fatal(err)
	}

}
