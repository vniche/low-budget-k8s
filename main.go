package main

import (
	computev1 "github.com/pulumi/pulumi-google-native/sdk/go/google/compute/v1"
	containerv1 "github.com/pulumi/pulumi-google-native/sdk/go/google/container/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	shortName := "lbk8s"
	pulumi.Run(func(ctx *pulumi.Context) error {
		// google-native configs
		googleConfig := config.New(ctx, "google-native")
		region := googleConfig.Require("region")
		zone := googleConfig.Require("zone")

		// low-budget-k8s configs
		conf := config.New(ctx, "low-budget-k8s")
		clusterName := conf.Require("cluster-name")
		nodesCIDR := conf.Require("nodes-cidr")
		podsCIDR := conf.Require("pods-cidr")
		servicesCIDR := conf.Require("services-cidr")
		maxPodsPerNode := conf.Require("max-pods-per-node")

		// create a VPC network
		network, err := computev1.NewNetwork(ctx, shortName, &computev1.NetworkArgs{
			AutoCreateSubnetworks: pulumi.Bool(false),
		})
		if err != nil {
			return err
		}

		// create the subnetwork for the VPC network
		subnetName := shortName + "-" + region
		subNetwork, err := computev1.NewSubnetwork(ctx, subnetName, &computev1.SubnetworkArgs{
			Network:               network.SelfLink,
			PrivateIpGoogleAccess: pulumi.Bool(true),
			Region:                pulumi.String(region),
			IpCidrRange:           pulumi.StringPtr(nodesCIDR),
			SecondaryIpRanges: computev1.SubnetworkSecondaryRangeArray{
				computev1.SubnetworkSecondaryRangeArgs{
					IpCidrRange: pulumi.StringPtr(podsCIDR),
					RangeName:   pulumi.String(subnetName + "-pods"),
				},
				computev1.SubnetworkSecondaryRangeArgs{
					IpCidrRange: pulumi.StringPtr(servicesCIDR),
					RangeName:   pulumi.String(subnetName + "-services"),
				},
			},
		})
		if err != nil {
			return err
		}

		// Create a Kubernetes cluster, associated with the VPC network
		// cluster depends on the above created subnetwork, configured here with https://www.pulumi.com/docs/intro/concepts/resources/options/dependson/
		cluster, err := containerv1.NewCluster(ctx, clusterName, &containerv1.ClusterArgs{
			NetworkConfig: &containerv1.NetworkConfigArgs{
				DatapathProvider: containerv1.NetworkConfigDatapathProviderAdvancedDatapath,
			},
			Subnetwork: subNetwork.Name.ToStringOutput(),
			Network:    network.SelfLink,
			IpAllocationPolicy: &containerv1.IPAllocationPolicyArgs{
				ClusterSecondaryRangeName:  pulumi.String(subnetName + "-pods"),
				ServicesSecondaryRangeName: pulumi.String(subnetName + "-services"),
				UseIpAliases:               pulumi.Bool(true),
			},
			Location: pulumi.StringPtr(zone),
			ReleaseChannel: &containerv1.ReleaseChannelArgs{
				Channel: containerv1.ReleaseChannelChannelStable,
			}, NodePools: containerv1.NodePoolTypeArray{
				&containerv1.NodePoolTypeArgs{
					Name:             pulumi.String("default"),
					InitialNodeCount: pulumi.Int(1),
					Autoscaling: &containerv1.NodePoolAutoscalingArgs{
						Enabled:      pulumi.Bool(true),
						MinNodeCount: pulumi.Int(1),
						MaxNodeCount: pulumi.Int(5),
					},
					Config: &containerv1.NodeConfigArgs{
						DiskSizeGb:  pulumi.Int(10),
						MachineType: pulumi.StringPtr("n1-standard-2"),
						Preemptible: pulumi.Bool(true),
					},
				},
			},
			DefaultMaxPodsConstraint: &containerv1.MaxPodsConstraintArgs{
				MaxPodsPerNode: pulumi.StringPtr(maxPodsPerNode),
			},
		}, pulumi.DependsOn([]pulumi.Resource{subNetwork}))
		if err != nil {
			return err
		}

		// Export the cluster self-link
		ctx.Export("cluster", cluster.SelfLink)

		return nil
	})
}
