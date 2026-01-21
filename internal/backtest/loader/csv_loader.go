package loader

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/iluyuns/alpha-trade/internal/domain/model"
)

// CsvLoader CSV历史数据加载器
// 支持标准K线格式：timestamp,open,high,low,close,volume
type CsvLoader struct {
	filePath string
	candles  []*model.Candle
	index    int
}

// NewCsvLoader 创建CSV加载器
func NewCsvLoader(filePath string) (*CsvLoader, error) {
	loader := &CsvLoader{
		filePath: filePath,
		index:    0,
	}

	if err := loader.load(); err != nil {
		return nil, err
	}

	return loader, nil
}

// load 加载CSV文件
func (l *CsvLoader) load() error {
	file, err := os.Open(l.filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// 跳过表头
	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}

	candles := make([]*model.Candle, 0, 1000)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read record: %w", err)
		}

		candle, err := l.parseRecord(record)
		if err != nil {
			continue // 跳过无效行
		}

		candles = append(candles, candle)
	}

	// 按时间排序（确保事件时间顺序）
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].OpenTime.Before(candles[j].OpenTime)
	})

	l.candles = candles
	return nil
}

// parseRecord 解析CSV行
// 格式：timestamp,open,high,low,close,volume
func (l *CsvLoader) parseRecord(record []string) (*model.Candle, error) {
	if len(record) < 6 {
		return nil, fmt.Errorf("invalid record length: %d", len(record))
	}

	// 解析时间戳（Unix毫秒或秒）
	timestamp, err := strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return nil, err
	}

	// 自动判断时间戳格式（秒 vs 毫秒）
	var openTime time.Time
	if timestamp > 1e12 {
		// 毫秒时间戳
		openTime = time.UnixMilli(timestamp)
	} else {
		// 秒时间戳
		openTime = time.Unix(timestamp, 0)
	}

	open, err := model.NewMoney(record[1])
	if err != nil {
		return nil, err
	}

	high, err := model.NewMoney(record[2])
	if err != nil {
		return nil, err
	}

	low, err := model.NewMoney(record[3])
	if err != nil {
		return nil, err
	}

	close, err := model.NewMoney(record[4])
	if err != nil {
		return nil, err
	}

	volume, err := model.NewMoney(record[5])
	if err != nil {
		return nil, err
	}

	return &model.Candle{
		Symbol:    "UNKNOWN", // 由外部设置
		Interval:  "1m",      // 默认1分钟
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
		OpenTime:  openTime,
		CloseTime: openTime.Add(1 * time.Minute),
		RecvTime:  time.Now(),
	}, nil
}

// Next 获取下一个K线
func (l *CsvLoader) Next() (*model.Candle, error) {
	if !l.HasNext() {
		return nil, io.EOF
	}

	candle := l.candles[l.index]
	l.index++
	return candle, nil
}

// HasNext 是否还有数据
func (l *CsvLoader) HasNext() bool {
	return l.index < len(l.candles)
}

// CurrentTime 当前回测时间（事件时间）
func (l *CsvLoader) CurrentTime() int64 {
	if l.index == 0 {
		return 0
	}
	if l.index > len(l.candles) {
		return l.candles[len(l.candles)-1].OpenTime.UnixMilli()
	}
	return l.candles[l.index-1].OpenTime.UnixMilli()
}

// Reset 重置迭代器
func (l *CsvLoader) Reset() {
	l.index = 0
}

// Count 总数据条数
func (l *CsvLoader) Count() int {
	return len(l.candles)
}

// SetSymbol 批量设置交易对
func (l *CsvLoader) SetSymbol(symbol string) {
	for _, candle := range l.candles {
		candle.Symbol = symbol
	}
}

// SetInterval 批量设置时间间隔
func (l *CsvLoader) SetInterval(interval string) {
	for _, candle := range l.candles {
		candle.Interval = interval
	}
}
