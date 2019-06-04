package cli

import (
	"flag"
	"fmt"

	"github.com/jawher/mow.cli/internal/lexer"

	"github.com/jawher/mow.cli/internal/container"
	"github.com/jawher/mow.cli/internal/values"
)

// BoolArg describes a boolean argument
type BoolArg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument
	EnvVar string
	// The argument's initial value
	Value bool
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a BoolArg) value(into *bool) (flag.Value, *bool) {
	if into == nil {
		into = new(bool)
	}
	return values.NewBool(into, a.Value), into
}

// StringArg describes a string argument
type StringArg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument
	EnvVar string
	// The argument's initial value
	Value string
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a StringArg) value(into *string) (flag.Value, *string) {
	if into == nil {
		into = new(string)
	}
	return values.NewString(into, a.Value), into
}

// IntArg describes an int argument
type IntArg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument
	EnvVar string
	// The argument's initial value
	Value int
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a IntArg) value(into *int) (flag.Value, *int) {
	if into == nil {
		into = new(int)
	}
	return values.NewInt(into, a.Value), into
}

// Float64Arg describes an float64 argument
type Float64Arg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument
	EnvVar string
	// The argument's initial value
	Value float64
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a Float64Arg) value(into *float64) (flag.Value, *float64) {
	if into == nil {
		into = new(float64)
	}
	return values.NewFloat64(into, a.Value), into
}

// StringsArg describes a string slice argument
type StringsArg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument.
	// The env variable should contain a comma separated list of values
	EnvVar string
	// The argument's initial value
	Value []string
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a StringsArg) value(into *[]string) (flag.Value, *[]string) {
	if into == nil {
		into = new([]string)
	}
	return values.NewStrings(into, a.Value), into
}

// IntsArg describes an int slice argument
type IntsArg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument.
	// The env variable should contain a comma separated list of values
	EnvVar string
	// The argument's initial value
	Value []int
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a IntsArg) value(into *[]int) (flag.Value, *[]int) {
	if into == nil {
		into = new([]int)
	}
	return values.NewInts(into, a.Value), into
}

// Floats64Arg describes an int slice argument
type Floats64Arg struct {
	// The argument name as will be shown in help messages
	Name string
	// The argument description as will be shown in help messages
	Desc string
	// A space separated list of environment variables names to be used to initialize this argument.
	// The env variable should contain a comma separated list of values
	EnvVar string
	// The argument's initial value
	Value []float64
	// A boolean to display or not the current value of the argument in the help message
	HideValue bool
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a Floats64Arg) value(into *[]float64) (flag.Value, *[]float64) {
	if into == nil {
		into = new([]float64)
	}
	return values.NewFloats64(into, a.Value), into
}

// VarArg describes an argument where the type and format of the value is controlled by the developer
type VarArg struct {
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
	// Set to true if this arg was set by the user (as opposed to being set from env or not set at all)
	SetByUser *bool
}

func (a VarArg) value() flag.Value {
	return a.Value
}

/*
BoolArg defines a boolean argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to a bool) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) BoolArg(name string, value bool, desc string) *bool {
	return c.Bool(BoolArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
BoolArgPtr defines a boolean argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a bool) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) BoolArgPtr(into *bool, name string, value bool, desc string) {
	c.BoolPtr(into, BoolArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringArg defines a string argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to a string) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringArg(name string, value string, desc string) *string {
	return c.String(StringArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringArgPtr defines a string argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a string) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringArgPtr(into *string, name string, value string, desc string) {
	c.StringPtr(into, StringArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntArg defines an int argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to an int) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntArg(name string, value int, desc string) *int {
	return c.Int(IntArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntArgPtr defines an int argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a int) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntArgPtr(into *int, name string, value int, desc string) {
	c.IntPtr(into, IntArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Float64Arg defines an float64 argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to an float64) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Float64Arg(name string, value float64, desc string) *float64 {
	return c.Float64(Float64Arg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Float64ArgPtr defines an float64 argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a float64) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Float64ArgPtr(into *float64, name string, value float64, desc string) {
	c.Float64Ptr(into, Float64Arg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringsArg defines a string slice argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to a string slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringsArg(name string, value []string, desc string) *[]string {
	return c.Strings(StringsArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
StringsArgPtr defines a string slice argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a string slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) StringsArgPtr(into *[]string, name string, value []string, desc string) {
	c.StringsPtr(into, StringsArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntsArg defines an int slice argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to an int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntsArg(name string, value []int, desc string) *[]int {
	return c.Ints(IntsArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
IntsArgPtr defines a int slice argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a int slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) IntsArgPtr(into *[]int, name string, value []int, desc string) {
	c.IntsPtr(into, IntsArg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Floats64Arg defines an float64 slice argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The result should be stored in a variable (a pointer to an float64 slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Floats64Arg(name string, value []float64, desc string) *[]float64 {
	return c.Floats64(Floats64Arg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
Floats64ArgPtr defines a float64 slice argument on the command c named `name`, with an initial value of `value` and a description of `desc` which will be used in help messages.

The into parameter points to a variable (a pointer to a float64 slice) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) Floats64ArgPtr(into *[]float64, name string, value []float64, desc string) {
	c.Floats64Ptr(into, Floats64Arg{
		Name:  name,
		Value: value,
		Desc:  desc,
	})
}

/*
VarArg defines an argument where the type and format is controlled by the developer on the command c named `name` and a description of `desc` which will be used in help messages.

The result will be stored in the value parameter (a value implementing the flag.Value interface) which will be populated when the app is run and the call arguments get parsed
*/
func (c *Cmd) VarArg(name string, value flag.Value, desc string) {
	c.mkArg(container.Container{Name: name, Desc: desc, Value: value})
}

func (c *Cmd) mkArg(arg container.Container) {
	if !validArgName(arg.Name) {
		panic(fmt.Sprintf("invalid argument name %q: must be in all caps", arg.Name))
	}
	if _, found := c.argsIdx[arg.Name]; found {
		panic(fmt.Sprintf("duplicate argument name %q", arg.Name))
	}

	arg.ValueSetFromEnv = values.SetFromEnv(arg.Value, arg.EnvVar)

	c.args = append(c.args, &arg)
	c.argsIdx[arg.Name] = &arg
}

func validArgName(n string) bool {
	tokens, err := lexer.Tokenize(n)
	if err != nil {
		return false
	}
	if len(tokens) != 1 {
		return false
	}

	return tokens[0].Typ == lexer.TTArg
}
