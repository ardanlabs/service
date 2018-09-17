package trace

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"go.opencensus.io/trace"
)

// Success and failure markers.
const (
	success = "\u2713"
	failed  = "\u2717"
)

// inputSpans represents spans of data for the tests.
var inputSpans = []*trace.SpanData{
	&trace.SpanData{Name: "span1"},
	&trace.SpanData{Name: "span2"},
	&trace.SpanData{Name: "span3"},
}

// inputSpansJSON represents a JSON representation of the span data.
var inputSpansJSON = `[{"TraceID":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"SpanID":[0,0,0,0,0,0,0,0],"TraceOptions":0,"ParentSpanID":[0,0,0,0,0,0,0,0],"SpanKind":0,"Name":"span1","StartTime":"0001-01-01T00:00:00Z","EndTime":"0001-01-01T00:00:00Z","Attributes":null,"Annotations":null,"MessageEvents":null,"Code":0,"Message":"","Links":null,"HasRemoteParent":false},{"TraceID":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"SpanID":[0,0,0,0,0,0,0,0],"TraceOptions":0,"ParentSpanID":[0,0,0,0,0,0,0,0],"SpanKind":0,"Name":"span2","StartTime":"0001-01-01T00:00:00Z","EndTime":"0001-01-01T00:00:00Z","Attributes":null,"Annotations":null,"MessageEvents":null,"Code":0,"Message":"","Links":null,"HasRemoteParent":false},{"TraceID":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0],"SpanID":[0,0,0,0,0,0,0,0],"TraceOptions":0,"ParentSpanID":[0,0,0,0,0,0,0,0],"SpanKind":0,"Name":"span3","StartTime":"0001-01-01T00:00:00Z","EndTime":"0001-01-01T00:00:00Z","Attributes":null,"Annotations":null,"MessageEvents":null,"Code":0,"Message":"","Links":null,"HasRemoteParent":false}]`

// =============================================================================

// logger is required to create an Exporter.
var logger = func(format string, v ...interface{}) {
	log.Printf(format, v)
}

// MakeExporter abstracts the error handling aspects of creating an Exporter.
func makeExporter(host string, batchSize int, sendInterval, sendTimeout time.Duration) *Exporter {
	exporter, err := NewExporter(logger, host, batchSize, sendInterval, sendTimeout)
	if err != nil {
		log.Fatalln("Unable to create exporter, ", err)
	}
	return exporter
}

// =============================================================================

var saveTests = []struct {
	name              string
	e                 *Exporter
	input             []*trace.SpanData
	output            []*trace.SpanData
	lastSaveDelay     time.Duration // The delay before the last save. For testing intervals.
	isInputMatchBatch bool          // If the input should match the internal exporter collection after the last save.
	isSendBatch       bool          // If the last save should return nil or batch data.
}{
	{"NoSend", makeExporter("test", 10, time.Minute, time.Second), inputSpans, nil, time.Nanosecond, true, false},
	{"SendOnBatchSize", makeExporter("test", 3, time.Minute, time.Second), inputSpans, inputSpans, time.Nanosecond, false, true},
	{"SendOnTime", makeExporter("test", 4, time.Millisecond, time.Second), inputSpans, inputSpans, 2 * time.Millisecond, false, true},
}

// TestSave validates the save batch functionality is working.
func TestSave(t *testing.T) {
	t.Log("Given the need to validate saving span data to a batch.")
	{
		for i, tt := range saveTests {
			t.Logf("\tTest: %d\tWhen running test: %s", i, tt.name)
			{
				// Save the input of span data.
				l := len(tt.input) - 1
				var batch []*trace.SpanData
				for i, span := range tt.input {

					// If this is the last save, take the configured delay.
					// We might be testing invertal based batching.
					if l == i {
						time.Sleep(tt.lastSaveDelay)
					}
					batch = tt.e.saveBatch(span)
				}

				// Compare the internal collection with what we saved.
				if tt.isInputMatchBatch {
					if len(tt.e.batch) != len(tt.input) {
						t.Log("\t\tGot :", len(tt.e.batch))
						t.Log("\t\tWant:", len(tt.input))
						t.Errorf("\t%s\tShould have the same number of spans as input.", failed)
					} else {
						t.Logf("\t%s\tShould have the same number of spans as input.", success)
					}
				} else {
					if len(tt.e.batch) != 0 {
						t.Log("\t\tGot :", len(tt.e.batch))
						t.Log("\t\tWant:", 0)
						t.Errorf("\t%s\tShould have zero spans.", failed)
					} else {
						t.Logf("\t%s\tShould have zero spans.", success)
					}
				}

				// Validate the return provided or didn't provide a batch to send.
				if !tt.isSendBatch && batch != nil {
					t.Errorf("\t%s\tShould not have a batch to send.", failed)
				} else if !tt.isSendBatch {
					t.Logf("\t%s\tShould not have a batch to send.", success)
				}
				if tt.isSendBatch && batch == nil {
					t.Errorf("\t%s\tShould have a batch to send.", failed)
				} else if tt.isSendBatch {
					t.Logf("\t%s\tShould have a batch to send.", success)
				}

				// Compare the batch to send.
				if !reflect.DeepEqual(tt.output, batch) {
					t.Log("\t\tGot :", batch)
					t.Log("\t\tWant:", tt.output)
					t.Errorf("\t%s\tShould have an expected match of the batch to send.", failed)
				} else {
					t.Logf("\t%s\tShould have an expected match of the batch to send.", success)
				}
			}
		}
	}
}

