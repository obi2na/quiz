package main

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"os"
	"strings"
	"testing"
	"time"
)

func TestValidatePath(t *testing.T) {
	// Setup a valid temp .csv file
	validFile, err := os.CreateTemp("", "*.csv")
	if err != nil {
		t.Fatalf("could not create temp file: %v", err)
	}
	defer os.Remove(validFile.Name())

	tests := []struct {
		name        string
		path        string
		expectError bool
		expectedErr string // optional: substring or exact message
	}{
		{
			name:        "valid csv file",
			path:        validFile.Name(),
			expectError: false,
		},
		{
			name:        "invalid file extension",
			path:        "bad.txt",
			expectError: true,
			expectedErr: "malformed file path: bad.txt contains no csv extension",
		},
		{
			name:        "non-existent file",
			path:        "does-not-exist.csv",
			expectError: true,
			expectedErr: "path does not exist",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePath(tc.path)

			if tc.expectError {
				fmt.Println("actual error:", err.Error())
				fmt.Println("expected substring:", tc.expectedErr)
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				// if error does not contain below text this is triggered
				if tc.expectedErr != "" && !strings.Contains(err.Error(), tc.expectedErr) {
					t.Errorf("expected error containing %q, got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPopulateSheet(t *testing.T) {

	tests := []struct {
		name           string
		path           string
		expectError    bool
		expectedError  string
		expectedResult *ProblemSheet
	}{
		{
			name:           "populateSheet with well formed csv",
			path:           "testdata/testdata.csv",
			expectError:    false,
			expectedResult: validProblemSheet(),
		},
		{
			name:          "populateSheet with malformed csv",
			path:          "testdata/testmalformeddata.csv",
			expectError:   true,
			expectedError: "wrong number of fields",
		},
		{
			name:          "empty CSV file",
			path:          "testdata/emptytestdata.csv",
			expectError:   true,
			expectedError: "empty csv file loaded",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.path
			sheet := NewProblemSheet()
			err := sheet.populateSheet(path)

			if tc.expectError {

				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				//run if this if does not contain right error
				if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expected error containing %q, got: %v", tc.expectedError, err)
				}

			} else {
				if diff := cmp.Diff(tc.expectedResult, sheet); diff != "" {
					t.Errorf("populateSheet() mismatch (-want +got):\n%s", diff)
				}
			}

		})
	}

}

func validProblemSheet() *ProblemSheet {
	return &ProblemSheet{
		Questions: []Problem{
			{Question: "2+2", Answer: "4"},
			{Question: "5+5", Answer: "10"},
			{Question: "3+1", Answer: "4"},
		},
	}
}

func TestStartQuiz(t *testing.T) {

	tests := []struct {
		name           string
		answers        string
		timelimit      time.Duration
		expectTimeout  bool
		expectedResult int
		expectedOutput string
	}{
		{
			name:           "test all correct answers",
			answers:        "\n4\n10\n4",
			timelimit:      10 * time.Second,
			expectTimeout:  false,
			expectedResult: 3,
		},
		{
			name:           "test one wrong answer",
			answers:        "\n4\n5\n4",
			timelimit:      10 * time.Second,
			expectTimeout:  false,
			expectedResult: 2,
		},
		{
			name:           "test timeout feature",
			answers:        "4\n10\n4",
			expectedResult: 0,
			timelimit:      10 * time.Nanosecond, // set to just 10 to force timeout
			expectTimeout:  true,
			expectedOutput: "quiz not completed before timeout",
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			sheet := validProblemSheet()

			input := strings.NewReader(tc.answers)
			var output strings.Builder

			sheet.StartQuiz(tc.timelimit, input, &output)

			if tc.expectTimeout {
				if !strings.Contains(output.String(), tc.expectedOutput) {
					t.Errorf("expected timeout message, got: %s", output.String())
				}
			}

			if tc.expectedResult != sheet.CorrectlyAnswered {
				t.Errorf("expected %d correct answers, got %d", tc.expectedResult, sheet.CorrectlyAnswered)
			}

		})
	}
}

func TestDisplayResults(t *testing.T) {
	tests := []struct {
		name           string
		sheet          *ProblemSheet
		expectedOutput string
	}{
		{
			name: "all correct",
			sheet: &ProblemSheet{
				Questions: []Problem{
					{Question: "2+2", Answer: "4"},
					{Question: "5+5", Answer: "10"},
				},
				CorrectlyAnswered: 2,
			},
			expectedOutput: "Quiz has ended\nYou answered 2/2 questions\n",
		},
		{
			name: "some correct",
			sheet: &ProblemSheet{
				Questions: []Problem{
					{Question: "2+2", Answer: "4"},
					{Question: "5+5", Answer: "10"},
				},
				CorrectlyAnswered: 1,
			},
			expectedOutput: "Quiz has ended\nYou answered 1/2 questions\n",
		},
		{
			name: "no questions",
			sheet: &ProblemSheet{
				Questions:         []Problem{},
				CorrectlyAnswered: 0,
			},
			expectedOutput: "Quiz has ended\nYou answered 0/0 questions\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var output strings.Builder

			tc.sheet.DisplayResults(&output)

			actual := output.String()
			if actual != tc.expectedOutput {
				t.Errorf("unexpected output:\nGot:\n%q\nWant:\n%q", actual, tc.expectedOutput)
			}
		})
	}
}
