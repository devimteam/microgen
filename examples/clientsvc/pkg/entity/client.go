package entity

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
)

type UUID uuid.UUID

type Client struct {
	Id            UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	IsInBlackList bool

	MainInfo  MainInfo
	Passports []Passport
	OtherInfo *json.RawMessage
	Audit
}

type Audit struct {
	CreatedBy    *UUID
	CreatedWhere *UUID
}

type MainInfo struct {
	Id        UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	ClientId  UUID

	RegAddress  *Address
	LiveAddress *Address
	JobAddress  *Address

	PersonalPhones []Phone
	RegPhones      []Phone
	LivePhones     []Phone
	JobPhones      []Phone

	Socials []Social
}

type Passport struct {
	Id        UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	ClientId  UUID

	FirstName    *string
	LastName     *string
	MiddleName   *string
	Series       *string
	Number       *string
	Birthday     *time.Time
	Gender       *string
	IssuedBy     *string
	IssuedWhere  *string
	IssuedAt     *time.Time
	PlaceOfBirth *string

	IsInvalid bool
	IsActive  bool
}

type Address struct {
	Id         UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	MainInfoId UUID

	FullAddress        *string
	Country            *string
	RegionType         *string
	RegionTypeFull     *string
	Region             *string
	AreaType           *string
	AreaTypeFull       *string
	Area               *string
	CityType           *string
	CityTypeFull       *string
	City               *string
	CityDistrict       *string
	SettlementType     *string
	SettlementTypeFull *string
	Settlement         *string
	StreetType         *string
	StreetTypeFull     *string
	Street             *string
	HouseType          *string
	HouseTypeFull      *string
	House              *string
	BlockType          *string
	BlockTypeFull      *string
	Block              *string
	FlatType           *string
	FlatTypeFull       *string
	Flat               *string
	PostalCode         *string
}

type PhoneType int

const (
	PhoneType_Personal PhoneType = iota
	PhoneType_Reg
	PhoneType_Live
	PhoneType_Job
)

type Phone struct {
	Id         UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	MainInfoId UUID

	Number string
	Type   PhoneType
}

type Social struct {
	Id         UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
	MainInfoId UUID

	Value     string
	Type      string
	OtherInfo *json.RawMessage
}