// =============================================================================

var sendTests = []struct {
	name  string
	e     *Exporter
	input []*trace.SpanData
	pass  bool
}{
	{"success", makeExporter("test", 3, time.Minute, time.Hour), inputSpans, true},
	{"failure", makeExporter("test", 3, time.Minute, time.Hour), inputSpans[:2], false},
	{"timeout", makeExporter("test", 3, time.Minute, time.Nanosecond), inputSpans, false},
}

// mockServer returns a pointer to a server to handle the mock get call.
func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		d, _ := ioutil.ReadAll(r.Body)
		data := string(d)
		if data != inputSpansJSON {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, data)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	return httptest.NewServer(http.HandlerFunc(f))
}

// TestSend validates spans can be sent to the sidecar.
func TestSend(t *testing.T) {
	s := mockServer()
	defer s.Close()

	t.Log("Given the need to validate sending span data to the sidecar.")
	{
		for i, tt := range sendTests {
			t.Logf("\tTest: %d\tWhen running test: %s", i, tt.name)
			{
				// Set the URL for the call.
				tt.e.host = s.URL

				// Send the span data.
				err := tt.e.send(tt.input)
				if tt.pass {
					if err != nil {
						t.Errorf("\t%s\tShould be able to send the batch successfully: %v", failed, err)
					} else {
						t.Logf("\t%s\tShould be able to send the batch successfully.", success)
					}
				} else {
					if err == nil {
						t.Errorf("\t%s\tShould not be able to send the batch successfully : %v", failed, err)
					} else {
						t.Logf("\t%s\tShould not be able to send the batch successfully.", success)
					}
				}
			}
		}
	}
}

// TestClose validates the flushing of the final batched spans.
func TestClose(t *testing.T) {
	s := mockServer()
	defer s.Close()

	t.Log("Given the need to validate flushing the remaining batched spans.")
	{
		t.Logf("\tTest: %d\tWhen running test: %s", 0, "FlushWithData")
		{
			e, err := NewExporter(logger, "test", 10, time.Minute, time.Hour)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create an Exporter : %v", failed, err)
			}
			t.Logf("\t%s\tShould be able to create an Exporter.", success)

			// Set the URL for the call.
			e.host = s.URL

			// Save the input of span data.
			for _, span := range inputSpans {
				e.saveBatch(span)
			}

			// Close the Exporter and we should get those spans sent.
			sent, err := e.Close()
			if err != nil {
				t.Fatalf("\t%s\tShould be able to flush the Exporter : %v", failed, err)
			}
			t.Logf("\t%s\tShould be able to flush the Exporter.", success)

			if sent != len(inputSpans) {
				t.Log("\t\tGot :", sent)
				t.Log("\t\tWant:", len(inputSpans))
				t.Fatalf("\t%s\tShould have flushed the expected number of spans.", failed)
			}
			t.Logf("\t%s\tShould have flushed the expected number of spans.", success)
		}

		t.Logf("\tTest: %d\tWhen running test: %s", 0, "FlushWithError")
		{
			e, err := NewExporter(logger, "test", 10, time.Minute, time.Hour)
			if err != nil {
				t.Fatalf("\t%s\tShould be able to create an Exporter : %v", failed, err)
			}
			t.Logf("\t%s\tShould be able to create an Exporter.", success)

			// Set the URL for the call.
			e.host = s.URL

			// Save the input of span data.
			for _, span := range inputSpans[:2] {
				e.saveBatch(span)
			}

			// Close the Exporter and we should get those spans sent.
			if _, err := e.Close(); err == nil {
				t.Fatalf("\t%s\tShould not be able to flush the Exporter.", failed)
			}
			t.Logf("\t%s\tShould not be able to flush the Exporter.", success)
		}
	}
}

// =============================================================================

// TestExporterFailure validates misuse cases are covered.
func TestExporterFailure(t *testing.T) {
	t.Log("Given the need to validate Exporter initializes properly.")
	{
		t.Logf("\tTest: %d\tWhen not passing a proper logger.", 0)
		{
			_, err := NewExporter(nil, "test", 10, time.Minute, time.Hour)
			if err == nil {
				t.Errorf("\t%s\tShould not be able to create an Exporter.", failed)
			} else {
				t.Logf("\t%s\tShould not be able to create an Exporter.", success)
			}
		}

		t.Logf("\tTest: %d\tWhen not passing a proper host.", 1)
		{
			_, err := NewExporter(logger, "", 10, time.Minute, time.Hour)
			if err == nil {
				t.Errorf("\t%s\tShould not be able to create an Exporter.", failed)
			} else {
				t.Logf("\t%s\tShould not be able to create an Exporter.", success)
			}
		}
	}
}
