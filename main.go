package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

type LoggingPrinter struct {
	logBuilder *strings.Builder
}

func (lp *LoggingPrinter) Printf(format string, a ...any) {
	line := fmt.Sprintf(format, a...)
	lp.logBuilder.Write([]byte(line))
	fmt.Print(line)
}

func (lp *LoggingPrinter) Println(a ...any) {
	line := fmt.Sprintln(a...)
	lp.logBuilder.Write([]byte(line))
	fmt.Print(line)
}

type LoggingScanner struct {
	scanner    *bufio.Scanner
	logBuilder *strings.Builder
}

func (ls *LoggingScanner) Scan() bool {
	return ls.scanner.Scan()
}

func (ls *LoggingScanner) Text() string {
	text := ls.scanner.Text()
	ls.logBuilder.WriteString(text + "\n")
	return text
}

type Flashcard struct {
	Term       string
	Definition string
	Mistakes   int
}

type Flashcards struct {
	elements map[int]Flashcard
}

func (fc *Flashcards) FindDefinitionByTerm(term string) (string, bool) {
	for _, flashcard := range fc.elements {
		if flashcard.Term == term {
			return flashcard.Definition, true
		}
	}
	return "", false
}

func (fc *Flashcards) FindTermByDefinition(definition string) (string, bool) {
	for _, flashcard := range fc.elements {
		if flashcard.Definition == definition {
			return flashcard.Term, true
		}
	}
	return "", false
}

func (fc *Flashcards) CreateOrUpdate(flashcard Flashcard) {
	for index, existingFlashcard := range fc.elements {
		if existingFlashcard.Term == flashcard.Term {
			fc.elements[index] = flashcard
			return
		}
	}
	fc.elements[len(fc.elements)] = flashcard
}

func (fc *Flashcards) RemoveByTerm(term string) {
	lastIndex := len(fc.elements) - 1
	for index, flashcard := range fc.elements {
		if flashcard.Term == term {
			fc.elements[index] = fc.elements[lastIndex]
			delete(fc.elements, lastIndex)
		}
	}
}

func (fc *Flashcards) GetRandomFc() Flashcard {
	return fc.elements[rand.Intn(len(fc.elements))]
}

func (fc *Flashcards) WriteCSV(filename string) int {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, flashcard := range fc.elements {
		record := []string{flashcard.Term, flashcard.Definition, strconv.Itoa(flashcard.Mistakes)}
		if err := writer.Write(record); err != nil {
			log.Fatal(err)
		}
	}
	return len(fc.elements)
}

func (fc *Flashcards) ReadCSV(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, record := range records {
		mistakes, _ := strconv.Atoi(record[2])
		loadedFlashcard := Flashcard{
			Term:       record[0],
			Definition: record[1],
			Mistakes:   mistakes,
		}
		fc.CreateOrUpdate(loadedFlashcard)
	}

	return len(records), nil
}

func (fc *Flashcards) ResetStats() {
	for i, flashcard := range fc.elements {
		flashcard.Mistakes = 0
		fc.elements[i] = flashcard
	}
}

func (fc *Flashcards) HardestCards() []Flashcard {
	maxMistakes := 0
	for _, flashcard := range fc.elements {
		maxMistakes = max(maxMistakes, flashcard.Mistakes)
	}

	var hardestCards []Flashcard

	if maxMistakes == 0 {
		return hardestCards
	}

	for _, flashcard := range fc.elements {
		if flashcard.Mistakes == maxMistakes {
			hardestCards = append(hardestCards, flashcard)
		}
	}

	return hardestCards
}

func (fc *Flashcards) IncrementMistakes(term string) {
	for i, flashcard := range fc.elements {
		if flashcard.Term == term {
			flashcard.Mistakes += 1
			fc.elements[i] = flashcard
		}
	}
}

func inputUniqueString(ls LoggingScanner, lp LoggingPrinter, checkExists func(string) bool, msgTmp string) string {
	ls.Scan()
	s := ls.Text()
	for exists := checkExists(s); exists; exists = checkExists(s) {
		lp.Printf(msgTmp, s)
		ls.Scan()
		s = ls.Text()
	}
	return s
}

func addFlashcard(ls LoggingScanner, lp LoggingPrinter, fc Flashcards) {
	lp.Println("The card:")
	term := inputUniqueString(ls, lp, func(s string) bool {
		_, exists := fc.FindDefinitionByTerm(s)
		return exists
	}, "The card \"%s\" already exists. Try again:\n")

	lp.Println("The definition of the card:")
	definition := inputUniqueString(ls, lp, func(s string) bool {
		_, exists := fc.FindTermByDefinition(s)
		return exists
	}, "The definition \"%s\" already exists. Try again:\n")

	newFlashcard := Flashcard{
		Term:       term,
		Definition: definition,
		Mistakes:   0,
	}
	fc.CreateOrUpdate(newFlashcard)
	lp.Printf("The pair (\"%s\":\"%s\") has been added.\n", term, definition)
}

