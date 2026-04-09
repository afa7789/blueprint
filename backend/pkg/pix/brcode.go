package pix

import (
	"fmt"
)

// GeneratePayload generates a PIX BRCode EMV payload (static QR, no fixed amount)
func GeneratePayload(key, beneficiary, city string, amountCents int64) string {
	tlv := func(id, val string) string {
		return fmt.Sprintf("%s%02d%s", id, len(val), val)
	}

	merchantAccount := tlv("00", "br.gov.bcb.pix") + tlv("01", key)

	payload := tlv("00", "01") + // format indicator
		tlv("01", "12") + // static QR
		tlv("26", merchantAccount) +
		tlv("52", "0000") + // merchant category
		tlv("53", "986") // BRL

	// If amount provided, add it
	if amountCents > 0 {
		amount := fmt.Sprintf("%.2f", float64(amountCents)/100)
		payload += tlv("54", amount)
	}

	payload += tlv("58", "BR") +
		tlv("59", truncate(beneficiary, 25)) +
		tlv("60", truncate(city, 15)) +
		tlv("62", tlv("05", "***"))

	// CRC16 placeholder
	payload += "6304"
	crc := crc16(payload)
	return payload + crc
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

func crc16(str string) string {
	crc := uint16(0xFFFF)
	for i := 0; i < len(str); i++ {
		crc ^= uint16(str[i]) << 8
		for j := 0; j < 8; j++ {
			if crc&0x8000 != 0 {
				crc = (crc << 1) ^ 0x1021
			} else {
				crc = crc << 1
			}
		}
	}
	return fmt.Sprintf("%04X", crc&0xFFFF)
}
