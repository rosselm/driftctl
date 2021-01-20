package aws

import "fmt"

type ListError interface {
	error
	SupplierError
	ListedTypeError() string
}
type SupplierError interface {
	error
	RootCause() error
	SupplierType() string
	Context() map[string]string
}

type BaseSupplierError struct {
	err          error
	context      map[string]string
	supplierType string
}

func NewBaseSupplierError(err error, context map[string]string, supplierType string) *BaseSupplierError {
	context["SupplierType"] = supplierType
	return &BaseSupplierError{err: err, context: context, supplierType: supplierType}
}

func (b *BaseSupplierError) Error() string {
	return fmt.Sprintf("error in supplier %s: %s", b.supplierType, b.err)
}

func (b *BaseSupplierError) RootCause() error {
	return b.err
}

func (b *BaseSupplierError) SupplierType() string {
	return b.supplierType
}

func (b *BaseSupplierError) Context() map[string]string {
	return b.context
}

type BaseListError struct {
	BaseSupplierError
	listedTypeError string
}

func NewBaseListError(error error, supplierType string, listedTypeError string) *BaseListError {
	context := map[string]string{
		"ListedTypeError": listedTypeError,
	}
	return &BaseListError{
		BaseSupplierError: *NewBaseSupplierError(error, context, supplierType),
		listedTypeError:   listedTypeError,
	}
}

func (b *BaseListError) ListedTypeError() string {
	return b.listedTypeError
}
