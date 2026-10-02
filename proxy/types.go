// Modified by two-barrels in 2026 for ARI v6 modernization.
// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"errors"
	"fmt"
	"time"

	"github.com/two-barrels/ari/v6"
)

// AnnouncementInterval is the amount of time to wait between periodic service availability announcements
var AnnouncementInterval = time.Minute

// EntityCheckInterval is the interval between checks against Asterisk entity ID
var EntityCheckInterval = time.Second * 10

// Announcement describes the structure of an ARI proxy's announcement of availability on the network.  These are sent periodically and upon request (by a Ping).
type Announcement struct {
	// Node indicates the Asterisk ID to which the proxy is connected
	Node string `json:"node"`

	// Application indicates the ARI application as which the proxy is connected
	Application string `json:"application"`
}

// AnnouncementSubject returns the MessageBus subject
func AnnouncementSubject(prefix string) string {
	return fmt.Sprintf("%sannounce", prefix)
}

// PingSubject returns the MessageBus subject for a cluster-wide proxy ping for presence
func PingSubject(prefix string) string {
	return fmt.Sprintf("%sping", prefix)
}

// EntityData is a response which returns the data for a specific entity.
type EntityData struct {
	Application     *ari.ApplicationData     `json:"application,omitempty"`
	Asterisk        *ari.AsteriskInfo        `json:"asterisk,omitempty"`
	AsteriskPing    *ari.AsteriskPing        `json:"asterisk_ping,omitempty"`
	Bridge          *ari.BridgeData          `json:"bridge,omitempty"`
	Channel         *ari.ChannelData         `json:"channel,omitempty"`
	Config          *ari.ConfigData          `json:"config,omitempty"`
	DeviceState     *ari.DeviceStateData     `json:"device_state,omitempty"`
	Endpoint        *ari.EndpointData        `json:"endpoint,omitempty"`
	LiveRecording   *ari.LiveRecordingData   `json:"live_recording,omitempty"`
	Log             *ari.LogData             `json:"log,omitempty"`
	Mailbox         *ari.MailboxData         `json:"mailbox,omitempty"`
	Module          *ari.ModuleData          `json:"module,omitempty"`
	Playback        *ari.PlaybackData        `json:"playback,omitempty"`
	Sound           *ari.SoundData           `json:"sound,omitempty"`
	StoredRecording *ari.StoredRecordingData `json:"stored_recording,omitempty"`
	TextMessage     *ari.TextMessageData     `json:"text_message,omitempty"`
	RTPStats        *ari.RTPStats            `json:"rtp_stats,omitempty"`

	Variable  string         `json:"variable,omitempty"`
	Variables map[string]any `json:"variables,omitempty"`
}

// ErrNotFound indicates that the operation did not return a result
var ErrNotFound = errors.New("Not found")

// Response is a response to a request.  This acts as a base type for more complicated responses, as well.
type Response struct {
	// Error is the error encountered
	Error string `json:"error"`

	// StatusCode preserves an ARI HTTP error status across the message bus.
	StatusCode int `json:"status_code,omitempty"`

	// RecordingFile carries at most one bounded binary chunk or open metadata.
	RecordingFile *RecordingFileResponse `json:"recording_file,omitempty"`

	// Data is the returned entity data, if applicable
	Data *EntityData `json:"data,omitempty"`

	// Key is the key of the returned entity, if applicable
	Key *ari.Key `json:"key,omitempty"`

	// Keys is the list of keys of any matching entities, if applicable
	Keys []*ari.Key `json:"keys,omitempty"`
}

// StatusError is an ARI HTTP error returned through the proxy.
type StatusError struct {
	Message string
	Status  int
}

func (e *StatusError) Error() string { return e.Message }
func (e *StatusError) Code() int     { return e.Status }

