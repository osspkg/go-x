/*
 *  Copyright (c) 2019-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package otp

import (
	"crypto/hmac"
	"crypto/rand"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

type OTP struct {
	hash      func() hash.Hash
	algorithm string

	period int64

	codeSize int
	codeTmpl string
}

func New(options ...Option) (*OTP, error) {
	obj := &OTP{}

	opts := make([]Option, 0, 10)
	opts = append(opts, OptHashSHA1(), OptPeriod(30), OptCode6Digits())
	opts = append(opts, options...)

	for _, opt := range opts {
		opt(obj)
	}

	return obj, nil
}

func (o *OTP) NewSecret(size int) (string, error) {
	if size < 1 {
		size = 10
	}
	secret := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, secret); err != nil {
		return "", err
	}
	return b32.EncodeToString(secret), nil
}

func (o *OTP) validateSecret(secret string) ([]byte, error) {
	secret = strings.TrimSpace(secret)
	if n := len(secret) % 8; n != 0 {
		secret = secret + strings.Repeat("=", 8-n)
	}
	secret = strings.ToUpper(secret)
	secretBytes, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return nil, fmt.Errorf("invalid secret: %w", err)
	}
	return secretBytes, nil
}

func (o *OTP) generate(secret string, counter uint64, delta int64) (string, error) {
	b, err := o.validateSecret(secret)
	if err != nil {
		return "", err
	}

	if delta >= 0 {
		if uint64(delta) > math.MaxUint64-counter {
			return "", fmt.Errorf("counter overflow")
		}
		counter += uint64(delta)
	} else {
		decrement := uint64(-(delta + 1)) + 1
		if decrement > counter {
			return "", fmt.Errorf("counter underflow")
		}
		counter -= decrement
	}

	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], counter)

	hm := hmac.New(o.hash, b)
	_, _ = hm.Write(counterBytes[:])
	var digest [64]byte
	timeHash := hm.Sum(digest[:0])

	offset := int(timeHash[len(timeHash)-1] & 0x0F)
	truncHash := int64(
		(int(timeHash[offset])&0x7f)<<24 |
			(int(timeHash[offset+1])&0xff)<<16 |
			(int(timeHash[offset+2])&0xff)<<8 |
			(int(timeHash[offset+3]) & 0xff))

	modulo := int64(1_000_000)
	if o.codeSize == 8 {
		modulo = 100_000_000
	}
	otp := truncHash % modulo

	var code [8]byte
	for i := o.codeSize - 1; i >= 0; i-- {
		code[i] = byte(otp%10) + '0'
		otp /= 10
	}
	return string(code[:o.codeSize]), nil
}

func (o *OTP) GenerateTOTP(secret string, delta int64) (string, error) {
	currentTime := time.Now().Unix()
	counter := uint64(currentTime / o.period)

	return o.generate(secret, counter, delta)
}

func (o *OTP) GenerateHOTP(secret string, counter uint64) (string, error) {
	return o.generate(secret, counter, 0)
}

func (o *OTP) UrlTOTP(secret, account, issuer string) string {
	secret = strings.TrimSpace(secret)
	params := url.Values{
		"secret":    []string{secret},
		"issuer":    []string{issuer},
		"algorithm": []string{o.algorithm},
		"digits":    []string{strconv.Itoa(o.codeSize)},
		"period":    []string{strconv.Itoa(int(o.period))},
	}

	uri := url.URL{
		Scheme:   "otpauth",
		Host:     "totp",
		Path:     "/" + account,
		RawQuery: params.Encode(),
	}

	return uri.String()
}

func (o *OTP) UrlHOTP(secret, account, issuer string, counter uint64) string {
	secret = strings.TrimSpace(secret)
	params := url.Values{
		"secret":    []string{secret},
		"issuer":    []string{issuer},
		"algorithm": []string{o.algorithm},
		"digits":    []string{strconv.Itoa(o.codeSize)},
		"period":    []string{strconv.Itoa(int(o.period))},
		"counter":   []string{strconv.FormatUint(counter, 10)},
	}

	uri := url.URL{
		Scheme:   "otpauth",
		Host:     "hotp",
		Path:     "/" + account,
		RawQuery: params.Encode(),
	}

	return uri.String()
}
