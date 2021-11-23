// Package env implements a library for loading environment variables.
package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// An Env value is used to load environment variables.
type Env struct {
	m      map[string]string // nil if the actual OS environment should be used
	prefix string
	vars   []variable
}

func (e *Env) lookup(key string) (string, bool) {
	if e.m != nil {
		val, ok := e.m[key]
		return val, ok
	}
	return os.LookupEnv(key)
}

// New returns an Env that loads values from OS environment variables.
func New() *Env {
	return new(Env)
}

// FromMap returns an Env that loads values from the map.
func FromMap(m map[string]string) *Env {
	return &Env{m: m}
}

// Prefix wraps the Env, adding the prefix to any keys it is given.
func (e *Env) Prefix(prefix string) *Env {
	return &Env{
		m:      e.m,
		prefix: e.prefix + prefix,
	}
}

// Load loads the values that have been specified. It returns an error if one is missing or invalid.
func (e *Env) Load() error {
	missing := []string{}
	for _, v := range e.vars {
		key := e.prefix + v.key()
		val, ok := e.lookup(key)
		if ok {
			if err := v.set(val); err != nil {
				return fmt.Errorf("invalid value for %s: %s", key, err)
			}
		} else {
			if ok := v.setDefault(); !ok {
				missing = append(missing, key)
			}
		}
	}
	if len(missing) != 0 {
		return fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Help returns a multi-line help text with variable names and descriptions.
func (e *Env) Help() string {
	var b strings.Builder
	for _, v := range e.vars {
		fmt.Fprintf(&b, "%s%s -- %s\n", e.prefix, v.key(), v.desc())
	}
	return b.String()
}

// String defines a variable of type string.
func (e *Env) String(key string, ptr *string, desc string) {
	e.vars = append(e.vars, &stringVar{
		basicVariable: basicVariable{key, desc, false},
		ptr:           ptr,
	})
}

// OptionalString defines a variable of type string, with a default value.
func (e *Env) OptionalString(key string, ptr *string, dflt, desc string) {
	e.vars = append(e.vars, &stringVar{
		basicVariable: basicVariable{key, desc, true},
		dflt:          dflt,
		ptr:           ptr,
	})
}

// Int defines a variable of type int.
func (e *Env) Int(key string, ptr *int, desc string) {
	e.vars = append(e.vars, &intVar{
		basicVariable: basicVariable{key, desc, false},
		ptr:           ptr,
	})
}

// OptionalInt defines a variable of type int, with a default value.
func (e *Env) OptionalInt(key string, ptr *int, dflt int, desc string) {
	e.vars = append(e.vars, &intVar{
		basicVariable: basicVariable{key, desc, true},
		dflt:          dflt,
		ptr:           ptr,
	})
}

// Float defines a variable of type float64.
func (e *Env) Float(key string, ptr *float64, desc string) {
	e.vars = append(e.vars, &floatVar{
		basicVariable: basicVariable{key, desc, false},
		ptr:           ptr,
	})
}

// OptionalFloat defines a variable of type float64, with a default value.
func (e *Env) OptionalFloat(key string, ptr *float64, dflt float64, desc string) {
	e.vars = append(e.vars, &floatVar{
		basicVariable: basicVariable{key, desc, true},
		dflt:          dflt,
		ptr:           ptr,
	})
}

// Bool defines a variable of type bool. Its value is expected to be "true" or "false".
func (e *Env) Bool(key string, ptr *bool, desc string) {
	e.vars = append(e.vars, &boolVar{
		basicVariable: basicVariable{key, desc, false},
		ptr:           ptr,
	})
}

// OptionalBool defines a variable of type bool, with a default value. Its value is expected to be
// "true" or "false".
func (e *Env) OptionalBool(key string, ptr *bool, dflt bool, desc string) {
	e.vars = append(e.vars, &boolVar{
		basicVariable: basicVariable{key, desc, true},
		dflt:          dflt,
		ptr:           ptr,
	})
}

// Flag defines a variable which is either set or not -- its value is ignored.
func (e *Env) Flag(key string, ptr *bool, desc string) {
	e.vars = append(e.vars, &flagVar{
		basicVariable: basicVariable{key, desc, true},
		ptr:           ptr,
	})
}

// List defines a variable that's a list of strings. The sep argument gives the separator between
// list elements, for example "," for a comma-separated list.
func (e *Env) List(key string, ptr *[]string, sep, desc string) {
	e.vars = append(e.vars, &listVar{
		basicVariable: basicVariable{key, desc, false},
		ptr:           ptr,
		sep:           sep,
	})
}

// OptionalList defines a variable that's a list of strings, with a default value. The sep argument
// gives the separator between list elements, for example "," for a comma-separated list.
func (e *Env) OptionalList(key string, ptr *[]string, sep string, dflt []string, desc string) {
	e.vars = append(e.vars, &listVar{
		basicVariable: basicVariable{key, desc, true},
		dflt:          dflt,
		ptr:           ptr,
		sep:           sep,
	})
}

// Set defines a variable that's a set of strings. It works the same way as List except duplicates
// are considered an error and it stores the result in a map, with each value set to true.
func (e *Env) Set(key string, ptr *map[string]bool, sep, desc string) {
	e.vars = append(e.vars, &setVar{
		basicVariable: basicVariable{key, desc, false},
		ptr:           ptr,
		sep:           sep,
	})
}

// OptionalSet defines a variable that's a set of strings, with a default value. It works the same
// way as OptionalList except duplicates are considered an error and it stores the result in a map,
// with each value set to true.
func (e *Env) OptionalSet(key string, ptr *map[string]bool, sep string, dflt map[string]bool, desc string) {
	e.vars = append(e.vars, &setVar{
		basicVariable: basicVariable{key, desc, true},
		dflt:          dflt,
		ptr:           ptr,
		sep:           sep,
	})
}

type variable interface {
	key() string
	desc() string
	set(val string) error
	setDefault() bool // returns false if there is no default
}

type basicVariable struct {
	k, d    string
	hasDflt bool
}

func (v *basicVariable) key() string {
	return v.k
}

func (v *basicVariable) desc() string {
	return v.d
}

type stringVar struct {
	basicVariable
	dflt string
	ptr  *string
}

func (v *stringVar) set(val string) error {
	if v.ptr != nil {
		*v.ptr = val
	}
	return nil
}

func (v *stringVar) setDefault() bool {
	if !v.hasDflt {
		return false
	}
	if v.ptr != nil {
		*v.ptr = v.dflt
	}
	return true
}

type intVar struct {
	basicVariable
	dflt int
	ptr  *int
}

func (v *intVar) set(val string) error {
	i, err := strconv.Atoi(val)
	if err != nil {
		return err
	}
	if v.ptr != nil {
		*v.ptr = i
	}
	return nil
}

func (v *intVar) setDefault() bool {
	if !v.hasDflt {
		return false
	}
	if v.ptr != nil {
		*v.ptr = v.dflt
	}
	return true
}

type floatVar struct {
	basicVariable
	dflt float64
	ptr  *float64
}

func (v *floatVar) set(val string) error {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return err
	}
	if v.ptr != nil {
		*v.ptr = f
	}
	return nil
}

func (v *floatVar) setDefault() bool {
	if !v.hasDflt {
		return false
	}
	if v.ptr != nil {
		*v.ptr = v.dflt
	}
	return true
}

type boolVar struct {
	basicVariable
	dflt bool
	ptr  *bool
}

func (v *boolVar) set(val string) error {
	if val != "true" && val != "false" {
		return fmt.Errorf(`invalid value "%s", should be "true" or "false"`, val)
	}
	if v.ptr != nil {
		*v.ptr = (val == "true")
	}
	return nil
}

func (v *boolVar) setDefault() bool {
	if !v.hasDflt {
		return false
	}
	if v.ptr != nil {
		*v.ptr = v.dflt
	}
	return true
}

type flagVar struct {
	basicVariable
	ptr *bool
}

func (v *flagVar) set(val string) error {
	if v.ptr != nil {
		*v.ptr = true
	}
	return nil
}

func (v *flagVar) setDefault() bool {
	if v.ptr != nil {
		*v.ptr = false
	}
	return true
}

type listVar struct {
	basicVariable
	dflt []string
	ptr  *[]string
	sep  string
}

func (v *listVar) set(val string) error {
	list := strings.Split(val, v.sep)
	if v.ptr != nil {
		*v.ptr = list
	}
	return nil
}

func (v *listVar) setDefault() bool {
	if !v.hasDflt {
		return false
	}
	if v.ptr != nil {
		*v.ptr = v.dflt
	}
	return true
}

type setVar struct {
	basicVariable
	dflt map[string]bool
	ptr  *map[string]bool
	sep  string
}

func (v *setVar) set(val string) error {
	list := strings.Split(val, v.sep)
	set := make(map[string]bool)
	for _, element := range list {
		ok := set[element]
		if ok {
			return fmt.Errorf("duplicate element: %s", element)
		}
		set[element] = true
	}
	if v.ptr != nil {
		*v.ptr = set
	}
	return nil
}

func (v *setVar) setDefault() bool {
	if !v.hasDflt {
		return false
	}
	if v.ptr != nil {
		*v.ptr = v.dflt
	}
	return true
}
