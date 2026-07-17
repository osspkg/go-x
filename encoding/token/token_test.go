/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package token

import (
	"crypto/md5"
	"encoding/binary"
	"math"
	"testing"
	"time"

	"go.osspkg.com/casecheck"
)

// TestSetTable проверяет установку таблицы кодирования и обработку ошибок.
func TestUnit_SetTable(t *testing.T) {
	// Сохраняем исходную таблицу для восстановления после теста.
	oldTable := table
	oldReverse := reverseTable
	defer func() {
		table = oldTable
		reverseTable = oldReverse
	}()

	// Успешная установка (длина 36, чётная, без дубликатов).
	const validTable = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	err := SetTable(validTable)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Проверяем, что кодирование работает с новой таблицей (нет паники).
	tok := NewByUint(0x0102030405060708)
	s := tok.String()
	if len(s) != 18 {
		t.Errorf("expected length 18, got %d", len(s))
	}
	// Проверяем, что распарсить обратно можно.
	if nt, ok := Parse(s); !ok {
		t.Error("failed to parse token encoded with custom table")
	} else {
		casecheck.Equal(t, uint64(0x0102030405060708), nt.Uint64())
	}

	// Ошибка: нечётная длина.
	err = SetTable("abc")
	if err == nil {
		t.Error("expected error for odd length")
	}
	// Ошибка: длина меньше 16.
	err = SetTable("abcdefghijklmno") // 15
	if err == nil {
		t.Error("expected error for length < 16")
	}
	// Ошибка: длина больше 255.
	long := make([]byte, 256)
	for i := range long {
		long[i] = byte('A' + i%26)
	}
	err = SetTable(string(long))
	if err == nil {
		t.Error("expected error for length > 255")
	}
	// Ошибка: дублирующиеся символы.
	err = SetTable("AABBCCDDEEFFGGHH")
	if err == nil {
		t.Error("expected error for duplicate chars")
	}
}

