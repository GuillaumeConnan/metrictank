package conf

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alyu/configparser"
)

// Aggregations holds the aggregation definitions
type Aggregations struct {
	Data               []Aggregation
	DefaultAggregation Aggregation
}

type Aggregation struct {
	Name              string
	Pattern           *regexp.Regexp
	XFilesFactor      float64
	AggregationMethod []Method
}

// NewAggregations create instance of Aggregations
func NewAggregations() Aggregations {
	return Aggregations{
		Data: make([]Aggregation, 0),
		DefaultAggregation: Aggregation{
			Name:              "default",
			Pattern:           regexp.MustCompile(".*"),
			XFilesFactor:      0.5,
			AggregationMethod: []Method{Avg},
		},
	}
}

// ReadAggregations returns the defined aggregations from a storage-aggregation.conf file
// and adds the default
func ReadAggregations(file string) (Aggregations, error) {
	config, err := configparser.Read(file)
	if err != nil {
		return Aggregations{}, err
	}
	sections, err := config.AllSections()
	if err != nil {
		return Aggregations{}, err
	}

	result := NewAggregations()

	for _, s := range sections {
		item := Aggregation{}
		item.Name = strings.Trim(strings.SplitN(s.String(), "\n", 2)[0], " []")
		if item.Name == "" || strings.HasPrefix(item.Name, "#") {
			continue
		}

		item.Pattern, err = regexp.Compile(s.ValueOf("pattern"))
		if err != nil {
			return Aggregations{}, fmt.Errorf("[%s]: failed to parse pattern %q: %s", item.Name, s.ValueOf("pattern"), err.Error())
		}

		item.XFilesFactor, err = strconv.ParseFloat(s.ValueOf("xFilesFactor"), 64)
		if err != nil {
			return Aggregations{}, fmt.Errorf("[%s]: failed to parse xFilesFactor %q: %s", item.Name, s.ValueOf("xFilesFactor"), err.Error())
		}

		aggregationMethodStr := s.ValueOf("aggregationMethod")
		methodStrs := strings.Split(aggregationMethodStr, ",")
		for _, methodStr := range methodStrs {
			agg, err := NewMethod(methodStr)
			if err != nil {
				return result, fmt.Errorf("[%s]: %s", item.Name, err.Error())
			}
			item.AggregationMethod = append(item.AggregationMethod, agg)
		}

		result.Data = append(result.Data, item)
	}

	return result, nil
}

// Match returns the correct aggregation setting for the given metric
// it can always find a valid setting, because there's a default catch all
// also returns the index of the setting, to efficiently reference it
func (a Aggregations) Match(metric string) (uint16, Aggregation) {
	for i, s := range a.Data {
		if s.Pattern.MatchString(metric) {
			return uint16(i), s
		}
	}
	return uint16(len(a.Data)), a.DefaultAggregation
}

// Get returns the aggregation setting corresponding to the given index
func (a Aggregations) Get(i uint16) Aggregation {
	if i+1 > uint16(len(a.Data)) {
		return a.DefaultAggregation
	}
	return a.Data[i]
}
