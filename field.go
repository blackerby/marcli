package main

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// Field represents a field inside a MARC record. Notice that the
// field could be a "control" field (tag 001-009) or a "data" field
// (any other tag)
//
// For example in:
//		=650  \0$aDiabetes$xComplications$zUnited States.
// Field would be:
// 		Field{
//			Tag: "650",
//			Value: ""
//			Indicator1: " ",
//			Indicator2: "0",
//			SubFields (see SubField definition above)
//	}
type Field struct {
	Tag        string     // for both Control and Data fields
	Value      string     // for Control fields
	Indicator1 string     // for Data fields
	Indicator2 string     // for Data fields
	SubFields  []SubField // for Data fields
}

// SubField contains a Code and a Value.
// For example in:
//		=650  \0$aDiabetes$xComplications$zUnited States.
// an example of SubFieldValue will be:
// 		SubField{
//			Code: "a",
//			Value: "Diabetes"
//		}
type SubField struct {
	Code  string
	Value string
}

// IsControlField returns true if the field is a control field (tag 001-009)
func (f Field) IsControlField() bool {
	return strings.HasPrefix(f.Tag, "00")
}

// Contains returns true if the field contains the passed string.
func (f Field) Contains(str string) bool {
	str = strings.ToLower(str)
	if f.IsControlField() {
		return strings.Contains(strings.ToLower(f.Value), str)
	}

	for _, sub := range f.SubFields {
		if strings.Contains(strings.ToLower(sub.Value), str) {
			return true
		}
	}
	return false
}

// MakeField creates a field objet with the data received.
func MakeField(tag string, data []byte) (Field, error) {
	f := Field{}
	f.Tag = tag

	// It's a control field
	if strings.HasPrefix(tag, "00") {
		f.Value = string(data)
		return f, nil
	}

	if len(data) > 2 {
		f.Indicator1 = string(data[0])
		f.Indicator2 = string(data[1])
	} else {
		return f, errors.New("Invalid Indicators detected")
	}

	for _, sf := range bytes.Split(data[3:], []byte{st}) {
		if len(sf) > 0 {
			f.SubFields = append(f.SubFields, SubField{string(sf[0]), string(sf[1:])})
		} else {
			return f, errors.New("Extraneous field terminator")
		}
	}
	return f, nil
}

func indicatorStr(indicator string) string {
	if indicator != " " && indicator != "" {
		return indicator
	}
	return "\\"
}

func (f Field) String() string {
	if f.IsControlField() {
		return fmt.Sprintf("=%s  %s", f.Tag, f.Value)
	}
	str := fmt.Sprintf("=%s  %s%s", f.Tag, indicatorStr(f.Indicator1), indicatorStr(f.Indicator2))
	for _, sub := range f.SubFields {
		str += fmt.Sprintf("$%s%s", sub.Code, sub.Value)
	}
	return str
}

type Fields struct {
	fields []Field
}

// func (v SubFieldValue) String() string {
// 	return fmt.Sprintf("$%s%s", v.SubField, v.Value)
// }

// func (f Fields) All() []Field {
// 	return f.fields
// }

// func (f *Fields) Add(field Field) {
// 	f.fields = append(f.fields, field)
// }

// func NewField(tag, valueStr string) Field {
// 	value := Field{Tag: tag}
// 	if tag <= "008" {
// 		// Control fields (001-008) don't have indicators or subfields
// 		// so we just get the value as-is.
// 		value.RawValue = valueStr
// 		return value
// 	}

// 	// Process the indicators and subfields
// 	if len(valueStr) >= 2 {
// 		value.Ind1 = string(valueStr[0])
// 		value.Ind2 = string(valueStr[1])
// 	}
// 	if len(valueStr) > 2 {
// 		// notice that we skip the indicators [0] and [1] because they are handled
// 		// above and valueStr[2] because that's a separator character
// 		value.RawValue = valueStr[3:]
// 	}
// 	value.SubFields = NewSubFieldValues(valueStr)
// 	return value
// }

// func NewSubFieldValues(valueStr string) []SubFieldValue {
// 	var values []SubFieldValue
// 	// valueStr comes with the indicators, we skip them:
// 	//   valueStr[0] indicator 1
// 	// 	 valueStr[1] indicator 2
// 	// 	 valueStr[2] separator (ascii 31/0x1f)
// 	separator := 0x1f
// 	tokens := strings.Split(valueStr[3:], string(separator))
// 	for _, token := range tokens {
// 		value := SubFieldValue{
// 			SubField: string(token[0]),
// 			Value:    token[1:],
// 		}
// 		values = append(values, value)
// 	}
// 	return values
// }

// func (f Field) String() string {
// 	ind1 := formatIndicator(f.Ind1)
// 	ind2 := formatIndicator(f.Ind2)
// 	strValue := ""
// 	if len(f.SubFields) > 0 {
// 		// use the subfield values
// 		for _, s := range f.SubFields {
// 			strValue += fmt.Sprintf("$%s%s", s.SubField, s.Value)
// 		}
// 	} else {
// 		// use the raw value
// 		strValue = f.RawValue
// 	}
// 	return fmt.Sprintf("=%s  %s%s%s", f.Tag, ind1, ind2, strValue)
// }

// func (f Field) SubFieldValue(subfield string) string {
// 	for _, s := range f.SubFields {
// 		if s.SubField == subfield {
// 			return s.Value
// 		}
// 	}
// 	return ""
// }

// // For a given value, extract the subfield values in the string
// // indicated. "subfields" is a plain string, like "abu", to
// // indicate subfields a, b, and u.
// func (f Field) SubFieldValues(subfields string) []SubFieldValue {
// 	var values []SubFieldValue
// 	for _, sub := range f.SubFields {
// 		if strings.Contains(subfields, sub.SubField) {
// 			value := SubFieldValue{
// 				SubField: sub.SubField,
// 				Value:    sub.Value,
// 			}
// 			values = append(values, value)
// 		}
// 	}
// 	return values
// }

func formatIndicator(value string) string {
	if value == " " {
		return "\\"
	}
	return value
}

// func (f Fields) Get(tag string) []Field {
// 	var fields []Field
// 	for _, field := range f.fields {
// 		if field.Tag == tag {
// 			fields = append(fields, field)
// 		}
// 	}
// 	return fields
// }

// func (f Fields) GetOne(tag string) (bool, Field) {
// 	fields := f.Get(tag)
// 	if len(fields) == 0 {
// 		return false, Field{}
// 	}
// 	return true, fields[0]
// }

// func (f Fields) GetValue(tag string, subfield string) string {
// 	value := ""
// 	found, field := f.GetOne(tag)
// 	if found {
// 		if subfield == "" {
// 			value = field.RawValue
// 		} else {
// 			value = field.SubFieldValue(subfield)
// 		}
// 	}
// 	return value
// }

// func (f Fields) GetValues(tag string, subfield string) []string {
// 	var values []string
// 	for _, field := range f.Get(tag) {
// 		var value string
// 		if subfield == "" {
// 			value = field.RawValue
// 		} else {
// 			value = field.SubFieldValue(subfield)
// 		}
// 		if value != "" {
// 			values = append(values, value)
// 		}
// 	}
// 	return values
// }
