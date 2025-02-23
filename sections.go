package konfig

import (
	"reflect"
)

type KonfigSection interface {
	Name() string
	Validate() error
	OnLoad() error
}

type konfigSectionImpl struct {
	data interface{} // the struct that contains the data that will be saved

	sectionNameFunc func() string
	validateFunc    func() error
	onLoadFunc      func() error
}

func (c *konfigSectionImpl) Name() string {
	if c.sectionNameFunc != nil {
		return c.sectionNameFunc()
	}
	typeName := reflect.TypeOf(c.data).Name()
	return typeName
}

func (c *konfigSectionImpl) Validate() error {
	if c.validateFunc != nil {
		return c.validateFunc()
	}
	return nil // Default: no validation
}

func (c *konfigSectionImpl) OnLoad() error {
	if c.onLoadFunc != nil {
		return c.onLoadFunc()
	}
	return nil // Default: no onload action
}

type SectionOption func(impl *konfigSectionImpl)

func WithSectionName(nameFn func() string) SectionOption {
	return func(impl *konfigSectionImpl) {
		impl.sectionNameFunc = nameFn
	}
}

func WithValidate(validateFn func() error) SectionOption {
	return func(impl *konfigSectionImpl) {
		impl.validateFunc = validateFn
	}
}

func WithOnLoad(onLoadFn func() error) SectionOption {
	return func(impl *konfigSectionImpl) {
		impl.onLoadFunc = onLoadFn
	}
}

func NewKonfigSection(
	data interface{},
	options ...SectionOption,
) KonfigSection {
	if data == nil {
		panic("data struct cannot be nil, pass in a pointer to your struct containing data")
	}

	if reflect.TypeOf(data).Kind() != reflect.Ptr || reflect.TypeOf(data).Elem().Kind() != reflect.Struct {
		panic("data must be a pointer to a struct, but is " + reflect.TypeOf(data).Kind().String())
	}

	impl := &konfigSectionImpl{
		data: data,

		sectionNameFunc: nil,
		validateFunc:    nil,
		onLoadFunc:      nil,
	}

	for _, opt := range options {
		opt(impl)
	}

	return impl
}
