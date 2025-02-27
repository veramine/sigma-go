package evaluator

import (
	"context"
	"testing"

	"github.com/bradleyjkemp/sigma-go"
)

func TestRuleEvaluator_Matches(t *testing.T) {
	rule := ForRule(sigma.Rule{
		Detection: sigma.Detection{
			Searches: map[string]sigma.Search{
				"foo": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "foo-field",
								Values: []string{
									"foo-value",
								},
							},
						},
					},
				},
				"bar": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "bar-field",
								Values: []string{
									"bar-value",
								},
							},
						},
					},
				},
				"baz": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "baz-field",
								Values: []string{
									"baz-value",
								},
							},
						},
					},
				},
				"null-field": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "non-existent-field",
								Values: []string{
									"null",
								},
							},
						},
					},
				},
			},
			Conditions: []sigma.Condition{
				{
					Search: sigma.And{
						sigma.SearchIdentifier{Name: "foo"},
						sigma.SearchIdentifier{Name: "bar"},
						sigma.SearchIdentifier{Name: "null-field"},
					},
				},
				{
					Search: sigma.AllOfThem{},
				},
			},
		},
	})

	result, err := rule.Matches(context.Background(), map[string]interface{}{
		"foo-field": "foo-value",
		"bar-field": "bar-value",
		"baz-field": "wrong-value",
	})
	switch {
	case err != nil:
		t.Fatal(err)
	case !result.Match:
		t.Error("rule should have matched", result.SearchResults)
	case !result.SearchResults["foo"] || !result.SearchResults["bar"] || result.SearchResults["baz"]:
		t.Error("expected foo and bar to be true but not baz")
	case !result.ConditionResults[0] || result.ConditionResults[1]:
		t.Error("expected first condition to be true and second condition to be false")
	}
}

func TestRuleEvaluator_Matches_WithPlaceholder(t *testing.T) {
	rule := ForRule(sigma.Rule{
		Detection: sigma.Detection{
			Searches: map[string]sigma.Search{
				"foo": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "foo-field",
								Values: []string{
									"%foo-placeholder%",
								},
							},
						},
					},
				},
			},
			Conditions: []sigma.Condition{
				{
					Search: sigma.SearchIdentifier{Name: "foo"},
				},
				{
					Search: sigma.AllOfThem{},
				},
			},
		},
	}, WithPlaceholderExpander(func(ctx context.Context, placeholderName string) ([]string, error) {
		if placeholderName != "%foo-placeholder%" {
			return nil, nil
		}

		return []string{"foo-value"}, nil
	}))

	result, err := rule.Matches(context.Background(), map[string]interface{}{
		"foo-field": "foo-value",
	})
	switch {
	case err != nil:
		t.Fatal(err)
	case !result.Match:
		t.Fatal("rule should have matched")
	}
}

func TestRuleEvaluator_Matches_Regex(t *testing.T) {
	rule := ForRule(sigma.Rule{
		Detection: sigma.Detection{
			Searches: map[string]sigma.Search{
				"foo": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field:     "foo-field",
								Modifiers: []string{"re"},
								Values: []string{
									"foo.*baz",
								},
							},
						},
					},
				},
			},
			Conditions: []sigma.Condition{
				{
					Search: sigma.SearchIdentifier{Name: "foo"},
				},
				{
					Search: sigma.AllOfThem{},
				},
			},
		},
	})

	result, err := rule.Matches(context.Background(), map[string]interface{}{
		"foo-field": "foo-bar-baz",
	})
	switch {
	case err != nil:
		t.Fatal(err)
	case !result.Match:
		t.Fatal("rule should have matched")
	}
}

func TestRuleEvaluator_Matches_CIDR(t *testing.T) {
	rule := ForRule(sigma.Rule{
		Detection: sigma.Detection{
			Searches: map[string]sigma.Search{
				"foo": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field:     "foo-field",
								Modifiers: []string{"cidr"},
								Values: []string{
									"10.0.0.0/8",
								},
							},
						},
					},
				},
			},
			Conditions: []sigma.Condition{
				{
					Search: sigma.SearchIdentifier{Name: "foo"},
				},
				{
					Search: sigma.AllOfThem{},
				},
			},
		},
	})

	result, err := rule.Matches(context.Background(), map[string]interface{}{
		"foo-field": "10.1.2.3",
	})
	switch {
	case err != nil:
		t.Fatal(err)
	case !result.Match:
		t.Fatal("rule should have matched")
	}
}

func TestRuleEvaluator_MatchesCaseInsensitive(t *testing.T) {
	rule := ForRule(sigma.Rule{
		Detection: sigma.Detection{
			Searches: map[string]sigma.Search{
				"foo": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "foo-field",
								Values: []string{
									"foo-value",
								},
							},
						},
					},
				},
				"bar": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "bar-field",
								Values: []string{
									"bAr-VaLuE",
								},
							},
						},
					},
				},
				"baz": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "baz-field",
								Values: []string{
									"baz-value",
								},
							},
						},
					},
				},
				"null-field": {
					EventMatchers: []sigma.EventMatcher{
						{
							{
								Field: "non-existent-field",
								Values: []string{
									"null",
								},
							},
						},
					},
				},
			},
			Conditions: []sigma.Condition{
				{
					Search: sigma.And{
						sigma.SearchIdentifier{Name: "foo"},
						sigma.SearchIdentifier{Name: "bar"},
						sigma.SearchIdentifier{Name: "null-field"},
					},
				},
				{
					Search: sigma.AllOfThem{},
				},
			},
		},
	})

	result, err := rule.Matches(context.Background(), map[string]interface{}{
		"foo-field": "FoO-vAlUe",
		"bar-field": "bar-value",
		"baz-field": "WrOnG-vAlUe",
	})
	switch {
	case err != nil:
		t.Fatal(err)
	case !result.Match:
		t.Error("rule should have matched", result.SearchResults)
	case !result.SearchResults["foo"] || !result.SearchResults["bar"] || result.SearchResults["baz"]:
		t.Error("expected foo and bar to be true but not baz")
	case !result.ConditionResults[0] || result.ConditionResults[1]:
		t.Error("expected first condition to be true and second condition to be false")
	}
}
