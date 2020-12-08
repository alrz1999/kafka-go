package kafka

import (
	"context"
	"net"
	"time"

	"github.com/segmentio/kafka-go/protocol/electleaders"
)

// ElectLeadersRequest is a request to the ElectLeaders API.
type ElectLeadersRequest struct {
	// Addr is the address of the kafka broker to send the request to.
	Addr net.Addr

	// Topic is the name of the topic to do the leader elections in.
	Topic string

	// Partitions is the list of partitions to run leader elections for.
	Partitions []int

	// Timeout is the amount of time to wait for the election to run.
	Timeout time.Duration
}

// ElectLeadersResponse is a response from the ElectLeaders API.
type ElectLeadersResponse struct {
	// ErrorCode is set to a non-zero value if a top-level error occurred.
	ErrorCode int

	// PartitionResults contains the results for each partition leader election.
	PartitionResults []ElectLeadersResponsePartitionResult
}

// ElectLeadersResponsePartitionResult contains the response details for a single partition.
type ElectLeadersResponsePartitionResult struct {
	// Partition is the ID of the partition.
	Partition int

	// ErrorCode is set to a non-zero value if an error happened for this partition's election.
	ErrorCode int

	// ErrorMessage describes the partition-specific election error that occurred.
	ErrorMessage string
}

func (c *Client) ElectLeaders(
	ctx context.Context,
	req *ElectLeadersRequest,
) (*ElectLeadersResponse, error) {
	partitions32 := []int32{}
	for _, partition := range req.Partitions {
		partitions32 = append(partitions32, int32(partition))
	}

	protoResp, err := c.roundTrip(
		ctx,
		req.Addr,
		&electleaders.Request{
			TopicPartitions: []electleaders.RequestTopicPartitions{
				{
					Topic:        req.Topic,
					PartitionIDs: partitions32,
				},
			},
			TimeoutMs: int32(req.Timeout.Milliseconds()),
		},
	)
	if err != nil {
		return nil, err
	}
	apiResp := protoResp.(*electleaders.Response)

	resp := &ElectLeadersResponse{
		ErrorCode: int(apiResp.ErrorCode),
	}

	for _, topicResult := range apiResp.ReplicaElectionResults {
		for _, partitionResult := range topicResult.PartitionResults {
			resp.PartitionResults = append(
				resp.PartitionResults,
				ElectLeadersResponsePartitionResult{
					Partition:    int(partitionResult.PartitionID),
					ErrorCode:    int(partitionResult.ErrorCode),
					ErrorMessage: partitionResult.ErrorMessage,
				},
			)
		}
	}

	return resp, nil
}