// Err returns an error from the Response.  If the response's Error is empty, a nil error is returned.  Otherwise, the error will be filled with the value of response.Error.
func (e *Response) Err() error {
	if e == nil {
		return nil
	}
	if e.Error != "" {
		if e.StatusCode != 0 {
			return &StatusError{Message: e.Error, Status: e.StatusCode}
		}
		return errors.New(e.Error)
	}
	return nil
}

// IsNotFound indicates that the retuned error response was a Not Found error response
func (e *Response) IsNotFound() bool {
	return e.Error == "Not found"
}

// NewErrorResponse wraps an error as an ErrorResponse
func NewErrorResponse(err error) *Response {
	if err == nil {
		return &Response{}
	}
	response := &Response{Error: err.Error()}
	var coded interface{ Code() int }
	if errors.As(err, &coded) {
		response.StatusCode = coded.Code()
	}
	return response
}

// Request describes a request which is sent from an ARI proxy Client to an ARI proxy Server
type Request struct {
	// Kind indicates the type of request
	Kind string `json:"kind"`

	// Key is the key or key filter on which this request should be processed
	Key *ari.Key `json:"key"`

	ApplicationSubscribe   *ApplicationSubscribe       `json:"application_subscribe,omitempty"`
	ApplicationEventFilter *ari.ApplicationEventFilter `json:"application_event_filter,omitempty"`

	AsteriskConfig         *AsteriskConfig          `json:"asterisk_config,omitempty"`
	AsteriskLoggingChannel *AsteriskLoggingChannel  `json:"asterisk_logging_channel,omitempty"`
	AsteriskVariableSet    *AsteriskVariableSet     `json:"asterisk_variable_set,omitempty"`
	AsteriskInfoOptions    *ari.AsteriskInfoOptions `json:"asterisk_info_options,omitempty"`

	BridgeAddChannel    *BridgeAddChannel    `json:"bridge_add_channel,omitempty"`
	BridgeVariable      *BridgeVariable      `json:"bridge_variable,omitempty"`
	BridgeVariables     *BridgeVariables     `json:"bridge_variables,omitempty"`
	BridgeCreate        *BridgeCreate        `json:"bridge_create,omitempty"`
	BridgeMOH           *BridgeMOH           `json:"bridge_moh,omitempty"`
	BridgePlay          *BridgePlay          `json:"bridge_play,omitempty"`
	BridgeRecord        *BridgeRecord        `json:"bridge_record,omitempty"`
	BridgeRemoveChannel *BridgeRemoveChannel `json:"bridge_remove_channel,omitempty"`
	BridgeVideoSource   *BridgeVideoSource   `json:"bridge_video_source,omitempty"`

	ChannelCreate           *ChannelCreate           `json:"channel_create,omitempty"`
	ChannelContinue         *ChannelContinue         `json:"channel_continue,omitempty"`
	ChannelMove             *ChannelMove             `json:"channel_move,omitempty"`
	ChannelDial             *ChannelDial             `json:"channel_dial,omitempty"`
	ChannelHangup           *ChannelHangup           `json:"channel_hangup,omitempty"`
	ChannelMOH              *ChannelMOH              `json:"channel_moh,omitempty"`
	ChannelMute             *ChannelMute             `json:"channel_mute,omitempty"`
	ChannelOriginate        *ChannelOriginate        `json:"channel_originate,omitempty"`
	ChannelPlay             *ChannelPlay             `json:"channel_play,omitempty"`
	ChannelRecord           *ChannelRecord           `json:"channel_record,omitempty"`
	ChannelSendDTMF         *ChannelSendDTMF         `json:"channel_send_dtmf,omitempty"`
	ChannelSnoop            *ChannelSnoop            `json:"channel_snoop,omitempty"`
	ChannelExternalMedia    *ChannelExternalMedia    `json:"channel_external_media,omitempty"`
	ChannelVariable         *ChannelVariable         `json:"channel_variable,omitempty"`
	ChannelVariables        *ChannelVariables        `json:"channel_variables,omitempty"`
	ChannelRedirect         *ChannelRedirect         `json:"channel_redirect,omitempty"`
	ChannelTransferProgress *ChannelTransferProgress `json:"channel_transfer_progress,omitempty"`
	ChannelUserevent        *ChannelUserevent        `json:"channel_user_event,omitempty"`

	DeviceStateUpdate *DeviceStateUpdate `json:"device_state_update,omitempty"`

	EndpointListByTech *EndpointListByTech       `json:"endpoint_list_by_tech,omitempty"`
	EndpointRefer      *ari.EndpointReferOptions `json:"endpoint_refer,omitempty"`
	EventClaim         *EventClaim               `json:"event_claim,omitempty"`

	MailboxUpdate *MailboxUpdate `json:"mailbox_update,omitempty"`

	PlaybackControl *PlaybackControl `json:"playback_control,omitempty"`

	RecordingStoredCopy *RecordingStoredCopy  `json:"recording_stored_copy,omitempty"`
	RecordingFile       *RecordingFileRequest `json:"recording_file,omitempty"`

	SoundList       *SoundList       `json:"sound_list,omitempty"`
	TextMessageSend *TextMessageSend `json:"text_message_send,omitempty"`
}

