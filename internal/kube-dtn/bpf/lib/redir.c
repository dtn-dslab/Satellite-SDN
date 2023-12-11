// SPDX-License-Identifier: GPL-2.0
/* Copyright (C) 2021 Intel Corporation */

#include "maps.h"
#include "sockops.h"
#include "vmlinux.h"
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

SEC("sk_msg")
int bpf_redir_proxy(struct sk_msg_md* msg) {
    uint32_t rc;
    uint32_t* debug_val_ptr;
    uint32_t debug_val;
    uint32_t debug_on_index = 0;
    uint32_t debug_pckts_index = 1;

    struct socket_4_tuple proxy_key = {};
    /* for inbound traffic */
    struct socket_4_tuple key = {};
    /* for outbound and envoy<->envoy traffic*/
    struct socket_4_tuple_extended* key_redir = NULL;
    sk_msg_extract4_keys(msg, &proxy_key, &key);

    if (key.local.ip4 == INBOUND_ENVOY_IP || key.remote.ip4 == INBOUND_ENVOY_IP) {
        rc = bpf_msg_redirect_hash(msg, &map_redir, &key, BPF_F_INGRESS);
    } else {
        key_redir = bpf_map_lookup_elem(&map_proxy, &proxy_key);
        if (key_redir == NULL || key_redir->flag == PROXY_DISABLED) {
            return SK_PASS;
        }

        if (key_redir->flag == PROXY_INIT) {
            // Enable proxy, but do not redirect this time
            key_redir->flag = PROXY_ENABLED;
            bpf_map_update_elem(&map_proxy, &proxy_key, key_redir, BPF_ANY);
            return SK_PASS;
        }

        // Proxy enabled, redirect
        rc = bpf_msg_redirect_hash(msg, &map_redir, &key_redir->tuple, BPF_F_INGRESS);
    }

    if (rc == SK_PASS) {
        debug_val_ptr = bpf_map_lookup_elem(&debug_map, &debug_on_index);
        if (debug_val_ptr && *debug_val_ptr == 1) {
            bpf_printk("data redirection succeed: [%x]->[%x]\n", proxy_key.local.ip4, proxy_key.remote.ip4);

            debug_val_ptr = bpf_map_lookup_elem(&debug_map, &debug_pckts_index);
            if (debug_val_ptr == NULL) {
                debug_val = 0;
                debug_val_ptr = &debug_val;
            }
            __sync_fetch_and_add(debug_val_ptr, 1);
            bpf_map_update_elem(&debug_map, &debug_pckts_index, debug_val_ptr, BPF_ANY);
        }
    }

    return SK_PASS;
}

char _license[] SEC("license") = "GPL";
