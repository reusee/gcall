package gcall

/*
#include <glib-object.h>
#include <stdlib.h>

static inline GType gvalue_get_type(GValue *v) {
	return G_VALUE_TYPE(v);
}

static inline GType gtype_get_fundamental(GType t) {
	return G_TYPE_FUNDAMENTAL(t);
}

extern void closureMarshal(GClosure*, GValue*, guint, GValue*, gpointer, gpointer);

GClosure* new_closure(gint64 *callbackKey) {
	GClosure *closure = g_closure_new_simple(sizeof(GClosure), NULL);
	g_closure_set_meta_marshal(closure, callbackKey, (GClosureMarshal)(closureMarshal));
	return closure;
}

gint64* newCallbackKeyPointer(gint64 n) {
	gint64 *p = (gint64*)malloc(sizeof(gint64));
	*p = n;
	return p;
}

*/
import "C"
import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

var (
	callbacks     = make(map[C.gint64]interface{})
	callbacksLock = new(sync.RWMutex)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Connect(obj unsafe.Pointer, signal string, cb interface{}) uint64 {
	key := C.newCallbackKeyPointer(C.gint64(rand.Int63()))
	callbacksLock.Lock()
	callbacks[*key] = cb
	callbacksLock.Unlock()
	closure := C.new_closure(key)
	id := C.g_signal_connect_closure(C.gpointer(obj), gs(signal), closure, C.FALSE)
	return uint64(id)
}

func fromGValue(v *C.GValue) (ret interface{}) {
	valueType := C.gvalue_get_type(v)
	fundamentalType := C.gtype_get_fundamental(valueType)
	switch fundamentalType {
	case C.G_TYPE_OBJECT:
		ret = unsafe.Pointer(C.g_value_get_object(v))
	case C.G_TYPE_STRING:
		ret = fromGStr(C.g_value_get_string(v))
	case C.G_TYPE_UINT:
		ret = int(C.g_value_get_uint(v))
	case C.G_TYPE_BOXED:
		ret = unsafe.Pointer(C.g_value_get_boxed(v))
	case C.G_TYPE_BOOLEAN:
		ret = C.g_value_get_boolean(v) == C.gboolean(1)
	default:
		fmt.Printf("from type %s %T\n", fromGStr(C.g_type_name(fundamentalType)), v)
		panic("FIXME") //TODO
	}
	return
}

func fromGStr(s *C.gchar) string {
	return C.GoString((*C.char)(unsafe.Pointer(s)))
}
