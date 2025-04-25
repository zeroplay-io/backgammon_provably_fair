// Package verifier checks a Backgammon “provably-fair” JSON
// report produced by the game server.
//
// A valid report proves that every die roll was generated from
// • serverSeed   – 16-byte random value chosen by the server
// • clientSeed A/B – 16-byte value from each player
// • nonce        – roll index (0,1,2…)
//
// The verifier reproduces the entire dice stream with
//
//	HMAC-SHA256(serverSeed, combinedSeed || nonce_BE)
//
// and compares it to the rolls recorded in the report.
package verifier

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// Report mirrors the JSON exported by the game server.
type Report struct {
	GameID         string        `json:"game_id"`
	ServerSeed     string        `json:"server_seed"`      // hex-encoded 16 B
	ServerSeedHash string        `json:"server_seed_hash"` // SHA-256(serverSeed) hex
	Rolls          string        `json:"rolls"`            // each element = AABBCC...
	Players        []PlayerEntry `json:"players"`          // length == 2
}

// PlayerEntry describes one player’s contribution to the seed mix.
type PlayerEntry struct {
	UID        int64  `json:"uid"`
	ClientSeed string `json:"client_seed"` // hex-encoded 16 B
	Source     string `json:"source"`      // "player" | "fallback"
}

func decodeHex(s, name string) ([]byte, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("%s is not valid hex", name)
	}
	return b, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// combinedSeed deterministically mixes two client seeds with their UIDs.
func combinedSeed(a, b PlayerEntry) ([]byte, error) {
	aUID := strconv.FormatInt(a.UID, 10)
	bUID := strconv.FormatInt(b.UID, 10)
	leftFirst := aUID < bUID
	header := aUID + ":" + bUID + ":"
	if !leftFirst {
		header = bUID + ":" + aUID + ":"
	}
	first := a.ClientSeed
	second := b.ClientSeed
	if aUID > bUID {
		first, second = second, first
	}
	ha, err := decodeHex(first, "clientSeed")
	if err != nil {
		return nil, err
	}
	hb, err := decodeHex(second, "clientSeed")
	if err != nil {
		return nil, err
	}
	buf := append([]byte(header), ha...)
	buf = append(buf, ':')
	buf = append(buf, hb...)
	sum := sha256.Sum256(buf)
	return sum[:], nil
}

// hmacBlock produces one 32-byte HMAC(serverSeed, combined||nonce_BE) block.
func hmacBlock(key, prefix []byte, nonce uint32) []byte {
	msg := make([]byte, len(prefix)+4)
	copy(msg, prefix)
	binary.BigEndian.PutUint32(msg[len(msg)-4:], nonce)

	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}

// drawDie returns one perfectly uniform die value 1-6 using rejection sampling.
func drawDie(key, prefix []byte, nonce *uint32, pool *[]byte, idx *int) int {
	for {
		if *idx >= len(*pool) {
			*pool = hmacBlock(key, prefix, *nonce)
			*idx, *nonce = 0, *nonce+1
		}
		v := (*pool)[*idx]
		*idx++
		if v < 252 { // 0–251 map evenly to 42×6 faces
			return int(v%6) + 1
		}
	}
}

// rollDice returns two dice and advances nonce.
func rollDice(serverSeed, combined []byte, nonce *uint32) (int, int) {
	pool := hmacBlock(serverSeed, combined, *nonce)
	*nonce++
	idx := 0
	d1 := drawDie(serverSeed, combined, nonce, &pool, &idx)
	d2 := drawDie(serverSeed, combined, nonce, &pool, &idx)
	return d1, d2
}

// VerifyBytes parses raw JSON and calls Verify.
func VerifyBytes(blob []byte) error {
	var rep Report
	if err := json.Unmarshal(blob, &rep); err != nil {
		return err
	}
	return Verify(rep)
}

// Verify checks every recorded roll. It returns nil when
// the entire sequence is provably correct.
func Verify(rep Report) error {
	if len(rep.Players) != 2 {
		return errors.New("report must contain exactly 2 players")
	}
	if len(rep.Rolls)%2 != 0 {
		return errors.New("report must contain exactly odds rolls")
	}
	srvSeed, err := decodeHex(rep.ServerSeed, "serverSeed")
	if err != nil {
		return err
	}
	if sha256Hex(srvSeed) != rep.ServerSeedHash {
		return errors.New("server_hash mismatch")
	}
	combined, err := combinedSeed(rep.Players[0], rep.Players[1])
	if err != nil {
		return err
	}

	var nonce uint32
	for i := 0; i < len(rep.Rolls); i += 2 {
		r1, r2 := rep.Rolls[i], rep.Rolls[i+1]
		d1, d2 := rollDice(srvSeed, combined, &nonce)
		e1 := byte(d1) + '0'
		e2 := byte(d2) + '0'
		if e1 != r1 || e2 != r2 {
			return fmt.Errorf("roll %d mismatch: expected (%v,%v), but got(%v,%v)", i, d1, d2, r1-'0', r2-'0')
		}
	}
	return nil
}
