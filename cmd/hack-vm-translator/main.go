package main

import (
	"bufio"
	"context"
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

	ctx := context.Background()
	asmCh := parsePipeline(ctx, inputFilePaths)
	writerCompleted := writeFilePipeline(ctx, asmCh, w)

	select {
	case <-ctx.Done():
		log.Fatal(ctx.Err())

	case <-writerCompleted:
		err = w.Flush()
		if err != nil {
			log.Fatal(err)
		}
	}
}

type WritePipelineResult struct {
	counter int64
}

func writePipeline(parentCtx context.Context, cmdCh <-chan translator.VmCommand, writerName string, counter int64, asmCh chan<- string) chan WritePipelineResult {
	ctx, cancel := context.WithCancelCause(parentCtx)
	writer := translator.NewWriter(writerName, counter)
	completed := make(chan WritePipelineResult)
	go func() {
		for cmd := range cmdCh {
			select {
			case <-ctx.Done():
				return
			default:
				asms, err := writer.Write(cmd)
				if err != nil {
					cancel(err)
					return
				}
				// write vm comment
				asmCh <- fmt.Sprintf("// %s\n", cmd.String())
				// write asm
				for _, a := range asms {
					asmCh <- a
				}
			}
		}
		completed <- WritePipelineResult{counter: writer.Counter()}
	}()
	return completed
}

func parsePipeline(parentCtx context.Context, inputFilePaths []string) <-chan string {
	counter := int64(0)
	ctx, cancel := context.WithCancelCause(parentCtx)
	asmCh := make(chan string)
	go func() {
		for _, inputFilePath := range inputFilePaths {
			inputFile, err := os.Open(inputFilePath)
			if err != nil {
				cancel(err)
				return
			}

			parser := translator.NewParser(inputFile)
			tokens := strings.Split(inputFilePath, "/")
			cmdCh := make(chan translator.VmCommand)
			completed := writePipeline(ctx, cmdCh, tokens[len(tokens)-1], counter, asmCh)

			for parser.HasMoreCommands() {
				select {
				case <-ctx.Done():
					return
				default:
					err = parser.Advance()
					if err != nil {
						cancel(err)
						return
					}
					cmd := parser.CurrentCommand()
					cmdCh <- cmd
				}
			}
			close(cmdCh)
			res := <-completed
			counter = res.counter
		}
		close(asmCh)
	}()
	return asmCh
}

func writeFilePipeline(parentCtx context.Context, asmCh <-chan string, w *bufio.Writer) <-chan struct{} {
	ctx, cancel := context.WithCancelCause(parentCtx)
	writerCompleted := make(chan struct{})
	go func() {
		for a := range asmCh {
			select {
			case <-ctx.Done():
				return

			default:
				_, err := w.WriteString(fmt.Sprintln(a))
				if err != nil {
					cancel(err)
					return
				}
			}
		}
		writerCompleted <- struct{}{}
	}()
	return writerCompleted
}
