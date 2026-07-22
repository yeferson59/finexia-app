package spreadsheet

import "testing"

func TestReadFileCSV(t *testing.T) {
	t.Run("comma-delimited with BOM", func(t *testing.T) {
		data := append([]byte{0xEF, 0xBB, 0xBF}, []byte("ticker,name\nAAPL,Apple\n")...)
		src, err := ReadFile(data, "assets.csv", "")
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if src.Sheet != "CSV" || len(src.Rows) != 2 {
			t.Fatalf("src = %+v", src)
		}
		if src.Rows[1][0] != "AAPL" || src.Rows[1][1] != "Apple" {
			t.Errorf("row = %v", src.Rows[1])
		}
	})

	t.Run("semicolon delimiter is detected", func(t *testing.T) {
		src, err := ReadFile([]byte("a;b;c\n1;2;3\n"), "x.csv", "")
		if err != nil {
			t.Fatalf("ReadFile: %v", err)
		}
		if len(src.Rows[0]) != 3 {
			t.Errorf("expected 3 columns, got %v", src.Rows[0])
		}
	})

	t.Run("unknown extension is rejected as a non-xlsx", func(t *testing.T) {
		if _, err := ReadFile([]byte("not a spreadsheet"), "data.txt", ""); err == nil {
			t.Error("expected an error opening a non-xlsx, non-csv file")
		}
	})
}

func TestNormKey(t *testing.T) {
	cases := map[string]string{
		"  Descripción ": "descripcion",
		"Asset_Type":     "asset type",
		"FROM/TO":        "from to",
		"Símbolo":        "simbolo",
	}
	for in, want := range cases {
		if got := NormKey(in); got != want {
			t.Errorf("NormKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRowIsEmpty(t *testing.T) {
	if !RowIsEmpty([]string{"", "  ", "\t"}) {
		t.Error("all-blank row should be empty")
	}
	if RowIsEmpty([]string{"", "x"}) {
		t.Error("row with a value should not be empty")
	}
}

func TestCellHelpers(t *testing.T) {
	row := []string{" a ", "b"}
	idx := 0
	if CellAt(row, &idx) != "a" {
		t.Errorf("CellAt trimmed = %q", CellAt(row, &idx))
	}
	if CellAt(row, nil) != "" {
		t.Error("CellAt(nil) should be empty")
	}
	if CellAtIdx(row, 1) != "b" || CellAtIdx(row, 9) != "" {
		t.Error("CellAtIdx out-of-range should be empty")
	}
}
