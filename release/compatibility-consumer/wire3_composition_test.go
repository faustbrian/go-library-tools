package compatibilityconsumer_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	internationalwire "github.com/faustbrian/go-international/v4/adapters/wire"
	"github.com/faustbrian/go-international/v4/country"
	internationallegacy "github.com/faustbrian/go-international/v4/internationalwire"
	"github.com/faustbrian/go-international/v4/locale"
	localized "github.com/faustbrian/go-localized/v5"
	localizedwire "github.com/faustbrian/go-localized/v5/adapters/wire"
	localizedlegacy "github.com/faustbrian/go-localized/v5/localizedwire"
	"github.com/faustbrian/go-math/decimal"
	measurement "github.com/faustbrian/go-measurement/v3"
	measurementwire "github.com/faustbrian/go-measurement/v3/adapters/wire"
	measurementlegacy "github.com/faustbrian/go-measurement/v3/measurementwire"
	openinghours "github.com/faustbrian/go-opening-hours/v4"
	openingwire "github.com/faustbrian/go-opening-hours/v4/adapters/wire"
	openinglegacy "github.com/faustbrian/go-opening-hours/v4/openinghourswire"
	wire "github.com/faustbrian/go-wire/v3"
	"github.com/faustbrian/go-wire/v3/jsonwire"
	"github.com/faustbrian/go-wire/v3/msgpackwire"
	"github.com/faustbrian/go-wire/v3/tomlwire"
	"github.com/faustbrian/go-wire/v3/yamlwire"
)

func TestWire3PublishedLocalized5Composition(t *testing.T) {
	value, err := localized.TextFromMap(map[string]string{"EN-us": "Hello", "fi": ""})
	if err != nil {
		t.Fatal(err)
	}
	finnish, err := locale.Parse("fi")
	if err != nil {
		t.Fatal(err)
	}
	swedish, err := locale.Parse("sv")
	if err != nil {
		t.Fatal(err)
	}
	for _, codec := range []struct {
		name           string
		encode         func(localized.Text) ([]byte, error)
		retainedEncode func(localized.Text) ([]byte, error)
		decode         func([]byte) (localized.Text, error)
		retainedDecode func([]byte) (localized.Text, error)
	}{
		{"json", func(v localized.Text) ([]byte, error) { return localizedwire.EncodeJSON(v, jsonwire.EncodeOptions{}) },
			func(v localized.Text) ([]byte, error) { return localizedlegacy.EncodeJSON(v, jsonwire.EncodeOptions{}) },
			func(b []byte) (localized.Text, error) { return localizedwire.DecodeJSON(b, jsonwire.DecodeOptions{}) },
			func(b []byte) (localized.Text, error) { return localizedlegacy.DecodeJSON(b, jsonwire.DecodeOptions{}) }},
		{"yaml", func(v localized.Text) ([]byte, error) { return localizedwire.EncodeYAML(v, yamlwire.EncodeOptions{}) },
			func(v localized.Text) ([]byte, error) { return localizedlegacy.EncodeYAML(v, yamlwire.EncodeOptions{}) },
			func(b []byte) (localized.Text, error) { return localizedwire.DecodeYAML(b, yamlwire.DecodeOptions{}) },
			func(b []byte) (localized.Text, error) { return localizedlegacy.DecodeYAML(b, yamlwire.DecodeOptions{}) }},
		{"toml", func(v localized.Text) ([]byte, error) { return localizedwire.EncodeTOML(v, tomlwire.EncodeOptions{}) },
			func(v localized.Text) ([]byte, error) { return localizedlegacy.EncodeTOML(v, tomlwire.EncodeOptions{}) },
			func(b []byte) (localized.Text, error) { return localizedwire.DecodeTOML(b, tomlwire.DecodeOptions{}) },
			func(b []byte) (localized.Text, error) { return localizedlegacy.DecodeTOML(b, tomlwire.DecodeOptions{}) }},
		{"messagepack", func(v localized.Text) ([]byte, error) {
			return localizedwire.EncodeMessagePack(v, msgpackwire.EncodeOptions{})
		},
			func(v localized.Text) ([]byte, error) {
				return localizedlegacy.EncodeMessagePack(v, msgpackwire.EncodeOptions{})
			},
			func(b []byte) (localized.Text, error) {
				return localizedwire.DecodeMessagePack(b, msgpackwire.DecodeOptions{})
			},
			func(b []byte) (localized.Text, error) {
				return localizedlegacy.DecodeMessagePack(b, msgpackwire.DecodeOptions{})
			}},
	} {
		t.Run(codec.name, func(t *testing.T) {
			payload, err := codec.encode(value)
			if err != nil {
				t.Fatal(err)
			}
			retained, err := codec.retainedEncode(value)
			if err != nil || !bytes.Equal(payload, retained) {
				t.Fatal("canonical/retained encoded identity differs")
			}
			if codec.name == "json" && string(payload) != `{"en-US":"Hello","fi":""}` {
				t.Fatalf("canonical localized JSON = %q", payload)
			}
			for name, decode := range map[string]func([]byte) (localized.Text, error){"canonical": codec.decode, "retained": codec.retainedDecode} {
				t.Run(name, func(t *testing.T) {
					input := bytes.Clone(payload)
					decoded, err := decode(input)
					if err != nil || !decoded.Equal(value) {
						t.Fatalf("published codec round-trip failed: %v", err)
					}
					for i := range input {
						input[i] = 'x'
					}
					if !decoded.Equal(value) {
						t.Fatal("decoded localized text aliased input")
					}
					if got, present := decoded.Get(finnish); !present || got != "" {
						t.Fatal("present-empty locale lost")
					}
					if _, present := decoded.Get(swedish); present {
						t.Fatal("missing locale gained fallback")
					}
				})
			}
		})
	}
}

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
