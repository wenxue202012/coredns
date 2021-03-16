package httpresolve

import (
	"context"
	"errors"
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/metrics"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
	"net"
	"strings"
)

type HttpResolve struct {
	Next      plugin.Handler
	Ttl       uint32
	SecretId  string
	SecretKey string
	Endpoint  string
	//Fall  fall.F
}

type Record struct {
	Address  string `json:"address"`
	HostName string `json:"dns_name,omitempty"`
}

type RecordsList struct {
	Records []Record `json:"results"`
}

var log = clog.NewWithPlugin("httpresolve")

func (h *HttpResolve) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	answers := []dns.RR{}
	state := request.Request{W: w, Req: r}

	// testtcr.tencentcloudcr.com
	qName := strings.TrimLeft(state.QName(), ".")
	instanceName := qName[0:strings.Index(qName, ".")]
	ipAddress, err := DescribeInstanceAll(h.Endpoint, h.SecretId, h.SecretKey, []string{instanceName})

	if err != nil || len(ipAddress) == 0 {
		ip, err := net.ResolveIPAddr("ip", state.QName())
		if err != nil {
			return plugin.NextOrFailure(h.Name(), h.Next, ctx, w, r)
		}
		ipAddress = ip.String()
	}

	rec := new(dns.A)
	rec.Hdr = dns.RR_Header{Name: state.QName(), Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: h.Ttl}
	rec.A = net.ParseIP(ipAddress)
	answers = append(answers, rec)
	m := new(dns.Msg)
	m.Answer = answers
	m.SetReply(r)
	w.WriteMsg(m)

	requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()

	return dns.RcodeSuccess, nil
}

func (h *HttpResolve) Name() string {
	return "httpresolve"
}

func DescribeInstanceAll(endpoint string, secretId string, secretKey string, registerNames []string) (string, error) {
	credential := common.NewCredential(
		secretId,
		secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = endpoint

	client, _ := tcr.NewClient(credential, regions.Chengdu, cpf)
	req := tcr.NewDescribeInstancesRequest()
	allRegion := true
	req.AllRegion = &allRegion
	req.WithApiInfo("tcr", "2019-09-24", "DescribeInstanceAll")
	req.Filters = []*tcr.Filter{
		&tcr.Filter{
			Name:   common.StringPtr("RegistryName"),
			Values: common.StringPtrs(registerNames),
		},
	}

	response := tcr.DescribeInstancesResponse{}
	err := client.Send(req, &response)
	if err != nil {
		log.Error("Failed to call tcr to resolve internal address . instanceName ={}, err={} ", registerNames, err)
		return "", err

	}
	if len(response.Response.Registries) <= 0 {
		log.Error("Call tcr to resolve the internal address is empty . instanceName = ", registerNames)
		return "", errors.New("Internal address is empty.")
	}
	ipAddressArray := make([]string, 0)
	for _, value := range response.Response.Registries {
		if len(*value.InternalEndpoint) > 0 {
			ipAddressArray = append(ipAddressArray, *value.InternalEndpoint)
		}
	}
	return strings.Join(ipAddressArray, ","), nil
}
