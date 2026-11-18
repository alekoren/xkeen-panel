package xkeen

// EnsureAPIConfigForRuntime must be used by startup code instead of the
// legacy migration: inactive cores must not have their config rewritten.
func EnsureAPIConfigForRuntime(rt Runtime, state PoolState, apiAddr string) (bool, error) {
	if !rt.Installed || rt.Core != CoreXray {
		return false, nil
	}
	return EnsureAPIConfig(state, apiAddr)
}
