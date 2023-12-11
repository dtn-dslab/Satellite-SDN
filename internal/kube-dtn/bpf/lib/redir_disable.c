#include "maps.h"
#include "sockops.h"
#include "vmlinux.h"
#include <arpa/inet.h>
#include <bpf/bpf_endian.h>
#include <bpf/bpf_helpers.h>
#include <iproute2/bpf_elf.h>
#include <linux/if_ether.h>
#include <linux/ip.h>
#include <linux/pkt_cls.h>
#include <linux/tcp.h>

SEC("egress")
int bpf_redir_disable(struct sk_buff_md* skb) {
    const int l3_off = ETH_HLEN;                       // IP header offset
    const int l4_off = l3_off + sizeof(struct iphdr);  // TCP header offset
    const int l7_off = l4_off + sizeof(struct tcphdr); // L7 (e.g. HTTP) header offset

    void* data = (void*)(long)skb->data;
    void* data_end = (void*)(long)skb->data_end;

    if (data_end < data + l7_off)
        return TC_ACT_OK; // Not our packet, handover to kernel

    struct ethhdr* eth = data;
    if (eth->h_proto != htons(ETH_P_IP))
        return TC_ACT_OK; // Not an IPv4 packet, handover to kernel

    struct iphdr* ip = (struct iphdr*)(data + l3_off);
    if (ip->protocol != IPPROTO_TCP)
        return TC_ACT_OK;

    struct tcphdr* tcp = (struct tcphdr*)(data + l4_off);

    uint32_t local_port = bpf_ntohl(tcp->source) >> 16, remote_port = bpf_ntohl(tcp->dest) >> 16;
    struct socket_4_tuple proxy_key = {};
    proxy_key.local.ip4 = ip->saddr;
    proxy_key.local.port = local_port;
    proxy_key.remote.ip4 = ip->daddr;
    proxy_key.remote.port = remote_port;

    struct socket_4_tuple_extended* key_redir = NULL;
    key_redir = bpf_map_lookup_elem(&map_proxy, &proxy_key);
    if (key_redir != NULL && key_redir->flag != PROXY_DISABLED) {
        key_redir->flag = PROXY_DISABLED;
        bpf_map_update_elem(&map_proxy, &proxy_key, key_redir, BPF_ANY);
    }
    return TC_ACT_OK;
}

char _license[] SEC("license") = "GPL";