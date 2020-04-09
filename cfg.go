package cfg

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const tagName = "cfg"

type decodeState struct {
	line    int
	scanner *bufio.Scanner
}

func (d *decodeState) init(data []byte) {
	d.scanner = bufio.NewScanner(bytes.NewReader(data))
}

type field struct {
	Tag     string
	IsArray bool
	IsInner bool
	Kind    reflect.Kind
	Value   reflect.Value
}

func extractFields(rv reflect.Value) []field {
	var fields []field
	v := rv

	if rv.Kind() == reflect.Ptr {
		v = rv.Elem()
	}

	for i := 0; i < v.Type().NumField(); i++ {
		fieldType := v.Type().Field(i)
		fieldValue := v.Field(i)
		fieldTag := fieldType.Tag.Get(tagName)
		fieldKind := fieldValue.Type().Kind()

		if len(fieldTag) == 0 {
			fieldTag = fieldType.Name
		}

		fieldTag = strings.TrimSpace(fieldTag)

		fields = append(fields, field{
			Tag:     fieldTag,
			IsArray: fieldKind == reflect.Slice,
			IsInner: fieldKind == reflect.Struct,
			Kind:    fieldKind,
			Value:   fieldValue,
		})
	}

	return fields
}

func (d *decodeState) readNextValidLine() (string, error) {
	for d.scanner.Scan() {
		d.line++

		line := d.scanner.Text()

		// comment line or empty line
		if len(line) < 1 || line[0] == '#' {
			continue
		}

		// remove comment from end of line
		commentPos := strings.Index(line, "#")
		if commentPos >= 0 {
			line = strings.TrimSpace(line[0:commentPos])
		}

		return line, nil
	}

	return "", errors.New(fmt.Sprintf("could not read line %d, end of file", d.line))
}

func (d *decodeState) unmarshal(v interface{}) error {
	rv := reflect.ValueOf(v)

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("decode target should be a pointer to a struct, got %s", reflect.TypeOf(v))
	}

	if rv.IsNil() {
		return fmt.Errorf("decode target is nil, %s", reflect.TypeOf(v))
	}

	if rv.Elem().Type().Kind() != reflect.Struct {
		return fmt.Errorf("decode target should point to a struct, got %s", reflect.TypeOf(rv.Elem()))
	}

	fields := extractFields(rv)

	for d.scanner.Scan() {
		d.line++
		line := strings.TrimSpace(d.scanner.Text())

		// remove comment line or empty line
		if len(line) < 1 || line[0] == '#' {
			continue
		}

		valueSeparatorPos := strings.Index(line, ":")

		// we have a key value pair (this can be easy)
		if valueSeparatorPos >= 0 {
			matches := strings.SplitN(line, ":", 2)

			// handle the case where key is the only thing on the line
			if len(matches) < 2 || len(matches[1]) == 0 {
				nextLine, err := d.readNextValidLine()
				if err != nil {
					return err
				}

				if strings.Contains(nextLine, ":") {
					return errors.New(fmt.Sprintf("could not decode line %d, expected value", d.line))
				}

				if len(matches) == 2 {
					matches[1] = nextLine
				} else {
					matches = append(matches, nextLine)
				}
			}

			// we have the key and value
			if len(matches) == 2 {
				key := strings.TrimSpace(matches[0])
				value := strings.TrimSpace(matches[1])

				// remove comments
				commentPos := strings.Index(value, "#")
				if commentPos >= 0 {
					value = strings.TrimSpace(value[0:commentPos])
				}

				startArrayPos := strings.Index(value, "[")
				endArrayPos := strings.Index(value, "]")
				startInnerPos := strings.Index(value, "{")
				endInnerPos := strings.Index(value, "}")

				var lines []string

				// multi-line array
				if startArrayPos >= 0 && endArrayPos < 0 {
					numberOfBrackets := 1

					for {
						line, err := d.readNextValidLine()
						if err != nil {
							return err
						}

						line = strings.TrimSpace(line)
						line = removeCommaFromEnd(line)

						if strings.Contains(line, "[") {
							numberOfBrackets++
						}

						endArrayPos := strings.Index(line, "]")
						if endArrayPos >= 0 {
							numberOfBrackets--
							line = line[0:endArrayPos]
						}

						if len(line) > 0 {
							lines = append(lines, line)
						}

						if numberOfBrackets == 0 {
							break
						}
					}
				}

				// one-line array
				if startArrayPos >= 0 && endArrayPos >= 0 {
					value = strings.TrimSpace(value[startArrayPos+1 : endArrayPos])
					lines = strings.Split(value, " ")

					// maybe lines without space?
					if len(lines) == 1 {
						lines = strings.Split(value, ",")
					}

					for i := range lines {
						lines[i] = removeCommaFromEnd(lines[i])
					}
				}

				var innerData string

				// multi-line inner
				if startInnerPos >= 0 && endInnerPos < 0 {
					numberOfCurlyBraces := 1

					for {
						line, err := d.readNextValidLine()
						if err != nil {
							return err
						}

						if strings.Contains(line, "{") {
							numberOfCurlyBraces++
						}

						endInnerPos := strings.Index(line, "}")
						if endInnerPos >= 0 {
							numberOfCurlyBraces--
							line = line[0:endInnerPos]
						}

						if len(line) > 0 {
							innerData += line + "\n"
						}

						if numberOfCurlyBraces == 0 {
							break
						}
					}
				}

				// one-line inner
				if startInnerPos >= 0 && endInnerPos >= 0 {
					innerData = strings.TrimSpace(value[startArrayPos+1 : endArrayPos])
				}

				isArray := len(lines) > 0
				isInner := len(innerData) > 0

				for _, f := range fields {
					if f.Tag == "-" {
						continue
					}

					if f.Tag == key {

						if isArray && !f.IsArray {
							return errors.New(fmt.Sprintf("could not parse array into non-slice type on line %d", d.line))
						}

						if isArray && f.IsArray {
							for _, line := range lines {
								err := setSliceValue(f, unwrapQuotationMarks(line))
								if err != nil {
									return errors.Wrap(err, fmt.Sprintf("could not parse line %d", d.line))
								}
							}
							continue
						}

						if isInner && !f.IsInner {
							return errors.New(fmt.Sprintf("could not parse inner struct into non-struct type on line %d", d.line))
						}

						if isInner && f.IsInner {
							err := Unmarshal([]byte(innerData), f.Value.Addr().Interface())
							if err != nil {
								return errors.Wrap(err, fmt.Sprintf("could not parse inner struct"))
							}
							continue
						}

						// simple value
						err := setValue(f, unwrapQuotationMarks(removeCommaFromEnd(value)))
						if err != nil {
							return errors.Wrap(err, fmt.Sprintf("could not parse line %d", d.line))
						}
					}
				}
			} else {
				return errors.New(fmt.Sprintf("could not parse line %d", d.line))
			}
		} else {
			return errors.New(fmt.Sprintf("could not decode line %d", d.line))
		}
	}

	return nil
}

