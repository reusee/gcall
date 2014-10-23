package gcall

/*
#include <girepository.h>
#include <girffi.h>

void arg_set_boolean(GIArgument *arg, gboolean b) {
	arg->v_boolean = b;
}

void arg_set_int8(GIArgument *arg, gint8 i) {
	arg->v_int8 = i;
}

void arg_set_int16(GIArgument *arg, gint16 i) {
	arg->v_int16 = i;
}

void arg_set_int32(GIArgument *arg, gint32 i) {
	arg->v_int32 = i;
}

void arg_set_int64(GIArgument *arg, gint64 i) {
	arg->v_int64 = i;
}

void arg_set_int(GIArgument *arg, gint i) {
  arg->v_int = i;
}

void arg_set_uint8(GIArgument *arg, guint8 i) {
	arg->v_uint8 = i;
}

void arg_set_uint16(GIArgument *arg, guint16 i) {
	arg->v_uint16 = i;
}

void arg_set_uint32(GIArgument *arg, guint32 i) {
	arg->v_uint32 = i;
}

void arg_set_uint64(GIArgument *arg, guint64 i) {
	arg->v_uint64 = i;
}

void arg_set_uint(GIArgument *arg, guint i) {
	arg->v_uint = i;
}

void arg_set_float(GIArgument *arg, gfloat f) {
	arg->v_float = f;
}

void arg_set_double(GIArgument *arg, gdouble d) {
	arg->v_double = d;
}

void arg_set_pointer(GIArgument *arg, void *p) {
	arg->v_pointer = p;
}

void arg_set_string(GIArgument *arg, gchar* s) {
	arg->v_string = s;
}

gboolean arg_get_boolean(GIArgument *arg) {
	return arg->v_boolean;
}

gint8 arg_get_int8(GIArgument *arg) {
	return arg->v_int8;
}

gint16 arg_get_int16(GIArgument *arg) {
	return arg->v_int16;
}

gint32 arg_get_int32(GIArgument *arg) {
	return arg->v_int32;
}

gint64 arg_get_int64(GIArgument *arg) {
	return arg->v_int64;
}

guint8 arg_get_uint8(GIArgument *arg) {
	return arg->v_uint8;
}

guint16 arg_get_uint16(GIArgument *arg) {
	return arg->v_uint16;
}

guint32 arg_get_uint32(GIArgument *arg) {
	return arg->v_uint32;
}

guint64 arg_get_uint64(GIArgument *arg) {
	return arg->v_uint64;
}

gfloat arg_get_float(GIArgument *arg) {
	return arg->v_float;
}

gdouble arg_get_double(GIArgument *arg) {
	return arg->v_double;
}

char* arg_get_string(GIArgument *arg) {
	return arg->v_string;
}

void* arg_get_pointer(GIArgument *arg) {
	return arg->v_pointer;
}

#cgo pkg-config: gobject-introspection-1.0
*/
import "C"
import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

var (
	pt = fmt.Printf
	sp = fmt.Sprintf

	repo = C.g_irepository_get_default()
)

type _Info struct {
	FnInfo     *C.GIFunctionInfo
	Dirs       []C.GIDirection
	ReturnType C.GITypeTag
	OutTypes   []C.GITypeTag
}

var infoCache = make(map[string]_Info)

func Require(ns, ver string) {
	var err *C.GError
	C.g_irepository_require(repo, gs(ns), gs(ver), 0, &err)
	if err != nil {
		panic(sp("%s", C.GoString((*C.char)(unsafe.Pointer(err.message)))))
	}
}

func Call(name string, args ...interface{}) []interface{} {
	ret, err := Pcall(name, args...)
	if err != nil {
		panic(err)
	}
	return ret
}

