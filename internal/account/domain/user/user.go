package user

import "modular_monolith/internal/shared/bizid"

type UserUUID string

func (uuid UserUUID) String() string { return string(uuid) }

type AddressUUID string

func (uuid AddressUUID) String() string { return string(uuid) }

type Status string

const (
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

type User struct {
	uuid      UserUUID
	name      string
	email     string
	status    Status
	addresses []Address
}

type Address struct {
	uuid     AddressUUID
	userUUID UserUUID
	receiver string
	phone    string
	city     string
	detail   string
}

func NewUser(name string, email string) (*User, error) {
	if name == "" || email == "" {
		return nil, ErrInvalidUser
	}
	return &User{uuid: UserUUID(bizid.New()), name: name, email: email, status: StatusActive}, nil
}

func Rehydrate(uuid UserUUID, name string, email string, status Status, addresses []Address) *User {
	return &User{uuid: uuid, name: name, email: email, status: status, addresses: append([]Address(nil), addresses...)}
}

func NewAddress(userUUID UserUUID, receiver string, phone string, city string, detail string) (Address, error) {
	if userUUID == "" || receiver == "" || phone == "" || city == "" || detail == "" {
		return Address{}, ErrInvalidAddress
	}
	return Address{uuid: AddressUUID(bizid.New()), userUUID: userUUID, receiver: receiver, phone: phone, city: city, detail: detail}, nil
}

func RehydrateAddress(uuid AddressUUID, userUUID UserUUID, receiver string, phone string, city string, detail string) Address {
	return Address{uuid: uuid, userUUID: userUUID, receiver: receiver, phone: phone, city: city, detail: detail}
}

func (u *User) AddAddress(receiver string, phone string, city string, detail string) (Address, error) {
	address, err := NewAddress(u.uuid, receiver, phone, city, detail)
	if err != nil {
		return Address{}, err
	}
	u.addresses = append(u.addresses, address)
	return address, nil
}

func (u *User) UUID() UserUUID       { return u.uuid }
func (u *User) Name() string         { return u.name }
func (u *User) Email() string        { return u.email }
func (u *User) Status() Status       { return u.status }
func (u *User) Addresses() []Address { return append([]Address(nil), u.addresses...) }

func (u *User) EnsureActive() error {
	if u.status == StatusDisabled {
		return ErrUserDisabled
	}
	return nil
}

func (a Address) UUID() AddressUUID  { return a.uuid }
func (a Address) UserUUID() UserUUID { return a.userUUID }
func (a Address) Receiver() string   { return a.receiver }
func (a Address) Phone() string      { return a.phone }
func (a Address) City() string       { return a.city }
func (a Address) Detail() string     { return a.detail }
