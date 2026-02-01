# Iterative Improvement Process
When proposing changes to the codebase, follow an iterative improvement process. Start with small, incremental changes that enhance code quality, readability, and maintainability.
After each iteration execute tests and linting to ensure code quality. If issues are found, fix them before proceeding to the next iteration. Use best practices for logging and error handling, following the conventions of the existing codebase.

# Refactoring Guidelines
Propose refactorings that improve code clarity and maintainability, and follow the most established Go best practices.
Proposed changes should be minimal yet effective, ensuring that the core functionality remains intact while enhancing the overall quality of the codebase.

# Code Duplication
Avoid duplicate code and ensure that logging statements are not redundant. Each log message should provide unique and valuable information about the application's state or behavior. If you find duplicate code, refactor it to eliminate redundancy.

# Configuration over magic numbers 
We do not want magic numbers or other configuration written directly in the code use and propose always to use the configuration pattern defined in config package.

# Testing
After making changes, run all existing tests to ensure that the modifications do not introduce regressions. If new functionality is added, write appropriate unit and integration tests to cover the new code paths. Ensure that tests are clear, concise, and follow the existing testing conventions in the codebase. After test execution check the coverage level to ensure at least a 70% of coverage for all packages, if this threashold is missed please add all needed and relevant unit test to hit the required score.

# Concurrency Tests
When modifying code that involves concurrency (e.g., goroutines, channels), ensure that proper synchronization mechanisms are in place to prevent race conditions. Use Go's race detector during testing to identify potential concurrency issues. 
    

# Logging and Error Handling Guidelines
Use the existing logger for logging messages. use appropriate log levels (Debug, Info, Warn, Error) based on the significance of the events being logged. Use Debug for detailed internal information, Info for general operational messages, Warn for potential issues, and Error for serious problems that need attention. 
Use consistent formatting for log messages to enhance readability and maintainability. use structured logging where possible, including relevant context in log entries. use the logger's WithComponent method to specify the component or module generating the log message. Ensure that log messages provide sufficient context to understand the event without being overly verbose. use error wrapping to provide more context when logging errors. Use fmt.Errorf("additional context: %w", err) to wrap errors before logging them.

# Documentation
If the changes affect the public API or behavior of the application, update the relevant documentation to reflect these changes. Ensure that any new configuration options or features are well-documented for users and developers. Update also documentation in README.md and other relevant files.

# Language Consistency
Ensure to write all comments and log messages in English. Also all the documentation shall be in English except for the content of file: "progetto.txt" 

# Configuration Management
When adding new configuration options, ensure they are properly defined in the configuration files and loaded into the application. Provide sensible default values and document the purpose and usage of each configuration option.