func Pcall(name string, args ...interface{}) (ret []interface{}, err error) {
	// get function info
	var info _Info
	var ok bool
	if info, ok = infoCache[name]; !ok {
		parts := strings.Split(name, ".")
		namespace := parts[0]
		var cinfo interface{}
		for i, name := range parts[1:] {
			if i == 0 {
				cinfo = C.g_irepository_find_by_name(repo, gs(namespace), gs(name))
			} else {
				switch ty := C.g_base_info_get_type((*C.GIBaseInfo)(unsafe.Pointer(reflect.ValueOf(cinfo).Pointer()))); ty {
				case C.GI_INFO_TYPE_OBJECT:
					cinfo = C.g_object_info_find_method(cinfo.(*C.GIFunctionInfo), gs(name))
				default:
					panic(sp("not handle base info type %v", ty))
				}
			}
		}
		if fnInfo, ok := cinfo.(*C.GIFunctionInfo); !ok {
			panic(sp("%s is not a function", name))
		} else {
			callable := (*C.GICallableInfo)(unsafe.Pointer(fnInfo))
			info = _Info{
				FnInfo:     fnInfo,
				ReturnType: C.g_type_info_get_tag(C.g_callable_info_get_return_type(callable)),
			}
			var dirs []C.GIDirection
			if C.g_callable_info_is_method(callable) == C.TRUE {
				dirs = append(dirs, C.GI_DIRECTION_IN)
			}
			var outTypes []C.GITypeTag
			nArgs := C.g_callable_info_get_n_args(callable)
			for i := C.gint(0); i < nArgs; i++ {
				argInfo := C.g_callable_info_get_arg(callable, i)
				dir := C.g_arg_info_get_direction(argInfo)
				dirs = append(dirs, dir)
				if dir == C.GI_DIRECTION_OUT || dir == C.GI_DIRECTION_INOUT {
					outTypes = append(outTypes, C.g_type_info_get_tag(C.g_arg_info_get_type(argInfo)))
				}
			}
			info.Dirs = dirs
			info.OutTypes = outTypes
			infoCache[name] = info
		}
	}

	// prepare arguments
	var inArgs, outArgs []C.GIArgument
	argIndex := 0
	for _, dir := range info.Dirs {
		switch dir {
		case C.GI_DIRECTION_IN:
			inArgs = append(inArgs, garg(args[argIndex]))
			argIndex++
		case C.GI_DIRECTION_OUT:
			var outArg C.GIArgument
			outArgs = append(outArgs, outArg)
		case C.GI_DIRECTION_INOUT:
			arg := garg(args[argIndex])
			argIndex++
			inArgs = append(inArgs, arg)
			outArgs = append(outArgs, arg)
		}
	}

	// invoke
	var retArg C.GIArgument
	var gerr *C.GError
	var ins, outs *C.GIArgument
	if len(inArgs) > 0 {
		ins = &inArgs[0]
	}
	if len(outArgs) > 0 {
		outs = &outArgs[0]
	}
	ok = C.g_function_info_invoke(info.FnInfo, ins, C.int(len(inArgs)), outs, C.int(len(outArgs)), &retArg, &gerr) == C.TRUE
	if !ok {
		panic(sp("%s", C.GoString((*C.char)(unsafe.Pointer(gerr.message)))))
	}

	// return value
	if info.ReturnType != C.GI_TYPE_TAG_VOID {
		ret = append(ret, fromGArg(info.ReturnType, &retArg))
	}

	// out args
	for i := 0; i < len(outArgs); i++ {
		ret = append(ret, fromGArg(info.OutTypes[i], &outArgs[i]))
	}

	// error
	if gerr != nil {
		err = fmt.Errorf("%s", C.GoString((*C.char)(unsafe.Pointer(gerr.message))))
		C.g_error_free(gerr)
	}

	return
}

func garg(v interface{}) (ret C.GIArgument) {
	switch v := v.(type) {
	case bool:
		if v {
			C.arg_set_boolean(&ret, C.TRUE)
		} else {
			C.arg_set_boolean(&ret, C.FALSE)
		}
	case int:
		C.arg_set_int(&ret, C.gint(v))
	case uint:
		C.arg_set_uint(&ret, C.guint(v))
	case int8:
		C.arg_set_int8(&ret, C.gint8(v))
	case uint8:
		C.arg_set_uint8(&ret, C.guint8(v))
	case int16:
		C.arg_set_int16(&ret, C.gint16(v))
	case uint16:
		C.arg_set_uint16(&ret, C.guint16(v))
	case int32:
		C.arg_set_int32(&ret, C.gint32(v))
	case uint32:
		C.arg_set_uint32(&ret, C.guint32(v))
	case int64:
		C.arg_set_int64(&ret, C.gint64(v))
	case uint64:
		C.arg_set_uint64(&ret, C.guint64(v))
	case float32:
		C.arg_set_float(&ret, C.gfloat(v))
	case float64:
		C.arg_set_double(&ret, C.gdouble(v))
	case string:
		C.arg_set_string(&ret, gs(v))
	case unsafe.Pointer:
		C.arg_set_pointer(&ret, v)
	case nil:
		C.arg_set_pointer(&ret, nil)
	default:
		panic(sp("not handled arg type %T", v))
	}
	return
}

func fromGArg(tag C.GITypeTag, arg *C.GIArgument) interface{} {
	switch tag {
	case C.GI_TYPE_TAG_VOID:
	case C.GI_TYPE_TAG_BOOLEAN:
		return C.arg_get_boolean(arg) == C.TRUE
	case C.GI_TYPE_TAG_INT8:
		return int8(C.arg_get_int8(arg))
	case C.GI_TYPE_TAG_UINT8:
		return uint8(C.arg_get_uint8(arg))
	case C.GI_TYPE_TAG_INT16:
		return int16(C.arg_get_int16(arg))
	case C.GI_TYPE_TAG_UINT16:
		return uint16(C.arg_get_uint16(arg))
	case C.GI_TYPE_TAG_INT32:
		return int32(C.arg_get_int32(arg))
	case C.GI_TYPE_TAG_UINT32:
		return uint32(C.arg_get_uint32(arg))
	case C.GI_TYPE_TAG_INT64:
		return int64(C.arg_get_int64(arg))
	case C.GI_TYPE_TAG_UINT64:
		return uint64(C.arg_get_uint64(arg))
	case C.GI_TYPE_TAG_FLOAT:
		return float32(C.arg_get_float(arg))
	case C.GI_TYPE_TAG_DOUBLE:
		return float64(C.arg_get_double(arg))
	case C.GI_TYPE_TAG_UTF8, C.GI_TYPE_TAG_FILENAME:
		return C.GoString(C.arg_get_string(arg))
	case C.GI_TYPE_TAG_INTERFACE, C.GI_TYPE_TAG_ARRAY:
		return C.arg_get_pointer(arg)
	default:
		panic(sp("not handled return type %v", tag))
	}
	return nil
}

var gcharCache = make(map[string]*C.gchar)

func gs(s string) *C.gchar {
	if gs, ok := gcharCache[s]; ok {
		return gs
	}
	gs := (*C.gchar)(unsafe.Pointer(C.CString(s)))
	gcharCache[s] = gs
	return gs
}
