package testutils

import "os"

// SavedEnv captures the previous state of an environment variable.
type SavedEnv struct {
	Key   string
	Had   bool
	Value string
}

// SetEnv sets an environment variable and returns its previous state.
func SetEnv(key, value string) SavedEnv {
	prev, had := os.LookupEnv(key)
	_ = os.Setenv(key, value)
	return SavedEnv{Key: key, Had: had, Value: prev}
}

// RestoreEnv restores environment variables to a previously saved state.
func RestoreEnv(envs []SavedEnv) {
	for _, env := range envs {
		if env.Had {
			_ = os.Setenv(env.Key, env.Value)
		} else {
			_ = os.Unsetenv(env.Key)
		}
	}
}
