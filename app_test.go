package main

import (
	"fmt"
	"testing"
)

func TestEvaluateLuceneQuery(t *testing.T) {
	app := &App{}

	// Test data - sample JSON records
	testRecord1 := JSONRecord{
		LineNumber: 1,
		Content: map[string]interface{}{
			"name":    "John Doe",
			"age":     30,
			"email":   "john.doe@example.com",
			"city":    "New York",
			"active":  true,
			"role":    "admin",
			"company": "TechCorp",
		},
		RawJSON: `{"name":"John Doe","age":30,"email":"john.doe@example.com","city":"New York","active":true,"role":"admin","company":"TechCorp"}`,
	}

	testRecord2 := JSONRecord{
		LineNumber: 2,
		Content: map[string]interface{}{
			"name":    "Jane Smith",
			"age":     25,
			"email":   "jane.smith@test.org",
			"city":    "Los Angeles",
			"active":  false,
			"role":    "user",
			"company": "StartupXYZ",
		},
		RawJSON: `{"name":"Jane Smith","age":25,"email":"jane.smith@test.org","city":"Los Angeles","active":false,"role":"user","company":"StartupXYZ"}`,
	}

	tests := []struct {
		name          string
		query         *LuceneQuery
		record        JSONRecord
		caseSensitive bool
		expected      bool
		description   string
	}{
		// Test nil query
		{
			name:          "NilQuery",
			query:         nil,
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false for nil query",
		},

		// Test field queries
		{
			name: "FieldQueryMatch",
			query: &LuceneQuery{
				Type:  "field",
				Field: "name",
				Value: "John",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match field value containing search term",
		},
		{
			name: "FieldQueryNoMatch",
			query: &LuceneQuery{
				Type:  "field",
				Field: "name",
				Value: "Bob",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should not match field value that doesn't contain search term",
		},
		{
			name: "FieldQueryCaseSensitive",
			query: &LuceneQuery{
				Type:  "field",
				Field: "name",
				Value: "john",
			},
			record:        testRecord1,
			caseSensitive: true,
			expected:      false,
			description:   "Should not match different case when case sensitive",
		},
		{
			name: "FieldQueryCaseInsensitive",
			query: &LuceneQuery{
				Type:  "field",
				Field: "name",
				Value: "john",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match different case when case insensitive",
		},
		{
			name: "FieldQueryNonExistentField",
			query: &LuceneQuery{
				Type:  "field",
				Field: "nonexistent",
				Value: "value",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false for non-existent field",
		},

		// Test term queries
		{
			name: "TermQueryMatch",
			query: &LuceneQuery{
				Type:  "term",
				Field: "email",
				Value: "example",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match term in specified field",
		},
		{
			name: "TermQueryGlobalMatch",
			query: &LuceneQuery{
				Type:  "term",
				Field: "",
				Value: "TechCorp",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match term in entire raw JSON when no field specified",
		},
		{
			name: "TermQueryGlobalNoMatch",
			query: &LuceneQuery{
				Type:  "term",
				Field: "",
				Value: "NotFound",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should not match term not found in raw JSON",
		},

		// Test phrase queries
		{
			name: "PhraseQueryFieldMatch",
			query: &LuceneQuery{
				Type:  "phrase",
				Field: "email",
				Value: "john.doe",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match phrase in specified field",
		},
		{
			name: "PhraseQueryGlobalMatch",
			query: &LuceneQuery{
				Type:  "phrase",
				Field: "",
				Value: "New York",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match phrase in entire raw JSON",
		},

		// Test wildcard queries
		{
			name: "WildcardQueryFieldMatch",
			query: &LuceneQuery{
				Type:  "wildcard",
				Field: "email",
				Value: "*.com",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match wildcard pattern in field",
		},
		{
			name: "WildcardQueryGlobalMatch",
			query: &LuceneQuery{
				Type:  "wildcard",
				Field: "",
				Value: "*TechCorp*",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match wildcard pattern in raw JSON",
		},
		{
			name: "WildcardQueryNoMatch",
			query: &LuceneQuery{
				Type:  "wildcard",
				Field: "email",
				Value: "*.org",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should not match wildcard pattern that doesn't fit",
		},

		// Test AND queries
		{
			name: "AndQueryBothMatch",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "John",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "New York",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should return true when both AND conditions match",
		},
		{
			name: "AndQueryLeftMatch",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "John",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "Paris",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false when only left AND condition matches",
		},
		{
			name: "AndQueryRightMatch",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "Bob",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "New York",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false when only right AND condition matches",
		},
		{
			name: "AndQueryNeitherMatch",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "Bob",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "Paris",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false when neither AND condition matches",
		},

		// Test OR queries
		{
			name: "OrQueryBothMatch",
			query: &LuceneQuery{
				Type: "or",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "John",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "New York",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should return true when both OR conditions match",
		},
		{
			name: "OrQueryLeftMatch",
			query: &LuceneQuery{
				Type: "or",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "John",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "Paris",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should return true when left OR condition matches",
		},
		{
			name: "OrQueryRightMatch",
			query: &LuceneQuery{
				Type: "or",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "Bob",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "New York",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should return true when right OR condition matches",
		},
		{
			name: "OrQueryNeitherMatch",
			query: &LuceneQuery{
				Type: "or",
				Left: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "Bob",
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "Paris",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false when neither OR condition matches",
		},

		// Test NOT queries
		{
			name: "NotQueryMatch",
			query: &LuceneQuery{
				Type: "not",
				Query: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "Bob",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should return true when NOT condition doesn't match",
		},
		{
			name: "NotQueryNoMatch",
			query: &LuceneQuery{
				Type: "not",
				Query: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "John",
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false when NOT condition matches",
		},

		// Test complex nested queries
		{
			name: "ComplexNestedQuery",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "or",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "John",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "Jane",
					},
				},
				Right: &LuceneQuery{
					Type: "not",
					Query: &LuceneQuery{
						Type:  "field",
						Field: "active",
						Value: "false",
					},
				},
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should handle complex nested queries: (name:John OR name:Jane) AND NOT active:false",
		},

		// Test with second record
		{
			name: "ComplexNestedQueryRecord2",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "or",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "John",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "Jane",
					},
				},
				Right: &LuceneQuery{
					Type: "not",
					Query: &LuceneQuery{
						Type:  "field",
						Field: "active",
						Value: "false",
					},
				},
			},
			record:        testRecord2,
			caseSensitive: false,
			expected:      false,
			description:   "Should handle complex nested queries with different data: (name:John OR name:Jane) AND NOT active:false",
		},

		// Test unknown query type
		{
			name: "UnknownQueryType",
			query: &LuceneQuery{
				Type:  "unknown",
				Field: "name",
				Value: "John",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      false,
			description:   "Should return false for unknown query type",
		},

		// Test numeric field matching
		{
			name: "NumericFieldMatch",
			query: &LuceneQuery{
				Type:  "field",
				Field: "age",
				Value: "30",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match numeric field converted to string",
		},

		// Test boolean field matching
		{
			name: "BooleanFieldMatch",
			query: &LuceneQuery{
				Type:  "field",
				Field: "active",
				Value: "true",
			},
			record:        testRecord1,
			caseSensitive: false,
			expected:      true,
			description:   "Should match boolean field converted to string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.evaluateLuceneQuery(tt.query, tt.record, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("Test %s failed: %s\nExpected: %v, Got: %v\nQuery: %+v\nRecord: %+v",
					tt.name, tt.description, tt.expected, result, tt.query, tt.record.Content)
			}
		})
	}
}

// Test helper functions individually
func TestMatchFieldValue(t *testing.T) {
	app := &App{}

	tests := []struct {
		name          string
		fieldValue    interface{}
		searchValue   string
		caseSensitive bool
		expected      bool
	}{
		{"String match", "Hello World", "World", false, true},
		{"String no match", "Hello World", "Foo", false, false},
		{"Case sensitive match", "Hello World", "world", true, false},
		{"Case insensitive match", "Hello World", "world", false, true},
		{"Nil field value", nil, "test", false, false},
		{"Number field", 123, "23", false, true},
		{"Boolean field", true, "true", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.matchFieldValue(tt.fieldValue, tt.searchValue, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for matchFieldValue(%v, %s, %v)",
					tt.expected, result, tt.fieldValue, tt.searchValue, tt.caseSensitive)
			}
		})
	}
}

func TestMatchPhrase(t *testing.T) {
	app := &App{}

	tests := []struct {
		name          string
		text          string
		phrase        string
		caseSensitive bool
		expected      bool
	}{
		{"Exact phrase match", "Hello World", "Hello World", false, true},
		{"Partial phrase match", "Hello Beautiful World", "Beautiful", false, true},
		{"No match", "Hello World", "Goodbye", false, false},
		{"Case sensitive match", "Hello World", "hello world", true, false},
		{"Case insensitive match", "Hello World", "hello world", false, true},
		{"Empty text", "", "test", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.matchPhrase(tt.text, tt.phrase, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for matchPhrase(%s, %s, %v)",
					tt.expected, result, tt.text, tt.phrase, tt.caseSensitive)
			}
		})
	}
}

func TestMatchWildcard(t *testing.T) {
	app := &App{}

	tests := []struct {
		name          string
		text          string
		pattern       string
		caseSensitive bool
		expected      bool
	}{
		{"Star wildcard", "hello.txt", "*.txt", false, true},
		{"Star wildcard no match", "hello.doc", "*.txt", false, false},
		{"Question mark wildcard", "test", "t?st", false, false}, // Current implementation doesn't support ? wildcards
		{"Multiple stars", "hello world test", "*world*", false, true},
		{"Star at end", "hello world", "hello*", false, true},
		{"Star at beginning", "hello world", "*world", false, true},
		{"No wildcards", "hello", "hello", false, true},
		{"Case sensitive", "Hello", "hello", true, false},
		{"Case insensitive", "Hello", "hello", false, true},
		{"Empty text", "", "*", false, false}, // Current implementation returns false for empty text regardless of pattern
		{"Match all", "anything", "*", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.matchWildcard(tt.text, tt.pattern, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for matchWildcard(%s, %s, %v)",
					tt.expected, result, tt.text, tt.pattern, tt.caseSensitive)
			}
		})
	}
}

func TestMatchTerm(t *testing.T) {
	app := &App{}

	tests := []struct {
		name          string
		text          string
		term          string
		caseSensitive bool
		expected      bool
	}{
		{"Term match", "Hello World", "World", false, true},
		{"Term no match", "Hello World", "Foo", false, false},
		{"Case sensitive match", "Hello World", "world", true, false},
		{"Case insensitive match", "Hello World", "world", false, true},
		{"Empty text", "", "test", false, false},
		{"Partial term match", "Hello", "ell", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.matchTerm(tt.text, tt.term, tt.caseSensitive)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for matchTerm(%s, %s, %v)",
					tt.expected, result, tt.text, tt.term, tt.caseSensitive)
			}
		})
	}
}

// Test complex queries with more than 2 conditions
func TestComplexMultiConditionQueries(t *testing.T) {
	app := &App{}

	// Test data
	testRecord := JSONRecord{
		LineNumber: 1,
		Content: map[string]interface{}{
			"name":    "John Doe",
			"age":     30,
			"email":   "john.doe@example.com",
			"city":    "New York",
			"active":  true,
			"role":    "admin",
			"company": "TechCorp",
			"status":  "employed",
		},
		RawJSON: `{"name":"John Doe","age":30,"email":"john.doe@example.com","city":"New York","active":true,"role":"admin","company":"TechCorp","status":"employed"}`,
	}

	tests := []struct {
		name        string
		query       *LuceneQuery
		expected    bool
		description string
	}{
		{
			name: "ThreeAndConditions",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "and",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "John",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "age",
						Value: "30",
					},
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "New York",
				},
			},
			expected:    true,
			description: "Should match all three AND conditions: name:John AND age:30 AND city:New York",
		},
		{
			name: "FourAndConditions",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "and",
					Left: &LuceneQuery{
						Type: "and",
						Left: &LuceneQuery{
							Type:  "field",
							Field: "name",
							Value: "John",
						},
						Right: &LuceneQuery{
							Type:  "field",
							Field: "age",
							Value: "30",
						},
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "city",
						Value: "New York",
					},
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "role",
					Value: "admin",
				},
			},
			expected:    true,
			description: "Should match all four AND conditions: name:John AND age:30 AND city:New York AND role:admin",
		},
		{
			name: "ThreeOrConditions",
			query: &LuceneQuery{
				Type: "or",
				Left: &LuceneQuery{
					Type: "or",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "John",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "Jane",
					},
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "name",
					Value: "Bob",
				},
			},
			expected:    true,
			description: "Should match any of three OR conditions: name:John OR name:Jane OR name:Bob",
		},
		{
			name: "MixedComplexQuery",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "or",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "role",
						Value: "admin",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "role",
						Value: "manager",
					},
				},
				Right: &LuceneQuery{
					Type: "and",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "active",
						Value: "true",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "status",
						Value: "employed",
					},
				},
			},
			expected:    true,
			description: "Should match complex query: (role:admin OR role:manager) AND (active:true AND status:employed)",
		},
		{
			name: "ComplexQueryWithNot",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "and",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "John",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "active",
						Value: "true",
					},
				},
				Right: &LuceneQuery{
					Type: "not",
					Query: &LuceneQuery{
						Type:  "field",
						Field: "status",
						Value: "unemployed",
					},
				},
			},
			expected:    true,
			description: "Should match complex query with NOT: name:John AND active:true AND NOT status:unemployed",
		},
		{
			name: "FiveConditionQuery",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "and",
					Left: &LuceneQuery{
						Type: "and",
						Left: &LuceneQuery{
							Type: "and",
							Left: &LuceneQuery{
								Type:  "field",
								Field: "name",
								Value: "John",
							},
							Right: &LuceneQuery{
								Type:  "field",
								Field: "age",
								Value: "30",
							},
						},
						Right: &LuceneQuery{
							Type:  "field",
							Field: "city",
							Value: "New York",
						},
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "role",
						Value: "admin",
					},
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "company",
					Value: "TechCorp",
				},
			},
			expected:    true,
			description: "Should match all five AND conditions",
		},
		{
			name: "ComplexFailingQuery",
			query: &LuceneQuery{
				Type: "and",
				Left: &LuceneQuery{
					Type: "and",
					Left: &LuceneQuery{
						Type:  "field",
						Field: "name",
						Value: "John",
					},
					Right: &LuceneQuery{
						Type:  "field",
						Field: "age",
						Value: "30",
					},
				},
				Right: &LuceneQuery{
					Type:  "field",
					Field: "city",
					Value: "Los Angeles", // This should fail
				},
			},
			expected:    false,
			description: "Should fail when one condition doesn't match: name:John AND age:30 AND city:Los Angeles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := app.evaluateLuceneQuery(tt.query, testRecord, false)
			if result != tt.expected {
				t.Errorf("Test %s failed: %s\nExpected: %v, Got: %v\nQuery structure: %s",
					tt.name, tt.description, tt.expected, result, formatQuery(tt.query))
			} else {
				t.Logf("✅ Test %s passed: %s", tt.name, tt.description)
			}
		})
	}
}

