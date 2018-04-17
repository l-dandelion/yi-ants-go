package module

import (
	"fmt"
	"net"
	"testing"

	"github.com/l-dandelion/yi-ants-go/lib/constant"
)

func TestRegNew(t *testing.T) {
	registrar := NewRegistrar()
	if registrar == nil {
		t.Fatal("Couldn't create registrar!")
	}
}

func TestRegRegister(t *testing.T) {
	mt := TYPE_DOWNLOADER
	ml := legalTypeLetterMap[mt]
	sn := DefaultSNGen.Get()
	addr, _ := NewAddr("http", "127.0.0.1", 8080)
	mid := MID(fmt.Sprintf(midTemplate, ml, sn, addr))

	// test:register the illegal module
	registrar := NewRegistrar()
	err := registrar.Register(nil)
	if err == nil {
		t.Fatal("No error when register module instance with nil module!")
	}
	// test:register the invaild module(type and module not match)
	var m Module
	for t, f := range fakeModuleFuncMap {
		if t != mt {
			m = f(mid)
			break
		}
	}
	err = registrar.Register(m)
	if err == nil {
		t.Fatalf("No error when register unmatched module instance! (type: %T)",
			m)
	}
	var midsAll []MID
	for _, mt := range legalTypes {
		var midsByType []MID
		for mip := range legalIPMap {
			ml = legalTypeLetterMap[mt]
			sn = DefaultSNGen.Get()
			addr, _ = NewAddr("http", mip, 8080)
			mid = MID(fmt.Sprintf(midTemplate, ml, sn, addr))
			midsByType = append(midsByType, mid)
			midsAll = append(midsAll, mid)
			m = fakeModuleFuncMap[mt](mid)
			err = registrar.Register(m)
			if err != nil {
				t.Fatalf("An error occurs when registering module instance: %s (MID: %s)",
					err, mid)
			}

			// register the same MID
			err = registrar.Register(m)
			if err != nil {
				t.Fatalf("An error occurs when registering module instance: %s (MID: %s)",
					err, mid)
			}

			// register module with illegal type letter
			sn = DefaultSNGen.Get()
			mid = MID(fmt.Sprintf(midTemplate, "M", sn, addr))
			m = fakeModuleFuncMap[mt](mid)
			err = registrar.Register(m)
			if err == nil {
				t.Fatalf("No error when register module instance with illegal MID %q!",
					mid)
			}
		}
		modules, err := registrar.GetAllByType(mt)
		if err != nil {
			t.Fatalf("An error occurs when getting all module instances: %s (type: %s)",
				err, mt)
		}
		for _, mid := range midsByType {
			if _, ok := modules[mid]; !ok {
				t.Fatalf("Not found the module instance! (MID: %s, type: %s)",
					mid, mt)
			}
		}
	}
	modules := registrar.GetAll()
	for _, mid := range midsAll {
		if _, ok := modules[mid]; !ok {
			t.Fatalf("Not found the module instance! (MID: %s)",
				mid)
		}
	}
	for _, mt := range illegalTypes {
		sn := DefaultSNGen.Get()
		addr, _ := NewAddr("http", "127.0.0.1", 8080)
		ml := legalTypeLetterMap[mt]
		mid := MID(fmt.Sprintf(midTemplate, ml, sn, addr))
		m := NewFakeDownloader(mid, nil)
		err := registrar.Register(m)
		if err == nil {
			t.Fatalf("No error when register module instance with illegal type %q!",
				mt)
		}
	}
}

func TestModuleUnregister(t *testing.T) {
	registrar := NewRegistrar()
	var mids []MID
	for _, mt := range legalTypes {
		for mip := range legalIPMap {
			sn := DefaultSNGen.Get()
			addr, _ := NewAddr("http", mip, 8080)
			mid, err := GenMID(mt, sn, addr)
			if err != nil {
				t.Fatalf("An error occurs when generating module ID: %s (type: %s, sn: %d, addr: %s)",
					err, mt, sn, addr)
			}
			m := fakeModuleFuncMap[mt](mid)
			err = registrar.Register(m)
			if err != nil {
				t.Fatalf("An error occurs when registering module instance: %s (type: %s, sn: %d, addr: %s)",
					err, mt, sn, addr)
			}
			mids = append(mids, mid)
		}
	}
	for _, mid := range mids {
		err := registrar.Unregister(mid)
		if err != nil {
			t.Fatalf("An error occurs when unregistering  module instance: %s (mid: %s)",
				err, mid)
		}
	}
	// unregister unregistered module
	for _, mid := range mids {
		err := registrar.Unregister(mid)
		if err.ErrNo != constant.ERR_MODULE_NOT_FOUND {
			t.Fatalf("An error occurs when unregistering  module instance: %s (mid: %s)",
				err, mid)
		}
	}
	for _, illegalMID := range illegalMIDs {
		err := registrar.Unregister(illegalMID)
		if err == nil {
			t.Fatalf("No error when unregister module instance with illegal MID %q!",
				illegalMID)
		}
	}
}