func removeFlashcard(ls LoggingScanner, lp LoggingPrinter, fc Flashcards) {
	lp.Println("Which card?")
	ls.Scan()
	term := ls.Text()
	if _, exists := fc.FindDefinitionByTerm(term); exists {
		fc.RemoveByTerm(term)
		lp.Println("The card has been removed.")
	} else {
		lp.Printf("Can't remove \"%s\": there is no such card.\n", term)
	}
}

func askFlashcards(ls LoggingScanner, lp LoggingPrinter, fc Flashcards) {
	lp.Println("How many times to ask?")
	ls.Scan()
	times, err := strconv.Atoi(ls.Text())
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < times; i++ {
		flashcard := fc.GetRandomFc()
		lp.Printf("Print the definition of \"%s\":\n", flashcard.Term)
		ls.Scan()
		inputDefinition := ls.Text()
		if flashcard.Definition == inputDefinition {
			lp.Println("Correct!")
			continue
		}
		if otherTerm, exists := fc.FindTermByDefinition(inputDefinition); exists {
			fc.IncrementMistakes(flashcard.Term)
			lp.Printf("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\"\n", flashcard.Definition, otherTerm)
			continue
		}
		fc.IncrementMistakes(flashcard.Term)
		lp.Printf("Wrong. The right answer is \"%s\".\n", flashcard.Definition)
	}
}

func importFlashcards(ls LoggingScanner, lp LoggingPrinter, fc Flashcards) {
	lp.Println("File name:")
	ls.Scan()
	filename := ls.Text()
	importFlashcardsFromFile(filename, lp, fc)
}

func importFlashcardsFromFile(filename string, lp LoggingPrinter, fc Flashcards) {
	loadedAmount, err := fc.ReadCSV(filename)
	if err != nil {
		lp.Println("File not found.")
		return
	}
	lp.Printf("%d cards have been loaded.\n", loadedAmount)
}

func exportFlashcards(ls LoggingScanner, lp LoggingPrinter, fc Flashcards) {
	lp.Println("File name:")
	ls.Scan()
	filename := ls.Text()
	savedAmount := fc.WriteCSV(filename)
	lp.Printf("%d cards have been saved.\n", savedAmount)
}

func checkHardestCards(lp LoggingPrinter, fc Flashcards) {
	hardestCards := fc.HardestCards()
	switch len(hardestCards) {
	case 0:
		lp.Println("There are no cards with errors.")
	case 1:
		lp.Printf("The hardest card is \"%s\". You have %d errors answering it.", hardestCards[0].Term, hardestCards[0].Mistakes)
	default:
		sb := strings.Builder{}
		for i, hardCard := range hardestCards {
			sb.Write([]byte("\""))
			sb.Write([]byte(hardCard.Term))
			sb.Write([]byte("\""))
			if i != len(hardestCards)-1 {
				sb.Write([]byte(", "))
			}
		}
		lp.Printf("The hardest cards are %s. You have %d errors answering them.", sb.String(), hardestCards[0].Mistakes)
	}
}

func resetStats(lp LoggingPrinter, fc Flashcards) {
	fc.ResetStats()
	lp.Println("Card statistics have been reset.")
}

func dumpLogs(ls LoggingScanner, lp LoggingPrinter, logBuilder *strings.Builder) {
	lp.Println("File name:")
	ls.Scan()
	filename := ls.Text()
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	successLog := "The log has been saved."
	lp.Println(successLog)
	_, err = file.WriteString(logBuilder.String())
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flashcards := Flashcards{elements: make(map[int]Flashcard)}
	scanner := bufio.NewScanner(os.Stdin)
	logBuilder := &strings.Builder{}
	ls := LoggingScanner{scanner: scanner, logBuilder: logBuilder}
	lp := LoggingPrinter{logBuilder: logBuilder}

	var importFilename, exportFilename string
	flag.StringVar(&importFilename, "import_from", "", "file to import from")
	flag.StringVar(&exportFilename, "export_to", "", "file to export to")
	flag.Parse()

	if importFilename != "" {
		importFlashcardsFromFile(importFilename, lp, flashcards)
	}

	action := ""
	for action != "exit" {
		lp.Println("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):")
		scanner.Scan()
		action = scanner.Text()

		switch action {
		case "exit":
			break
		case "add":
			addFlashcard(ls, lp, flashcards)
		case "remove":
			removeFlashcard(ls, lp, flashcards)
		case "ask":
			askFlashcards(ls, lp, flashcards)
		case "import":
			importFlashcards(ls, lp, flashcards)
		case "export":
			exportFlashcards(ls, lp, flashcards)
		case "log":
			dumpLogs(ls, lp, logBuilder)
		case "hardest card":
			checkHardestCards(lp, flashcards)
		case "reset stats":
			resetStats(lp, flashcards)
		default:
			lp.Println("Unknown command!")
		}

		lp.Println()
	}

	if exportFilename != "" {
		savedAmount := flashcards.WriteCSV(exportFilename)
		lp.Printf("%d cards have been saved.\n", savedAmount)
	}
	lp.Println("Bye bye!")
}
