package user

import (
	"reflect"
	"testing"
)

func TestUser_AddAddressRequiresReceiver(t *testing.T) {
	u, err := NewUser("Ada", "ada@example.com")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	_, err = u.AddAddress("", "13800000000", "Shanghai", "Road 1")
	if err == nil {
		t.Fatal("AddAddress() error = nil, want error")
	}
}

func TestUser_AddAddressStoresAddress(t *testing.T) {
	u, err := NewUser("Ada", "ada@example.com")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}

	addr, err := u.AddAddress("Ada", "13800000000", "Shanghai", "Road 1")
	if err != nil {
		t.Fatalf("AddAddress() error = %v", err)
	}
	if addr.UserUUID() != u.UUID() {
		t.Fatalf("addr.UserUUID() = %q, want %q", addr.UserUUID(), u.UUID())
	}
}

func TestUserUsesUUIDTerminology(t *testing.T) {
	u, err := NewUser("Ada", "ada@example.com")
	if err != nil {
		t.Fatalf("NewUser() error = %v", err)
	}
	var uuid UserUUID = u.UUID()
	if uuid == "" {
		t.Fatal("UUID() is empty")
	}
	if _, ok := reflect.TypeOf(u).MethodByName("ID"); ok {
		t.Fatal("User exposes ID(), want UUID()")
	}
}

func TestAddressUsesUUIDTerminology(t *testing.T) {
	address, err := NewAddress(UserUUID("user-uuid"), "Ada", "13800000000", "Shanghai", "Road 1")
	if err != nil {
		t.Fatalf("NewAddress() error = %v", err)
	}
	var uuid AddressUUID = address.UUID()
	if uuid == "" {
		t.Fatal("UUID() is empty")
	}
	if _, ok := reflect.TypeOf(address).MethodByName("ID"); ok {
		t.Fatal("Address exposes ID(), want UUID()")
	}
}
