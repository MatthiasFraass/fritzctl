package fritz

import (
	"strconv"
)

func fmtTemperatureHkr(th string, min, max float64) string {
	f, err := strconv.ParseFloat(th, 64)
	if err != nil {
		return ""
	}
	f = cappedRawTemperature(f, min, max)

	switch {
	case f == 255:
		return "?"
	case f == 254:
		return "ON"
	case f == 253:
		return "OFF"
	default:
		return strconv.FormatFloat(f*0.5, 'f', -1, 64)
	}
}

// cappedRawTemperature limits the raw temperature values rage to be withing specific limits
// as described in:
// https://avm.de/fileadmin/user_upload/Global/Service/Schnittstellen/AHA-HTTP-Interface.pdf.
func cappedRawTemperature(v, min, max float64) float64 {
	// values above 253 are considered special
	if v > max && v < 253 {
		v = max
	} else if v < min {
		v = min
	}
	return v
}
