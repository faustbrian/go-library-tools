package compatibilityconsumer_test

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"reflect"
	"testing"

	tabularv1 "github.com/faustbrian/go-tabular"
	tabularv2 "github.com/faustbrian/go-tabular/v2"
)

func TestPublishedTabularMajorsDecodeXLSXWithPatchedExcelize(t *testing.T) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	for _, part := range []struct{ name, content string }{
		{"[Content_Types].xml", `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/></Types>`},
		{"_rels/.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`},
		{"xl/workbook.xml", `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="Sheet1" sheetId="1" r:id="rId1"/></sheets></workbook>`},
		{"xl/_rels/workbook.xml.rels", `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/></Relationships>`},
		{"xl/worksheets/sheet1.xml", `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData><row r="1"><c r="A1" t="inlineStr"><is><t>Name</t></is></c><c r="B1" t="inlineStr"><is><t>City</t></is></c></row><row r="2"><c r="A2" t="inlineStr"><is><t>Ada</t></is></c><c r="B2" t="inlineStr"><is><t>Helsinki</t></is></c></row></sheetData></worksheet>`},
	} {
		entry, err := archive.Create(part.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(entry, part.content); err != nil {
			t.Fatal(err)
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	data := buffer.Bytes()
	check := func(t *testing.T, read func() ([]string, error), closeReader func() error) {
		t.Helper()
		t.Cleanup(func() {
			if err := closeReader(); err != nil {
				t.Error(err)
			}
		})
		for _, want := range [][]string{{"Name", "City"}, {"Ada", "Helsinki"}} {
			got, err := read()
			if err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("published spreadsheet row = %v, %v; want %v", got, err, want)
			}
		}
		if _, err := read(); !errors.Is(err, io.EOF) {
			t.Fatalf("published spreadsheet end = %v; want EOF", err)
		}
	}
	t.Run("v1", func(t *testing.T) {
		reader, err := tabularv1.OpenSpreadsheet(bytes.NewReader(data), int64(len(data)), tabularv1.SpreadsheetConfig{
			Format: tabularv1.FormatXLSX, FieldsPerRecord: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		check(t, func() ([]string, error) {
			row, err := reader.Read()
			return []string(row), err
		}, reader.Close)
	})
	t.Run("v2", func(t *testing.T) {
		reader, err := tabularv2.OpenSpreadsheet(bytes.NewReader(data), int64(len(data)), tabularv2.SpreadsheetConfig{
			Format: tabularv2.FormatXLSX, FieldsPerRecord: 2,
		})
		if err != nil {
			t.Fatal(err)
		}
		check(t, func() ([]string, error) {
			row, err := reader.Read()
			return []string(row), err
		}, reader.Close)
	})
}
