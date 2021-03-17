package gCache

import "testing"

func TestGetGroup(t *testing.T) {
	groupName := "scores"
	NewGroup(groupName, 2<<10)
	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group %s not exist", groupName)
	}

	if group := GetGroup(groupName + "111"); group != nil {
		t.Fatalf("expect nil, but %s got", group.name)
	}
}
