package transport

import (
	"encoding/json"
	"google.golang.org/protobuf/proto"
	"jian-unified-system/hephaestus/hephaestus-rpc/income"
	"testing"
)

func TestEmptyRowsAfterRPCWireRoundTrip(t *testing.T) {
	wire, err := proto.Marshal(&income.BatchRows{})
	if err != nil {
		t.Fatal(err)
	}
	var rpc income.BatchRows
	if err = proto.Unmarshal(wire, &rpc); err != nil {
		t.Fatal(err)
	}
	type row struct {
		Errors []string `json:"errors"`
		Gross  *int64   `json:"gross"`
	}
	type response struct {
		Items []row `json:"items"`
	}
	got, err := Convert[response](&rpc)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(got)
	if string(b) != `{"items":[]}` {
		t.Fatalf("empty RPC list: %s", b)
	}
	nested, err := Convert[response](map[string]any{"items": []any{map[string]any{"errors": nil, "gross": nil}}})
	if err != nil || nested.Items[0].Errors == nil || nested.Items[0].Gross != nil {
		t.Fatalf("nullable wage or errors changed: %+v %v", nested, err)
	}
}

func TestSavedBatchMetadataAndSheetCellsSurviveRPC(t *testing.T) {
	input := &income.Batch{Id: "10", Deleted: true, Filename: "work.csv", Spreadsheet: "saved-sheet", Range: "Sheet!A1:B2", Mapping: &income.MappingReq{Sheet: "Sheet", HeaderRow: 1, Columns: []*income.ColumnMapping{{Field: "gross", Column: 0}}, Decisions: []*income.RowDecision{{Row: 2, Action: "exclude"}}}}
	wire, e := proto.Marshal(input)
	if e != nil {
		t.Fatal(e)
	}
	var got income.Batch
	if e = proto.Unmarshal(wire, &got); e != nil {
		t.Fatal(e)
	}
	if !proto.Equal(input, &got) {
		t.Fatal("batch metadata lost on RPC wire")
	}
	sheet := &income.SheetData{Sheet: "Sheet", Total: 1, Items: []*income.SheetRow{{Row: 1, Cells: []string{"", "中文", "=literal"}}}}
	wire, e = proto.Marshal(sheet)
	if e != nil {
		t.Fatal(e)
	}
	var out income.SheetData
	if e = proto.Unmarshal(wire, &out); e != nil || !proto.Equal(sheet, &out) {
		t.Fatal("raw cells lost on RPC wire", e)
	}
}
