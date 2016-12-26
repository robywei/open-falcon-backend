package nqm

import (
	"encoding/json"
	"encoding/hex"
	ojson "github.com/Cepave/open-falcon-backend/common/json"
	sjson "github.com/bitly/go-simplejson"
	nqmModel "github.com/Cepave/open-falcon-backend/common/model/nqm"
	ocheck "github.com/Cepave/open-falcon-backend/common/testing/check"
	t "github.com/Cepave/open-falcon-backend/common/testing"
	"github.com/Cepave/open-falcon-backend/common/utils"
	. "gopkg.in/check.v1"
	"reflect"
	"time"
)

type TestQuerySuite struct{}

var _ = Suite(&TestQuerySuite{})

// Tests marshalling of JSON
func (suite *TestQuerySuite) TestJsonMarshal(c *C) {
	testCases := []struct {
		sampleTimeFilter *TimeFilter
	} {
		{
			&TimeFilter{
				timeRangeType: TimeRangeAbsolute,
				StartTime: ojson.JsonTime(time.Unix(2908001, 0)),
				EndTime: ojson.JsonTime(time.Unix(2909001, 0)),
				ToNow: &TimeWithUnit { "", 0 },
			},
		},
		{
			&TimeFilter{
				timeRangeType: TimeRangeRelative,
				StartTime: ojson.JsonTime{},
				EndTime: ojson.JsonTime{},
				ToNow: &TimeWithUnit { "d", 17 },
			},
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		sampleQuery := buildSampleQuery(testCase.sampleTimeFilter)

		jsonResult, e := json.Marshal(sampleQuery)
		c.Assert(e, IsNil, comment)

		c.Logf("[Case %d] JSON: %s", i + 1, string(jsonResult))

		testedQuery := NewCompoundQuery()
		testedQuery.UnmarshalJSON(jsonResult)
		testedQuery.SetupDefault()

		c.Assert(testedQuery.Filters.Time, DeepEquals, sampleQuery.Filters.Time, Commentf("query.filters is not equal"))
		c.Assert(testedQuery.Grouping, DeepEquals, sampleQuery.Grouping, Commentf("query.grouping is not equal"))
		c.Assert(testedQuery.Output, DeepEquals, sampleQuery.Output, Commentf("query.output is not equal"))
	}
}

func buildSampleQuery(timeFilter *TimeFilter) *CompoundQuery {
	return &CompoundQuery {
		Filters: &CompoundQueryFilter {
			Time: timeFilter,
			Agent: &nqmModel.AgentFilter{
				Name: []string { "CB1", "KC2" },
				Hostname: []string { "GA3", "ZC0" },
				IpAddress: []string { "10.9", "11.56.71.89" },
				ConnectionId: []string { "AB@13", "AC@13" },
				IspIds: []int16 { 11, 12 },
				ProvinceIds: []int16 { 5, 8, 9 },
				CityIds: []int16 { 31, 34 },
				NameTagIds: []int16 { 10, 19 },
				GroupTagIds: []int32 { 45, 51 },
			},
			Target: &nqmModel.TargetFilter{
				Name: []string { "CB1", "KC2" },
				Host: []string { "GA3", "ZC0" },
				IspIds: []int16 { 13, 17 },
				ProvinceIds: []int16 { 24, 39, 81 },
				CityIds: []int16 { 14, 23 },
				NameTagIds: []int16 { 39, 46 },
				GroupTagIds: []int32 { 61, 63 },
			},
			Metrics: "$max > 100 or $min < 30",
		},
		Grouping: &QueryGrouping {
			Agent: []string { AgentGroupingName, GroupingProvince },
			Target: []string { GroupingIsp },
		},
		Output: &QueryOutput {
			Metrics: []string { "min", "loss" },
		},
	}
}

// Tests the compression of query
func (suite *TestQuerySuite) TestGetCompressedQuery(c *C) {
	sampleQuery := buildSampleQuery(
		&TimeFilter { ToNow: &TimeWithUnit{ Unit: TimeUnitHour, Value: 77 } },
	)

	compressedQuery := sampleQuery.GetCompressedQuery()

	testedQuery := NewCompoundQuery()
	testedQuery.UnmarshalFromCompressedQuery(compressedQuery)

	c.Assert(testedQuery, DeepEquals, testedQuery)
}

// Tests the loading of filters.metrics
func (suite *TestQuerySuite) TestLoadMetricsOfFilters(c *C) {
	testCases := []struct {
		sampleJson string
		expectedResult string
	} {
		{
			`{ "filters": { "metrics": " $mAx > 20 aNd $min < 40 " } }`,
			"$max > 20 and $min < 40",
		},
		{
			`{ "filters": { "metrics": "" } }`, "",
		},
		{
			`{}`, "",
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedQuery := loadQueryObject(c, testCase.sampleJson, comment)
		c.Assert(testedQuery.Filters.Metrics, DeepEquals, testCase.expectedResult, comment)
	}
}

// Tests the loading of filters.time
func (suite *TestQuerySuite) TestLoadFiltersOfTime(c *C) {
	testCases := []struct {
		jsonSource string
		expectedResult *TimeFilter
	} {
		{
			`
			{
				"filters": {
					"time": {
						"start_time": 8977123,
						"end_time": 19082711
					}
				}
			}
			`,
			&TimeFilter {
				timeRangeType: TimeRangeAbsolute,
				StartTime: ojson.JsonTime(time.Unix(8977123, 0)),
				EndTime: ojson.JsonTime(time.Unix(19082711, 0)),
			},
		},
		{
			`
			{
				"filters": {
					"time": {
						"to_now": {
							"unit": "m",
							"value": 3
						}
					}
				}
			}
			`,
			&TimeFilter {
				timeRangeType: TimeRangeRelative,
				ToNow: &TimeWithUnit{
					Unit: TimeUnitMonth,
					Value: 3,
				},
			},
		},
		{ `{ "filters": {} }`, &TimeFilter{ timeRangeType: 0 }, },
		{ // "Zero" value of time.Time
			`{ "filters": { "start_time": -62135596800, "end_time": -62135596800, "to_now": { "unit": "", "value": -1 } } }`,
			&TimeFilter{ timeRangeType: 0 },
		},
		{ //
			`{ "filters": { "to_now": {} } }`,
			&TimeFilter{ timeRangeType: 0 },
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := loadQueryObject(c, testCase.jsonSource, comment).Filters.Time
		expectedResult := testCase.expectedResult

		c.Assert(testedResult.timeRangeType, Equals, expectedResult.timeRangeType, comment)
		switch testedResult.timeRangeType {
		case TimeRangeAbsolute:
			c.Assert(testedResult.StartTime, Equals, expectedResult.StartTime, comment)
			c.Assert(testedResult.EndTime, Equals, expectedResult.EndTime, comment)
		case TimeRangeRelative:
			testedTimeWithUnit := testedResult.ToNow
			c.Assert(testedTimeWithUnit.Unit, Equals, expectedResult.ToNow.Unit, comment)
			c.Assert(testedTimeWithUnit.Value, Equals, expectedResult.ToNow.Value, comment)
		}
	}
}

// Tests the loading of filters.agent
func (suite *TestQuerySuite) TestLoadFiltersOfAgent(c *C) {
	testCases := []struct {
		jsonSource string
		expectedFilter *nqmModel.AgentFilter
	} {
		{
			`
			{ "filters": {
				"agent": {
					"name": [ "G1", "C2", "K3", "G1"],
					"hostname": [ "hs-3", "hs-1", "hs-3" ],
					"ip_address": [ "10.20", "9.7", "10.20" ],
					"connection_id": [ "conn-id-3", "conn-id-1", "conn-id-3" ],
					"isp_ids": [ 20, 17, 20 ],
					"province_ids": [ 31, 22, 31 ],
					"city_ids": [ 32, 7, 32 ],
					"name_tag_ids": [ 77, 9, 77 ],
					"group_tag_ids": [ 16, 8, 16 ]
				}
			} }
			`,
			&nqmModel.AgentFilter {
				Name: []string { "C2", "G1", "K3" },
				Hostname: []string{ "hs-1", "hs-3" },
				IpAddress: []string{ "10.20", "9.7" },
				ConnectionId: []string{ "conn-id-1", "conn-id-3" },
				IspIds: []int16{ 17, 20 },
				ProvinceIds: []int16{ 22, 31 },
				CityIds: []int16{ 7, 32 },
				NameTagIds: []int16{ 9, 77 },
				GroupTagIds: []int32{ 8, 16 },
			},
		},
		{
			`{ "filters": { "agent": {} } }`,
			&nqmModel.AgentFilter {
				Name: []string {},
				Hostname: []string{},
				IpAddress: []string{},
				ConnectionId: []string{},
				IspIds: []int16{},
				ProvinceIds: []int16{},
				CityIds: []int16{},
				NameTagIds: []int16{},
				GroupTagIds: []int32{},
			},
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := loadQueryObject(c, testCase.jsonSource, comment)
		c.Assert(testedResult.Filters.Agent, DeepEquals, testCase.expectedFilter, comment)
	}
}

// Tests the loading of filters.target
func (suite *TestQuerySuite) TestLoadFiltersOfTarget(c *C) {
	testCases := []struct {
		jsonSource string
		expectedFilter *nqmModel.TargetFilter
	} {
		{
			`
			{ "filters": {
				"target": {
					"name": [ "G1", "C2", "K3", "G1"],
					"host": [ "hs-3", "hs-1", "hs-3" ],
					"isp_ids": [ 20, 17, 20 ],
					"province_ids": [ 31, 22, 31 ],
					"city_ids": [ 32, 7, 32 ],
					"name_tag_ids": [ 77, 9, 77 ],
					"group_tag_ids": [ 16, 8, 16 ]
				}
			} }
			`,
			&nqmModel.TargetFilter {
				Name: []string { "C2", "G1", "K3" },
				Host: []string{ "hs-1", "hs-3" },
				IspIds: []int16{ 17, 20 },
				ProvinceIds: []int16{ 22, 31 },
				CityIds: []int16{ 7, 32 },
				NameTagIds: []int16{ 9, 77 },
				GroupTagIds: []int32{ 8, 16 },
			},
		},
		{
			`{ "filters": {"target": {} } }`,
			&nqmModel.TargetFilter {
				Name: []string {},
				Host: []string{},
				IspIds: []int16{},
				ProvinceIds: []int16{},
				CityIds: []int16{},
				NameTagIds: []int16{},
				GroupTagIds: []int32{},
			},
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := loadQueryObject(c, testCase.jsonSource, comment)
		c.Assert(testedResult.Filters.Target, DeepEquals, testCase.expectedFilter, comment)
	}
}

// Tests the query of loading output
func (suite *TestQuerySuite) TestLoadOutput(c *C) {
	testCases := []*struct{
		sampleJson string
		expectedResult []string
	} {
		{
			`{ "output": { "metrics": [ "max", "min", "avg" ] } }`,
			[]string { "max", "min", "avg" },
		},
		{
			`{ "output": { "metrics": [] } }`,
			[]string {},
		},
		{ // No output property
			`{}`,
			[]string {},
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedQuery := loadQueryObject(c, testCase.sampleJson, comment)

		c.Assert(testedQuery.Output.Metrics, DeepEquals, testCase.expectedResult, comment)
	}
}

// Tests the loading of grouping
func (suite *TestQuerySuite) TestLoadGrouping(c *C) {
	testCases := []struct {
		sampleJson string
		expectedAgentGrouping []string
		expectedTargetGrouping []string
	} {
		{
			`{ "grouping": { "agent": [ "isp", "province" ], "target": [ "name_tag"] } }`,
			[]string{ "isp", "province" },
			[]string{ "name_tag" },
		},
		{
			`{ "grouping": { "agent": [], "target": [] } }`,
			[]string{},
			[]string{},
		},
		{ // No output property
			`{}`,
			[]string{},
			[]string{},
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedQuery := loadQueryObject(c, testCase.sampleJson, comment)

		c.Assert(testedQuery.Grouping.Agent, DeepEquals, testCase.expectedAgentGrouping, comment)
		c.Assert(testedQuery.Grouping.Target, DeepEquals, testCase.expectedTargetGrouping, comment)
	}
}

// Tests the purifying of json's array of numbers
func (suite *TestQuerySuite) TestPurifyNumberArrayOfJson(c *C) {
	testCases := []struct {
		jsonSource string
		targetType reflect.Type
		expectedResult interface{}
	} {
		{
			`[ 38, 29, 40, 38, 29 ]`,
			utils.TypeOfInt8,
			[]int8 { 29, 38, 40 },
		},
		{
			`[ 78781, 10981, 78781 ]`,
			utils.TypeOfUint64,
			[]uint64 { 10981, 78781 },
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := purifyNumberArrayOfJson(
			loadJson(c, testCase.jsonSource),
			testCase.targetType,
		)
		c.Assert(testedResult, DeepEquals, testCase.expectedResult, comment)
	}
}

// Tests the purifying for array of strings(by doamin)
func (suite *TestQuerySuite) TestPurifyStringArrayOfJsonForDomain(c *C) {
	testCases := []struct {
		jsonSource string
		sampleDomain map[string]bool
		expectedData []string
	} {
		{
			`[ "c9", " A1 ", "  ", "no-1", "C1", "c2", "no-3", "A1", " C1 " ]`,
			map[string]bool { "a1": true, "c1": true, "c9": true },
			[]string{ "c9", "a1", "c1" },
		},
		{ `[ "A1", "A2" ]`, map[string]bool {}, []string{}, },
		{ `[ "A1", "A2" ]`, nil, []string{}, },
		{ `[ "  ", "" ]`, map[string]bool { "k1": true }, []string{}, },
		{ "", map[string]bool { "g1": true }, []string{}, },
		{ "", nil, []string{}, },
		{ "null", nil, []string{}, },
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := purifyStringArrayOfJsonForDomain(
			loadJson(c, testCase.jsonSource),
			testCase.sampleDomain,
		)

		c.Assert(testedResult, DeepEquals, testCase.expectedData, comment)
	}
}

// Tests the purifying for array of strings(for values)
func (suite *TestQuerySuite) TestPurifyStringArrayOfJsonForValues(c *C) {
	testCases := []struct {
		jsonSource string
		exepctedResult []string
	} {
		{
			`[ " A1 ", " b1 ", "A1", "B2", " a1 ", "B2", "C3" ]`,
			[]string{ "A1", "B2", "C3", "a1", "b1" },
		},
		{ `[ "  ", ""]`, []string{}, },
		{ "", []string{}, },
		{ "null", []string{}, },
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := purifyStringArrayOfJsonForValues(
			loadJson(c, testCase.jsonSource),
		)
		c.Assert(testedResult, DeepEquals, testCase.exepctedResult, comment)
	}
}

// Tests the diguest for content of query
func (suite *TestQuerySuite) TestGetDigestValue(c *C) {
	sampleQuery := NewCompoundQuery()
	sampleQuery.Filters.Time.StartTime = ojson.JsonTime(time.Unix(12890090, 0))
	sampleQuery.Filters.Time.EndTime = ojson.JsonTime(time.Unix(12930090, 0))
	sampleQuery.Filters.Time.timeRangeType = TimeRangeAbsolute
	sampleQuery.Filters.Agent.CityIds = []int16 { 18, 92, 154 }
	sampleQuery.Filters.Target.IspIds = []int16 { 8, 192, 103 }
	sampleQuery.Grouping.Agent = []string { AgentGroupingIpAddress }
	sampleQuery.Grouping.Target = []string { GroupingCity }
	sampleQuery.Output.Metrics = []string{ MetricMax, MetricLoss, MetricNumAgent }

	hexValue := hex.EncodeToString(sampleQuery.GetDigestValue())
	c.Logf("Query digest: [%s]", hexValue)

	c.Assert(hexValue, Equals, "a70a2928c1d5226f9ba262aafc910f4c")
}

// Tests the digesting for time filter
func (suite *TestQuerySuite) TestDigestingOfTimeFilter(c *C) {
	testCases := []struct {
		sampleFilter *TimeFilter
		expectedDigest string
	} {
		{
			&TimeFilter {
				StartTime: ojson.JsonTime(time.Unix(789907610, 0)),
				EndTime: ojson.JsonTime(time.Unix(789937610, 0)),
				timeRangeType: TimeRangeAbsolute,
			},
			"61b6d73fb22673ff28746c847eaef593",
		},
		{
			&TimeFilter {
				ToNow: &TimeWithUnit { Unit: TimeUnitYear, Value: 3, },
				timeRangeType: TimeRangeRelative,
			},
			"580f6dacf6ac8d59d5ad86d7f0286cf6",
		},
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedDigestValue := hex.EncodeToString(testCase.sampleFilter.GetDigest())
		c.Logf("Time filter: [%v]. Digest value: [%s]", testCase.sampleFilter, testedDigestValue)

		c.Assert(testedDigestValue, Equals, testCase.expectedDigest, comment)
	}
}

// Tests the getting of time range of net
func (suite *TestQuerySuite) TestGetRelativeTimeRangeOfNet(c *C) {
	testCases := []struct {
		sampleTimeWithUnit *TimeWithUnit
		expectedStartTime string
		expectedEndTime string
	} {
		{
			&TimeWithUnit{ Unit: TimeUnitYear, Value: 2 },
			"2012-01-01T00:00:00Z", "2014-01-01T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitYear, Value: 0 },
			"2014-01-01T00:00:00Z", "2015-01-01T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitMonth, Value: 2 },
			"2014-04-01T00:00:00Z", "2014-06-01T00:00:00Z",
		},
		{ // Cross year for months
			&TimeWithUnit{ Unit: TimeUnitMonth, Value: 6 },
			"2013-12-01T00:00:00Z", "2014-06-01T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitMonth, Value: 0 },
			"2014-06-01T00:00:00Z", "2014-07-01T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitWeek, Value: 2 },
			"2014-05-26T00:00:00Z", "2014-06-09T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitWeek, Value: 0 },
			"2014-06-09T00:00:00Z", "2014-06-16T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitDay, Value: 2 },
			"2014-06-08T00:00:00Z", "2014-06-10T00:00:00Z",
		},
		{ // Cross month for days
			&TimeWithUnit{ Unit: TimeUnitDay, Value: 10 },
			"2014-05-31T00:00:00Z", "2014-06-10T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitDay, Value: 0 },
			"2014-06-10T00:00:00Z", "2014-06-11T00:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitHour, Value: 2 },
			"2014-06-10T08:00:00Z", "2014-06-10T10:00:00Z",
		},
		{ // Cross day for hours
			&TimeWithUnit{ Unit: TimeUnitHour, Value: 11 },
			"2014-06-09T23:00:00Z", "2014-06-10T10:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitHour, Value: 0 },
			"2014-06-10T10:00:00Z", "2014-06-10T11:00:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitMinute, Value: 7 },
			"2014-06-10T10:13:00Z", "2014-06-10T10:20:00Z",
		},
		{ // Corss hour for minutes
			&TimeWithUnit{ Unit: TimeUnitMinute, Value: 25 },
			"2014-06-10T09:55:00Z", "2014-06-10T10:20:00Z",
		},
		{
			&TimeWithUnit{ Unit: TimeUnitMinute, Value: 0 },
			"2014-06-10T10:20:00Z", "2014-06-10T10:21:00Z",
		},
	}

	baseTime := t.ParseTime(c, "2014-06-10T10:20:30Z")
	c.Logf("Base time: %s", baseTime.Format(time.RFC3339))
	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d.", i + 1)

		testedTime := &TimeFilter {
			ToNow: testCase.sampleTimeWithUnit,
		}

		expectedStartTime, expectedEndTime :=
			t.ParseTime(c, testCase.expectedStartTime),
			t.ParseTime(c, testCase.expectedEndTime)
		testedStartTime, testedEndTime := testedTime.getRelativeTimeRangeOfNet(baseTime)

		c.Logf(
			"Case [%d]. [%v] Start time: %s. End time: %s",
			i + 1, testedTime.ToNow,
			testedStartTime.Format(time.RFC3339), testedEndTime.Format(time.RFC3339),
		)
		c.Assert(
			testedStartTime, ocheck.TimeEquals, expectedStartTime,
			Commentf("%s Start time.", comment.CheckCommentString()),
		)
		c.Assert(
			testedEndTime, ocheck.TimeEquals, expectedEndTime,
			Commentf("%s End time.", comment.CheckCommentString()),
		)
	}
}

// Tests the whether the grouping is for each agent
func (suite *TestQuerySuite) TestIsForEachAgentOfGrouping(c *C) {
	testCases := []struct {
		agentGrouping []string
		expected bool
	} {
		{ []string { GroupingIsp, GroupingCity, GroupingProvince, GroupingNameTag }, false, },
		{ []string { GroupingCity, AgentGroupingName }, true, },
		{ []string { GroupingCity, AgentGroupingHostname }, true, },
		{ []string { GroupingCity, AgentGroupingIpAddress }, true, },
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := (&QueryGrouping{ Agent: testCase.agentGrouping }).IsForEachAgent()
		c.Assert(testedResult, Equals, testCase.expected, comment)
	}
}

// Tests the whether the grouping is for each target
func (suite *TestQuerySuite) TestIsForEachTargetOfGrouping(c *C) {
	testCases := []struct {
		targetGrouping []string
		expected bool
	} {
		{ []string { GroupingIsp, GroupingCity, GroupingProvince, GroupingNameTag }, false, },
		{ []string { GroupingCity, TargetGroupingName }, true, },
		{ []string { GroupingCity, TargetGroupingHost }, true, },
	}

	for i, testCase := range testCases {
		comment := Commentf("Test Case: %d", i + 1)

		testedResult := (&QueryGrouping{ Target: testCase.targetGrouping }).IsForEachTarget()
		c.Assert(testedResult, Equals, testCase.expected, comment)
	}
}

func loadJson(c *C, jsonSource string) *sjson.Json {
	if jsonSource == "" {
		return nil
	}

	sampleJson, err := sjson.NewJson(([]byte)(jsonSource))
	c.Assert(err, IsNil)

	return sampleJson
}

func loadQueryObject(c *C, json string, comment CommentInterface) *CompoundQuery {
	testedQuery := NewCompoundQuery()
	err := testedQuery.UnmarshalJSON(([]byte)(json))
	c.Assert(err, IsNil, comment)

	return testedQuery
}
