package cwlogs

import (
	"errors"
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/versent/syslog-cloudlogs/pkg/batching"
	"github.com/versent/syslog-cloudlogs/pkg/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var sequenceMatcher = regexp.MustCompile(`The given sequenceToken is invalid. The next expected sequenceToken is: (.+)`)

// Dispatcher dispatches logs to cloudwatch
type Dispatcher struct {
	config        *config.SyslogConfig
	session       *session.Session
	sequenceToken string
	lock          *sync.Mutex // just to be safe with sequenceToken
	svc           *cloudwatchlogs.CloudWatchLogs
}

// NewDispatcher create a new dispatcher
func NewDispatcher(config *config.SyslogConfig) (*Dispatcher, error) {

	var options session.Options

	if config.Region != "" {
		options = session.Options{
			Config: aws.Config{
				Region: aws.String(config.Region),
			},
		}
	}

	if config.Profile != "" {
		options.Profile = config.Profile
	}

	sess := session.Must(session.NewSessionWithOptions(options))

	return &Dispatcher{
		config:  config,
		session: sess,
		lock:    &sync.Mutex{},
		svc:     cloudwatchlogs.New(sess),
	}, nil
}

// SetupCloudwatch create cloudwatch group and stream
func (d *Dispatcher) SetupCloudwatch() error {

	_, err := d.svc.CreateLogGroup(&cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(d.config.Group),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != "ResourceAlreadyExistsException" {
				logrus.WithError(err).Fatalf("cloudwatch log group creation failed")
			}
			logrus.WithError(err).Warn("cloudwatch log group already exists")
		}
	}

	_, err = d.svc.CreateLogStream(&cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(d.config.Group),
		LogStreamName: aws.String(d.config.Stream),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() != "ResourceAlreadyExistsException" {
				logrus.WithError(err).Fatalf("cloudwatch log stream creation failed")
			}
			logrus.WithError(err).Warn("cloudwatch log stream already exists")
		}
	}

	return nil
}

// Dispatch handle entries and send them to cloudwatch
func (d *Dispatcher) Dispatch(entries []*batching.LogEntry) {

	logrus.Info("dispatch")

	events := make([]*cloudwatchlogs.InputLogEvent, len(entries))

	for n, entry := range entries {
		events[n] = &cloudwatchlogs.InputLogEvent{
			Message:   aws.String(entry.Message),
			Timestamp: aws.Int64(entry.MilliTimestamp),
		}
	}

	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  aws.String(d.config.Group),
		LogStreamName: aws.String(d.config.Stream),
	}

	// first request has no SequenceToken - in all subsequent request we set it
	if d.sequenceToken != "" {
		params.SequenceToken = aws.String(d.sequenceToken)
	}

	resp, err := d.putLogEvents(params)
	if err != nil {
		logrus.Fatalln(err)
	}

	d.lock.Lock()
	d.sequenceToken = *resp.NextSequenceToken
	d.lock.Unlock()

	logrus.WithField("sequenceToken", d.sequenceToken).Info("cwlogs sequence update")
}

func (d *Dispatcher) putLogEvents(input *cloudwatchlogs.PutLogEventsInput) (*cloudwatchlogs.PutLogEventsOutput, error) {
	var (
		seq  string
		err  error
		resp *cloudwatchlogs.PutLogEventsOutput
	)

	resp, err = d.svc.PutLogEvents(input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "InvalidSequenceTokenException" {

				logrus.WithError(err).Warn("retry sending due to sequence resume")

				// this section just pulls the sequence out of the message
				// and then set it in the request
				seq, err = extractSeq(awsErr.Message())
				if err != nil {
					logrus.Fatalln(err)
				}

				input.SequenceToken = aws.String(seq)

				resp, err = d.svc.PutLogEvents(input)
				if err != nil {
					logrus.Fatalln(err)
				}

			}
		}
	}

	return resp, err
}

func extractSeq(msg string) (string, error) {
	res := sequenceMatcher.FindStringSubmatch(msg)

	if len(res) != 2 {
		return "", errors.New("Missing sequence in message")
	}

	return res[1], nil
}