func TestModuleGet(t *testing.T) {
	registrar := NewRegistrar()
	mt := illegalTypes[0]
	m1, err := registrar.Get(mt)
	if err == nil {
		t.Fatalf("No error when get module instance with illegal type %q!",
			mt)
	}
	if m1 != nil {
		t.Fatalf("It still can get module instance with illegal type %q!",
			mt)
	}
	mt = TYPE_DOWNLOADER
	m1, err = registrar.Get(mt)
	if err == nil {
		t.Fatal("No error when get nonexistent module instance!")
	}
	if m1 != nil {
		t.Fatalf("It still can get nonexistent module instance!")
	}
	addr, _ := NewAddr("http", "127.0.0.1", 8080)
	mid := MID(fmt.Sprintf(
		midTemplate,
		legalTypeLetterMap[mt],
		DefaultSNGen.Get(),
		addr))
	m := defaultFakeModuleMap[mt]
	err = registrar.Register(m)
	if err != nil {
		t.Fatalf("An error occurs when registering module instance: %s (mid: %s)",
			err, mid)
	}
	m1, err = registrar.Get(mt)
	if err != nil {
		t.Fatalf("An error occurs when getting module instance: %s (mid: %s)",
			err, mid)
	}
	if m1 == nil {
		t.Fatalf("Couldn't get module instance with MID %q!", mid)
	}
	if m1.ID() != m.ID() {
		t.Fatalf("Inconsistent MID: expected: %s, actual: %s",
			m.ID(), m1.ID())
	}
}

func TestModuleAllInParallel(t *testing.T) {
	baseSize := 1000
	basePort := 8000
	legalTypesLen := len(legalTypes)
	sLen := baseSize * legalTypesLen
	types := make([]int8, sLen)
	sns := make([]uint64, sLen)
	addrs := make([]net.Addr, sLen)
	for i := 0; i < sLen; i++ {
		types[i] = legalTypes[i%legalTypesLen]
		port := uint64(basePort + basePort%legalTypesLen)
		addrs[i], _ = NewAddr("http", "127.0.0.1", port)
		sns[i] = DefaultSNGen.Get()
	}
	registrar := NewRegistrar()
	t.Run("All in parallel", func(t *testing.T) {
		t.Run("Register", func(t *testing.T) {
			t.Parallel()
			for i, addr := range addrs {
				mt := types[i]
				sn := DefaultSNGen.Get()
				mid, err := GenMID(mt, sn, addr)
				if err != nil {
					t.Fatalf("An error occurs when generating module ID: %s (type: %s, sn: %d, addr: %s)",
						err, mt, sn, addr)
				}
				m := fakeModuleFuncMap[mt](mid)
				err = registrar.Register(m)
				if err != nil {
					t.Fatalf("An error occurs when registering module instance: %s (type: %s, sn: %d, addr: %s)",
						err, mt, sn, addr)
				}
			}
		})
		t.Run("Unregister", func(t *testing.T) {
			t.Parallel()
			for i, addr := range addrs {
				mt := types[i]
				sn := sns[i]
				mid, _ := GenMID(mt, sn, addr)
				err := registrar.Unregister(mid)
				if err != nil && err.ErrNo != constant.ERR_MODULE_NOT_FOUND {
					t.Fatalf("An error occurs when unregistering  module instance: %s (mid: %s)",
						err, mid)
				}
			}
		})
		t.Run("Get", func(t *testing.T) {
			t.Parallel()
			for _, mt := range types {
				m, err := registrar.Get(mt)
				if err != nil && err.ErrNo != constant.ERR_MODULE_NOT_FOUND {
					t.Fatalf("An error occurs when getting module instance: %s (mid: %s)",
						err, m.ID())
				}
			}
		})
	})
}
