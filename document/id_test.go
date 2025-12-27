// Package document_test contains tests for the document package.
package document_test

import (
	"image"
	"testing"

	"github.com/ubavic/bas-celik/v2/document"
)

var documentID1 = document.IDDocument{}
var documentID2 = document.IDDocument{
	GivenName:       "Петар",
	ParentGivenName: "Арсеније",
	Surname:         "Петровић",
	Street:          "Његошева",
	HouseNumber:     "9",
	HouseLetter:     "Б",
	Place:           "Подгорица",
	PlaceOfBirth:    "Његуши",
	StateOfBirth:    "Црна Гора",
}
var documentID3 = document.IDDocument{
	GivenName:        "Pablo Diego",
	Surname:          "Ruiz Picasso",
	HouseNumber:      "7",
	HouseLetter:      "A",
	Street:           "Rue des Grands-Augustins",
	ApartmentNumber:  "21",
	Floor:            "6",
	Community:        "Saint-Germain-des-Prés",
	Place:            "Paris",
	PlaceOfBirth:     "Málaga",
	CommunityOfBirth: "Andalucía",
	StateOfBirth:     "Reino de España",
}

// Test_GetFullName_ID tests the GetFullName method of the IdDocument struct.
func Test_GetFullName_ID(t *testing.T) {
	testCases := []struct {
		value    document.IDDocument
		expected string
	}{
		{
			value:    documentID1,
			expected: "",
		},
		{
			value:    documentID2,
			expected: "Петар, Арсеније, Петровић",
		},
		{
			value:    documentID3,
			expected: "Pablo Diego, Ruiz Picasso",
		},
	}

	for _, testCase := range testCases {
		result := testCase.value.GetFullName()
		if result != testCase.expected {
			t.Errorf("Expected '%s' but got '%s'", testCase.expected, result)
		}
	}
}

// Test_GetFullAddress_ID tests the GetFullAddress method of the IdDocument struct.
func Test_GetFullAddress_ID(t *testing.T) {
	testCases := []struct {
		value            document.IDDocument
		expected         string
		expectedReversed string
	}{
		{
			value:            documentID1,
			expected:         "",
			expectedReversed: "",
		},
		{
			value:            documentID2,
			expected:         "Његошева 9Б, Подгорица",
			expectedReversed: "Подгорица, Његошева 9Б",
		},
		{
			value:            documentID3,
			expected:         "Rue des Grands-Augustins 7A/6/21, Saint-Germain-des-Prés, Paris",
			expectedReversed: "Paris, Saint-Germain-des-Prés, Rue des Grands-Augustins 7A/6/21",
		},
	}

	for _, testCase := range testCases {
		result := testCase.value.GetFullAddress(false)
		if result != testCase.expected {
			t.Errorf("Expected '%s' but got '%s'", testCase.expected, result)
		}

		resultReversed := testCase.value.GetFullAddress(true)
		if resultReversed != testCase.expectedReversed {
			t.Errorf("Expected '%s' but got '%s'", testCase.expectedReversed, resultReversed)
		}
	}
}

// Test_GetFullPlaceOfBirth_ID tests the GetFullPlaceOfBirth method of the IdDocument struct.
func Test_GetFullPlaceOfBirth_ID(t *testing.T) {
	testCases := []struct {
		value    document.IDDocument
		expected string
	}{
		{
			value:    documentID1,
			expected: "",
		},
		{
			value:    documentID2,
			expected: "Његуши, Црна Гора",
		},
		{
			value:    documentID3,
			expected: "Málaga, Andalucía, Reino de España",
		},
	}

	for _, testCase := range testCases {
		result := testCase.value.GetFullPlaceOfBirth()
		if result != testCase.expected {
			t.Errorf("Expected '%s' but got '%s'", testCase.expected, result)
		}
	}
}

// Test_BuildPDFID tests the BuildPdf method of the IdDocument struct.
func Test_BuildPDFID(t *testing.T) {
	unsetDocumentConfig()

	_, _, err := documentID1.BuildPdf()
	if err == nil {
		t.Errorf("Expected error but got %v", err)
	}

	setDocumentConfigFromLocalFiles(t)

	_, _, err = documentID1.BuildPdf()
	if err == nil {
		t.Errorf("Expected error but got %v", err)
	}

	rect := image.Rect(0, 0, 200, 200)
	img := image.NewRGBA(rect)
	documentID1.Portrait = img
	documentID2.Portrait = img

	_, _, err = documentID1.BuildPdf()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}

	_, _, err = documentID2.BuildPdf()
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
}