// BridgeVariable carries a single bridge variable name, value, and event option.
type BridgeVariable struct {
	Name         string `json:"name"`
	Value        string `json:"value"`
	ReportEvents *bool  `json:"report_events,omitempty"`
}

// BridgeVariables carries repeated names for a read or assignments for a write.
type BridgeVariables struct {
	Names  []string                                `json:"names,omitempty"`
	Values map[string]ari.BridgeVariableAssignment `json:"values,omitempty"`
}

// ChannelVariables carries repeated names for a read or assignments for a write.
type ChannelVariables struct {
	Names  []string                          `json:"names,omitempty"`
	Values map[string]ari.VariableAssignment `json:"values,omitempty"`
}

type ChannelRedirect struct {
	Endpoint string `json:"endpoint"`
}

type ChannelTransferProgress struct {
	States string `json:"states"`
}

// ApplicationSubscribe describes a request to subscribe/unsubscribe a particular ARI application to an EventSource
type ApplicationSubscribe struct {
	// EventSource is the ARI event source to which the subscription is requested.  This should be one of:
	//  - channel:<channelId>
	//  - bridge:<bridgeId>
	//  - endpoint:<tech>/<resource> (e.g. SIP/102)
	//  - deviceState:<deviceName>
	EventSource string `json:"event_source"`
}

// AsteriskVariableSet is the request type for setting an asterisk variable
type AsteriskVariableSet struct {
	// Value is the value to set
	Value string `json:"value"`
}

// BridgeAddChannel is the request type for adding a channel to a bridge
type BridgeAddChannel struct {
	// Channel is the channel ID to add to the bridge
	Channel string `json:"channel"`

	// AbsorbDTMF indicates that DTMF coming from this channel will not be passed through to the bridge
	AbsorbDTMF bool `json:"absorbDTMF,omitempty"`

	// Mute indicates that the channel should be muted, preventing audio from it passing through to the bridge
	Mute bool `json:"mute,omitempty"`

	// Role indicates the channel's role in the bridge
	Role                        string `json:"role,omitempty"`
	InhibitConnectedLineUpdates *bool  `json:"inhibit_connected_line_updates,omitempty"`
}

// BridgeCreate is the request type for creating a bridge
type BridgeCreate struct {
	// Type is the comma-separated list of bridge type attributes (mixing,
	// holding, dtmf_events, proxy_media).  If not set, the default (mixing)
	// will be used.
	Type     string `json:"type"`
	BridgeID string `json:"bridge_id,omitempty"`

	// Name is the name to assign to the bridge (optional)
	Name      string                              `json:"name,omitempty"`
	Variables map[string]ari.BridgeCreateVariable `json:"variables,omitempty"`
}

// BridgeMOH is the request type for playing Music on Hold to a bridge
type BridgeMOH struct {
	// Class is the Music On Hold class to be played
	Class string `json:"class"`
}

