package novelai

import "testing"

func TestUpstreamErrorIncludesMessage(t *testing.T) {
	err := (&UpstreamError{
		StatusCode: 400,
		Body:       []byte(`{"statusCode":400,"message":"bad request: model 'kayra' doesn't exist"}`),
	}).Error()

	want := "novelai upstream error (status=400): bad request: model 'kayra' doesn't exist"
	if err != want {
		t.Fatalf("got %q want %q", err, want)
	}
}

func TestUpstreamErrorFallsBackToStatus(t *testing.T) {
	err := (&UpstreamError{
		StatusCode: 500,
		Body:       nil,
	}).Error()

	want := "novelai upstream error (status=500)"
	if err != want {
		t.Fatalf("got %q want %q", err, want)
	}
}
