package queue

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azqueue"
	"github.com/Azure/moby-packaging/pkg/archive"
)

const (
	defaultAccountName = "moby"
	defaultQueueName   = "moby-packaging-signing-and-publishing"
)

var (
	twoMinutesInSeconds int32 = 60 * 2
)

type Message struct {
	Artifact ArtifactInfo `json:"artifact"`
	Spec     archive.Spec `json:"spec"`
}

type ArtifactInfo struct {
	Name      string `json:"name"`
	URI       string `json:"uri"`
	Sha256Sum string `json:"sha256sum"`
}

type Messages struct {
	Messages []*azqueue.DequeuedMessage
}

type Client struct {
	c *azqueue.QueueClient
}

func (c *Client) GetAllMessages(ctx context.Context) (*Messages, error) {
	var (
		allMessages = []*azqueue.DequeuedMessage{}

		max           int32 = 32 // maximum number of messages for request
		failures      int   = 0
		totalFailures int   = 0
		errs          error
		allErrs       error

		dqOpts = azqueue.DequeueMessagesOptions{
			NumberOfMessages:  &max,
			VisibilityTimeout: &twoMinutesInSeconds,
		}
	)

	// Temporarily dequeue all the messages to ensure we don't enqueue a duplicate
	for m, err := c.c.DequeueMessages(ctx, &dqOpts); len(m.Messages) != 0; m, err = c.c.DequeueMessages(ctx, &dqOpts) {
		if err != nil {
			errs = errors.Join(errs, err)
			allErrs = errors.Join(allErrs, err)
			totalFailures++
			failures++

			if failures > 4 || totalFailures > 10 {
				fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]failed to examine messages: %s\n", errs)
				break
			}
			continue
		}

		allMessages = append(allMessages, m.Messages...)
		errs = nil
		failures = 0
	}

	return &Messages{Messages: allMessages}, allErrs
}

// used by trigger
func (m *Messages) ContainsBuild(spec archive.Spec) (bool, error) {
	failures := 0
	for _, rawMessage := range m.Messages {
		if failures > 4 {
			return false, fmt.Errorf("too many failures inspecting builds")
		}

		messageID := "unknown"
		if rawMessage.MessageID != nil {
			messageID = *rawMessage.MessageID
		}

		if rawMessage.MessageText == nil {
			failures++
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]nil message with ID: %s\n", messageID)
			continue
		}

		b, err := base64.StdEncoding.DecodeString(*rawMessage.MessageText)
		if err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]error decoding base64 string for message with ID: %s\n", messageID)
			continue
		}

		var m Message
		if err := json.Unmarshal(b, &m); err != nil {
			failures++
			fmt.Fprintf(os.Stderr, "##vso[task.logissue type=error;]error unmarshaling message with ID: %s\n", messageID)
			continue
		}

		if m.Spec == spec {
			return true, nil
		}
	}

	return false, nil
}

func NewDefaultSignQueueClient() (*Client, error) {
	return NewClient(defaultAccountName, defaultQueueName)
}

func NewClient(accountName, queueName string) (*Client, error) {
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	serviceURL := fmt.Sprintf("https://%s.queue.core.windows.net", accountName)
	sClient, err := azqueue.NewServiceClient(serviceURL, credential, nil)
	if err != nil {
		return nil, err
	}

	return &Client{c: sClient.NewQueueClient(queueName)}, nil
}