// BridgePlay is the request type for playing audio on the bridge
type BridgePlay struct {
	// PlaybackID is the unique identifier for this playback
	PlaybackID string `json:"playback_id"`

	// MediaURI is the URI from which to obtain the playback media
	MediaURI string `json:"media_uri"`
	// MediaURIs carries multiple media items without changing the legacy field.
	MediaURIs []string               `json:"media_uris,omitempty"`
	Options   *ari.BridgePlayOptions `json:"options,omitempty"`
}

func (p *BridgePlay) URIs() []string {
	if len(p.MediaURIs) > 0 {
		return p.MediaURIs
	}
	return []string{p.MediaURI}
}

// BridgeRecord is the request for recording a bridge
type BridgeRecord struct {
	// Name is the name for the recording
	Name string `json:"name"`

	// Options is the list of recording Options
	Options *ari.RecordingOptions `json:"options,omitempty"`
}

// BridgeRemoveChannel is the request for removing a channel on the bridge
type BridgeRemoveChannel struct {
	// Channel is the name of the channel to remove
	Channel string `json:"channel"`
}

// BridgeVideoSource describes the details of a request to set the video source of a bridge explicitly
type BridgeVideoSource struct {
	// Channel is the name of the channel to use as the explicit video source
	Channel string `json:"channel"`
}

// ChannelCreate describes a request to create a new channel
type ChannelCreate struct {
	// ChannelCreateRequest is the request for creating the channel
	ChannelCreateRequest ari.ChannelCreateRequest `json:"channel_create_request"`
}

// ChannelContinue describes a request to continue an ARI application
type ChannelContinue struct {
	// Context is the context into which the channel should be continued
	Context string `json:"context"`

	// Extension is the extension into which the channel should be continued
	Extension string `json:"extension"`

	// Priority is the priority at which the channel should be continued
	Priority int                         `json:"priority"`
	Options  *ari.ChannelContinueOptions `json:"options,omitempty"`
}

// ChannelMove describes a move to another Stasis application. Proxy-owned
// fields use snake_case; the native client maps app_args to ARI's appArgs.
type ChannelMove struct {
	App     string `json:"app"`
	AppArgs string `json:"app_args,omitempty"`
}

// ChannelDial describes a request to dial
type ChannelDial struct {
	// Caller is the channel ID of the "caller" channel; if specified, the media parameters of the dialing channel will be matched to the "caller" channel.
	Caller string `json:"caller"`

	// Timeout is the maximum time which should be allowed for the dial to complete
	Timeout time.Duration `json:"timeout"`
}

// ChannelHangup is the request for hanging up a channel
type ChannelHangup struct {
	// Reason is the reason the channel is being hung up
	Reason     string `json:"reason"`
	ReasonCode string `json:"reason_code,omitempty"`
}

// ChannelMOH is the request playing hold on music on a channel
type ChannelMOH struct {
	// Music is the music to play
	Music string `json:"music"`
}

// ChannelMute is the request for muting or unmuting a channel
type ChannelMute struct {
	// Direction is the direction to mute
	Direction ari.Direction `json:"direction,omitempty"`
}

// ChannelOriginate is the request for creating a channel
type ChannelOriginate struct {
	// OriginateRequest contains the information for originating a channel
	OriginateRequest ari.OriginateRequest `json:"originate_request"`
}

// ChannelPlay is the request for playing audio on a channel
type ChannelPlay struct {
	// PlaybackID is the unique identifier for this playback
	PlaybackID string `json:"playback_id"`

	// MediaURI is the URI from which to obtain the playback media
	MediaURI  string                  `json:"media_uri"`
	MediaURIs []string                `json:"media_uris,omitempty"`
	Options   *ari.ChannelPlayOptions `json:"options,omitempty"`
}

func (p *ChannelPlay) URIs() []string {
	if len(p.MediaURIs) > 0 {
		return p.MediaURIs
	}
	return []string{p.MediaURI}
}

