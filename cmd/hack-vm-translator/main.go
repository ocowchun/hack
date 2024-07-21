package main

import (
	"fmt"
	"hack/vm/translator"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the vm file")
	}
	inputFilePath := os.Args[1]
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
