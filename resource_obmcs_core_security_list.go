// Copyright (c) 2017, Oracle and/or its affiliates. All rights reserved.

package main

import (
	"github.com/MustWin/baremetal-sdk-go"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/oracle/terraform-provider-oci/crud"
)

var transportSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"min": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	},
}

var icmpSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"code": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	},
}

func SecurityListResource() *schema.Resource {
	return &schema.Resource{
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: crud.DefaultTimeout,
		Create:   createSecurityList,
		Read:     readSecurityList,
		Update:   updateSecurityList,
		Delete:   deleteSecurityList,
		Schema: map[string]*schema.Schema{
			"compartment_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"display_name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"egress_security_rules": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination": {
							Type:     schema.TypeString,
							Required: true,
						},
						"icmp_options": icmpSchema,
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"tcp_options": transportSchema,
						"udp_options": transportSchema,
						"stateless": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ingress_security_rules": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"icmp_options": icmpSchema,
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"source": {
							Type:     schema.TypeString,
							Required: true,
						},
						"tcp_options": transportSchema,
						"udp_options": transportSchema,
						"stateless": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"time_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vcn_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func createSecurityList(d *schema.ResourceData, m interface{}) (e error) {
	client := m.(*baremetal.Client)
	crd := &SecurityListResourceCrud{}
	crd.D = d
	crd.Client = client
	return crud.CreateResource(d, crd)
}

func readSecurityList(d *schema.ResourceData, m interface{}) (e error) {
	client := m.(*baremetal.Client)
	crd := &SecurityListResourceCrud{}
	crd.D = d
	crd.Client = client
	return crud.ReadResource(crd)
}

func updateSecurityList(d *schema.ResourceData, m interface{}) (e error) {
	client := m.(*baremetal.Client)
	crd := &SecurityListResourceCrud{}
	crd.D = d
	crd.Client = client
	return crud.UpdateResource(d, crd)
}

func deleteSecurityList(d *schema.ResourceData, m interface{}) (e error) {
	client := m.(*baremetal.Client)
	crd := &SecurityListResourceCrud{}
	crd.D = d
	crd.Client = client
	return crud.DeleteResource(d, crd)
}

type SecurityListResourceCrud struct {
	crud.BaseCrud
	Res *baremetal.SecurityList
}

func (s *SecurityListResourceCrud) ID() string {
	return s.Res.ID
}

func (s *SecurityListResourceCrud) CreatedPending() []string {
	return []string{baremetal.ResourceProvisioning}
}

func (s *SecurityListResourceCrud) CreatedTarget() []string {
	return []string{baremetal.ResourceAvailable}
}

func (s *SecurityListResourceCrud) DeletedPending() []string {
	return []string{baremetal.ResourceTerminating}
}

func (s *SecurityListResourceCrud) DeletedTarget() []string {
	return []string{baremetal.ResourceTerminated}
}

func (s *SecurityListResourceCrud) State() string {
	return s.Res.State
}

func (s *SecurityListResourceCrud) Create() (e error) {
	compartmentID := s.D.Get("compartment_id").(string)
	egress := s.buildEgressRules()
	ingress := s.buildIngressRules()
	vcnID := s.D.Get("vcn_id").(string)

	opts := &baremetal.CreateOptions{}
	opts.DisplayName = s.D.Get("display_name").(string)

	s.Res, e = s.Client.CreateSecurityList(compartmentID, vcnID, egress, ingress, opts)

	return
}

func (s *SecurityListResourceCrud) Get() (e error) {
	res, e := s.Client.GetSecurityList(s.D.Id())
	if e == nil {
		s.Res = res
	}
	return
}

func (s *SecurityListResourceCrud) Update() (e error) {
	opts := &baremetal.UpdateSecurityListOptions{}

	if displayName, ok := s.D.GetOk("display_name"); ok {
		opts.DisplayName = displayName.(string)
	}

	if egress := s.buildEgressRules(); egress != nil {
		opts.EgressRules = egress
	}
	if ingress := s.buildIngressRules(); ingress != nil {
		opts.IngressRules = ingress
	}

	s.Res, e = s.Client.UpdateSecurityList(s.D.Id(), opts)
	return
}

func (s *SecurityListResourceCrud) SetData() {
	s.D.Set("compartment_id", s.Res.CompartmentID)
	s.D.Set("display_name", s.Res.DisplayName)

	confEgressRules := []map[string]interface{}{}
	for _, egressRule := range s.Res.EgressSecurityRules {
		confEgressRule := map[string]interface{}{}
		confEgressRule["destination"] = egressRule.Destination
		confEgressRule = buildConfRule(
			confEgressRule,
			egressRule.Protocol,
			egressRule.ICMPOptions,
			egressRule.TCPOptions,
			egressRule.UDPOptions,
			&egressRule.IsStateless,
		)
		confEgressRules = append(confEgressRules, confEgressRule)
	}
	s.D.Set("egress_security_rules", confEgressRules)

	confIngressRules := []map[string]interface{}{}
	for _, ingressRule := range s.Res.IngressSecurityRules {
		confIngressRule := map[string]interface{}{}
		confIngressRule["source"] = ingressRule.Source
		confIngressRule = buildConfRule(
			confIngressRule,
			ingressRule.Protocol,
			ingressRule.ICMPOptions,
			ingressRule.TCPOptions,
			ingressRule.UDPOptions,
			&ingressRule.IsStateless,
		)
		confIngressRules = append(confIngressRules, confIngressRule)
	}
	s.D.Set("ingress_security_rules", confIngressRules)

	s.D.Set("state", s.Res.State)
	s.D.Set("time_created", s.Res.TimeCreated.String())
	s.D.Set("vcn_id", s.Res.VcnID)
}

func (s *SecurityListResourceCrud) Delete() (e error) {
	return s.Client.DeleteSecurityList(s.D.Id(), nil)
}

func (s *SecurityListResourceCrud) buildEgressRules() (sdkRules []baremetal.EgressSecurityRule) {
	sdkRules = []baremetal.EgressSecurityRule{}
	for _, val := range s.D.Get("egress_security_rules").([]interface{}) {
		confRule := val.(map[string]interface{})

		sdkRule := baremetal.EgressSecurityRule{
			Destination: confRule["destination"].(string),
			ICMPOptions: s.buildICMPOptions(confRule),
			Protocol:    confRule["protocol"].(string),
			TCPOptions:  s.buildTCPOptions(confRule),
			UDPOptions:  s.buildUDPOptions(confRule),
			IsStateless: confRule["stateless"].(bool),
		}

		sdkRules = append(sdkRules, sdkRule)
	}
	return
}

func (s *SecurityListResourceCrud) buildIngressRules() (sdkRules []baremetal.IngressSecurityRule) {
	sdkRules = []baremetal.IngressSecurityRule{}
	for _, val := range s.D.Get("ingress_security_rules").([]interface{}) {
		confRule := val.(map[string]interface{})

		sdkRule := baremetal.IngressSecurityRule{
			ICMPOptions: s.buildICMPOptions(confRule),
			Protocol:    confRule["protocol"].(string),
			Source:      confRule["source"].(string),
			TCPOptions:  s.buildTCPOptions(confRule),
			UDPOptions:  s.buildUDPOptions(confRule),
			IsStateless: confRule["stateless"].(bool),
		}

		sdkRules = append(sdkRules, sdkRule)
	}
	return
}

func (s *SecurityListResourceCrud) buildICMPOptions(conf map[string]interface{}) (opts *baremetal.ICMPOptions) {
	l := conf["icmp_options"].([]interface{})
	if len(l) > 0 {
		confOpts := l[0].(map[string]interface{})
		opts = &baremetal.ICMPOptions{
			Code: uint64(confOpts["code"].(int)),
			Type: uint64(confOpts["type"].(int)),
		}
	}
	return
}

func (s *SecurityListResourceCrud) buildTCPOptions(conf map[string]interface{}) (opts *baremetal.TCPOptions) {
	l := conf["tcp_options"].([]interface{})
	if len(l) > 0 {
		confOpts := l[0].(map[string]interface{})
		opts = &baremetal.TCPOptions{
			baremetal.PortRange{
				Max: uint64(confOpts["max"].(int)),
				Min: uint64(confOpts["min"].(int)),
			},
		}
	}
	return
}

func (s *SecurityListResourceCrud) buildUDPOptions(conf map[string]interface{}) (opts *baremetal.UDPOptions) {
	l := conf["udp_options"].([]interface{})
	if len(l) > 0 {
		confOpts := l[0].(map[string]interface{})
		opts = &baremetal.UDPOptions{
			baremetal.PortRange{
				Max: uint64(confOpts["max"].(int)),
				Min: uint64(confOpts["min"].(int)),
			},
		}
	}
	return
}

func buildConfICMPOptions(opts *baremetal.ICMPOptions) (list []interface{}) {
	confOpts := map[string]interface{}{
		"code": int(opts.Code),
		"type": int(opts.Type),
	}
	return []interface{}{confOpts}
}

func buildConfTransportOptions(portRange baremetal.PortRange) (list []interface{}) {
	confOpts := map[string]interface{}{
		"max": int(portRange.Max),
		"min": int(portRange.Min),
	}
	return []interface{}{confOpts}
}

func buildConfRule(
	confRule map[string]interface{},
	protocol string,
	icmpOpts *baremetal.ICMPOptions,
	tcpOpts *baremetal.TCPOptions,
	udpOpts *baremetal.UDPOptions,
	stateless *bool,
) map[string]interface{} {
	confRule["protocol"] = protocol
	if icmpOpts != nil {
		confRule["icmp_options"] = buildConfICMPOptions(icmpOpts)
	}
	if tcpOpts != nil {
		confRule["tcp_options"] = buildConfTransportOptions(tcpOpts.DestinationPortRange)
	}
	if udpOpts != nil {
		confRule["udp_options"] = buildConfTransportOptions(udpOpts.DestinationPortRange)
	}
	if stateless != nil {
		confRule["stateless"] = *stateless
	}
	return confRule
}
