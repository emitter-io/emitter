package cli

import (
	"flag"
	"fmt"
	"strings"

	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/values"
)

// BoolOpt describes a boolean option
type BoolOpt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option
	EnvVar string
	// The option's initial value
	Value bool
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o BoolOpt) value(into *bool) (flag.Value, *bool) {
	if into == nil {
		into = new(bool)
	}
	return values.NewBool(into, o.Value), into
}

// StringOpt describes a string option
type StringOpt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option
	EnvVar string
	// The option's initial value
	Value string
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o StringOpt) value(into *string) (flag.Value, *string) {
	if into == nil {
		into = new(string)
	}
	return values.NewString(into, o.Value), into
}

// IntOpt describes an int option
type IntOpt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option
	EnvVar string
	// The option's initial value
	Value int
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o IntOpt) value(into *int) (flag.Value, *int) {
	if into == nil {
		into = new(int)
	}
	return values.NewInt(into, o.Value), into
}

// Float64Opt describes an float64 option
type Float64Opt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option
	EnvVar string
	// The option's initial value
	Value float64
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o Float64Opt) value(into *float64) (flag.Value, *float64) {
	if into == nil {
		into = new(float64)
	}
	return values.NewFloat64(into, o.Value), into
}

// StringsOpt describes a string slice option
type StringsOpt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option.
	// The env variable should contain a comma separated list of values
	EnvVar string
	// The option's initial value
	Value []string
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o StringsOpt) value(into *[]string) (flag.Value, *[]string) {
	if into == nil {
		into = new([]string)
	}
	return values.NewStrings(into, o.Value), into
}

// IntsOpt describes an int slice option
type IntsOpt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option.
	// The env variable should contain a comma separated list of values
	EnvVar string
	// The option's initial value
	Value []int
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o IntsOpt) value(into *[]int) (flag.Value, *[]int) {
	if into == nil {
		into = new([]int)
	}
	return values.NewInts(into, o.Value), into

}

// Floats64Opt describes an int slice option
type Floats64Opt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option.
	// The env variable should contain a comma separated list of values
	EnvVar string
	// The option's initial value
	Value []float64
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o Floats64Opt) value(into *[]float64) (flag.Value, *[]float64) {
	if into == nil {
		into = new([]float64)
	}
	return values.NewFloats64(into, o.Value), into

}

// VarOpt describes an option where the type and format of the value is controlled by the developer
type VarOpt struct {
	// A space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
	// The one letter names will then be called with a single dash (short option), the others with two (long options).
	Name string
	// The option description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this option
	EnvVar string
	// A value implementing the flag.Value type (will hold the final value)
	Value flag.Value
	// A boolean to display or not the current value of the option in the help message
	HideValue bool
	// Set to true if this option was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (o VarOpt) value() flag.Value {
	return o.Value
}

/*
BoolOpt defines a boolean option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to a bool) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) BoolOpt(name string, value bool, desc string) *bool {
	return c.Bool(BoolOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
BoolOptPtr defines a bool option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to a int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) BoolOptPtr(into *bool, name string, value bool, desc string) {
	c.BoolPtr(into, BoolOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringOpt defines a string option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to a string) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringOpt(name string, value string, desc string) *string {
	return c.String(StringOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringOptPtr defines a string option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to a int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringOptPtr(into *string, name string, value string, desc string) {
	c.StringPtr(into, StringOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntOpt defines an int option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to an int) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntOpt(name string, value int, desc string) *int {
	return c.Int(IntOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntOptPtr defines a int option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to an int) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntOptPtr(into *int, name string, value int, desc string) {
	c.IntPtr(into, IntOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Float64Opt defines an float64 option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to an float64) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Float64Opt(name string, value float64, desc string) *float64 {
	return c.Float64(Float64Opt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Float64OptPtr defines a float64 option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to a float64) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Float64OptPtr(into *float64, name string, value float64, desc string) {
	c.Float64Ptr(into, Float64Opt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringsOpt defines a string slice option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to a string slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringsOpt(name string, value []string, desc string) *[]string {
	return c.Strings(StringsOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringsOptPtr defines a string slice option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to a int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringsOptPtr(into *[]string, name string, value []string, desc string) {
	c.StringsPtr(into, StringsOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntsOpt defines an int slice option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to an int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntsOpt(name string, value []int, desc string) *[]int {
	return c.Ints(IntsOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntsOptPtr defines a int slice option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to a int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntsOptPtr(into *[]int, name string, value []int, desc string) {
	c.IntsPtr(into, IntsOpt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Floats64Opt defines an float64 slice option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result should be stored in a variable (a pointer to an float64 slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Floats64Opt(name string, value []float64, desc string) *[]float64 {
	return c.Floats64(Floats64Opt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Floats64OptPtr defines a int slice option on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The into parameter points to a variable (a pointer to a int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Floats64OptPtr(into *[]float64, name string, value []float64, desc string) {
	c.Floats64Ptr(into, Floats64Opt{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
VarOpt defines an option where the type and format is controlled by the developer.

The name is a space separated list of the option names *WITHOUT* the dashes, e.g. `f force` and *NOT* `-f --force`.
The one letter names will then be called with a single dash (short option), the others with two (long options).


The result will be stored in the value parameter (a value implementing the flag.Value interface) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) VarOpt(name string, value flag.Value, desc string) {
	c.mkOpt(container.Container{Name: name, Desc: desc, Value: value})
}

func mkOptStrs(optName string) []string {
	res := strings.Fields(optName)
	for i, name := range res {
		prefix := "-"
		if len(name) > 1 {
			prefix = "--"
		}
		res[i] = prefix + name
	}
	return res
}

func (c *Cmd) mkOpt(opt container.Container) {
	opt.ValueSetFromEnv = values.SetFromEnv(opt.Value, opt.EnvVar)

	opt.Names = mkOptStrs(opt.Name)

	c.options = append(c.options, &opt)
	for _, name := range opt.Names {
		if _, found := c.optionsIdx[name]; found {
			panic(fmt.Sprintf("duplicate option name %q", name))
		}
		c.optionsIdx[name] = &opt
	}
}
