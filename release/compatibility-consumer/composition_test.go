package compatibilityconsumer_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	apiquery "github.com/faustbrian/go-api-query/v4"
	queryjsonapi "github.com/faustbrian/go-api-query/v4/adapters/jsonapi"
	queryjsonapilegacy "github.com/faustbrian/go-api-query/v4/apiqueryjsonapi"
	calendar "github.com/faustbrian/go-calendar/v2"
	configdecode "github.com/faustbrian/go-config/v2/decode"
	idempotency "github.com/faustbrian/go-idempotency/v2"
	idempotencymemory "github.com/faustbrian/go-idempotency/v2/memory"
	"github.com/faustbrian/go-international/v3/locale"
	jsonapi "github.com/faustbrian/go-jsonapi/v2"
	moneyobjective "github.com/faustbrian/go-knapsack/objective/money/v3"
	"github.com/faustbrian/go-knapsack/v2"
	localized "github.com/faustbrian/go-localized/v4"
	localizedquery "github.com/faustbrian/go-localized/v4/adapters/query"
	localizedquerylegacy "github.com/faustbrian/go-localized/v4/localizedquery"
	log "github.com/faustbrian/go-log/v2"
	mpt "github.com/faustbrian/go-merkle-patricia-trie/v2"
	mptmemory "github.com/faustbrian/go-merkle-patricia-trie/v2/memory"
	"github.com/faustbrian/go-money/v2"
	"github.com/faustbrian/go-money/v2/moneytest"
	openinghours "github.com/faustbrian/go-opening-hours/v3"
	openingcalendar "github.com/faustbrian/go-opening-hours/v3/adapters/calendar"
	openingtemporal "github.com/faustbrian/go-opening-hours/v3/adapters/temporal"
	openingcalendarlegacy "github.com/faustbrian/go-opening-hours/v3/openinghourscalendar"
	openingtemporallegacy "github.com/faustbrian/go-opening-hours/v3/openinghourstemporal"
	openrpc "github.com/faustbrian/go-openrpc/v2"
	"github.com/faustbrian/go-openrpc/v2/builder"
	"github.com/faustbrian/go-openrpc/v2/validate"
	tabular "github.com/faustbrian/go-tabular/v2"
	temporal "github.com/faustbrian/go-temporal/v2"
	temporalconfig "github.com/faustbrian/go-temporal/v2/adapters/config"
	temporalvalidation "github.com/faustbrian/go-temporal/v2/adapters/validation"
	"github.com/faustbrian/go-temporal/v2/dateperiod"
	"github.com/faustbrian/go-temporal/v2/instant"
	"github.com/faustbrian/go-temporal/v2/timeofday"
	"github.com/faustbrian/go-tenancy/v2"
	validation "github.com/faustbrian/go-validation/v2"
)

func TestLocalizedV4PublishedAPIQueryV4Composition(t *testing.T) {
	text, err := localized.TextFromMap(map[string]string{"en": "Hello", "fi": ""})
	if err != nil {
		t.Fatal(err)
	}
	english, err := locale.Parse("en")
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
	for _, adapter := range []struct {
		name      string
		value     func(localized.Text, locale.Tag) (apiquery.Value, bool)
		predicate func(string, apiquery.Operator, localized.Text, locale.Tag) (*apiquery.Predicate, error)
	}{
		{"canonical", localizedquery.ExactValue, localizedquery.ExactPredicate},
		{"retained", localizedquerylegacy.ExactValue, localizedquerylegacy.ExactPredicate},
	} {
		t.Run(adapter.name, func(t *testing.T) {
			if value, present := adapter.value(text, english); !present || value.Type() != apiquery.TypeString || value.String() != "Hello" {
				t.Fatal("exact localized value lost APIQuery v4 string semantics")
			}
			if value, present := adapter.value(text, finnish); !present || value.Type() != apiquery.TypeString || value.String() != "" {
				t.Fatal("present-empty localized value became absent")
			}
			predicate, err := adapter.predicate("title", apiquery.OpEqual, text, english)
			want := &apiquery.Predicate{Name: "title", Operator: apiquery.OpEqual, Values: []apiquery.Value{apiquery.StringValue("Hello")}}
			if err != nil || !reflect.DeepEqual(predicate, want) {
				t.Fatal("exact localized predicate lost APIQuery v4 nominal semantics")
			}
			if _, present := adapter.value(text, swedish); present {
				t.Fatal("exact localized lookup applied fallback")
			}
			if _, err := adapter.predicate("title", apiquery.OpEqual, text, swedish); !errors.Is(err, localized.ErrMissingLocale) {
				t.Fatal("missing localized predicate lost sentinel semantics")
			}
		})
	}
}

