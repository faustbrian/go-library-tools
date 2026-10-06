package compatibilityconsumer_test

import (
	"errors"
	"testing"
	"time"

	internationalwire "github.com/faustbrian/go-international/v4/adapters/wire"
	"github.com/faustbrian/go-international/v4/country"
	internationallegacy "github.com/faustbrian/go-international/v4/internationalwire"
	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v3"
	measurementwire "github.com/faustbrian/go-measurement/v3/adapters/wire"
	measurementlegacy "github.com/faustbrian/go-measurement/v3/measurementwire"
	openinghours "github.com/faustbrian/go-opening-hours/v4"
	openingwire "github.com/faustbrian/go-opening-hours/v4/adapters/wire"
	openinglegacy "github.com/faustbrian/go-opening-hours/v4/openinghourswire"
	wire "github.com/faustbrian/go-wire/v3"
)

func TestWire3PublishedInternational4Composition(t *testing.T) {
	finland, err := country.Parse("FI")
	if err != nil {
		t.Fatal(err)
	}
	type address struct {
		Country country.Code `json:"country"`
	}
	payload, err := internationalwire.Encode(wire.FormatJSON, address{Country: finland})
	if err != nil || string(payload) != `{"country":"FI"}` {
		t.Fatalf("published JSON = %q, error = %v", payload, err)
	}
	for name, decode := range map[string]func(wire.Format, []byte, any) error{
		"canonical": internationalwire.Decode,
		"retained":  internationallegacy.Decode,
	} {
		t.Run(name, func(t *testing.T) {
			var result address
			if err := decode(wire.FormatJSON, payload, &result); err != nil || result.Country.String() != "FI" {
				t.Fatalf("published country = %v, error = %v", result.Country, err)
			}
			if err := decode(wire.FormatSOAP, nil, &result); !errors.Is(err, internationalwire.ErrUnsupportedFormat) {
				t.Fatalf("unsupported format identity = %v", err)
			}
		})
	}
}

func TestWire3PublishedMeasurement3Composition(t *testing.T) {
	original := measurement.MustNew(decimal.MustParse("12.50"), measurement.Kilogram)
	options := measurementwire.Options{MaxBytes: 1024}
	for _, format := range []wire.Format{wire.FormatJSON, wire.FormatXML} {
		t.Run(string(format), func(t *testing.T) {
			payload, err := measurementwire.Encode(original, format, options)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := measurementwire.Decode(payload, format, options)
			if err != nil || decoded.String() != original.String() {
				t.Fatalf("canonical quantity = %v, error = %v", decoded, err)
			}
			retained, err := measurementlegacy.Decode(payload, format, measurementlegacy.Options{MaxBytes: 1024})
			if err != nil || retained.String() != original.String() {
				t.Fatalf("retained quantity = %v, error = %v", retained, err)
			}
			if _, err := measurementwire.Decode(payload, format, measurementwire.Options{MaxBytes: 1}); !errors.Is(err, wire.ErrSizeLimit) {
				t.Fatalf("published byte limit identity = %v", err)
			}
		})
	}
}

func TestWire3PublishedOpening4Composition(t *testing.T) {
	start, err := openinghours.NewLocalTime(9, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	end, err := openinghours.NewLocalTime(17, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	interval, err := openinghours.NewRange(start, end)
	if err != nil {
		t.Fatal(err)
	}
	monday, err := openinghours.OpenRanges([]openinghours.Range{interval}, openinghours.RejectOverlapAndAdjacent)
	if err != nil {
		t.Fatal(err)
	}
	schedule, err := openinghours.NewSchedule(openinghours.Config{
		Timezone: "UTC",
		Weekly:   map[time.Weekday]openinghours.DayRule{time.Monday: monday},
	})
	if err != nil {
		t.Fatal(err)
	}
	var format wire.Format = openingwire.WireFormat
	if format != wire.Format("opening-hours+json;v=1") {
		t.Fatalf("published format = %q", format)
	}
	if retained := openinglegacy.WireFormat; retained != format {
		t.Fatalf("retained format = %q, want %q", retained, format)
	}
	payload, err := (openingwire.Codec{}).Encode(schedule)
	if err != nil {
		t.Fatal(err)
	}
	for name, decode := range map[string]func([]byte) (openinghours.Schedule, error){
		"canonical": (openingwire.Codec{}).Decode,
		"retained":  (openinglegacy.Codec{}).Decode,
	} {
		t.Run(name, func(t *testing.T) {
			decoded, err := decode(payload)
			if err != nil {
				t.Fatal(err)
			}
			for _, test := range []struct {
				hour int
				open bool
			}{{9, true}, {16, true}, {17, false}} {
				result, err := decoded.IsOpen(time.Date(2026, time.January, 5, test.hour, 0, 0, 0, time.UTC))
				if err != nil || result.Open != test.open {
					t.Fatalf("Monday hour %d open = %v, want %v, error = %v", test.hour, result.Open, test.open, err)
				}
			}
		})
	}
}