// Helper function to format query structure for debugging
func formatQuery(q *LuceneQuery) string {
	if q == nil {
		return "nil"
	}

	switch q.Type {
	case "field", "term", "phrase", "wildcard":
		if q.Field != "" {
			return fmt.Sprintf("%s:%s:%s", q.Type, q.Field, q.Value)
		}
		return fmt.Sprintf("%s:%s", q.Type, q.Value)
	case "and", "or":
		return fmt.Sprintf("(%s %s %s)", formatQuery(q.Left), q.Type, formatQuery(q.Right))
	case "not":
		return fmt.Sprintf("NOT %s", formatQuery(q.Query))
	default:
		return fmt.Sprintf("unknown:%s", q.Type)
	}
}

// Test the query parser with multi-condition queries
func TestParseLuceneQueryMultiCondition(t *testing.T) {
	tests := []struct {
		name        string
		query       string
		expectNil   bool
		description string
	}{
		{
			name:        "ThreeAndConditions",
			query:       "name:John AND age:30 AND city:NewYork",
			expectNil:   false,
			description: "Should parse three AND conditions",
		},
		{
			name:        "FourAndConditions",
			query:       "name:John AND age:30 AND city:NewYork AND role:admin",
			expectNil:   false,
			description: "Should parse four AND conditions",
		},
		{
			name:        "ThreeOrConditions",
			query:       "name:John OR name:Jane OR name:Bob",
			expectNil:   false,
			description: "Should parse three OR conditions",
		},
		{
			name:        "TwoConditions",
			query:       "name:John AND age:30",
			expectNil:   false,
			description: "Should parse two AND conditions (this works)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLuceneQuery(tt.query)
			if tt.expectNil && result != nil {
				t.Errorf("Expected nil but got %s", formatQuery(result))
			} else if !tt.expectNil && result == nil {
				t.Errorf("Expected non-nil result but got nil for query: %s", tt.query)
			} else if result != nil {
				t.Logf("✅ Parsed query '%s' into: %s", tt.query, formatQuery(result))
			} else {
				t.Logf("❌ Failed to parse query: %s", tt.query)
			}
		})
	}
}

