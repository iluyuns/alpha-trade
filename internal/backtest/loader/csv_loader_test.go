package loader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCsvLoader(t *testing.T) {
	// 创建临时CSV文件
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test.csv")

	csvContent := `timestamp,open,high,low,close,volume
1609459200000,29000.00,29500.00,28500.00,29200.00,100.5
1609459260000,29200.00,29800.00,29000.00,29500.00,150.2
1609459320000,29500.00,30000.00,29300.00,29800.00,200.8`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 加载CSV
	loader, err := NewCsvLoader(csvPath)
	if err != nil {
		t.Fatalf("NewCsvLoader failed: %v", err)
	}

	loader.SetSymbol("BTCUSDT")
	loader.SetInterval("1m")

	// 检查总数
	if loader.Count() != 3 {
		t.Errorf("Count = %d, want 3", loader.Count())
	}

	// 迭代验证
	count := 0
	for loader.HasNext() {
		candle, err := loader.Next()
		if err != nil {
			t.Fatalf("Next failed: %v", err)
		}

		if candle.Symbol != "BTCUSDT" {
			t.Errorf("Symbol = %s, want BTCUSDT", candle.Symbol)
		}

		count++
	}

	if count != 3 {
		t.Errorf("Iterated %d candles, want 3", count)
	}

	// 验证 CurrentTime
	currentTime := loader.CurrentTime()
	if currentTime == 0 {
		t.Error("CurrentTime should not be 0 after iteration")
	}

	// 重置测试
	loader.Reset()
	if !loader.HasNext() {
		t.Error("HasNext should be true after reset")
	}
}

func TestCsvLoader_InvalidFile(t *testing.T) {
	_, err := NewCsvLoader("/nonexistent/file.csv")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
