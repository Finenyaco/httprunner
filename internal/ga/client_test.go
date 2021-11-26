package ga

import (
	"testing"
)

func TestSendEvents(t *testing.T) {
	event := EventTracking{
		Category: "unittest",
		Action:   "SendEvents",
		Value:    "123",
	}
	err := gaClient.SendEvent(event)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStructToUrlValues(t *testing.T) {
	event := EventTracking{
		Category: "unittest",
		Action:   "convert",
		Label:    "StructToUrlValues",
		Value:    "123",
	}
	val := structToUrlValues(event)
	if val.Encode() != "ea=convert&ec=unittest&el=StructToUrlValues&ev=123" {
		t.Fail()
	}
}