// Test end-to-end parsing and evaluation with real query strings
func TestEndToEndMultiConditionQueries(t *testing.T) {
	app := &App{}

	// Test data
	testRecord := JSONRecord{
		LineNumber: 1,
		Content: map[string]interface{}{
			"name":    "John Doe",
			"age":     30,
			"email":   "john.doe@example.com",
			"city":    "New York",
			"active":  true,
			"role":    "admin",
			"company": "TechCorp",
			"status":  "employed",
		},
		RawJSON: `{"name":"John Doe","age":30,"email":"john.doe@example.com","city":"New York","active":true,"role":"admin","company":"TechCorp","status":"employed"}`,
	}

	tests := []struct {
		name        string
		queryString string
		expected    bool
		description string
	}{
		{
			name:        "ThreeAndConditionsString",
			queryString: "name:John AND age:30 AND city:New",
			expected:    true,
			description: "Three AND conditions should work with query string",
		},
		{
			name:        "FourAndConditionsString",
			queryString: "name:John AND age:30 AND city:New AND role:admin",
			expected:    true,
			description: "Four AND conditions should work with query string",
		},
		{
			name:        "FiveAndConditionsString",
			queryString: "name:John AND age:30 AND city:New AND role:admin AND status:employed",
			expected:    true,
			description: "Five AND conditions should work with query string",
		},
		{
			name:        "ThreeOrConditionsString",
			queryString: "name:John OR name:Jane OR name:Bob",
			expected:    true,
			description: "Three OR conditions should work with query string",
		},
		{
			name:        "MixedConditionsString",
			queryString: "name:John AND age:30 OR role:admin",
			expected:    true,
			description: "Mixed AND/OR conditions should work (left-associative: (name:John AND age:30) OR role:admin)",
		},
		{
			name:        "FailingMultiCondition",
			queryString: "name:John AND age:30 AND city:Paris",
			expected:    false,
			description: "Should fail when one condition in multi-condition query doesn't match",
		},
		{
			name:        "ComplexMixedQuery",
			queryString: "role:admin AND active:true AND status:employed",
			expected:    true,
			description: "Complex query with multiple boolean/string fields should work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the query string
			luceneQuery := parseLuceneQuery(tt.queryString)
			if luceneQuery == nil {
				t.Fatalf("Failed to parse query string: %s", tt.queryString)
			}

			// Evaluate the parsed query
			result := app.evaluateLuceneQuery(luceneQuery, testRecord, false)

			if result != tt.expected {
				t.Errorf("Test %s failed: %s\nQuery: %s\nParsed as: %s\nExpected: %v, Got: %v",
					tt.name, tt.description, tt.queryString, formatQuery(luceneQuery), tt.expected, result)
			} else {
				t.Logf("✅ Test %s passed: %s\nQuery: %s\nParsed as: %s",
					tt.name, tt.description, tt.queryString, formatQuery(luceneQuery))
			}
		})
	}
}
