package gcall

import (
	"testing"
	"time"
	"unsafe"
)

func TestAll(t *testing.T) {
	Require("Gtk", "3.0")

	pt("major version %v\n", Call("Gtk.get_major_version"))
	Call("Gtk.init", 0, nil)
	win := Call("Gtk.Window.new", 0)[0].(unsafe.Pointer)
	Call("Gtk.Widget.show_all", win)

	go func() {
		<-time.After(time.Second * 2)
		Call("Gtk.main_quit")
	}()
	Call("Gtk.main")
}
