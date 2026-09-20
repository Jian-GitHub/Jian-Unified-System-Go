package files

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/xml"
	"io"
	a "jian-unified-system/hephaestus/hephaestus-rpc/internal/application"
	d "jian-unified-system/hephaestus/hephaestus-rpc/internal/domain"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MaxFile = 10 << 20
const MaxExpanded = 50 << 20

var safeID = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Parser struct{ Root string }

func (p *Parser) Parse(ctx context.Context, owner int64, upload, filename string) (a.Grid, string, error) {
	if !safeID.MatchString(upload) || owner <= 0 {
		return a.Grid{}, "", d.Invalid("invalid upload handle")
	}
	name := filepath.Join(p.Root, strconv.FormatInt(owner, 10), upload)
	f, e := os.Open(name)
	if e != nil {
		return a.Grid{}, "", d.Invalid("upload is missing")
	}
	defer f.Close()
	defer os.Remove(name)
	b, e := io.ReadAll(io.LimitReader(f, MaxFile+1))
	if e != nil {
		return a.Grid{}, "", e
	}
	if len(b) > MaxFile {
		return a.Grid{}, "", d.Fail("TOO_LARGE", "file exceeds 10 MiB")
	}
	h := sha256.Sum256(b)
	hash := hex.EncodeToString(h[:])
	g, e := ParseBytes(ctx, b, strings.ToLower(filepath.Ext(filename)))
	return g, hash, e
}
func ParseBytes(ctx context.Context, b []byte, ext string) (a.Grid, error) {
	g := a.Grid{Sheets: map[string][][]string{}}
	if len(b) > MaxFile {
		return g, d.Fail("TOO_LARGE", "file exceeds 10 MiB")
	}
	switch ext {
	case ".csv":
		if !utf8.Valid(b) || bytes.ContainsRune(b, 0) {
			return g, d.Invalid("CSV must be UTF-8 text")
		}
		b = bytes.TrimPrefix(b, []byte{0xef, 0xbb, 0xbf})
		r := csv.NewReader(bytes.NewReader(b))
		r.FieldsPerRecord = -1
		rows := [][]string{}
		for {
			if e := ctx.Err(); e != nil {
				return g, e
			}
			row, e := r.Read()
			if e == io.EOF {
				break
			}
			if e != nil {
				return g, d.Invalid("invalid CSV quoting or row structure")
			}
			if e = checkRow(row); e != nil {
				return g, e
			}
			rows = append(rows, row)
			if len(rows) > 10001 {
				return g, d.Invalid("file exceeds 10000 data rows")
			}
		}
		if len(rows) < 1 {
			return g, d.Invalid("empty CSV")
		}
		g.Sheets["CSV"] = rows
		return g, nil
	case ".xlsx":
		return parseXLSX(ctx, b)
	default:
		return g, d.Invalid("only CSV and XLSX are supported")
	}
}
func checkRow(row []string) error {
	if len(row) > 100 {
		return d.Invalid("sheet exceeds 100 columns")
	}
	for _, c := range row {
		if len(c) > 32768 || strings.ContainsRune(c, 0) || !utf8.ValidString(c) {
			return d.Invalid("invalid or oversized cell")
		}
	}
	return nil
}

type relation struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
	Mode   string `xml:"TargetMode,attr"`
}
type relationships struct {
	Items []relation `xml:"Relationship"`
}

