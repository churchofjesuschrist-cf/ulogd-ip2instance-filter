package plugin

import (
	"fmt"
	"github.com/lds-cf/ulogd-ip2instance-filter/resolver"
	"log"
	"net"
	"unsafe"
)

/*
#include <ulogd/ulogd.h>

void ulogd_register_plugin(struct ulogd_plugin *me) __attribute__((weak));
int config_parse_file(const char *section, struct config_keyset *kset) __attribute__((weak));

char *configString(struct ulogd_pluginstance *pi, int k);
int configInt(struct ulogd_pluginstance *pi, int k);

static inline int isValid(struct ulogd_key *keys, unsigned int kidx) {
	return pp_is_valid(keys, kidx);
}

static inline struct ulogd_pluginstance *input_plugin(struct ulogd_pluginstance_stack *ps) {
	return llist_entry(ps->list.next, struct ulogd_pluginstance, list);
}

static struct ulogd_key i_pkt[] = {
	{
		.type = ULOGD_RET_IPADDR,
		.name = "ip.saddr"
	},
	{
		.type = ULOGD_RET_IPADDR,
		.name = "ip.daddr"
	},
};

static inline void setup_pkt(struct ulogd_keyset *in_keyset) {
	in_keyset->keys = i_pkt;
	in_keyset->num_keys = ARRAY_SIZE(i_pkt);
	in_keyset->type = ULOGD_DTYPE_PACKET;
}

static struct ulogd_key i_flow[] = {
	// TODO: figure out which ikeys are actually useful; only should need two, but don't know which two.
	{
		.type = ULOGD_RET_IPADDR,
		.name = "orig.ip.saddr"
	},
	{
		.type = ULOGD_RET_IPADDR,
		.name = "orig.ip.daddr"
	},
	{
		.type = ULOGD_RET_IPADDR,
		.name = "reply.ip.saddr"
	},
	{
		.type = ULOGD_RET_IPADDR,
		.name = "reply.ip.daddr"
	},
};

static inline void setup_flow(struct ulogd_keyset *in_keyset) {
	in_keyset->keys = i_flow;
	in_keyset->num_keys = ARRAY_SIZE(i_flow);
	in_keyset->type = ULOGD_DTYPE_FLOW;
}

*/
import "C"

// int (*configure)(struct ulogd_pluginstance *instance,
//                  struct ulogd_pluginstance_stack *stack)
//export configurePlugin
func configurePlugin(pi *C.struct_ulogd_pluginstance, ps *C.struct_ulogd_pluginstance_stack) C.int {
	log.Printf(">>> configurePlugin ip2instance pi=%x", unsafe.Pointer(pi))
	defer log.Printf("configurePlugin ip2instance pi=%x <<<", unsafe.Pointer(pi))

	parseResult := C.config_parse_file(&(pi.id[0]), pi.config_kset)
	if 0 != parseResult {
		log.Printf("failed parsing pluginstance config")
		return -1
	}

	inPluginstance := C.input_plugin(ps)
	inPluginstanceOutputType := (*(*inPluginstance).plugin).output._type
	switch {
	case C.ULOGD_DTYPE_FLOW == inPluginstanceOutputType&C.ULOGD_DTYPE_FLOW:
		log.Printf("configuring ULOGD_DTYPE_FLOW ikeys")
		C.setup_flow(&(*pi).input)
	case C.ULOGD_DTYPE_RAW == inPluginstanceOutputType&C.ULOGD_DTYPE_RAW:
		log.Printf("configuring ULOGD_DTYPE_PACKET ikeys")
		C.setup_pkt(&(*pi).input)
	default:
		log.Printf("unrecognized input plugin in stack")
		return -1
	}

	return 0
}

// int (*start)(struct ulogd_pluginstance *pi)
//export startPlugin
func startPlugin(pi *C.struct_ulogd_pluginstance) C.int {
	log.Printf(">>> startPlugin ip2instance pi=%x", unsafe.Pointer(pi))
	defer log.Printf("startPlugin ip2instance pi=%x <<<", unsafe.Pointer(pi))

	return 0
}

