A ulogd plugin (builds a Linux .so) that knows how to obtain the app instance id, given an IPv4 address.
Uses (input, string) keys: `ip.saddr.str` and `ip.daddr.str`.
Produces (output, string) keys: `cf.sinstance.id` and `cf.dinstance.id`.