// Unmarshal parse the data provided an try to populate the struct pointer
func Unmarshal(data []byte, v interface{}) error {
	var d decodeState
	d.init(data)
	return d.unmarshal(v)
}

// Marshal returns the CFG encoding of v
func Marshal(v interface{}) ([]byte, error) {
	var output bytes.Buffer

	rv := reflect.ValueOf(v)

	if rv.Type().Kind() != reflect.Struct {
		return output.Bytes(), fmt.Errorf("encode target should be a struct, got %s", reflect.TypeOf(rv.Elem()))
	}

	fields := extractFields(rv)

	var lines []string

	for _, field := range fields {
		if field.Tag == "-" {
			continue
		}

		if field.IsArray {
			var arr []string

			for i := 0; i < field.Value.Len(); i++ {
				arr = append(arr, fmt.Sprintf("  '%v'", field.Value.Index(i)))
			}

			lines = append(lines, fmt.Sprintf("%s: [\n%s\n]", field.Tag, strings.Join(arr, ",\n")))
			continue
		}

		if field.IsInner {
			out, err := Marshal(field.Value.Interface())
			if err != nil {
				return output.Bytes(), errors.Wrap(err, "could not marshal inner struct")
			}

			lines = append(lines, fmt.Sprintf("%s: {\n  %s\n}", field.Tag, strings.ReplaceAll(string(out), "\n", "\n  ")))
			continue
		}

		line, err := marshalField(field)
		if err != nil {
			return output.Bytes(), errors.Wrap(err, "could not marshal field")
		}

		lines = append(lines, line)
	}

	output.WriteString(strings.Join(lines, ",\n"))
	return output.Bytes(), nil
}

