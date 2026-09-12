# Encoding, OTP, and compact tokens

## Base62: `encoding/base62`

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/encoding/base62"
)

func main() {
	codec := base62.New("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	encoded := codec.Encode(123456)
	decoded := codec.Decode(encoded)
	fmt.Println(encoded, decoded)
}
```

`New` requires exactly 62 unique bytes and panics for an invalid alphabet.
`Encode(0)` returns an empty string, and `Decode("")` returns `0`. Invalid or
overflowing input to `Decode` also becomes `0`, so the API cannot distinguish
the value `0` from an error; validate the format or check a round-trip when that
distinction matters.

After `Base62` is created, its tables do not change, so one codec can be reused
for encoding and decoding.

## TOTP/HOTP: `encoding/otp`

Import `go.osspkg.com/algorithms/encoding/otp`. The defaults are SHA-1, a
30-second TOTP period, and 6 digits:

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/encoding/otp"
)

func main() {
	gen, err := otp.New(otp.OptHashSHA256(), otp.OptCode8Digits(), otp.OptPeriod(60))
	if err != nil {
		panic(err)
	}

	secret, err := gen.NewSecret(20)
	if err != nil {
		panic(err)
	}
	code, err := gen.GenerateTOTP(secret, 0)
	if err != nil {
		panic(err)
	}
	uri := gen.UrlTOTP(secret, "alice@example.com", "Example")
	fmt.Println(code, uri)
}
```

`NewSecret` creates a Base32 secret without padding; validation also accepts
spaces, lowercase letters, and unpadded Base32. `delta` in `GenerateTOTP` is an
offset measured in periods, not seconds. For HOTP, use
`GenerateHOTP(secret, counter)` and `UrlHOTP(secret, account, issuer, counter)`.
`OptPeriod` clamps the period to 30–120 seconds, while an invalid secret or
counter returns an error.

The secret and `otpauth` URI contain the key: do not log them or expose them in
ordinary traces or client responses without an explicit need.

## `T64`: `encoding/token`

`T64` is an 8-byte value with an 18-character text representation in the
`XXXXXX-XXXX-XXXXXX` format:

```go
package main

import (
	"fmt"

	"go.osspkg.com/algorithms/encoding/token"
)

func main() {
	id := token.NewRandom()
	wire := id.String()
	parsed, ok := token.Parse(wire)
	if !ok {
		panic("invalid token")
	}
	fmt.Println(parsed == id, id.Uint64())
}
```

Factories:

- `NewRandom()` — a random `T64`;
- `NewByTime()` and `NewFormTime(time.Time)` — UnixNano in big-endian order;
- `NewByUint(uint64)` — a number in big-endian order;
- `NewByBytes([]byte)` and `NewByString(string)` — the first 8 bytes of an MD5
  hash, suitable for a deterministic identifier but not for storing a secret.

`Parse`/`ParseBytes` return `(T64, bool)`. `SetTable` changes the process-wide
representation table: it must have an even length from 16 to 254, contain
unique bytes, and exclude `0xff`. Configure it once at startup, before tokens
are exchanged; the byte value of `T64` itself does not change.

`T64` provides text/binary and SQL `Scanner`/`Valuer` integration: `Value`
returns a string, while `Scan` accepts `nil`, empty values, `string`, and
`[]byte`. For JSON, do not rely solely on the presence of `MarshalJSON`: the
current implementation delegates to `MarshalText` without adding JSON quotes.
For JSON fields, use an explicit string wrapper and check the round-trip with
`Parse` until this contract is separately confirmed.

Verified examples in the checkout: `encoding/base62/base62_test.go`,
`encoding/otp/totp_test.go`, and `encoding/token/token_test.go`.
