package ai

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// TesseractVision parses bank screenshots using local Tesseract OCR.
// Completely free, runs inside the Docker container with no external API calls.
type TesseractVision struct{}

// NewTesseractVision returns an error if tesseract is not installed.
func NewTesseractVision() (*TesseractVision, error) {
	if _, err := exec.LookPath("tesseract"); err != nil {
		return nil, fmt.Errorf("tesseract not found — install tesseract-ocr package")
	}
	return &TesseractVision{}, nil
}

// ParseScreenshot extracts transactions from a bank app screenshot using OCR.
func (t *TesseractVision) ParseScreenshot(ctx context.Context, imageData []byte, mimeType string) ([]ParsedTransaction, error) {
	// Normalise to PNG so tesseract always gets a format it handles reliably.
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		// Fall back to writing raw bytes if Go can't decode (e.g. WebP without x/image).
		return t.runTesseract(ctx, imageData, extensionFor(mimeType))
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("re-encode png: %w", err)
	}
	return t.runTesseract(ctx, buf.Bytes(), ".png")
}

func (t *TesseractVision) runTesseract(ctx context.Context, data []byte, ext string) ([]ParsedTransaction, error) {
	tmp, err := os.CreateTemp("", "ocr-*"+ext)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	// --psm 6: treat image as a single uniform block of text.
	// -l rus+eng: Russian + English character set.
	cmd := exec.CommandContext(ctx, "tesseract", tmp.Name(), "stdout", "-l", "rus+eng", "--psm", "6")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("tesseract: %w", err)
	}

	return parseOCRText(string(out)), nil
}

func extensionFor(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	case "image/heic", "image/heif":
		return ".heic"
	default:
		return ".jpg"
	}
}

// ── OCR text parser ───────────────────────────────────────────────────────────

var (
	// ИТОГО / total / к оплате / сумма — highest priority for receipts.
	totalRe = regexp.MustCompile(`(?i)(?:итого|total|к\s*оплате|сумма)[:\s]+([−\-]?\s*[\d][0-9\s]*[,.][\d]{2})`)

	// Standalone ruble amount on its own line, possibly with leading minus/em-dash.
	// Matches: -1 426,98  |  —847,50  |  1426.98  (₽ optional, captured separately)
	amountRe = regexp.MustCompile(`(?m)^[[:space:]]*([−\-—]?)\s*([\d][\d\s]{0,10}[,.][\d]{2})\s*[₽РрPp]?[[:space:]]*$`)

	// MCC code.
	mccRe = regexp.MustCompile(`(?i)mcc[\s:]+(\d{4})`)

	// Russian date: "17 марта 2026" (optionally followed by "г.")
	ruDateRe = regexp.MustCompile(`(\d{1,2})\s+(января|февраля|марта|апреля|мая|июня|июля|августа|сентября|октября|ноября|декабря)\s+(\d{4})`)

	// Numeric date: 18.03.2026 or 18/03/2026
	numDateRe = regexp.MustCompile(`(\d{2})[./](\d{2})[./](\d{4})`)

	// Time: 09:06
	timeRe = regexp.MustCompile(`\b(\d{2}):(\d{2})\b`)

	// Lines that are clearly not merchant names.
	skipLineRe = regexp.MustCompile(`(?i)операция|статус|категори|mcc|итого|сумма|к\s*оплате|кассов|зачислен|пополнен|возврат|баланс|дата|время|подтвержд|успешн|выполнен`)

	// Lines made mostly of digits, punctuation, or single letters.
	junkLineRe = regexp.MustCompile(`^[\d\s.,₽%+\-—−/:()«»"']+$`)
)

var ruMonths = map[string]time.Month{
	"января": time.January, "февраля": time.February, "марта": time.March,
	"апреля": time.April, "мая": time.May, "июня": time.June,
	"июля": time.July, "августа": time.August, "сентября": time.September,
	"октября": time.October, "ноября": time.November, "декабря": time.December,
}

func parseOCRText(text string) []ParsedTransaction {
	amountCents, ok := extractOCRAmount(text)
	if !ok {
		return nil
	}

	occurredAt := extractOCRDate(text)
	merchant := extractOCRMerchant(strings.Split(text, "\n"))

	mcc := 0
	if m := mccRe.FindStringSubmatch(text); len(m) > 1 {
		mcc, _ = strconv.Atoi(m[1])
	}

	if merchant == "" {
		merchant = "Операция"
	}

	return []ParsedTransaction{{
		Description: merchant,
		AmountCents: amountCents,
		MCC:         mcc,
		OccurredAt:  occurredAt,
		Currency:    "RUB",
	}}
}

func extractOCRAmount(text string) (int64, bool) {
	// Priority 1: ИТОГО / К оплате line (receipts).
	if m := totalRe.FindStringSubmatch(text); len(m) > 1 {
		v, ok := rubleStringToCents(m[1], false)
		if ok {
			return -v, true // receipt total → expense
		}
	}

	// Priority 2: standalone amount on its own line.
	matches := amountRe.FindAllStringSubmatch(text, -1)
	// Pick the largest amount (usually the total/main one).
	var best int64
	found := false
	for _, m := range matches {
		neg := m[1] != "" // leading minus or em-dash
		v, ok := rubleStringToCents(m[2], false)
		if !ok {
			continue
		}
		if neg {
			v = -v
		}
		if !found || abs64(v) > abs64(best) {
			best = v
			found = true
		}
	}
	return best, found
}

func rubleStringToCents(s string, alreadyNeg bool) (int64, bool) {
	s = strings.TrimSpace(s)
	// Strip unicode/em minus at the start, we track sign separately.
	neg := alreadyNeg || strings.HasPrefix(s, "-") || strings.HasPrefix(s, "−") || strings.HasPrefix(s, "—")
	s = strings.TrimLeft(s, "-−— \t")
	// Remove thousands separators (spaces, nbsp).
	s = strings.Join(strings.Fields(s), "")
	// Normalise decimal separator.
	s = strings.ReplaceAll(s, ",", ".")
	// Strip any leftover non-numeric characters except dot.
	var b strings.Builder
	for _, r := range s {
		if (r >= '0' && r <= '9') || r == '.' {
			b.WriteRune(r)
		}
	}
	f, err := strconv.ParseFloat(b.String(), 64)
	if err != nil || f == 0 {
		return 0, false
	}
	cents := int64(math.Round(f * 100))
	if neg {
		cents = -cents
	}
	return cents, true
}

func extractOCRDate(text string) time.Time {
	now := time.Now()
	hour, minute := 12, 0
	if m := timeRe.FindStringSubmatch(text); len(m) > 2 {
		hour, _ = strconv.Atoi(m[1])
		minute, _ = strconv.Atoi(m[2])
	}

	if m := ruDateRe.FindStringSubmatch(text); len(m) > 3 {
		day, _ := strconv.Atoi(m[1])
		month := ruMonths[strings.ToLower(m[2])]
		year, _ := strconv.Atoi(m[3])
		return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
	}

	if m := numDateRe.FindStringSubmatch(text); len(m) > 3 {
		day, _ := strconv.Atoi(m[1])
		month, _ := strconv.Atoi(m[2])
		year, _ := strconv.Atoi(m[3])
		return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
	}

	return now
}

func extractOCRMerchant(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		runes := []rune(line)
		if len(runes) < 3 || len(runes) > 80 {
			continue
		}
		if skipLineRe.MatchString(line) {
			continue
		}
		if junkLineRe.MatchString(line) {
			continue
		}
		return line
	}
	return ""
}

func abs64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