// ChannelRecord is the request for recording a channel
type ChannelRecord struct {
	// Name is the name for the recording
	Name string `json:"name"`

	// Options is the list of recording Options
	Options *ari.RecordingOptions `json:"options,omitempty"`
}

// ChannelSendDTMF is the request for sending a DTMF event to a channel
type ChannelSendDTMF struct {
	// DTMF is the series of DTMF inputs to send
	DTMF string `json:"dtmf"`

	// Options are the DTMF options
	Options *ari.DTMFOptions `json:"options,omitempty"`
}

// ChannelSnoop is the request for snooping on a channel
type ChannelSnoop struct {
	// SnoopID is the ID to use for the snoop channel which will be created.
	SnoopID string `json:"snoop_id"`

	// Options describe the parameters for the snoop session
	Options *ari.SnoopOptions `json:"options,omitempty"`
}

// ChannelExternalMedia describes the request for an external media channel
type ChannelExternalMedia struct {
	Options ari.ExternalMediaOptions `json:"options"`
}

// ChannelVariable is the request type to read or modify a channel variable
type ChannelVariable struct {
	// Name is the name of the channel variable
	Name string `json:"name"`

	// Value is the value to set to the channel variable
	Value        string `json:"value,omitempty"`
	ReportEvents *bool  `json:"report_events,omitempty"`
}

// DeviceStateUpdate describes the request for updating the device state
type DeviceStateUpdate struct {
	// State is the new state of the device to set
	State string `json:"state"`
}

// EndpointListByTech describes the request for listing endpoints by technology
type EndpointListByTech struct {
	// Tech is the technology for the endpoint
	Tech string `json:"tech"`
}

// EventClaim identifies the broadcast channel to claim.
type EventClaim struct {
	ChannelID string `json:"channel_id"`
}

// TextMessageSend carries a text message and its technology variables.
type TextMessageSend struct {
	From      string            `json:"from"`
	To        string            `json:"to,omitempty"`
	Body      string            `json:"body"`
	Variables map[string]string `json:"variables"`
}

// RecordingFileRequest identifies an open recording stream and requested chunk.
type RecordingFileRequest struct {
	Session  string `json:"session,omitempty"`
	Sequence uint64 `json:"sequence,omitempty"`
}

// RecordingFileResponse carries stream metadata or one bounded file chunk.
type RecordingFileResponse struct {
	Session     string `json:"session"`
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size,omitempty"`
	Sequence    uint64 `json:"sequence,omitempty"`
	Data        []byte `json:"data,omitempty"`
	EOF         bool   `json:"eof,omitempty"`
}

// MailboxUpdate describes the request for updating a mailbox
type MailboxUpdate struct {
	// New is the number of New (unread) messages in the mailbox
	New int `json:"new"`

	// Old is the number of Old (read) messages in the mailbox
	Old int `json:"old"`
}

// PlaybackControl describes the request for performing a playback command
type PlaybackControl struct {
	// Command is the playback control command to run
	Command string `json:"command"`
}

// RecordingStoredCopy describes the request for copying a stored recording
type RecordingStoredCopy struct {
	// Destination is the destination location to copy to
	Destination string `json:"destination"`
}

// SoundList describes the request for listing the sounds
type SoundList struct {
	// Filters are the filters to apply when listing the sounds
	Filters map[string]string `json:"filters"`
}

// AsteriskConfig describes the request relating to asterisk configuration
type AsteriskConfig struct {
	// Tuples is the list of configuration tuples to update
	Tuples []ari.ConfigTuple `json:"tuples,omitempty"`
}

// AsteriskLoggingChannel describes a request relating to an asterisk logging channel
type AsteriskLoggingChannel struct {
	// Levels is the set of logging levels for this logging channel (comma-separated string)
	Levels string `json:"config"`
}

// ChannelUserevent - "User-generated event with additional user-defined fields in the object."
type ChannelUserevent struct {
	UserEvent ari.ChannelUserevent
}
