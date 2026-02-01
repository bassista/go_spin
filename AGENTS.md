# Iterative Improvement Process
When proposing changes to the codebase, always follow an iterative improvement process. Start with small, incremental changes that enhance code quality, readability, and maintainability.
After each iteration, execute tests and linting to ensure code quality. If issues are found, fix them before proceeding to the next iteration. Use best practices for logging and error handling, following the conventions of the existing codebase.
# Refactoring Guidelines
Propose refactorings that improve code clarity and maintainability, and follow the most established Go best practices.
Proposed changes should be minimal yet effective, ensuring that the core functionality remains intact while enhancing the overall quality of the codebase.
# Code Duplication
Avoid duplicate code and ensure that logging statements are not redundant. Each log message should provide unique and valuable information about the application's state or behavior. If you find duplicate code, refactor it to eliminate redundancy.
# Configuration over magic numbers
Do not use magic numbers or other configuration written directly in the code. Always use and propose the configuration pattern defined in the config package.
# Testing
After making changes, run all existing tests to ensure that the modifications do not introduce regressions. If new functionality is added, write appropriate unit and integration tests to cover the new code paths. Ensure that tests are clear, concise, and follow the existing testing conventions in the codebase. After test execution, check the coverage level to ensure at least 70% coverage for all packages. If this threshold is missed, please add all needed and relevant unit tests to hit the required score.
# Concurrency Tests
When modifying code that involves concurrency (e.g., goroutines, channels), ensure that proper synchronization mechanisms are in place to prevent race conditions. Use Go's race detector during testing to identify potential concurrency issues.
# Logging and Error Handling Guidelines
Use the existing logger for logging messages. Use appropriate log levels (Debug, Info, Warn, Error) based on the significance of the events being logged. Use Debug for detailed internal information, Info for general operational messages, Warn for potential issues, and Error for serious problems that need attention.
# Documentation
If the changes affect the public API, configuration, or runtime behavior, update the relevant documentation to reflect these changes:
- **README.md**: Document all new endpoints, configuration options, and usage patterns. For new endpoints, add a table row in the API Endpoints section. For new config options, update the example YAML and environment variable sections.
- **docs/progetto.txt**: Update the architectural description and workflows if new patterns, endpoints, or scheduling/runtime behaviors are introduced (this file remains in Italian).
- **AGENTS.md**: Update this guide if new process rules, test conventions, or documentation policies are introduced.
When adding a new endpoint:
- Document the HTTP method, path, and a short description in README.md (API Endpoints table)
- Describe the expected input/output and error codes
- If the endpoint uses new config, document the config key and its default/expected values
When adding new configuration options:
- Add the option to the YAML example in README.md
- Document the environment variable override (if any)
- Update progetto.txt if the option impacts scheduling, runtime, or persistence behavior
When changing runtime or scheduling behavior:
- Update the architecture and workflow sections in progetto.txt
- Add a note in README.md if user-facing behavior changes
All documentation must be kept in sync with the codebase after every relevant change.
# Language Consistency
Ensure to write all comments and log messages in English. All documentation must be in English except for the content of the file: "progetto.txt" (which remains in Italian).
# Configuration Management
When adding new configuration options, ensure they are properly defined in the configuration files and loaded into the application. Provide sensible default values and document the purpose and usage of each configuration option.