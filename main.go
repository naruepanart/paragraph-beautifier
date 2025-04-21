package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	fillerWords = regexp.MustCompile(`\b(uh|um)\b`)
	multiSpace  = regexp.MustCompile(`\s+`)
	pool        = sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	}
)

var version = "dev"

func formatParagraph(text string, buf *bytes.Buffer) string {
	buf.Reset()
	cleaned := fillerWords.ReplaceAllString(text, "")
	cleaned = multiSpace.ReplaceAllString(cleaned, " ")
	cleaned = strings.TrimSpace(cleaned)

	sentences := strings.FieldsFunc(cleaned, func(r rune) bool {
		return r == '.' || r == '!' || r == '?'
	})

	paragraphs := make([]string, 0, len(sentences)/3+1)
	currentParagraph := strings.Builder{}
	currentParagraph.Grow(256)
	sentenceCount := 0
	wordCount := 0

	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}

		if len(sentence) > 0 {
			sentence = strings.ToUpper(sentence[:1]) + sentence[1:] + "."
		}

		if currentParagraph.Len() > 0 {
			currentParagraph.WriteByte(' ')
		}
		currentParagraph.WriteString(sentence)
		sentenceCount++
		wordCount += strings.Count(sentence, " ") + 1

		if sentenceCount >= 3 || wordCount > 50 {
			paragraphs = append(paragraphs, currentParagraph.String())
			currentParagraph.Reset()
			sentenceCount = 0
			wordCount = 0
		}
	}

	if currentParagraph.Len() > 0 {
		paragraphs = append(paragraphs, currentParagraph.String())
	}

	return strings.Join(paragraphs, "\n\n")
}

func processFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %v", filename, err)
	}
	defer file.Close()

	var input bytes.Buffer
	if _, err := io.Copy(&input, bufio.NewReader(file)); err != nil {
		return fmt.Errorf("error reading file %s: %v", filename, err)
	}

	buf := pool.Get().(*bytes.Buffer)
	defer pool.Put(buf)

	formattedText := formatParagraph(input.String(), buf)

	if err := os.WriteFile(filename, []byte(formattedText), 0644); err != nil {
		return fmt.Errorf("error writing to file %s: %v", filename, err)
	}

	return nil
}

func main() {
	fmt.Printf("App version: %s\n", version)

	files, err := filepath.Glob("*.txt")
	if err != nil {
		fmt.Printf("Error finding .txt files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No .txt files found")
		return
	}

	for _, file := range files {
		fmt.Printf("Processing %s ", file)
		if err := processFile(file); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("Done")
		}
	}
}
