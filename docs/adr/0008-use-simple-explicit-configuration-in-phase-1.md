# Use simple explicit configuration in Phase 1

Phase 1 uses a small explicit `Config` struct loaded from YAML with environment-variable expansion, rather than introducing Viper or a larger configuration framework. This keeps startup validation, required secret checks, test configuration, and provider-specific requirements visible while the configuration surface is still small.
