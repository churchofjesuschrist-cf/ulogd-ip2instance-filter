#include "_cgo_export.h"

#define CONN_NET_UNIX "unix"
#define CONN_ADDR_SOCK_PATH ""
#define INSTANCES_JSON_PATH ""

static struct config_keyset c = {
	.num_ces = 0,
	.ces = {
		{
			.key = "diego_conn_network",
			.type = CONFIG_TYPE_STRING,
			.options = CONFIG_OPT_NONE,
			.u = { .string = CONN_NET_UNIX },
		}, {
			.key = "diego_conn_address",
			.type = CONFIG_TYPE_STRING,
			.options = CONFIG_OPT_NONE,
			.u = { .string = CONN_ADDR_SOCK_PATH },
		},
	},
};

static struct ulogd_key o[] = {
	{
		.type = ULOGD_RET_STRING,
		.name = "cf.sinstance",
		.flags = ULOGD_RETF_FREE,
	},
	{
		.type = ULOGD_RET_STRING,
		.name = "cf.dinstance",
		.flags = ULOGD_RETF_FREE,
	},
};

static struct ulogd_plugin plugin = {
	.name = "",
	.input = {
		.keys = NULL,
		.num_keys = 0,
		.type = 0,
	},
	.output = {
		.keys = o,
		.num_keys = ARRAY_SIZE(o),
		.type = ULOGD_DTYPE_PACKET | ULOGD_DTYPE_FLOW,
	},
	.config_kset = &c,
	.priv_size = 0,
	.configure = &configurePlugin,
	.start = &startPlugin,
	.stop = &stopPlugin,
	.interp = &doFilter,
	.version = VERSION,
};

void __attribute__ ((constructor)) init(void);
void init(void) {
	memset(&plugin.name[0], 0, ULOGD_MAX_KEYLEN + 1);

	char *name = pluginRegisterName();
	strncpy(&plugin.name[0], name, ULOGD_MAX_KEYLEN);
	free(name);

	ulogd_register_plugin(&plugin);
}

char *configString(struct ulogd_pluginstance *pi, int k) {
	if (k >= pi->config_kset->num_ces) return NULL;
	return pi->config_kset->ces[k].u.string;
}

int configInt(struct ulogd_pluginstance *pi, int k) {
	if (k >= pi->config_kset->num_ces) return -1;
	return pi->config_kset->ces[k].u.value;
}

