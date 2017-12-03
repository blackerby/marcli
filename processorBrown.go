package main

import (
	"fmt"
	"strings"
)

type ProcessorBrown struct {
	Filters     FieldFilters
	SearchValue string
}

type BrownRecord struct {
	Bib   string
	Title string
	Items []BrownItem
}

type BrownItem struct {
	Callnumber string
	Barcode    string
}

func NewBrownRecord(r Record) BrownRecord {
	b := BrownRecord{}
	b.Bib = bib(r)
	b.Title = pad(r.Fields.GetValue("245", "a"))
	b.Items = items(r)
	return b
}

func (p ProcessorBrown) Header() {
	header := ""
	if len(p.Filters.Fields) == 0 {
		header = "bib\ttitle\tcallnumber\tbarcode"
	} else {
		header = p.outputString("bib", "title", "callnumber", "barcode")
	}
	fmt.Printf("%s\r\n", header)
}

func (p ProcessorBrown) Footer() {
}

func (p ProcessorBrown) ProcessRecord(f *MarcFile, r Record) {
	b := NewBrownRecord(r)
	if len(b.Items) == 0 {
		// fmt.Printf("%s\t%s\t%s\r\n", b.Bib, b.Title, "--")
	} else {
		for _, item := range b.Items {
			output := p.outputString(b.Bib, b.Title, item.Callnumber, item.Barcode)
			fmt.Printf("%s\r\n", output)
		}
	}
}

func (p ProcessorBrown) Separator() {
}

func notEmpty(str string) string {
	if len(str) == 0 {
		return "-"
	}
	return str
}

func (p ProcessorBrown) outputString(bib, title, callnumber, barcode string) string {
	output := ""
	allFields := len(p.Filters.Fields) == 0

	if allFields || p.Filters.IncludeField("bib") {
		output = notEmpty(bib)
	}
	if allFields || p.Filters.IncludeField("tit") {
		output = concatTab(output, pad(notEmpty(title)))
	}
	if allFields || p.Filters.IncludeField("cal") {
		output = concatTab(output, notEmpty(callnumber))
	}
	if allFields || p.Filters.IncludeField("bar") {
		output = concatTab(output, notEmpty(barcode))
	}
	return output
}

func bib(r Record) string {
	bib := r.Fields.GetValue("907", "a")
	if bib != "" {
		bib = bib[1:(len(bib) - 1)]
	}
	return bib
}

func baseCallNumber(r Record) (bool, Field) {
	// 090 ab            LC CALL NO(c)
	if found, field := r.Fields.GetOne("090"); found {
		return true, field
	}

	// 091 ab            HARRIS CALL NO(e)
	if found, field := r.Fields.GetOne("091"); found {
		return true, field
	}

	// 092 ab            JCB CALL NO(f)
	if found, field := r.Fields.GetOne("092"); found {
		return true, field
	}

	// 096 ab           SUDOCS CALL NO(v)
	if found, field := r.Fields.GetOne("096"); found {
		return true, field
	}

	// 099 ab            OTHER BROWN CALL (l)
	if found, field := r.Fields.GetOne("099"); found {
		return true, field
	}

	return false, Field{}
}

func barcode(f Field) string {
	barcode := f.SubFieldValue("i")
	barcode = removeSpaces(barcode)
	if barcode == "" {
		return "N/A"
	}
	return barcode
}

func items(r Record) []BrownItem {
	var items []BrownItem

	marcItems := r.Fields.Get("945")
	if len(marcItems) == 0 {
		return items
	}

	// Base call number from the 09X field
	found, f_090 := baseCallNumber(r)
	if !found {
		return items
	}

	f_090a := f_090.SubFieldValue("a")
	f_090b := f_090.SubFieldValue("b")
	f_090f := f_090.SubFieldValue("f") // 1-SIZE
	// get the call numbers from the items
	for _, f_945 := range marcItems {
		barcode := barcode(f_945)
		base := concat3(f_090f, f_090a, f_090b)
		f_945a := f_945.SubFieldValue("a")
		f_945b := f_945.SubFieldValue("b")
		if f_945a != "" {
			// use the values in the item record
			base = concat3(f_090f, f_945a, f_945b)
		}
		volume := f_945.SubFieldValue("c")
		copy := f_945.SubFieldValue("g")
		if copy == "1" {
			copy = ""
		} else if copy > "1" {
			copy = "c. " + copy
		}
		number := concat3(base, volume, copy)
		if strings.HasSuffix(number, "\\") {
			number = number[0 : len(number)-1]
		}
		item := BrownItem{Callnumber: number, Barcode: barcode}
		items = append(items, item)
	}
	return items
}