func parseXLSX(ctx context.Context, b []byte) (a.Grid, error) {
	g := a.Grid{Sheets: map[string][][]string{}}
	z, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		return g, d.Invalid("invalid XLSX archive")
	}
	if len(z.File) > 2048 {
		return g, d.Invalid("too many archive entries")
	}
	entries := map[string][]byte{}
	total := int64(0)
	for _, f := range z.File {
		if e = ctx.Err(); e != nil {
			return g, e
		}
		if strings.HasSuffix(f.Name, "/") {
			continue
		}
		name := path.Clean(f.Name)
		if name != f.Name || strings.HasPrefix(name, "/") || strings.HasPrefix(name, "../") || strings.Contains(name, "\\") {
			return g, d.Invalid("unsafe archive path")
		}
		if _, ok := entries[name]; ok {
			return g, d.Invalid("duplicate archive entry")
		}
		if f.UncompressedSize64 > MaxExpanded || total+int64(f.UncompressedSize64) > MaxExpanded {
			return g, d.Fail("TOO_LARGE", "expanded XLSX exceeds limit")
		}
		lower := strings.ToLower(name)
		if strings.Contains(lower, "vbaproject") || strings.Contains(lower, "externallinks") {
			return g, d.Invalid("macros and external links are not supported")
		}
		r, err := f.Open()
		if err != nil {
			return g, d.Invalid("invalid archive entry")
		}
		data, err := io.ReadAll(io.LimitReader(r, MaxExpanded-total+1))
		r.Close()
		if err != nil {
			return g, d.Invalid("invalid archive entry")
		}
		total += int64(len(data))
		if total > MaxExpanded {
			return g, d.Fail("TOO_LARGE", "expanded XLSX exceeds limit")
		}
		entries[name] = data
		if strings.HasSuffix(name, ".rels") {
			var rel relationships
			if xml.Unmarshal(data, &rel) != nil {
				return g, d.Invalid("invalid relationships")
			}
			for _, v := range rel.Items {
				if strings.EqualFold(v.Mode, "External") {
					return g, d.Invalid("external relationships are not supported")
				}
			}
		}
	}
	if bytes.Contains(bytes.ToLower(entries["[Content_Types].xml"]), []byte("macroenabled")) {
		return g, d.Invalid("macro-enabled workbook is not supported")
	}
	var workbook struct {
		Props struct {
			Date1904 string `xml:"date1904,attr"`
		} `xml:"workbookPr"`
		Sheets []struct {
			Name string `xml:"name,attr"`
			ID   string `xml:"id,attr"`
		} `xml:"sheets>sheet"`
	}
	if xml.Unmarshal(entries["xl/workbook.xml"], &workbook) != nil || len(workbook.Sheets) == 0 || len(workbook.Sheets) > 20 {
		return g, d.Invalid("invalid workbook or too many sheets")
	}
	g.Date1904 = workbook.Props.Date1904 == "1" || workbook.Props.Date1904 == "true"
	var rel relationships
	if xml.Unmarshal(entries["xl/_rels/workbook.xml.rels"], &rel) != nil {
		return g, d.Invalid("missing workbook relationships")
	}
	targets := map[string]string{}
	for _, v := range rel.Items {
		target := path.Clean(path.Join("xl", v.Target))
		if strings.HasPrefix(v.Target, "/") {
			target = strings.TrimPrefix(v.Target, "/")
		}
		if !strings.HasPrefix(target, "xl/") {
			return g, d.Invalid("unsafe workbook target")
		}
		targets[v.ID] = target
	}
	shared := []string{}
	if data, ok := entries["xl/sharedStrings.xml"]; ok {
		var ss struct {
			Items []struct {
				T    string `xml:"t"`
				Runs []struct {
					T string `xml:"t"`
				} `xml:"r"`
			} `xml:"si"`
		}
		if xml.Unmarshal(data, &ss) != nil {
			return g, d.Invalid("invalid shared strings")
		}
		for _, v := range ss.Items {
			s := v.T
			for _, r := range v.Runs {
				s += r.T
			}
			if len(s) > 32768 {
				return g, d.Invalid("shared string too long")
			}
			shared = append(shared, s)
		}
	}
	totalRows, totalText := 0, 0
	for _, sheet := range workbook.Sheets {
		if e = ctx.Err(); e != nil {
			return g, e
		}
		if _, ok := g.Sheets[sheet.Name]; ok {
			return g, d.Invalid("duplicate sheet name")
		}
		data, ok := entries[targets[sheet.ID]]
		if !ok {
			return g, d.Invalid("missing worksheet")
		}
		var ws struct {
			Rows []struct {
				N     int `xml:"r,attr"`
				Cells []struct {
					Ref     string  `xml:"r,attr"`
					Type    string  `xml:"t,attr"`
					Value   *string `xml:"v"`
					Formula *string `xml:"f"`
					Inline  struct {
						T    string `xml:"t"`
						Runs []struct {
							T string `xml:"t"`
						} `xml:"r"`
					} `xml:"is"`
				} `xml:"c"`
			} `xml:"sheetData>row"`
		}
		if xml.Unmarshal(data, &ws) != nil {
			return g, d.Invalid("invalid worksheet XML")
		}
		rows := [][]string{}
		last := 0
		for _, row := range ws.Rows {
			if row.N <= last || row.N > 10001 {
				return g, d.Invalid("invalid or oversized worksheet row index")
			}
			for len(rows) < row.N {
				rows = append(rows, []string{})
			}
			last = row.N
			cells := []string{}
			seen := map[int]bool{}
			for _, c := range row.Cells {
				col, rowNumber, err := cellPosition(c.Ref)
				if err != nil || rowNumber != row.N || seen[col] {
					return g, d.Invalid("invalid or duplicate cell coordinate")
				}
				seen[col] = true
				for len(cells) <= col {
					cells = append(cells, "")
				}
				value := ""
				if c.Value != nil {
					value = *c.Value
				}
				if c.Formula != nil && (c.Value == nil || value == "") {
					return g, d.Invalid("formula has no cached result; recalculate and export again")
				}
				switch c.Type {
				case "s":
					n, err := strconv.Atoi(value)
					if err != nil || n < 0 || n >= len(shared) {
						return g, d.Invalid("invalid shared string index")
					}
					value = shared[n]
				case "inlineStr":
					value = c.Inline.T
					for _, r := range c.Inline.Runs {
						value += r.T
					}
				case "e":
					return g, d.Invalid("worksheet contains an Excel error cell")
				case "", "n", "str", "b", "d":
				default:
					return g, d.Invalid("unsupported cell type")
				}
				cells[col] = value
				totalText += len(value)
				if totalText > MaxFile {
					return g, d.Fail("TOO_LARGE", "normalized worksheet text exceeds 10 MiB")
				}
			}
			if e = checkRow(cells); e != nil {
				return g, e
			}
			rows[row.N-1] = cells
		}
		totalRows += len(rows)
		if totalRows > 10020 {
			return g, d.Invalid("workbook exceeds row budget")
		}
		g.Sheets[sheet.Name] = rows
	}
	return g, nil
}
func cellPosition(s string) (int, int, error) {
	i, col := 0, 0
	for i < len(s) && s[i] >= 'A' && s[i] <= 'Z' {
		col = col*26 + int(s[i]-'A'+1)
		i++
		if col > 100 {
			return 0, 0, d.Invalid("column exceeds limit")
		}
	}
	row, e := strconv.Atoi(s[i:])
	if i == 0 || e != nil || row < 1 || row > 10001 {
		return 0, 0, d.Invalid("invalid coordinate")
	}
	return col - 1, row, nil
}
