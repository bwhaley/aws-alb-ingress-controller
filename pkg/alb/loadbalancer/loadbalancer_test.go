package loadbalancer

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/coreos/alb-ingress-controller/pkg/annotations"
	"github.com/coreos/alb-ingress-controller/pkg/util/log"
	"github.com/coreos/alb-ingress-controller/pkg/util/types"
)

const (
	clusterName = "cluster1"
	namespace   = "namespace1"
	ingressName = "ingress1"
	sg1         = "sg-123"
	sg2         = "sg-abc"
	tag1Key     = "tag1"
	tag1Value   = "value1"
	tag2Key     = "tag2"
	tag2Value   = "value2"
)

var (
	logr     *log.Logger
	lbScheme *string
	tags     types.Tags
)

func init() {
	logr = log.New("test")
	lbScheme = aws.String("internal")
	tags = types.Tags{
		{
			Key:   aws.String(tag1Key),
			Value: aws.String(tag1Value),
		},
		{
			Key:   aws.String(tag2Key),
			Value: aws.String(tag2Value),
		},
	}
}

func TestNewDesiredLoadBalancer(t *testing.T) {
	anno := &annotations.Annotations{
		Scheme:         lbScheme,
		SecurityGroups: types.AWSStringSlice{aws.String(sg1), aws.String(sg2)},
	}

	opts := &NewDesiredLoadBalancerOptions{
		ClusterName: clusterName,
		Namespace:   namespace,
		Logger:      logr,
		Annotations: anno,
		Tags:        tags,
		IngressName: ingressName,
	}

	expectedID := createLBName(namespace, ingressName, clusterName)
	lb := NewDesiredLoadBalancer(opts)

	key1, _ := lb.DesiredTags.Get(tag1Key)
	switch {
	case *lb.Desired.LoadBalancerName != expectedID:
		t.Errorf("LB ID was wrong. Expected: %s | Actual: %s", expectedID, lb.ID)
	case *lb.Desired.Scheme != *lbScheme:
		t.Errorf("LB scheme was wrong. Expected: %s | Actual: %s", *lbScheme, *lb.Desired.Scheme)
	case *lb.Desired.SecurityGroups[0] == sg2: // note sgs are sorted during checking for modification needs.
		t.Errorf("Secruity group was wrong. Expected: %s | Actual: %s", sg2, *lb.Desired.SecurityGroups[0])
	case key1 != tag1Value:
		t.Errorf("Tag was invalid. Expected: %s | Actual: %s", tag1Value, key1)

	}
}

func TestNewCurrentLoadBalancer(t *testing.T) {
	expectedName := createLBName(namespace, ingressName, clusterName)
	existing := &elbv2.LoadBalancer{
		LoadBalancerName: aws.String(expectedName),
	}
	tags := types.Tags{
		{
			Key:   aws.String("IngressName"),
			Value: aws.String(ingressName),
		},
		{
			Key:   aws.String("Namespace"),
			Value: aws.String(namespace),
		},
	}

	opts := &NewCurrentLoadBalancerOptions{
		LoadBalancer: existing,
		Logger:       logr,
		Tags:         tags,
		ClusterName:  clusterName,
	}

	lb, err := NewCurrentLoadBalancer(opts)
	if err != nil {
		t.Errorf("Failed to create LoadBalancer object from existing elbv2.LoadBalancer."+
			"Error: %s", err.Error())
	}

	switch {
	case *lb.Current.LoadBalancerName != expectedName:
		t.Errorf("Current LB created returned improper LoadBalancerName. Expected: %s | "+
			"Desired: %s", expectedName, *lb.Current.LoadBalancerName)
	}

}
