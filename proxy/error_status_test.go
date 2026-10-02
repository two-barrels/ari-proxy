// Created by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestHTTPStatusSurvivesResponseEncoding(t *testing.T) {
	original := NewErrorResponse(&StatusError{Message: "channel already claimed", Status: http.StatusConflict})
	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Response
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	var coded interface{ Code() int }
	if !errors.As(decoded.Err(), &coded) || coded.Code() != http.StatusConflict {
		t.Fatalf("decoded error = %v", decoded.Err())
	}
	if plain := NewErrorResponse(errors.New("other error")); plain.StatusCode != 0 || plain.Err().Error() != "other error" {
		t.Fatalf("plain error = %+v", plain)
	}
}
