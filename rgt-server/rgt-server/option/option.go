package option

import (
	"fmt"
	"reflect"
	"rgt-server/log"
	"strconv"
	"strings"
	"time"
)

type Integer interface {
	~int8 | ~int16 | ~int32 | ~int64
}

type UInteger interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Float interface {
	~float32 | ~float64
}

type Option interface {
	Name() string
	Names() []string
	HasName(name string) bool
	GetString() string
	SetString(val string)
}

type TypedOption[T any] interface {
	Option
	Init(val T, name string, otherNames ...string)
	Get() T
	Set(val T)
	SetHook(hook func(val T))
}

// -------------------------------
// BaseOption: Option base fields
// -------------------------------
type BaseOption[T any] struct {
	value T
	names []string
	hook  func(val T)
}

func (o *BaseOption[T]) Init(val T, name string, otherNames ...string) {
	o.value = val
	o.names = make([]string, 0, len(otherNames)+1)
	o.names = append(o.names, name)
	o.names = append(o.names, otherNames...)
}

func (o *BaseOption[T]) Get() T {
	return o.value
}

func (o *BaseOption[T]) Name() string {
	if len(o.names) == 0 {
		return ""
	}
	return o.names[0]
}

func (o *BaseOption[T]) Names() []string {
	return o.names
}

func (o *BaseOption[T]) Set(val T) {
	o.value = val
	o.execHook()
}

func (o *BaseOption[T]) execHook() {
	if o.hook != nil {
		o.hook(o.value)
	}
}

func (o *BaseOption[T]) HasName(name string) bool {
	for _, optName := range o.names {
		if name == optName {
			return true
		}
	}
	return false
}

func (o *BaseOption[T]) SetHook(hook func(val T)) {
	o.hook = hook
}

// ------------------------------------
// UIntOption: unsigned integer option
// ------------------------------------
type UintOption[T UInteger] struct {
	BaseOption[T]
}

func NewUint[T UInteger](val T, name string, otherNames ...string) TypedOption[T] {
	option := &UintOption[T]{}
	option.Init(val, name, otherNames...)
	return option
}

func (v *UintOption[T]) GetString() string {
	return fmt.Sprint(v.value)
}

func (v *UintOption[T]) SetString(val string) {
	uintVal, _ := strconv.ParseUint(strings.TrimSpace(val), 10, 64)
	v.value = T(uintVal)
	v.execHook()
}

// -------------------------------
// IntOption: integer Option
// -------------------------------
type IntOption[T Integer] struct {
	BaseOption[T]
}

func (v *IntOption[T]) GetString() string {
	return fmt.Sprint(v.value)
}

func (v *IntOption[T]) SetString(val string) {
	intVal, _ := strconv.ParseInt(strings.TrimSpace(val), 10, 64)
	v.value = T(intVal)
	v.execHook()
}

// -------------------------------
// FloatOption: Float Option
// -------------------------------
type FloatOption[T Float] struct {
	BaseOption[T]
}

func (v *FloatOption[T]) GetString() string {
	return fmt.Sprint(v.value)
}

func (v *FloatOption[T]) SetString(value string) {
	var tmp T
	var val float64
	val, _ = strconv.ParseFloat(strings.TrimSpace(value), reflect.TypeOf(tmp).Bits())
	v.value = T(val)
	v.execHook()
}

// -------------------------------
// StringOption: String Option
// -------------------------------
type StringOption struct {
	BaseOption[string]
}

func NewString(val string, name string, otherNames ...string) TypedOption[string] {
	op := &StringOption{}
	op.Init(val, name, otherNames...)
	return op
}

func (v *StringOption) GetString() string {
	return v.value
}

func (v *StringOption) SetString(val string) {
	v.value = val
	v.execHook()
}

// --------------------------------
// LogLevelOption: LogLevel Option
// --------------------------------
type LogLevelOption struct {
	BaseOption[log.LogLevel]
}

func NewLogLevel(val log.LogLevel, name string, otherNames ...string) TypedOption[log.LogLevel] {
	option := &LogLevelOption{}
	option.Init(val, name, otherNames...)
	return option
}

func (v *LogLevelOption) GetString() string {
	return v.value.Name()
}

func (v *LogLevelOption) SetString(val string) {
	v.value = log.LogLevelFromName(strings.TrimSpace(val))
	v.execHook()
}

// -------------------------------------
// DurationOption: Time duration Option
// -------------------------------------
type DurationOption struct {
	BaseOption[time.Duration]
}

func NewDuration(val time.Duration, name string, otherNames ...string) TypedOption[time.Duration] {
	option := &DurationOption{}
	option.Init(val, name, otherNames...)
	return option
}

func (v *DurationOption) GetString() string {
	return v.value.String()
}

func (v *DurationOption) SetString(str string) {
	duration, err := time.ParseDuration(strings.TrimSpace(str))
	if err == nil {
		v.value = duration
	} else {
		v.value = 0
	}
	v.execHook()
}

// --------------------------
// BoolOption: Boolean Option
// --------------------------
type BoolOption struct {
	BaseOption[bool]
}

func NewBool(val bool, name string, otherNames ...string) TypedOption[bool] {
	option := &BoolOption{}
	option.Init(val, name, otherNames...)
	return option
}

func (v *BoolOption) GetString() string {
	if v.value {
		return "true"
	} else {
		return "false"
	}
}

func (v *BoolOption) SetString(str string) {
	v.value = strings.EqualFold(strings.TrimSpace(str), "true")
	v.execHook()
}

// --------------------------------
// CustomOption: Custom type Option
// --------------------------------
type CustomOption[T any] struct {
	BaseOption[T]
	toString func(val T) string
	toValue  func(val string) T
}

func NewCustom[T any](val T, toString func(val T) string, toValue func(val string) T, name string, otherNames ...string) TypedOption[T] {
	op := &CustomOption[T]{
		toString: toString,
		toValue:  toValue}
	op.Init(val, name, otherNames...)
	return op
}

func (v *CustomOption[T]) GetString() string {
	return v.toString(v.value)
}

func (v *CustomOption[T]) SetString(str string) {
	v.value = v.toValue(str)
	v.execHook()
}

// -------------------
// Options set
// -------------------
type Options struct {
	options         map[string]Option
	allNamesOptions map[string]Option
}

func NewOptions() *Options {
	return &Options{
		options:         make(map[string]Option),
		allNamesOptions: make(map[string]Option)}
}

func (options *Options) List() []Option {
	optionsList := make([]Option, 0, len(options.options))
	for _, o := range options.options {
		optionsList = append(optionsList, o)
	}
	return optionsList
}

func (o *Options) Add(option Option) {
	o.options[option.Name()] = option
	for _, name := range option.Names() {
		o.allNamesOptions[name] = option
	}
}

func (options *Options) Get(name string) Option {
	option, found := options.allNamesOptions[name]
	if found {
		return option
	}
	return nil
}

func (options *Options) Delete(name string) Option {
	option, found := options.allNamesOptions[name]
	if found {
		delete(options.options, option.Name())
		for _, name := range option.Names() {
			delete(options.allNamesOptions, name)
		}
		return option
	}
	return nil
}
