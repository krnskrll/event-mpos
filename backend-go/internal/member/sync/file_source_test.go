package sync

import "testing"

func TestFileSourceReadsSampleCSV(t *testing.T) {
	members, err := FileSource{Path: "../../../samples/members/members.csv"}.FetchMembers()
	if err != nil {
		t.Fatal(err)
	}
	if len(members) == 0 || members[0].QRCode == "" {
		t.Errorf("members = %+v", members)
	}
}