var res resolver.Resolver = resolver.Get()

func translate(ikeyIdx, okeyIdx C.uint, pi *C.struct_ulogd_pluginstance) error {
	const ULOGD_RETF_FREE C.uint = 0x0002
	ikey, err := nthKey(ikeyIdx, &(*pi).input)

	if nil != err {
		return fmt.Errorf("expected ikeyIdx=%d", ikeyIdx)
	}

	ikeyName := C.GoString(&ikey.name[0])

	// it is not an error if the input key is not marked as valid:
	// in that case, the corresponding output key will also be not
	// marked as valid
	if 0 != C.isValid((*pi).input.keys, ikeyIdx) {
		if (*ikey)._type != 0x100 {
			return fmt.Errorf("input key `%v` could not be parsed as an IP address", ikeyName)
		}

		quad := C.ikey_get_u32(ikey)
		var d, c, b, a byte = byte((quad & C.u_int32_t(0xff000000)) >> 24),
			byte((quad & C.u_int32_t(0x00ff0000)) >> 16),
			byte((quad & C.u_int32_t(0x0000ff00)) >> 8),
			byte(quad & C.u_int32_t(0x000000ff))

		addr := net.IPv4(a, b, c, d)
		ai, err := res.Resolve(addr)
		if nil != err {
			return nil
		}

		okey, err := nthKey(okeyIdx, &(*pi).output)

		if nil != err {
			return fmt.Errorf("expected okeyIdx=%d", okeyIdx)
		}

		// okeyName := C.GoString(&okey.name[0])
		// log.Printf("resolved `%v` = %v", okeyName, ai.Guid)

		cstr := C.CString(fmt.Sprintf("%s/%d", ai.Guid, ai.InstanceIndex))
		C.okey_set_ptr(okey, unsafe.Pointer(cstr))
		(*okey).flags |= C.u_int16_t(ULOGD_RETF_FREE)
	} else {
		// log.Printf("address ikey %v not marked `valid`", ikeyName)
	}

	return nil
}

// int (*interp)(struct ulogd_pluginstance *instance)
//export doFilter
func doFilter(pi *C.struct_ulogd_pluginstance) C.int {
	// log.Printf(">>>doFilter ip2instance pi=%x", unsafe.Pointer(pi))
	// defer log.Printf("doFilter ip2instance pi=%x <<<", unsafe.Pointer(pi))

	err := translate(0, 0, pi)
	if err != nil {
		// log.Printf("computing `cf.sinstance`: %v", err)
		return -1
	}
	err = translate(1, 1, pi)
	if err != nil {
		// log.Printf("computing `cf.dinstance`: %v", err)
		return -1
	}

	return 0
}

// int (*stop)(struct ulogd_pluginstance *pi)
//export stopPlugin
func stopPlugin(pi *C.struct_ulogd_pluginstance) C.int {
	log.Printf(">>>stopPlugin ip2instance pi=%x", unsafe.Pointer(pi))
	defer log.Printf("stopPlugin ip2instance pi=%x <<<", unsafe.Pointer(pi))

	return 0
}

func nthKey(n C.uint, keyset *C.struct_ulogd_keyset) (*C.struct_ulogd_key, error) {
	if n >= (*keyset).num_keys {
		return nil, fmt.Errorf("index=%d out of range for num_keys=%d", n, (*keyset).num_keys)
	}

	sz := (C.uint)(C.sizeof_struct_ulogd_key)
	k0 := unsafe.Pointer((*keyset).keys)
	pk := (*C.struct_ulogd_key)(unsafe.Pointer(uintptr(k0) + uintptr(n*sz)))

	return pk, nil
}

//export pluginRegisterName
func pluginRegisterName() *C.char {
	return registrationName()
}
