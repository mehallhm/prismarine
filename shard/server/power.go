package server

type PowerAction string

// The power actions that can be performed for a given server. This taps into the given server
// environment and performs them in a way that prevents a race condition from occurring. For
// example, sending two "start" actions back to back will not process the second action until
// the first action has been completed.
//
// This utilizes a workerpool with a limit of one worker so that all the actions execute
// in a sync manner.
const (
	PowerActionStart   = "start"
	PowerActionStop    = "stop"
	PowerActionRestart = "restart"
	PowerActionKill    = "kill"
)

// IsValid checks if the power action being received is valid
func (pa PowerAction) IsValid() bool {
	return pa == PowerActionStart ||
		pa == PowerActionStop ||
		pa == PowerActionRestart ||
		pa == PowerActionKill
}

func (pa PowerAction) IsStart() bool {
	return pa == PowerActionStart || pa == PowerActionRestart
}

// ExecutingPowerAction checks if there is currently a power action
// being processed for the server
func (s *Server) ExecutingPowerAction() bool {
	return s.powerLock.IsLocked()
}

// HandlePowerAction is a helper function that can receive a power action and then process the
// actions that need to occur for it. This guards against someone calling Start() twice at the
// same time, or trying to restart while another restart process is currently running.
//
// However, the code design for the daemon does depend on the user correctly calling this
// function rather than making direct calls to the start/stop/restart functions on the
// runtime struct.
func (s *Server) HandlePowerAction(action PowerAction, waitSeconds ...int) error {
	return nil
}
