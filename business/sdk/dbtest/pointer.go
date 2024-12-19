package dbtest

import (
	"github.com/ardanlabs/service/business/types/money"
	"github.com/ardanlabs/service/business/types/name"
	"github.com/ardanlabs/service/business/types/quantity"
)

// StringPointer is a helper to get a *string from a string. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func StringPointer(s string) *string {
	return &s
}

// IntPointer is a helper to get a *int from a int. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func IntPointer(i int) *int {
	return &i
}

// FloatPointer is a helper to get a *float64 from a float64. It is in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func FloatPointer(f float64) *float64 {
	return &f
}

// BoolPointer is a helper to get a *bool from a bool. It is in the tests package
// because we normally don't want to deal with pointers to basic types but it's
// useful in some tests.
func BoolPointer(b bool) *bool {
	return &b
}

// NamePointer is a helper to get a *Name from a string. It's in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func NamePointer(value string) *name.Name {
	name := name.MustParse(value)
	return &name
}

// NameNullPointer is a helper to get a *EmptyName from a string. It's in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func NameNullPointer(value string) *name.Null {
	name := name.MustParseNull(value)
	return &name
}

// MoneyPointer is a helper to get a *Money from a float. It's in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func MoneyPointer(value float64) *money.Money {
	money := money.MustParse(value)
	return &money
}

// QuantityPointer is a helper to get a *Quantity from an int. It's in the tests
// package because we normally don't want to deal with pointers to basic types
// but it's useful in some tests.
func QuantityPointer(value int) *quantity.Quantity {
	quantity := quantity.MustParse(value)
	return &quantity
}
