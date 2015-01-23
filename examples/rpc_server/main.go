package main

import (
	"errors"
	"github.com/funny/link"
	"log"
	"reflect"
)

// NOTE: Thid code copy from default rpc package.
// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

type RPCServer struct {
	server   *link.Server
	services []rpcService
}

type rpcService struct {
	Name     string
	Receiver interface{}
	Methods  []rpcMethod
}

type rpcMethod struct {
	Name      string
	Method    reflect.Method
	ArgsType  reflect.Type
	ReplyType reflect.Type
}

func (server *RPCServer) Register(service interface{}) error {
	return server.register("", service)
}

func (server *RPCServer) RegisterName(name string, service interface{}) error {
	return server.register(name, service)
}

func (server *RPCServer) register(name string, service interface{}) error {
	var (
		serviceType  = reflect.TypeOf(service)
		serviceValue = reflect.ValueOf(service)
	)

	sname := name

	if sname == "" {
		sname = reflect.Indirect(service).Type().Name()
	}

	if sname == "" {
		err := "RPC no service name for type: " + serviceType.String()
		log.Println(err)
		return errors.New(err)
	}

	if !isExported(sname) && name == "" {
		err := "RPC service type " + sname + " is not exported"
		log.println(err)
		return errors.New(err)
	}

	for i := 0; i < len(server.services); i++ {
		if server.services[i].Name == sname {
			return errors.New("RPC service already defined: " + sname)
		}
	}

	serviceInfo := rpcService{
		Name:     sname,
		Receiver: service,
		Methods:  getRpcMethods(),
	}

	if len(methods) == 0 {
		err := ""
		// To help the user, see if a pointer receiver would work.
		methods := getRpcMethods(reflect.PtrTo(serviceType), false)
		if len(methods) != 0 {
			err = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			err = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		log.Print(err)
		return errors.New(err)
	}

	server.services = append(server.services, serviceInfo)

	return nil
}

func getRpcMethods() []rpcMethod {
	methods := make([]rpcMethod, 0, serviceType.NumMethod())
	for i := 0; i < len(methods); i++ {
		var (
			method     = serviceType.Method(i)
			methodType = method.Type
			methodInfo = rpcMethod{
				Name:      method.Name,
				Method:    methodType,
				ArgsType:  methodType.In(1),
				ReplyType: methodType.In(2),
			}
		)

		if !isExportedOrBuiltinType(methodInfo.ArgsType) {
			log.Println("RPC method", method.Name, "argument type not exported:", methodInfo.ArgsType)
			continue
		}

		if !isExportedOrBuiltinType(methodInfo.ReplyType) {
			log.Println("RPC method", method.Name, "reply type not exported:", methodInfo.ReplyType)
			continue
		}

		if methodInfo.ReplyType.Kind() != reflect.Ptr {
			log.Println("RPC method", method.Name, "reply type not a pointer:", methodInfo.ReplyType)
			continue
		}

		if methodType.NumOut() != 1 {
			log.Println("RPC method", method.Name, "wrong number of return value:", methodType.NumOut())
			continue
		}

		if returnType := methodType.Out(0); returnType != typeOfError {
			log.Println("RPC method", method.Name, "return type not error:", returnType)
			continue
		}

		methods = append(serviceInfo.Methods, methodInfo)
	}
	return methods
}

// NOTE: Thid code copy from default rpc package.
// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// NOTE: Thid code copy from default rpc package.
// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

func main() {

}
