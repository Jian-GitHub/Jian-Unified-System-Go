package files

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"testing"
)

func TestContractorWorkbookCopy(t *testing.T) {
	b, e := os.ReadFile("../../../../docs/Jian-Auckland Contractor Installation .xlsx")
	if os.IsNotExist(e) {
		t.Skip("local user workbook is not distributed")
	}
	if e != nil {
		t.Fatal(e)
	}
	g, e := ParseBytes(context.Background(), b, ".xlsx")
	if e != nil {
		t.Fatal(e)
	}
	rows := g.Sheets["Sheet1"]
	if len(rows) < 27 || rows[0][2] != "Job description" || rows[0][3] != "Job description" || rows[0][7] != "Net Wage" || rows[0][8] != "Payment Status" {
		t.Fatal("unexpected worksheet structure")
	}
}

func workbook(cell string, extra map[string]string) []byte {
	var b bytes.Buffer
	z := zip.NewWriter(&b)
	files := map[string]string{"[Content_Types].xml": "<Types/>", "xl/workbook.xml": `<workbook xmlns:r="urn:r"><workbookPr date1904="true"/><sheets><sheet name="Data" r:id="rId1"/></sheets></workbook>`, "xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId1" Target="worksheets/sheet1.xml"/></Relationships>`, "xl/worksheets/sheet1.xml": `<worksheet><sheetData><row r="1">` + cell + `</row></sheetData></worksheet>`}
	for k, v := range extra {
		files[k] = v
	}
	for k, v := range files {
		w, _ := z.Create(k)
		w.Write([]byte(v))
	}
	z.Close()
	return b.Bytes()
}
func TestCSVAndXLSXReadOnly(t *testing.T) {
	ctx := context.Background()
	g, e := ParseBytes(ctx, []byte("\xef\xbb\xbfDate,Gross\r\n2026-08-01,10.50\r\n"), ".csv")
	if e != nil || g.Sheets["CSV"][1][1] != "10.50" {
		t.Fatal(g, e)
	}
	g, e = ParseBytes(ctx, workbook(`<c r="A1" t="inlineStr"><is><t>Date</t></is></c><c r="B1"><f>1+1</f><v>2</v></c>`, nil), ".xlsx")
	if e != nil || !g.Date1904 || g.Sheets["Data"][0][1] != "2" {
		t.Fatal(g, e)
	}
}
func TestHostileOrUnsupportedWorkbooks(t *testing.T) {
	ctx := context.Background()
	for _, b := range [][]byte{workbook(`<c r="A1"><f>1+1</f></c>`, nil), workbook(`<c r="CW1"><v>1</v></c>`, nil), workbook(`<c r="A1"><v>1</v></c>`, map[string]string{"xl/_rels/workbook.xml.rels": `<Relationships><Relationship Id="rId1" Target="http://127.0.0.1" TargetMode="External"/></Relationships>`}), workbook(`<c r="A1"><v>1</v></c>`, map[string]string{"../escape": "bad"})} {
		if _, e := ParseBytes(ctx, b, ".xlsx"); e == nil {
			t.Fatal("unsafe workbook accepted")
		}
	}
	if _, e := ParseBytes(ctx, []byte{0xff, 0x00}, ".csv"); e == nil {
		t.Fatal("binary CSV accepted")
	}
}
