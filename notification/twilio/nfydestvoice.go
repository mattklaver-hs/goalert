package twilio

import (
	"context"
	"regexp"

	"github.com/nyaruka/phonenumbers"
	"github.com/target/goalert/config"
	"github.com/target/goalert/notification/nfydest"
	"github.com/target/goalert/validation"
)

const (
	DestTypeTwilioVoice  = "builtin-twilio-voice"
	FallbackIconURLVoice = "builtin://phone-voice"

	// FieldExtension is an optional destination field for a DTMF extension
	// dialed automatically after the call connects.
	FieldExtension = "extension"
)

var reDTMFSequence = regexp.MustCompile(`^[0-9*#wW]+$`)

var _ nfydest.Provider = (*Voice)(nil)

func (v *Voice) ID() string { return DestTypeTwilioVoice }
func (v *Voice) TypeInfo(ctx context.Context) (*nfydest.TypeInfo, error) {
	cfg := config.FromContext(ctx)
	return &nfydest.TypeInfo{
		Type:                       DestTypeTwilioVoice,
		Name:                       "Voice Call",
		Enabled:                    cfg.Twilio.Enable,
		UserDisclaimer:             cfg.General.NotificationDisclaimer,
		SupportsAlertNotifications: true,
		SupportsUserVerification:   true,
		SupportsStatusUpdates:      true,
		UserVerificationRequired:   true,
		RequiredFields: []nfydest.FieldConfig{
			{
				FieldID:            FieldPhoneNumber,
				Label:              "Phone Number",
				Hint:               "Include country code e.g. +1 (USA), +91 (India), +44 (UK)",
				PlaceholderText:    "11235550123",
				Prefix:             "+",
				InputType:          "tel",
				SupportsValidation: true,
			},
			{
				FieldID:            FieldExtension,
				Label:              "Extension / DTMF Sequence (optional)",
				Hint:               "Sent after call connects. Use digits, *, #, w (0.5s pause), or W (1s pause). Leave blank if not needed.",
				PlaceholderText:    "ww1234",
				InputType:          "text",
				SupportsValidation: true,
			},
		},
	}, nil
}

func (v *Voice) ValidateField(ctx context.Context, fieldID, value string) error {
	switch fieldID {
	case FieldPhoneNumber:
		n, err := phonenumbers.Parse(value, "")
		if err != nil {
			return validation.WrapError(err)
		}
		if !phonenumbers.IsValidNumber(n) {
			return validation.NewGenericError("invalid phone number")
		}
		return nil
	case FieldExtension:
		if value == "" {
			return nil
		}
		if !reDTMFSequence.MatchString(value) {
			return validation.NewGenericError("extension must contain only digits, *, #, w (0.5s pause), or W (1s pause)")
		}
		return nil
	}

	return validation.NewGenericError("unknown field ID")
}

func (v *Voice) DisplayInfo(ctx context.Context, args map[string]string) (*nfydest.DisplayInfo, error) {
	if args == nil {
		args = make(map[string]string)
	}

	n, err := phonenumbers.Parse(args[FieldPhoneNumber], "")
	if err != nil {
		return nil, validation.WrapError(err)
	}

	text := phonenumbers.Format(n, phonenumbers.INTERNATIONAL)
	if ext := args[FieldExtension]; ext != "" {
		text += " ext. " + ext
	}

	return &nfydest.DisplayInfo{
		IconURL:     FallbackIconURLVoice,
		IconAltText: "Voice Call",
		Text:        text,
	}, nil
}
