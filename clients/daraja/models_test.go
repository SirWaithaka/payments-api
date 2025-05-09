package daraja

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func TestResponseCode_MarshalJSON(t *testing.T) {

	type ts struct {
		ResCode ResponseCode `json:"ResponseCode"`
	}

	tcs := []struct {
		input    ts
		expected string
	}{
		{ts{ResCode: SuccessSubmission}, fmt.Sprintf(`{"ResponseCode":"%s"}`, SuccessSubmission.String())},
		{ts{ResCode: InvalidAccessToken}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAccessToken.String())},
		{ts{ResCode: InvalidAuthHeader}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthHeader.String())},
		{ts{ResCode: InvalidAuthType}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthType.String())},
		{ts{ResCode: InvalidGrantType}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidGrantType.String())},
		{ts{ResCode: InternalServerError}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InternalServerError.String())},
	}

	for _, tc := range tcs {
		var buf bytes.Buffer
		err := jsoniter.NewEncoder(&buf).Encode(tc.input)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		result := bytes.TrimRight(buf.Bytes(), "\n")

		if string(result) != tc.expected {
			t.Errorf("\nexp %s\ngot %s\n", tc.expected, string(result))
		}

	}

}

func TestResponseCode_MarshalText(t *testing.T) {

	type ts struct {
		ResCode ResponseCode `json:"ResponseCode"`
	}

	tcs := []struct {
		input    ts
		expected string
	}{
		{ts{ResCode: SuccessSubmission}, fmt.Sprintf(`{"ResponseCode":"%s"}`, SuccessSubmission.String())},
		{ts{ResCode: InvalidAccessToken}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAccessToken.String())},
		{ts{ResCode: InvalidAuthHeader}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthHeader.String())},
		{ts{ResCode: InvalidAuthType}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthType.String())},
		{ts{ResCode: InvalidGrantType}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidGrantType.String())},
		{ts{ResCode: InternalServerError}, fmt.Sprintf(`{"ResponseCode":"%s"}`, InternalServerError.String())},
	}

	for _, tc := range tcs {
		result, err := jsoniter.Marshal(&tc.input)
		if err != nil {
			t.Errorf("expected nil error, got %v", err)
		}
		if string(result) != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, string(result))
		}
	}

}

func TestResponseCode_UnmarshalText(t *testing.T) {
	type ts struct {
		ResCode ResponseCode `json:"ResponseCode"`
	}

	tcs := []struct {
		input    string
		expected ResponseCode
	}{
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, SuccessSubmission.String()), SuccessSubmission},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAccessToken.String()), InvalidAccessToken},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthHeader.String()), InvalidAuthHeader},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthType.String()), InvalidAuthType},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidGrantType.String()), InvalidGrantType},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InternalServerError.String()), InternalServerError},
	}

	for _, tc := range tcs {
		var result ts
		if err := jsoniter.Unmarshal([]byte(tc.input), &result); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		if result.ResCode != tc.expected {
			t.Errorf("expected %v, got %v", tc.expected, result.ResCode)
		}

	}

}

func TestResponseCode_UnmarshalJSON(t *testing.T) {

	type ts struct {
		ResCode ResponseCode `json:"ResponseCode"`
	}

	tcs := []struct {
		input    string
		expected ResponseCode
	}{
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, SuccessSubmission.String()), SuccessSubmission},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAccessToken.String()), InvalidAccessToken},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthHeader.String()), InvalidAuthHeader},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidAuthType.String()), InvalidAuthType},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InvalidGrantType.String()), InvalidGrantType},
		{fmt.Sprintf(`{"ResponseCode":"%s"}`, InternalServerError.String()), InternalServerError},
	}

	for _, tc := range tcs {
		var result ts
		if err := jsoniter.NewDecoder(strings.NewReader(tc.input)).Decode(&result); err != nil {
			t.Errorf("expected nil error, got %v", err)
		}

		if result.ResCode != tc.expected {
			t.Errorf("expected %v, got %v", tc.expected, result.ResCode)
		}

	}

}

func TestToResponseCode(t *testing.T) {

	result := ToResponseCode(InvalidGrantType.String())
	if result != InvalidGrantType {
		t.Errorf("expected %v, got %v", InvalidGrantType, result)
	}

}
