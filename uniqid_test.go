package uniqid

import "testing"

func TestUniqid(t *testing.T) {
	SetServerID(77)

	adid := Append(nil)
	if len(adid) != 16 {
		t.Fatalf("unexpected adid length: %d", len(adid))
	}

	v := GetServerID(adid)
	if v != 77 {
		t.Fatalf("unexpected server id: %d", v)
	}

}
