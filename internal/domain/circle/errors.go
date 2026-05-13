package circle

import "errors"

var (
	ErrCircleNotFound   = errors.New("circle not found")
	ErrNotOrganizer     = errors.New("only the organizer can perform this action")
	ErrCircleNotActive  = errors.New("circle is not active")
	ErrCircleFull       = errors.New("circle is full")
	ErrAlreadyMember    = errors.New("already a member of this circle")
	ErrNotMember        = errors.New("not a member of this circle")
	ErrMoiScoreTooLow   = errors.New("MoiScore is too low for this circle")
	ErrInvalidInvite    = errors.New("invalid or expired invite code")
	ErrMaxStrikes       = errors.New("maximum strikes reached")
	ErrParticipantLimit = errors.New("circle must have at least 2 members")
	ErrInvalidUUID      = errors.New("invalid UUID format")
)
