package sbom

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	spdx_json "github.com/spdx/tools-golang/json"
	"github.com/spdx/tools-golang/spdx"
	"github.com/stretchr/testify/assert"
)

const testSpdxJSONDocument = "./test/example.spdx.json"

// testingSpdxDocument returns a SPDX document for testing purposes.
func testingSpdxDocument() *spdx.Document {
	f, err := os.Open(testSpdxJSONDocument)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	doc, err := spdx_json.Read(f)
	if err != nil {
		panic(err)
	}
	return doc
}

// testingSpdxJSONRawDocument loads testSpdxJSONDocument and returns the raw JSON data.
func testingSpdxJSONRawDocument() []byte {
	data, err := os.ReadFile(testSpdxJSONDocument)
	if err != nil {
		panic(err)
	}
	return data
}

func TestNewDocumentFromSpdxJSON(t *testing.T) {
	data := testingSpdxJSONRawDocument()
	doc, err := NewDocumentFromSpdxJSON(data)
	assert.NoError(t, err)
	assert.NotNil(t, doc)

	assert.Equal(t, StandardTypeSpdx, doc.docType)
	assert.IsType(t, &spdx.Document{}, doc.document)
}

func TestToSpdxJSON(t *testing.T) {
	data := testingSpdxJSONRawDocument()
	doc, err := NewDocumentFromSpdxJSON(data)
	assert.NoError(t, err)
	assert.NotNil(t, doc)

	writter := bytes.NewBuffer(nil)
	err = doc.ToSpdxJSON(writter)
	assert.NoError(t, err)

	// we can't compare the raw JSON data because the order of the elements may change
	// so compare unmarshaled interface{} objects instead
	var expectedSpdxDoc interface{}
	var gotSpdxDoc interface{}

	err = json.Unmarshal(data, &expectedSpdxDoc)
	assert.NoError(t, err)
	err = json.Unmarshal(writter.Bytes(), &gotSpdxDoc)
	assert.NoError(t, err)
	assert.Equal(t, expectedSpdxDoc, gotSpdxDoc)
}

func TestToSpdxDocument(t *testing.T) {
	data := testingSpdxJSONRawDocument()
	doc, err := NewDocumentFromSpdxJSON(data)
	assert.NoError(t, err)
	assert.NotNil(t, doc)

	expectedSpdxDoc := testingSpdxDocument()
	gotSpdxDoc, err := doc.ToSpdxDocument()
	assert.NoError(t, err)

	assert.Equal(t, expectedSpdxDoc, gotSpdxDoc)
}
