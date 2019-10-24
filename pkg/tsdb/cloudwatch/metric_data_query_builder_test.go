package cloudwatch

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMetricDataQueryBuilder(t *testing.T) {
	Convey("TestMetricDataQueryBuilder", t, func() {
		const (
			maxNoOfSearchExpressions = 2
			maxNoOfMetricDataQueries = 10
		)
		mdib := &metricDataInputBuilder{maxNoOfSearchExpressions, maxNoOfMetricDataQueries}

		Convey("buildMetricDataQueries", func() {
			Convey("and one GetMetricDataInput is generated for each query statistic", func() {
				dimensions := make(map[string][]string)
				dimensions["InstanceId"] = []string{"i-12345678"}
				query := &cloudWatchQuery{
					RefId:      "A",
					Region:     "us-east-1",
					Namespace:  "AWS/EC2",
					MetricName: "CPUUtilization",
					Dimensions: dimensions,
					Statistics: []*string{aws.String("Average"), aws.String("Sum")},
					Period:     300,
					Id:         "id1",
					Identifier: "id1",
					Expression: "",
					MatchExact: true,
				}

				res, err := mdib.buildMetricDataQueries(query)
				So(err, ShouldBeNil)
				So(len(res), ShouldEqual, 2)
				So(*res[0].Id, ShouldEqual, "id1_____0")
				So(*res[1].Id, ShouldEqual, "id1_____1")
			})
		})

		Convey("buildSearchExpression", func() {
			Convey("and query should be matched exact", func() {
				matchExact := true
				Convey("and query has three dimension values for a given dimension key", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"lb1", "lb2", "lb3"}
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: dimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('{AWS/EC2,LoadBalancer} MetricName="CPUUtilization" "LoadBalancer"=("lb1" OR "lb2" OR "lb3")', 'Average', 300))`)
				})

				Convey("and query has three dimension values for two given dimension keys", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"lb1", "lb2", "lb3"}
					dimensions["InstanceId"] = []string{"i-123", "i-456", "i-789"}
					sortedDimensions := sortDimensions(dimensions)
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: sortedDimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('{AWS/EC2,InstanceId,LoadBalancer} MetricName="CPUUtilization" "InstanceId"=("i-123" OR "i-456" OR "i-789") "LoadBalancer"=("lb1" OR "lb2" OR "lb3")', 'Average', 300))`)
				})

				Convey("and no OR operator was added if a star was used for dimension value", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"*"}
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: dimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldNotContainSubstring, "OR")
				})

				Convey("and query has one dimension key with a * value", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"*"}
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: dimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('{AWS/EC2,LoadBalancer} MetricName="CPUUtilization"', 'Average', 300))`)
				})

				Convey("and query has three dimension values for two given dimension keys, and one value is a star", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"lb1", "lb2", "lb3"}
					dimensions["InstanceId"] = []string{"i-123", "*", "i-789"}
					sortedDimensions := sortDimensions(dimensions)
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: sortedDimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('{AWS/EC2,InstanceId,LoadBalancer} MetricName="CPUUtilization" "LoadBalancer"=("lb1" OR "lb2" OR "lb3")', 'Average', 300))`)
				})
			})

			Convey("and query should not be matched exact", func() {
				matchExact := false
				Convey("and query has three dimension values for a given dimension key", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"lb1", "lb2", "lb3"}
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: dimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('Namespace="AWS/EC2" MetricName="CPUUtilization" "LoadBalancer"=("lb1" OR "lb2" OR "lb3")', 'Average', 300))`)
				})

				Convey("and query has three dimension values for two given dimension keys", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"lb1", "lb2", "lb3"}
					dimensions["InstanceId"] = []string{"i-123", "i-456", "i-789"}
					sortedDimensions := sortDimensions(dimensions)
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: sortedDimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('Namespace="AWS/EC2" MetricName="CPUUtilization" "InstanceId"=("i-123" OR "i-456" OR "i-789") "LoadBalancer"=("lb1" OR "lb2" OR "lb3")', 'Average', 300))`)
				})

				Convey("and query has one dimension key with a * value", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"*"}
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: dimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('Namespace="AWS/EC2" MetricName="CPUUtilization" "LoadBalancer"', 'Average', 300))`)
				})

				Convey("and query has three dimension values for two given dimension keys, and one value is a star", func() {
					dimensions := make(map[string][]string)
					dimensions["LoadBalancer"] = []string{"lb1", "lb2", "lb3"}
					dimensions["InstanceId"] = []string{"i-123", "*", "i-789"}
					sortedDimensions := sortDimensions(dimensions)
					query := &cloudWatchQuery{
						Namespace:  "AWS/EC2",
						MetricName: "CPUUtilization",
						Dimensions: sortedDimensions,
						Period:     300,
						Identifier: "id1",
						Expression: "",
						MatchExact: matchExact,
					}

					res := buildSearchExpression(query, "Average")
					So(res, ShouldEqual, `REMOVE_EMPTY(SEARCH('Namespace="AWS/EC2" MetricName="CPUUtilization" "LoadBalancer"=("lb1" OR "lb2" OR "lb3") "InstanceId"', 'Average', 300))`)
				})
			})
		})
	})
}