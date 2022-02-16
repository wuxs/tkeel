// Code generated by protoc-gen-go-errors. DO NOT EDIT.

package v1

import (
	errors "github.com/tkeel-io/kit/errors"
	codes "google.golang.org/grpc/codes"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the ego package it is being compiled against.
const _ = errors.SupportPackageIsVersion1

var errUnknown *errors.TError
var errInvalidArgument *errors.TError
var errInternalStore *errors.TError
var errInternalError *errors.TError
var errPermissionNotFound *errors.TError
var errRoleNotFound *errors.TError
var errRoleHasBeenExsist *errors.TError

func init() {
	errUnknown = errors.New(int(codes.Unknown), "io.tkeel.rudder.api.rbac.v1.ERR_UNKNOWN", Error_ERR_UNKNOWN.String())
	errors.Register(errUnknown)
	errInvalidArgument = errors.New(int(codes.InvalidArgument), "io.tkeel.rudder.api.rbac.v1.ERR_INVALID_ARGUMENT", Error_ERR_INVALID_ARGUMENT.String())
	errors.Register(errInvalidArgument)
	errInternalStore = errors.New(int(codes.Internal), "io.tkeel.rudder.api.rbac.v1.ERR_INTERNAL_STORE", Error_ERR_INTERNAL_STORE.String())
	errors.Register(errInternalStore)
	errInternalError = errors.New(int(codes.Internal), "io.tkeel.rudder.api.rbac.v1.ERR_INTERNAL_ERROR", Error_ERR_INTERNAL_ERROR.String())
	errors.Register(errInternalError)
	errPermissionNotFound = errors.New(int(codes.NotFound), "io.tkeel.rudder.api.rbac.v1.ERR_PERMISSION_NOT_FOUND", Error_ERR_PERMISSION_NOT_FOUND.String())
	errors.Register(errPermissionNotFound)
	errRoleNotFound = errors.New(int(codes.NotFound), "io.tkeel.rudder.api.rbac.v1.ERR_ROLE_NOT_FOUND", Error_ERR_ROLE_NOT_FOUND.String())
	errors.Register(errRoleNotFound)
	errRoleHasBeenExsist = errors.New(int(codes.InvalidArgument), "io.tkeel.rudder.api.rbac.v1.ERR_ROLE_HAS_BEEN_EXSIST", Error_ERR_ROLE_HAS_BEEN_EXSIST.String())
	errors.Register(errRoleHasBeenExsist)
}

func ErrUnknown() errors.Error {
	return errUnknown
}

func ErrInvalidArgument() errors.Error {
	return errInvalidArgument
}

func ErrInternalStore() errors.Error {
	return errInternalStore
}

func ErrInternalError() errors.Error {
	return errInternalError
}

func ErrPermissionNotFound() errors.Error {
	return errPermissionNotFound
}

func ErrRoleNotFound() errors.Error {
	return errRoleNotFound
}

func ErrRoleHasBeenExsist() errors.Error {
	return errRoleHasBeenExsist
}
