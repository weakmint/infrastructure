package main

import (
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/dns"
	"github.com/pulumi/pulumi-gcp/sdk/v7/go/gcp/projects"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type DNSRecord struct {
	Name  pulumi.String
	Value pulumi.StringArray
	Type  pulumi.String
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clouddnsAPI, err := projects.NewService(ctx, "clouddns", &projects.ServiceArgs{
			DisableDependentServices: pulumi.Bool(true),
			Service:                  pulumi.String("dns.googleapis.com"),
		})
		if err != nil {
			return err
		}

		zone, err := dns.NewManagedZone(ctx, "weakmint-dot-dev", &dns.ManagedZoneArgs{
			Description: pulumi.String("DNS zone for weakmint.dev"),
			DnsName:     pulumi.String("weakmint.dev."),
		}, pulumi.DependsOn([]pulumi.Resource{clouddnsAPI}))
		if err != nil {
			return err
		}

		records := []DNSRecord{
			{Name: "protonmail._domainkey.", Value: pulumi.StringArray{pulumi.String("protonmail.domainkey.duhlkrlkno2jq3jkwrixvxob2kobtttwo4wzyudh5652lrkhpmuwa.domains.proton.ch.")}, Type: "CNAME"},
			{Name: "protonmail2._domainkey.", Value: pulumi.StringArray{pulumi.String("protonmail2.domainkey.duhlkrlkno2jq3jkwrixvxob2kobtttwo4wzyudh5652lrkhpmuwa.domains.proton.ch.")}, Type: "CNAME"},
			{Name: "protonmail3._domainkey.", Value: pulumi.StringArray{pulumi.String("protonmail3.domainkey.duhlkrlkno2jq3jkwrixvxob2kobtttwo4wzyudh5652lrkhpmuwa.domains.proton.ch.")}, Type: "CNAME"},
			{Name: "@", Value: pulumi.StringArray{pulumi.String("protonmail-verification=0e3a27fa62fdd2ca6cd0d25deb580adccfe979cc"), pulumi.String("\"v=spf1 include:_spf.protonmail.ch ~all\"")}, Type: "TXT"},
			{Name: "@", Value: pulumi.StringArray{pulumi.String("10 mail.protonmail.ch."), pulumi.String("20 mailsec.protonmail.ch.")}, Type: "MX"},
			{Name: "_dmarc.", Value: pulumi.StringArray{pulumi.String("\"v=DMARC1; p=none\"")}, Type: "TXT"},
		}

		for i, x := range records {
			name := x.Name
			var fqdn pulumi.StringInput = zone.DnsName.ApplyT(func(dnsName string) (string, error) {
				return fmt.Sprintf("%v%v", name, dnsName), nil
			}).(pulumi.StringOutput)
			if name == "@" {
				fqdn = zone.DnsName
			}
			ttl := pulumi.Int(300)
			if x.Type == "MX" {
				ttl = pulumi.Int(3600)
			}
			_, err = dns.NewRecordSet(ctx, fmt.Sprintf("weakmint-dot-dev-record-set-%v", i), &dns.RecordSetArgs{
				Name:        fqdn,
				ManagedZone: zone.Name,
				Type:        x.Type,
				Ttl:         ttl,
				Rrdatas:     x.Value,
			})
			if err != nil {
				return err
			}
		}

		ctx.Export("nameservers", zone.NameServers)
		return nil
	})
}
