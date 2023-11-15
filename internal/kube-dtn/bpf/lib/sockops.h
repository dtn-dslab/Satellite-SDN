// SPDX-License-Identifier: GPL-2.0
/* Copyright (C) 2021 Intel Corporation */

#pragma once

#ifndef AF_INET
#define AF_INET 2
#endif

#ifndef NULL
#define NULL ((void*)0)
#endif

#define INBOUND_ENVOY_IP 0x600007f
#define SOCKOPS_MAP_SIZE 65535

#include "vmlinux.h"
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>

#define PROXY_INIT 1
#define PROXY_DISABLED 0
#define PROXY_ENABLED 2

#define TC_ACT_OK 0

struct addr_2_tuple {
    uint32_t ip4;
    uint32_t port;
};

struct socket_4_tuple {
    struct addr_2_tuple local;
    struct addr_2_tuple remote;
};

struct socket_4_tuple_extended {
    struct socket_4_tuple tuple;
    uint32_t flag;
};

static __inline__ void sk_ops_extract4_key(struct bpf_sock_ops* ops, struct socket_4_tuple* key) {
    key->local.ip4 = ops->local_ip4;
    key->local.port = ops->local_port;
    key->remote.ip4 = ops->remote_ip4;
    key->remote.port = bpf_ntohl(ops->remote_port);
}

static __inline__ void sk_msg_extract4_keys(struct sk_msg_md* msg, struct socket_4_tuple* proxy_key,
                                            struct socket_4_tuple* key) {
    proxy_key->local.ip4 = msg->local_ip4;
    proxy_key->local.port = msg->local_port;
    proxy_key->remote.ip4 = msg->remote_ip4;
    proxy_key->remote.port = bpf_ntohl(msg->remote_port);
    key->local.ip4 = msg->remote_ip4;
    key->local.port = bpf_ntohl(msg->remote_port);
    key->remote.ip4 = msg->local_ip4;
    key->remote.port = msg->local_port;
}

static __inline__ void sk_buff_extract4_keys(struct sk_buff_md* skb, struct socket_4_tuple* proxy_key,
                                             struct socket_4_tuple* key) {
    proxy_key->local.ip4 = skb->local_ip4;
    proxy_key->local.port = skb->local_port;
    proxy_key->remote.ip4 = skb->remote_ip4;
    proxy_key->remote.port = bpf_ntohl(skb->remote_port);
    key->local.ip4 = skb->remote_ip4;
    key->local.port = bpf_ntohl(skb->remote_port);
    key->remote.ip4 = skb->local_ip4;
    key->remote.port = skb->local_port;
}
