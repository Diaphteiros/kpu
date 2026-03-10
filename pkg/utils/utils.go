package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Fatal prints the given error message to stderr and then exits with the given exit code.
func Fatal(code int, format string, values ...any) {
	fmt.Fprintf(os.Stderr, format, values...)
	os.Exit(code)
}

// PromptForConfirmation prompts the user for confirmation.
func PromptForConfirmation(prompt string, newline bool) bool {
	lastRune := prompt[len(prompt)-1]
	if lastRune != ' ' && lastRune != '\n' {
		prompt += " "
	}
	prompt += "Confirm with 'y' or 'yes' (case-insensitive): "
	if newline {
		prompt += "\n"
	}
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while reading user input")
		return false
	}
	input = strings.TrimSuffix(input, "\n")
	input = strings.ToLower(input)
	return input == "y" || input == "yes"
}

// PromptForInput asks the user for information and returns it.
func PromptForInput(prompt string, newline bool, transform func(string) string) string {
	if newline {
		prompt += "\n"
	}
	fmt.Print(prompt)
	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "error while reading user input")
		return ""
	}
	input = strings.TrimSuffix(input, "\n")
	if transform != nil {
		input = transform(input)
	}
	return input
}

// NaturalLanguageJoin works like strings.Join, but it separates the last element via ' and ' instead of comma.
// For exactly two elements, the result is '<elem> and <elem>'.
func NaturalLanguageJoin(data []string, separator string, oxfordComma bool) string {
	if len(data) == 1 {
		return data[0]
	} else if len(data) == 2 {
		return fmt.Sprintf("%s and %s", data[0], data[1])
	}
	sb := strings.Builder{}
	sb.WriteString(strings.Join(data[:len(data)-1], ", "))
	if oxfordComma {
		sb.WriteString(",")
	}
	sb.WriteString(" and ")
	sb.WriteString(data[len(data)-1])
	return sb.String()
}

// Project takes a list and converts/projects each element of the list to another type using the given project function.
func Project[X any, Y any](data []X, project func(X) Y) []Y {
	res := make([]Y, len(data))
	for i, elem := range data {
		res[i] = project(elem)
	}
	return res
}

// FilterSlice takes a slice and filters it using the given predicate function.
// Does not modify the original slice, but values are not deep-copied, so changes to the values might influence the original slice, depending on the type.
func FilterSlice[X any](data []X, predicate func(X) bool) []X {
	var res []X
	for _, elem := range data {
		if predicate(elem) {
			res = append(res, elem)
		}
	}
	return res
}
