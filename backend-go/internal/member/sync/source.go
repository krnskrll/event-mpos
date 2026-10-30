// Package sync is the pluggable boundary to the external member system — same swap-without-rewrite
// design as the ERP product source. Selected by app.member.source: mock (default) | file | rest.
package sync

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/krnskrll/event-mpos/internal/member"
)

// Record is a normalized member record from the external system, regardless of transport.
type Record struct {
	MemberCode string
	QRCode     string
	Name       string
	Tier       member.MemberTier
	Region     *string
	Active     bool
	Phone      *string
}

// Source pulls the full current member list from the external system.
type Source interface {
	Description() string
	FetchMembers() ([]Record, error)
}

// NewSource selects the member source from config (mock | file | rest).
func NewSource(kind, filePath, restURL string) Source {
	switch kind {
	case "file":
		return FileSource{Path: filePath}
	case "rest":
		return RestSource{URL: restURL}
	default:
		return MockSource{}
	}
}

func parseTier(s string) member.MemberTier {
	switch member.MemberTier(strings.ToUpper(strings.TrimSpace(s))) {
	case member.TierSilver:
		return member.TierSilver
	case member.TierGold:
		return member.TierGold
	case member.TierPlatinum:
		return member.TierPlatinum
	default:
		return member.TierStandard
	}
}

func strptr(s string) *string { return &s }

func blankToNil(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// ---- mock ----

// MockSource is the default dev source. QR content is "MBR-xxxxx".
type MockSource struct{}

func (MockSource) Description() string { return "mock" }

func (MockSource) FetchMembers() ([]Record, error) {
	return []Record{
		{MemberCode: "M00001", QRCode: "MBR-00001", Name: "Aisha Rahman", Tier: member.TierPlatinum, Region: strptr("Central"), Active: true},
		{MemberCode: "M00002", QRCode: "MBR-00002", Name: "Bryan Tan", Tier: member.TierGold, Region: strptr("Central"), Active: true},
		{MemberCode: "M00003", QRCode: "MBR-00003", Name: "Chloe Lim", Tier: member.TierSilver, Region: strptr("North"), Active: true},
		{MemberCode: "M00004", QRCode: "MBR-00004", Name: "Devan Nair", Tier: member.TierStandard, Region: strptr("South"), Active: true},
		{MemberCode: "M00005", QRCode: "MBR-00005", Name: "Elena Wong", Tier: member.TierGold, Region: strptr("Central"), Active: true},
		{MemberCode: "M00006", QRCode: "MBR-00006", Name: "Former Customer", Tier: member.TierStandard, Region: strptr("East"), Active: false},
	}, nil
}

// ---- file (CSV) ----

// FileSource reads a CSV export. Columns: memberCode,qrCode,name,tier,region,active.
type FileSource struct{ Path string }

func (f FileSource) Description() string { return "file:" + f.Path }

func (f FileSource) FetchMembers() ([]Record, error) {
	path := strings.TrimPrefix(f.Path, "classpath:")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("member file not found at %s: %w", f.Path, err)
	}
	defer file.Close()

	var out []Record
	scanner := bufio.NewScanner(file)
	first := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if first {
			first = false
			continue
		}
		c := strings.Split(line, ",")
		for i := range c {
			c[i] = strings.TrimSpace(c[i])
		}
		if len(c) < 6 {
			return nil, fmt.Errorf("malformed member CSV row (need 6 columns): %s", line)
		}
		out = append(out, Record{
			MemberCode: c[0], QRCode: c[1], Name: c[2],
			Tier: parseTier(c[3]), Region: blankToNil(c[4]),
			Active: strings.EqualFold(c[5], "true"),
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ---- rest ----

// RestSource calls a configured endpoint returning member JSON.
type RestSource struct{ URL string }

func (r RestSource) Description() string { return "rest:" + r.URL }

type memberJSON struct {
	MemberCode string  `json:"memberCode"`
	QRCode     string  `json:"qrCode"`
	Name       string  `json:"name"`
	Tier       string  `json:"tier"`
	Region     *string `json:"region"`
	Active     bool    `json:"active"`
}

func (r RestSource) FetchMembers() ([]Record, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(r.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("member REST returned status %d", resp.StatusCode)
	}
	var raw []memberJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]Record, 0, len(raw))
	for _, m := range raw {
		out = append(out, Record{
			MemberCode: m.MemberCode, QRCode: m.QRCode, Name: m.Name,
			Tier: parseTier(m.Tier), Region: m.Region, Active: m.Active,
		})
	}
	return out, nil
}
