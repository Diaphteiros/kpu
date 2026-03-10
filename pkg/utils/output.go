package utils

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OutputFormat string

const (
	OUTPUT_TEXT OutputFormat = "text"
	OUTPUT_YAML OutputFormat = "yaml"
	OUTPUT_JSON OutputFormat = "json"
)

// Set implements pflag.Value
func (o *OutputFormat) Set(val string) error {
	*o = OutputFormat(val)
	return nil
}

// Type implements pflag.Value
func (o *OutputFormat) Type() string {
	return "string"
}

// String implements pflag.Value
func (o *OutputFormat) String() string {
	return string(*o)
}

var DefaultOutputFormats = []OutputFormat{OUTPUT_JSON, OUTPUT_TEXT, OUTPUT_YAML}

// AddOutputFlag adds a --output/-o flag to the given FlagSet, binding the result to the given string variable.
func AddOutputFlag(fs *pflag.FlagSet, outputVar *OutputFormat, defaultOutput OutputFormat, validOutputs ...OutputFormat) {
	*outputVar = defaultOutput
	if len(validOutputs) == 0 {
		validOutputs = DefaultOutputFormats
	}
	fs.VarP(outputVar, "output", "o", fmt.Sprintf("Output format. Valid formats are [%s].", strings.Join(Project(validOutputs, func(in OutputFormat) string { return string(in) }), ", ")))
}

// ValidateOutputFormat verifies that a valid output format was specified and exits with error otherwise.
func ValidateOutputFormat(output OutputFormat, validOutputs ...OutputFormat) {
	if len(validOutputs) == 0 {
		validOutputs = DefaultOutputFormats
	}
	if !slices.Contains(validOutputs, output) {
		UnknownOutputFatal(output)
	}
}

// UnknownOutputFatal prints an error message that the specified output format is not known and then exits with status 1.
func UnknownOutputFatal(output OutputFormat) {
	Fatal(1, "unknown output format '%s'", string(output))
}

// Table is a helper struct to format output as table.
type Table[T any] struct {
	columns []tableColumn[T]
	data    []T
}

type tableColumn[T any] struct {
	name     string
	getValue func(T) string
	maxLen   int
}

// NewOutputTable returns a new Table for constructing output formatted as a table.
func NewOutputTable[T any]() *Table[T] {
	return &Table[T]{
		columns: []tableColumn[T]{},
		data:    []T{},
	}
}

// WithColumn adds a new column with the given name and getter function to the table.
func (t *Table[T]) WithColumn(name string, getValue func(obj T) string) *Table[T] {
	t.columns = append(t.columns, tableColumn[T]{
		name:     name,
		getValue: getValue,
		maxLen:   len(name),
	})
	return t
}

// WithData adds a one or more new rows to the table.
func (t *Table[T]) WithData(obj ...T) *Table[T] {
	t.data = append(t.data, obj...)
	return t
}

func (t *Table[T]) String() string {
	spacingBetweenColumns := 3
	stringData := make([][]string, len(t.data))
	for i, obj := range t.data {
		stringData[i] = make([]string, len(t.columns))
		for j := range t.columns {
			c := &t.columns[j]
			s := c.getValue(obj)
			slen := len(s)
			if slen > c.maxLen {
				c.maxLen = slen
			}
			stringData[i][j] = s
		}
	}

	sb := strings.Builder{}
	for i, c := range t.columns {
		s := strings.ToUpper(c.name)
		if i < len(t.columns)-1 {
			// if this is not the last column, add padding to format the columns properly
			s = StringPadding(s, c.maxLen+spacingBetweenColumns)
		}
		sb.WriteString(s)
	}
	sb.WriteString("\n")
	for _, row := range stringData {
		for j, cVal := range row {
			if j < len(t.columns)-1 {
				// if this is not the last column, add padding to format the columns properly
				cVal = StringPadding(cVal, t.columns[j].maxLen+spacingBetweenColumns)
			}
			sb.WriteString(cVal)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// StringPadding returns the given string filled with spaces to the given length.
// If the string already has the required or a greater length, it is returned unmodified.
func StringPadding(s string, l int) string {
	diff := l - len(s)
	if diff <= 0 {
		return s
	}
	sb := strings.Builder{}
	sb.WriteString(s)
	for range diff {
		sb.WriteString(" ")
	}
	return sb.String()
}

func GVKToString(gvk schema.GroupVersionKind) string {
	sb := strings.Builder{}
	sb.WriteString(gvk.Kind)
	sb.WriteString(".")
	sb.WriteString(gvk.Version)
	if gvk.Group != "" {
		sb.WriteString(".")
		sb.WriteString(gvk.Group)
	}
	return sb.String()
}

func ResourceIdentifier(obj client.Object) string {
	sb := strings.Builder{}
	sb.WriteString(GVKToString(obj.GetObjectKind().GroupVersionKind()))
	sb.WriteString(" ")
	if obj.GetNamespace() != "" {
		sb.WriteString(obj.GetNamespace())
		sb.WriteString("/")
	}
	sb.WriteString(obj.GetName())
	return sb.String()
}
