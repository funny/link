package rpc

import (
	"encoding/json"
	"errors"
	"github.com/funny/link"
	"log"
	"reflect"
	"unicode"
	"unicode/utf8"
)

// NOTE: This code copy from default rpc package.
// Precompute the reflect type for error.  Can't use error directly
// because Typeof takes an empty interface value.  This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()
var typeOfSession = reflect.TypeOf((*link.Session)(nil))

type Server struct {
	server   *link.Server
	services []rpcService
}

type rpcService struct {
	Name     string
	Receiver reflect.Value
	Methods  []rpcMethod
}

type rpcMethod struct {
	Name      string
	Method    reflect.Method
	ArgsType  reflect.Type
	ReplyType reflect.Type
}

func NewServer(network, address string) (*Server, error) {
	server, err := link.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &Server{server, nil}, nil
}

func (server *Server) Stop() {
	server.server.Stop()
}

func (server *Server) Serve() error {
	return server.server.Serve(func(session *link.Session) {
		session.Process(func(msg *link.InBuffer) error {
			var err error
			defer func() {
				if e := recover(); e != nil {
					log.Println("RPC error:", e)
					e = errors.New("RPC failed")
				}
			}()
			service := msg.ReadString(int(msg.ReadUvarint()))
			method := msg.ReadString(int(msg.ReadUvarint()))
			seqNum := msg.ReadUint32LE()
			for i := 0; i < len(server.services); i++ {
				s := &server.services[i]
				if s.Name == service {
					for j := 0; j < len(s.Methods); j++ {
						m := &s.Methods[j]
						if m.Name == method {
							var argv reflect.Value
							argIsValue := false
							if m.ArgsType.Kind() == reflect.Ptr {
								argv = reflect.New(m.ArgsType.Elem())
							} else {
								argv = reflect.New(m.ArgsType)
								argIsValue = true
							}
							if err := json.NewDecoder(msg).Decode(argv.Interface()); err != nil {
								log.Println("RPC decode request argument failed:", err)
								return err
							}
							if argIsValue {
								argv = argv.Elem()
							}
							replyv := reflect.New(m.ReplyType.Elem())
							returnValues := m.Method.Func.Call([]reflect.Value{
								s.Receiver,
								argv,
								replyv,
							})
							err := ""
							if errInterface := returnValues[0].Interface(); errInterface != nil {
								err = errInterface.(error).Error()
							}
							return session.SendFunc(func(buffer *link.OutBuffer) error {
								buffer.WriteUint32LE(seqNum)
								buffer.WriteUint32LE(uint32(len(err)))
								buffer.WriteString(err)
								return json.NewEncoder(buffer).Encode(replyv.Interface()) // TODO
							})
						}
					}
				}
			}
			err = errors.New("RPC service not exists: " + service + "." + method)
			return err
		})
	})
}

func (server *Server) Register(service interface{}) error {
	return server.register("", service)
}

func (server *Server) RegisterName(name string, service interface{}) error {
	return server.register(name, service)
}

func (server *Server) register(name string, service interface{}) error {
	var (
		serviceType  = reflect.TypeOf(service)
		serviceValue = reflect.ValueOf(service)
	)

	sname := name

	if sname == "" {
		sname = reflect.Indirect(serviceValue).Type().Name()
	}

	if sname == "" {
		err := "RPC no service name for type: " + serviceType.String()
		log.Println(err)
		return errors.New(err)
	}

	if !isExported(sname) && name == "" {
		err := "RPC service type " + sname + " is not exported"
		log.Println(err)
		return errors.New(err)
	}

	for i := 0; i < len(server.services); i++ {
		if server.services[i].Name == sname {
			return errors.New("RPC service already defined: " + sname)
		}
	}

	serviceInfo := rpcService{
		Name:     sname,
		Receiver: serviceValue,
		Methods:  getRpcMethods(serviceType, true),
	}

	if len(serviceInfo.Methods) == 0 {
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

func getRpcMethods(serviceType reflect.Type, reportError bool) []rpcMethod {
	methods := make([]rpcMethod, 0, serviceType.NumMethod())
	for i := 0; i < cap(methods); i++ {
		var (
			method     = serviceType.Method(i)
			methodType = method.Type
		)

		if methodType.NumIn() != 3 {
			if reportError {
				log.Println("RPC method", method.Name, "has wrong number of parameter:", methodType.NumIn())
			}
			continue
		}

		var methodInfo = rpcMethod{
			Name:      method.Name,
			Method:    method,
			ArgsType:  methodType.In(1),
			ReplyType: methodType.In(2),
		}

		if !isExportedOrBuiltinType(methodInfo.ArgsType) {
			if reportError {
				log.Println("RPC method", method.Name, "argument type not exported:", methodInfo.ArgsType)
			}
			continue
		}

		if !isExportedOrBuiltinType(methodInfo.ReplyType) {
			if reportError {
				log.Println("RPC method", method.Name, "reply type not exported:", methodInfo.ReplyType)
			}
			continue
		}

		if methodInfo.ReplyType.Kind() != reflect.Ptr {
			if reportError {
				log.Println("RPC method", method.Name, "reply type not a pointer:", methodInfo.ReplyType)
			}
			continue
		}

		if methodType.NumOut() != 1 {
			if reportError {
				log.Println("RPC method", method.Name, "wrong number of return value:", methodType.NumOut())
			}
			continue
		}

		if returnType := methodType.Out(0); returnType != typeOfError {
			if reportError {
				log.Println("RPC method", method.Name, "return type not error:", returnType)
			}
			continue
		}

		methods = append(methods, methodInfo)
	}
	return methods
}

// NOTE: This code copy from default rpc package.
// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// NOTE: This code copy from default rpc package.
// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}
