package model

type CompletionResponse struct {
	Text string `json:"text"`
}

type ModelProbeResponse struct {
	Model            string `json:"model"`
	OAAvailable      bool   `json:"oa_available"`
	NativeRecognized bool   `json:"native_recognized"`
	NativeStatusCode int    `json:"native_status_code,omitempty"`
	NativeMessage    string `json:"native_message,omitempty"`
}
