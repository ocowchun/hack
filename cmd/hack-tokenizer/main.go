package main

import (
	"fmt"
	"hack/compiler"
	"io"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please specify the jack file")
	}
	inputFilePath := os.Args[1]
	inputFile, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatal(err)
	}
	tokenizer := compiler.NewTokenizer(inputFile)
	fmt.Println("<tokens>")
	for {
		token, err := tokenizer.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(token)
	}
	fmt.Println("</tokens>")
}