func TestTabularV2PublishedBoundedCSVComposition(t *testing.T) {
	reader, err := tabular.NewCSVReader(strings.NewReader("Name,City\nAda,Helsinki\n"), tabular.DelimitedConfig{
		FieldsPerRecord: 2,
		MaxSourceBytes:  128,
		MaxRecordBytes:  64,
		MaxFieldBytes:   32,
		Header: &tabular.HeaderConfig{
			Case: tabular.HeaderCaseLower, RejectEmpty: true, RejectDuplicates: true,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	header, err := reader.Header()
	if err != nil || !reflect.DeepEqual(header, tabular.Row{"name", "city"}) {
		t.Fatalf("published Tabular v2 normalized header = %v, %v", header, err)
	}
	row, err := reader.Read()
	if err != nil || !reflect.DeepEqual(row, tabular.Row{"Ada", "Helsinki"}) {
		t.Fatalf("published Tabular v2 finite CSV row = %v, %v", row, err)
	}
	if _, err := reader.Read(); !errors.Is(err, io.EOF) {
		t.Fatalf("published Tabular v2 clean input end = %v", err)
	}
}

func TestAPIQueryV4PublishedJSONAPIV2Composition(t *testing.T) {
	pagination, err := jsonapi.NewCursorPagination(jsonapi.CursorPaginationConfig{DefaultSize: 5, MaxSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	query, err := jsonapi.ParseQuery(url.Values{
		"fields[orders]": {"status"}, "filter[status]": {"paid"}, "sort": {"id"}, "page[size]": {"10"},
	})
	if err != nil {
		t.Fatal(err)
	}
	decodeFilter := func(family jsonapi.ParameterFamily) (*apiquery.FilterExpr, error) {
		return &apiquery.FilterExpr{Predicate: &apiquery.Predicate{
			Name: "status", Operator: apiquery.OpEqual, Values: []apiquery.Value{apiquery.StringValue(family["filter[status]"][0])},
		}}, nil
	}
	decodePage := func(family jsonapi.ParameterFamily) (apiquery.PageRequest, error) {
		page, err := pagination.Parse(family)
		return apiquery.PageRequest{Mode: apiquery.PageCursor, Size: page.Size}, err
	}
	preferred, err := queryjsonapi.FromQuery(query, queryjsonapi.Config{Resource: "orders", DecodeFilter: decodeFilter, DecodePage: decodePage})
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := queryjsonapilegacy.FromQuery(query, queryjsonapilegacy.Config{Resource: "orders", DecodeFilter: decodeFilter, DecodePage: decodePage})
	if err != nil {
		t.Fatal(err)
	}
	want := apiquery.Request{
		Fields: apiquery.Present([]string{"status"}),
		Sorts:  apiquery.Present([]apiquery.SortTerm{{Name: "id", Direction: apiquery.Ascending}}),
		Filter: &apiquery.FilterExpr{Predicate: &apiquery.Predicate{Name: "status", Operator: apiquery.OpEqual, Values: []apiquery.Value{apiquery.StringValue("paid")}}},
		Page:   apiquery.PageRequest{Mode: apiquery.PageCursor, Size: 10},
	}
	if !reflect.DeepEqual(preferred, want) || !reflect.DeepEqual(legacy, want) {
		t.Fatal("published APIQuery v4 bridges lost JSONAPI v2 field, filter, sort or finite page semantics")
	}
}

func TestJSONAPIV2PublishedFiniteCursorComposition(t *testing.T) {
	pagination, err := jsonapi.NewCursorPagination(jsonapi.CursorPaginationConfig{DefaultSize: 5, MaxSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	query, err := jsonapi.ParseQuery(url.Values{
		"fields[articles]": {"title"}, "sort": {"id"}, "page[size]": {"10"},
	})
	if err != nil {
		t.Fatal(err)
	}
	page, err := pagination.ParseQuery(query)
	if err != nil {
		t.Fatal(err)
	}
	if page.Size != 10 || !page.SizePresent || page.PageMember != "page" || page.Range || page.AfterPresent || page.BeforePresent {
		t.Fatal("published JSONAPI v2 lost finite initial-page request semantics")
	}
	defaults, err := pagination.Parse(jsonapi.ParameterFamily{})
	if err != nil || defaults.Size != 5 || defaults.SizePresent {
		t.Fatal("published JSONAPI v2 lost configured default page size")
	}
	if len(query.Fields["articles"]) != 1 || query.Fields["articles"][0] != "title" || len(query.Sort) != 1 || query.Sort[0].Name != "id" {
		t.Fatal("cursor parsing changed ordinary query fields or sorting")
	}
}

func TestMPTV2PublishedMemoryComposition(t *testing.T) {
	ctx := context.Background()
	original, err := mpt.NewRawTrie(mpt.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	key, input := []byte("key"), []byte("ordinary")
	changed, err := original.Update(ctx, key, input)
	if err != nil {
		t.Fatal(err)
	}
	input[0] = 'X'
	if _, err := original.Get(ctx, key); !errors.Is(err, mpt.ErrAbsentKey) {
		t.Fatalf("Update changed original snapshot: %v", err)
	}
	value, err := changed.Get(ctx, key)
	if err != nil || !bytes.Equal(value, []byte("ordinary")) {
		t.Fatalf("Update retained caller value alias: %q, %v", value, err)
	}
	store := mptmemory.New()
	committed, err := changed.Commit(ctx, store)
	if err != nil {
		t.Fatal(err)
	}
	changedRoot, err := changed.Root()
	if err != nil {
		t.Fatal(err)
	}
	committedRoot, err := committed.Root()
	if err != nil {
		t.Fatal(err)
	}
	if store.Root() != committedRoot || committedRoot != changedRoot {
		t.Fatal("Commit changed immutable root identity")
	}
	loaded, err := mpt.LoadRawTrie(store.Root(), store, mpt.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	value, err = loaded.Get(ctx, key)
	if err != nil || !bytes.Equal(value, []byte("ordinary")) {
		t.Fatalf("public stored round trip: %q, %v", value, err)
	}
	value[0] = 'X'
	again, err := loaded.Get(ctx, key)
	if err != nil || !bytes.Equal(again, []byte("ordinary")) {
		t.Fatalf("Get exposed retained value storage: %q, %v", again, err)
	}
}

func TestOpeningV3PublishedNominalComposition(t *testing.T) {
	date := calendar.MustDate(2026, time.December, 25)
	canonicalDate, err := openingcalendar.FromDate(date)
	if err != nil || canonicalDate != date || canonicalDate.String() != "2026-12-25" {
		t.Fatalf("Opening v3 Calendar v2 date = %v, %v", canonicalDate, err)
	}
	legacyDate, err := openingcalendarlegacy.FromDate(date)
	if err != nil || legacyDate != canonicalDate {
		t.Fatalf("retained calendar facade date = %v, %v", legacyDate, err)
	}
	start, err := timeofday.New(9, 0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	end, err := timeofday.New(17, 0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	interval, err := timeofday.Between(start, end, temporal.ClosedOpen)
	if err != nil {
		t.Fatal(err)
	}
	canonicalRange, err := openingtemporal.RangeFromInterval(interval)
	if err != nil {
		t.Fatal(err)
	}
	legacyRange, err := openingtemporallegacy.RangeFromInterval(interval)
	if err != nil {
		t.Fatal(err)
	}
	wantStart, err := openinghours.NewLocalTime(9, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	wantEnd, err := openinghours.NewLocalTime(17, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if canonicalRange != legacyRange || canonicalRange.Start() != wantStart || canonicalRange.End() != wantEnd || canonicalRange.Overnight() {
		t.Fatal("Opening v3 lost literal 09:00–17:00 local range")
	}
	roundTrip, err := openingtemporal.IntervalFromRange(canonicalRange, 0)
	if err != nil || !roundTrip.Equal(interval) || roundTrip.Bounds() != temporal.ClosedOpen {
		t.Fatalf("Temporal v2 half-open round trip = %v, %v", roundTrip, err)
	}
	if openingcalendar.ErrInvalidInput != openingcalendarlegacy.ErrInvalidInput || openingtemporal.ErrLossyMapping != openingtemporallegacy.ErrLossyMapping {
		t.Fatal("retained facade changed sentinel identity")
	}
}

func TestTemporalV2PublishedNominalCompositionAndAdmission(t *testing.T) {
	date := calendar.MustDate(2026, time.October, 5)
	civil, err := dateperiod.New(date, date, temporal.Closed)
	if err != nil || civil.Start() != date || civil.Bounds() != temporal.Closed || civil.Days() != 1 {
		t.Fatalf("Calendar v2 closed singleton = %v, %v", civil, err)
	}
	var rule validation.Validator[dateperiod.Period] = temporalvalidation.DateNonEmpty()
	if report := rule.Validate(validation.Context{}, civil); !report.Empty() {
		t.Fatalf("Validation v2 singleton report = %v", report)
	}
	empty, err := dateperiod.New(date, date, temporal.Open)
	if err != nil || !rule.Validate(validation.Context{}, empty).HasCode("temporal_empty") {
		t.Fatalf("Validation v2 empty period classification = %v", err)
	}
	encoded, err := temporalconfig.NewDatePeriod(civil).MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var decoded temporalconfig.DatePeriod
	if err := configdecode.Value(string(encoded), &decoded); err != nil || decoded.Value() != civil {
		t.Fatalf("Config v2 date/bounds composition = %v, %v", decoded.Value(), err)
	}
	start := time.Date(2026, time.October, 5, 0, 0, 0, 0, time.UTC)
	first, err := instant.New(start, start.Add(time.Hour), temporal.OpenClosed)
	if err != nil || first.Includes(start) || !first.Includes(start.Add(time.Hour)) {
		t.Fatalf("instant endpoint bounds = %v, %v", first, err)
	}
	second, err := instant.New(start.Add(2*time.Hour), start.Add(3*time.Hour), temporal.Closed)
	if err != nil {
		t.Fatal(err)
	}
	left, err := instant.NewSet(temporal.Limits{InputPeriods: 1}, first)
	if err != nil {
		t.Fatal(err)
	}
	right, err := instant.NewSet(temporal.Limits{InputPeriods: 1}, second)
	if err != nil {
		t.Fatal(err)
	}
	refused, err := left.Union(right)
	var limit *temporal.LimitError
	if !errors.Is(err, temporal.ErrLimit) || !errors.As(err, &limit) || limit.Field != "input_periods" || limit.Value != 2 || limit.Max != 1 || refused.Len() != 0 {
		t.Fatalf("Union refusal = %v, %v; want empty typed 2/1 refusal", refused, err)
	}
	if left.Len() != 1 || right.Len() != 1 || left.Periods()[0] != first || right.Periods()[0] != second {
		t.Fatal("refused Union changed an operand")
	}
}

func TestTenancyV2PublishedTenantScopeComposition(t *testing.T) {
	tenant, err := tenancy.ParseTenantID("tenant-reference")
	if err != nil {
		t.Fatal(err)
	}
	scope, err := tenancy.NewTenantScope(tenant, tenancy.Metadata{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := tenancy.WithScope(context.Background(), scope)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := tenancy.RequireTenant(ctx)
	if err != nil || resolved.Value() != "tenant-reference" {
		t.Fatalf("published tenant scope identity = %v, %v", resolved, err)
	}
}

func TestCalendarV2PublishedCivilDateComposition(t *testing.T) {
	date, err := calendar.NewDate(2024, time.February, 29)
	if err != nil {
		t.Fatal(err)
	}
	if date.String() != "2024-02-29" || date.Year() != 2024 ||
		date.Month() != time.February || date.Day() != 29 ||
		date.Weekday() != time.Thursday || date.DayOfYear() != 60 {
		t.Fatal("published Calendar v2 lost literal leap-day identity")
	}
	text, err := date.MarshalText()
	if err != nil || string(text) != "2024-02-29" {
		t.Fatalf("MarshalText() = %q, %v", text, err)
	}
	var fromText calendar.Date
	if err := fromText.UnmarshalText([]byte("2024-02-29")); err != nil || !fromText.Equal(date) {
		t.Fatalf("UnmarshalText() = %v, %v", fromText, err)
	}
	encoded, err := json.Marshal(date)
	if err != nil || string(encoded) != `"2024-02-29"` {
		t.Fatalf("Marshal() = %s, %v", encoded, err)
	}
	var fromJSON calendar.Date
	if err := json.Unmarshal([]byte(`"2024-02-29"`), &fromJSON); err != nil || !fromJSON.Equal(date) {
		t.Fatalf("Unmarshal() = %v, %v", fromJSON, err)
	}
	next, err := date.AddDays(1)
	if err != nil || next.String() != "2024-03-01" || date.String() != "2024-02-29" {
		t.Fatalf("AddDays() = %v, %v; original = %v", next, err, date)
	}
}

type idempotencyCompositionClock struct{}

func (idempotencyCompositionClock) Now() time.Time {
	return time.Date(2026, time.October, 3, 0, 0, 0, 0, time.UTC)
}

func TestIdempotencyV2PatchAcquireCompleteReplayComposition(t *testing.T) {
	store, err := idempotencymemory.New(idempotencymemory.Options{
		Clock:       idempotencyCompositionClock{},
		OwnerTokens: func() (string, error) { return "00112233445566778899aabbccddeeff", nil },
		MaxRecords:  1,
	})
	if err != nil {
		t.Fatal(err)
	}
	service, err := idempotency.NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	key, err := idempotency.NewKey("orders", "tenant-one", "create", "application", "request-one")
	if err != nil {
		t.Fatal(err)
	}
	fingerprint, err := idempotency.NewFingerprint("order-v1", []byte("order-one"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	request := idempotency.BeginRequest{Acquire: idempotency.AcquireRequest{Key: key, Fingerprint: fingerprint, Lease: time.Minute}}
	first, err := service.Begin(ctx, request)
	if err != nil || !first.Execute || first.Outcome != idempotency.OutcomeAcquired {
		t.Fatalf("Begin() = %#v, %v", first, err)
	}
	want := []byte("created-order-one")
	completed, err := service.Complete(ctx, idempotency.CompleteRequest{Ownership: first.Record.Ownership(), Result: want})
	if err != nil || completed.State != idempotency.StateCompleted || !bytes.Equal(completed.Result, want) {
		t.Fatalf("Complete() = %#v, %v", completed, err)
	}
	replay, err := service.Begin(ctx, request)
	if err != nil || replay.Execute || replay.Outcome != idempotency.OutcomeReplayed || !bytes.Equal(replay.Record.Result, want) {
		t.Fatalf("Begin() replay = %#v, %v", replay, err)
	}
}

func TestOpenRPCV2BuilderValidationAndCanonicalComposition(t *testing.T) {
	version, err := openrpc.ParseVersion("1.4.1")
	if err != nil {
		t.Fatal(err)
	}
	info, err := openrpc.NewInfo(openrpc.InfoInput{Title: "Calculator", Version: "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	method, err := openrpc.NewMethod(openrpc.MethodInput{Name: "add", Params: []openrpc.ContentDescriptorOrReference{}})
	if err != nil {
		t.Fatal(err)
	}
	documentBuilder, err := builder.NewDocument(version, info)
	if err != nil {
		t.Fatal(err)
	}
	documentBuilder, err = documentBuilder.WithMethod(method)
	if err != nil {
		t.Fatal(err)
	}
	document, err := documentBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}
	report := validate.Document(context.Background(), document, validate.Options{MaxMethods: 1, MaxDiagnostics: 1})
	if !report.Valid() || report.Truncated() || len(report.Diagnostics()) != 0 {
		t.Fatalf("ordinary document validation = %#v", report)
	}
	encoded, err := openrpc.MarshalCanonical(document)
	if err != nil {
		t.Fatal(err)
	}
	const expected = `{"info":{"title":"Calculator","version":"1.0.0"},"methods":[{"name":"add","params":[]}],"openrpc":"1.4.1"}`
	if string(encoded) != expected {
		t.Fatalf("canonical output = %s, want %s", encoded, expected)
	}
}

func TestLogV2DefaultAndTrustedComposition(t *testing.T) {
	for _, test := range []struct {
		name    string
		trusted bool
		message string
	}{
		{name: "default", message: "[REDACTED]"},
		{name: "trusted", trusted: true, message: "application.ready"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var output bytes.Buffer
			construct := log.New
			if test.trusted {
				construct = log.TrustedNew
			}
			logger, err := construct(slog.NewJSONHandler(&output, nil))
			if err != nil {
				t.Fatal(err)
			}
			logger.Info("application.ready", slog.String("component", "worker"))
			var record map[string]any
			if err := json.Unmarshal(output.Bytes(), &record); err != nil {
				t.Fatal(err)
			}
			if record["msg"] != test.message || record["level"] != "INFO" {
				t.Fatalf("record message/level = %v/%v, want %s/INFO", record["msg"], record["level"], test.message)
			}
			component, present := record["component"]
			if present != test.trusted || (present && component != "worker") {
				t.Fatalf("component = %v (present=%t), trusted=%t", component, present, test.trusted)
			}
		})
	}
}

func TestMoneyV2CanonicalObjectiveV3Composition(t *testing.T) {
	fixture := moneytest.CurrencyFixtures()[0]
	euro, moneyContext := fixture.Code, fixture.Context
	unitCost, err := money.Parse("0.60", euro, moneyContext)
	if err != nil {
		t.Fatal(err)
	}
	input := map[string]money.Money{"box": unitCost}
	costs, err := moneyobjective.New(input)
	if err != nil {
		t.Fatal(err)
	}
	clear(input)
	plan, err := knapsack.NewPlan(knapsack.PlanSpec{
		Containers: []knapsack.ContainerInstance{
			{ID: "box-1", TypeID: "box"},
			{ID: "box-2", TypeID: "box"},
		},
		Status:      knapsack.StatusFeasible,
		Termination: knapsack.TerminationCompleted,
	})
	if err != nil {
		t.Fatal(err)
	}
	var total money.Money
	total, err = costs.Total(plan)
	if err != nil {
		t.Fatal(err)
	}
	if total.String() != "1.20 EUR" || total.Currency() != euro || total.Context() != moneyContext {
		t.Fatalf("Total after caller map mutation = %s, currency=%v, context=%v", total.String(), total.Currency(), total.Context())
	}
}