func marshalField(field field) (string, error) {
	if field.Kind == reflect.String {
		return fmt.Sprintf("%s: '%s'", field.Tag, field.Value.String()), nil
	}

	if field.Kind == reflect.Int || field.Kind == reflect.Int8 ||
		field.Kind == reflect.Int16 || field.Kind == reflect.Int32 ||
		field.Kind == reflect.Int64 {
		return fmt.Sprintf("%s: %d", field.Tag, field.Value.Int()), nil
	}

	if field.Kind == reflect.Uint ||
		field.Kind == reflect.Uint8 || field.Kind == reflect.Uint16 ||
		field.Kind == reflect.Uint32 || field.Kind == reflect.Uint64 {
		return fmt.Sprintf("%s: %d", field.Tag, field.Value.Uint()), nil
	}

	if field.Kind == reflect.Float32 || field.Kind == reflect.Float64 {
		return fmt.Sprintf("%s: %f", field.Tag, field.Value.Float()), nil
	}

	if field.Kind == reflect.Bool {
		value := "false"

		if field.Value.Bool() {
			value = "true"
		}

		return fmt.Sprintf("%s: %s", field.Tag, value), nil
	}

	return "", errors.New(fmt.Sprintf("could not encode %v to string", field.Kind))
}

func setValue(f field, value string) error {
	if f.IsArray {
		panic("wrong implementation, use setSliceValue")
	}

	if f.Kind == reflect.String {
		f.Value.SetString(value)
		return nil
	}

	if f.Kind == reflect.Bool {
		f.Value.SetBool(boolValue(value))
		return nil
	}

	if f.Kind == reflect.Int || f.Kind == reflect.Int8 ||
		f.Kind == reflect.Int16 || f.Kind == reflect.Int32 ||
		f.Kind == reflect.Int64 {
		n, err := intValue(value)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not convert %q to %q", value, f.Kind))
		}

		f.Value.SetInt(n)
		return nil
	}

	if f.Kind == reflect.Uint || f.Kind == reflect.Uint8 ||
		f.Kind == reflect.Uint16 || f.Kind == reflect.Uint32 ||
		f.Kind == reflect.Uint64 {
		n, err := uintValue(value)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not convert %q to %q", value, f.Kind))
		}

		f.Value.SetUint(n)
		return nil
	}

	if f.Kind == reflect.Float32 || f.Kind == reflect.Float64 {
		n, err := floatValue(value, f.Value.Type().Bits())

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not convert %q to %q", value, f.Kind))
		}

		f.Value.SetFloat(n)
		return nil
	}

	return nil
}

func setSliceValue(f field, value string) error {
	if !f.IsArray {
		panic("wrong implementation, use setValue")
	}

	kind := f.Value.Type().Elem().Kind()

	if kind == reflect.String {
		f.Value.Set(reflect.Append(f.Value, reflect.ValueOf(value)))
		return nil
	}

	if kind == reflect.Bool {
		f.Value.Set(reflect.Append(f.Value, reflect.ValueOf(boolValue(value))))
		return nil
	}

	if kind == reflect.Int || kind == reflect.Int8 ||
		kind == reflect.Int16 || kind == reflect.Int32 ||
		kind == reflect.Int64 {
		n, err := intValue(value)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not convert %q to %q", value, kind))
		}

		f.Value.Set(reflect.Append(f.Value, reflect.ValueOf(n)))
		return nil
	}

	if kind == reflect.Uint || kind == reflect.Uint8 ||
		kind == reflect.Uint16 || kind == reflect.Uint32 ||
		kind == reflect.Uint64 {
		n, err := uintValue(value)

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not convert %q to %q", value, kind))
		}

		f.Value.Set(reflect.Append(f.Value, reflect.ValueOf(n)))
		return nil
	}

	if kind == reflect.Float32 || kind == reflect.Float64 {
		n, err := floatValue(value, f.Value.Type().Bits())

		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not convert %q to %q", value, kind))
		}

		f.Value.Set(reflect.Append(f.Value, reflect.ValueOf(n)))
		return nil
	}

	return errors.New(fmt.Sprintf("invalid type %s", kind))
}

func intValue(s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, nil
	}

	return i, nil
}

func uintValue(s string) (uint64, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, nil
	}

	return i, nil
}

func floatValue(s string, bits int) (float64, error) {
	i, err := strconv.ParseFloat(s, bits)
	if err != nil {
		return 0.0, nil
	}

	return i, nil
}

func boolValue(s string) bool {
	v := false

	switch strings.ToLower(s) {
	case "t", "true", "y", "yes", "1":
		v = true
	}

	return v
}

func unwrapQuotationMarks(input string) string {
	last := len(input) - 1

	if (input[0] == '\'' && input[last] == '\'') ||
		(input[0] == '"' && input[last] == '"') {
		return input[1:last]
	}

	return input
}

func removeCommaFromEnd(input string) string {
	if input[len(input)-1] == ',' {
		input = input[0 : len(input)-1]
	}
	return input
}
