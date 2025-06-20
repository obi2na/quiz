package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var file string // flag -f for file
	flag.StringVar(&file, "f", "problems.csv", "file location of quiz")

	var timeLimit int
	flag.IntVar(&timeLimit, "t", 30, "time limit for the quiz")

	//parse the flag
	flag.Parse()

	// validate correct file is used
	if err := validatePath(file); err != nil {
		log.Fatal(err)
	}

	problems := NewProblemSheet()
	err := problems.populateSheet(file)
	if err != nil {
		log.Fatal(err)
	}

	problems.StartQuiz(time.Duration(timeLimit)*time.Second, os.Stdin, os.Stdout)
	problems.DisplayResults(os.Stdout)
}

func validatePath(path string) error {
	// validate extension is correct
	if filepath.Ext(path) != ".csv" {
		return fmt.Errorf("malformed file path: %s contains no csv extension", path)
	}

	//check that file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %w", err)
	} else if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}

	return nil
}

type Problem struct {
	Question string
	Answer   string
}

type ProblemSheet struct {
	Questions         []Problem
	CorrectlyAnswered int
}

func (p *ProblemSheet) populateSheet(file string) error {
	//open file
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	lines, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("failed to read csv: %w", err)
	}

	if len(lines) == 0 {
		return fmt.Errorf("empty csv file loaded")
	}

	for _, line := range lines {

		newProblem := Problem{
			Question: strings.TrimSpace(line[0]),
			Answer:   strings.TrimSpace(line[1]),
		}

		p.Questions = append(p.Questions, newProblem)
	}
	return nil
}

func NewProblemSheet() *ProblemSheet {
	return &ProblemSheet{}
}

func (p *ProblemSheet) StartQuiz(limit time.Duration, r io.Reader, w io.Writer) {

	reader := bufio.NewReader(r)

	fmt.Fprintln(w, "Press Enter to begin quiz")
	reader.ReadString('\n')

	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()

	for _, problem := range p.Questions {
		answerCh := make(chan string, 1)

		go func(q string) {
			fmt.Fprintf(w, "\n\nQuestion: %s\n", q)
			fmt.Fprint(w, "answer: ")
			userAns, _ := reader.ReadString('\n')
			answerCh <- strings.TrimSpace(userAns)
		}(problem.Question)

		select {
		case userAnswer := <-answerCh:
			if userAnswer == problem.Answer {
				p.CorrectlyAnswered++
			}
		case <-ctx.Done():
			fmt.Fprintln(w, "\nquiz not completed before timeout")
			return
		}
	}
}

func (p *ProblemSheet) DisplayResults(w io.Writer) {
	fmt.Fprintln(w, "Quiz has ended")
	fmt.Fprintf(w, "You answered %d/%d questions\n", p.CorrectlyAnswered, len(p.Questions))
}

//type workerFunc func() string
//
//func timeLimit(worker workerFunc, limit time.Duration) (string, error) {
//	out := make(chan string, 1)
//	ctx, cancel := context.WithTimeout(context.Background(), limit)
//	defer cancel()
//
//	go func() {
//		fmt.Println("goroutine started")
//		out <- worker()
//	}()
//
//	select {
//	case result := <-out:
//		return result, nil
//	case <-ctx.Done():
//		return "quiz not completed before timeout", errors.New("work timed out")
//	}
//}

//func (p *problemSheet) StartQuiz() string {
//	for _, problem := range p.questions {
//		fmt.Printf("\n\nQuestion: %s\n", problem.question)
//		fmt.Print("answer: ")
//		var answer string
//		fmt.Scanln(&answer)
//
//		if answer == problem.answer {
//			p.correctlyAnswered++
//		}
//	}
//
//	return "Quiz completed before timeout"
//}
