package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var readLocaltimeLink = os.Readlink

func formatLocalTimeDisplay(value time.Time) string {
	localValue := value.In(localTimeLocation())
	zoneName, _ := localValue.Zone()
	locationName := localTimezoneName(localValue.Location(), zoneName)

	if locationName == "" {
		return localValue.Format(time.RFC3339)
	}

	return fmt.Sprintf("%s (%s)", localValue.Format(time.RFC3339), locationName)
}

func formatLocalTimezoneDisplay(now time.Time) string {
	localNow := now.In(localTimeLocation())
	zoneName, offsetSeconds := localNow.Zone()
	locationName := localTimezoneName(localNow.Location(), zoneName)
	offset := formatUTCOffset(offsetSeconds)

	switch {
	case locationName != "" && locationName != "Local" && locationName != zoneName && zoneName != "":
		return fmt.Sprintf("%s (%s, %s)", locationName, zoneName, offset)
	case locationName != "" && locationName != "Local":
		return fmt.Sprintf("%s (%s)", locationName, offset)
	case zoneName != "":
		return fmt.Sprintf("%s (%s)", zoneName, offset)
	default:
		return offset
	}
}

func localTimeLocation() *time.Location {
	if time.Local == nil {
		return time.UTC
	}

	return time.Local
}

func localTimezoneName(location *time.Location, zoneName string) string {
	if location != nil {
		name := location.String()
		if name != "" && name != "Local" && strings.Contains(name, "/") {
			return name
		}
	}

	if readLocaltimeLink != nil {
		localtimePath, err := readLocaltimeLink("/etc/localtime")
		if err == nil {
			if name := parseLocaltimeZonePath(localtimePath); name != "" {
				return name
			}
		}
	}

	if location != nil {
		name := location.String()
		if name != "" && name != "Local" {
			return name
		}
	}

	return strings.TrimSpace(zoneName)
}

func parseLocaltimeZonePath(localtimePath string) string {
	const zoneInfoMarker = "/zoneinfo/"

	index := strings.Index(localtimePath, zoneInfoMarker)
	if index == -1 {
		return ""
	}

	return strings.TrimPrefix(localtimePath[index+len(zoneInfoMarker):], "/")
}

func formatUTCOffset(offsetSeconds int) string {
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		offsetSeconds = -offsetSeconds
	}

	hours := offsetSeconds / 3600
	minutes := (offsetSeconds % 3600) / 60

	return fmt.Sprintf("UTC%s%02d:%02d", sign, hours, minutes)
}