// TestT64String проверяет преобразование токена в строку.
func TestUnit_T64String(t *testing.T) {
	casecheck.NoError(t, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	// Используем таблицу по умолчанию "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789".
	// Задаём байты: 0,1,35,36,37,70,71,255.
	var tok T64
	tok[0] = 0
	tok[1] = 1
	tok[2] = 35
	tok[3] = 36
	tok[4] = 37
	tok[5] = 70
	tok[6] = 71
	tok[7] = 255

	expected := "AAABA9-BABB-B8B9HD"
	s := tok.String()
	if s != expected {
		t.Errorf("String() = %q, want %q", s, expected)
	}

	// Проверка Nil (все байты нулевые).
	var nilTok T64
	nilStr := nilTok.String()
	if len(nilStr) != 18 {
		t.Errorf("Nil String length = %d, want 18", len(nilStr))
	}
	// Парсим обратно и сравниваем с Nil.
	parsed, ok := Parse(nilStr)
	if !ok {
		t.Error("failed to parse Nil string")
	}
	if parsed != Nil {
		t.Errorf("parse of Nil string = %v, want Nil", parsed)
	}
}

// TestParseBytes проверяет разбор токена из среза байт.
func TestUnit_ParseBytes(t *testing.T) {
	casecheck.NoError(t, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	valid := []byte("AAABA9-BABB-B8B9HD")
	tok, ok := ParseBytes(valid)
	if !ok {
		t.Fatal("ParseBytes failed")
	}
	expected := T64{0, 1, 35, 36, 37, 70, 71, 255}
	if tok != expected {
		t.Errorf("ParseBytes = %v, want %v", tok, expected)
	}

	// Неправильная длина.
	if _, ok := ParseBytes([]byte("short")); ok {
		t.Error("expected false for short input")
	}
	// Отсутствие дефисов.
	if _, ok := ParseBytes([]byte("AAABAZBABBB8B9HD")); ok {
		t.Error("expected false without hyphens")
	}
	// Дефисы не на своих местах.
	if _, ok := ParseBytes([]byte("AA-ABAZ-BABBB8B9H")); ok {
		t.Error("expected false for wrong hyphen positions")
	}
	// Недопустимые символы.
	if _, ok := ParseBytes([]byte("AAABAZ-BABB-B8B9H@")); ok {
		t.Error("expected false for invalid character")
	}
}

// TestParse проверяет разбор токена из строки.
func TestUnit_Parse(t *testing.T) {
	casecheck.NoError(t, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	valid := "AAABA9-BABB-B8B9HD"
	tok, ok := Parse(valid)
	if !ok {
		t.Fatal("Parse failed")
	}
	expected := T64{0, 1, 35, 36, 37, 70, 71, 255}
	if tok != expected {
		t.Errorf("Parse = %v, want %v", tok, expected)
	}
	if _, ok := Parse("invalid"); ok {
		t.Error("expected false for invalid string")
	}
}

// TestNewByTime проверяет создание токена по текущему времени.
func TestUnit_NewByTime(t *testing.T) {
	now := time.Now().UnixNano()
	tok := NewByTime()
	val := binary.BigEndian.Uint64(tok[:])
	// Допускаем небольшое расхождение из-за времени выполнения.
	if val < uint64(now) || val > uint64(now)+10000 {
		t.Errorf("NewByTime value %d not close to now %d", val, now)
	}
}

// TestNewFormTime проверяет создание токена по заданному времени.
func TestUnit_NewFormTime(t *testing.T) {
	tt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tok := NewFormTime(tt)
	val := binary.BigEndian.Uint64(tok[:])
	expected := uint64(tt.UnixNano())
	if val != expected {
		t.Errorf("NewFormTime = %d, want %d", val, expected)
	}
}

// TestNewByUint проверяет создание токена из uint64.
func TestUnit_NewByUint(t *testing.T) {
	v := uint64(0x0102030405060708)
	tok := NewByUint(v)
	val := binary.BigEndian.Uint64(tok[:])
	if val != v {
		t.Errorf("NewByUint = %x, want %x", val, v)
	}
}

// TestNewByBytes проверяет создание токена из среза байт (MD5).
func TestUnit_NewByBytes(t *testing.T) {
	b := []byte("hello")
	tok := NewByBytes(b)
	hash := md5.Sum(b)
	var expected T64
	copy(expected[:], hash[:8])
	if tok != expected {
		t.Errorf("NewByBytes = %v, want %v", tok, expected)
	}
}

// TestNewByString проверяет создание токена из строки (MD5).
func TestUnit_NewByString(t *testing.T) {
	s := "hello"
	tok := NewByString(s)
	expected := NewByBytes([]byte(s))
	if tok != expected {
		t.Errorf("NewByString = %v, want %v", tok, expected)
	}
}

// TestNewRandom проверяет генерацию случайного токена.
func TestUnit_NewRandom(t *testing.T) {
	tok := NewRandom()
	if tok == Nil {
		t.Error("NewRandom returned Nil")
	}
	s := tok.String()
	if len(s) != 18 {
		t.Errorf("String length = %d, want 18", len(s))
	}
	// Парсим обратно.
	tok2, ok := Parse(s)
	if !ok {
		t.Error("failed to parse random token")
	}
	if tok != tok2 {
		t.Error("parsed token does not equal original")
	}
}

// TestMarshalUnmarshal проверяет методы Marshal/Unmarshal для текста, бинарного и JSON.
func TestUnit_MarshalUnmarshal(t *testing.T) {
	orig := NewByUint(0x0102030405060708)

	// MarshalText / UnmarshalText
	data, err := orig.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var t1 T64
	err = t1.UnmarshalText(data)
	if err != nil {
		t.Fatal(err)
	}
	if orig != t1 {
		t.Errorf("UnmarshalText mismatch: got %v, want %v", t1, orig)
	}

	// MarshalBinary / UnmarshalBinary
	bData, err := orig.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	var t2 T64
	err = t2.UnmarshalBinary(bData)
	if err != nil {
		t.Fatal(err)
	}
	if orig != t2 {
		t.Errorf("UnmarshalBinary mismatch: got %v, want %v", t2, orig)
	}

	// MarshalJSON / UnmarshalJSON
	jData, err := orig.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var t3 T64
	err = t3.UnmarshalJSON(jData)
	if err != nil {
		t.Fatal(err)
	}
	if orig != t3 {
		t.Errorf("UnmarshalJSON mismatch: got %v, want %v", t3, orig)
	}

	// Ошибки при невалидных данных.
	var t4 T64
	err = t4.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Error("expected error for invalid text")
	}
	err = t4.UnmarshalJSON([]byte(`"invalid"`))
	if err == nil {
		t.Error("expected error for invalid JSON text")
	}
}

// TestSQL проверяет реализацию sql.Scanner и driver.Valuer.
func TestUnit_SQL(t *testing.T) {
	orig := NewByUint(0x0102030405060708)

	// Value
	val, err := orig.Value()
	if err != nil {
		t.Fatal(err)
	}
	str, ok := val.(string)
	if !ok {
		t.Fatalf("Value returned %T, want string", val)
	}
	if str != orig.String() {
		t.Errorf("Value = %q, want %q", str, orig.String())
	}

	// Scan из string
	var t1 T64
	err = t1.Scan(str)
	if err != nil {
		t.Fatal(err)
	}
	if t1 != orig {
		t.Errorf("Scan from string mismatch: got %v, want %v", t1, orig)
	}

	// Scan из []byte
	var t2 T64
	err = t2.Scan([]byte(str))
	if err != nil {
		t.Fatal(err)
	}
	if t2 != orig {
		t.Errorf("Scan from []byte mismatch: got %v, want %v", t2, orig)
	}

	// Scan nil -> Nil
	var t3 T64
	err = t3.Scan(nil)
	if err != nil {
		t.Fatal(err)
	}
	if t3 != Nil {
		t.Errorf("Scan(nil) = %v, want Nil", t3)
	}

	// Scan пустая строка -> Nil
	var t4 T64
	err = t4.Scan("")
	if err != nil {
		t.Fatal(err)
	}
	if t4 != Nil {
		t.Errorf("Scan(\"\") = %v, want Nil", t4)
	}

	// Scan пустой []byte -> Nil
	var t5 T64
	err = t5.Scan([]byte{})
	if err != nil {
		t.Fatal(err)
	}
	if t5 != Nil {
		t.Errorf("Scan([]byte{}) = %v, want Nil", t5)
	}

	// Scan неверный тип
	var t6 T64
	err = t6.Scan(123)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

/*
goos: linux
goarch: amd64
pkg: go.osspkg.com/algorithms/encoding/token
cpu: 12th Gen Intel(R) Core(TM) i9-12900KF
Benchmark_ByTime
Benchmark_ByTime-4     	197569156	         6.013 ns/op	       0 B/op	       0 allocs/op
Benchmark_FormTime
Benchmark_FormTime-4   	1000000000	         0.06160 ns/op	       0 B/op	       0 allocs/op
Benchmark_ByUint
Benchmark_ByUint-4     	1000000000	         0.06557 ns/op	       0 B/op	       0 allocs/op
Benchmark_ByString
Benchmark_ByString-4   	46054270	        25.78 ns/op	      64 B/op	       2 allocs/op
Benchmark_Random
Benchmark_Random-4     	16827362	        74.25 ns/op	       0 B/op	       0 allocs/op
Benchmark_Parse
Benchmark_Parse-4      	552649376	         2.166 ns/op	       0 B/op	       0 allocs/op
*/

func Benchmark_ByTime(b *testing.B) {
	casecheck.NoError(b, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewByTime()
		}
	})
}

func Benchmark_FormTime(b *testing.B) {
	casecheck.NoError(b, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewFormTime(time.Time{})
		}
	})
}

func Benchmark_ByUint(b *testing.B) {
	casecheck.NoError(b, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewByUint(math.MaxUint64)
		}
	})
}

func Benchmark_ByString(b *testing.B) {
	casecheck.NoError(b, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewByString("udbfvaudsfvaounucnuruhn875ro87ugrneimuacgnomiue")
		}
	})
}

func Benchmark_Random(b *testing.B) {
	casecheck.NoError(b, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			NewRandom()
		}
	})
}

func Benchmark_Parse(b *testing.B) {
	casecheck.NoError(b, SetTable("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"))
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Parse("012345-6789-abcdef")
		}
	})
}
