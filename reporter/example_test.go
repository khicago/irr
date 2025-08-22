package reporter_test

import (
	"context"
	"fmt"

	"github.com/khicago/irr"
	"github.com/khicago/irr/reporter"
)

// Example: Solving the cache miss logging problem
// This shows how ErrorReporter solves the intermediate process logging issue
func ExampleStructuredReporter_cacheWithFallback() {
	// Setup logger and reporter
	logger := &ExampleLogger{}
	config := reporter.DefaultConfig()
	errorReporter := reporter.NewStructuredReporter(logger, config)

	ctx := context.Background()

	// Simulate a service method that tries cache first, then database
	result, err := processUserDataWithFallback(ctx, "user123", errorReporter)
	if err != nil {
		fmt.Printf("Final error: %v\n", err)
		return
	}

	fmt.Printf("Result: %s\n", result)
	fmt.Printf("Logged messages: %d\n", len(logger.messages))

	// Output:
	// Result: user data from database
	// Logged messages: 1
}

// Example service function that demonstrates intermediate error reporting
func processUserDataWithFallback(ctx context.Context, userID string, reporter reporter.ErrorReporter) (string, error) {
	// Try to get data from cache
	data, err := getFromCache(userID)
	if err != nil {
		// 🎯 KEY SOLUTION: Report cache miss as warning but continue processing
		warningErr := irr.Wraps(err, "cache miss for user %s, falling back to database", userID).
			Tag("user_id", userID).
			Tag("fallback", "database").
			Tag("cache_type", "redis").
			BuildWithContext(ctx)

		// This reports the warning but doesn't terminate the function
		reporter.ReportWarning(warningErr)

		// Try database fallback
		data, err = getFromDatabase(userID)
		if err != nil {
			// This is a real error that we need to return
			return "", irr.Wraps(err, "failed to get user data from both cache and database").
				Tag("user_id", userID).
				BuildWithContext(ctx)
		}
	}

	return data, nil
}

// Simulated cache function
func getFromCache(userID string) (string, error) {
	// Simulate cache miss
	return "", fmt.Errorf("cache miss: key %s not found", userID)
}

// Simulated database function
func getFromDatabase(userID string) (string, error) {
	// Simulate successful database query
	return "user data from database", nil
}

// Example logger implementation
type ExampleLogger struct {
	messages []string
}

func (l *ExampleLogger) Debug(msg string, fields ...reporter.Field) {
	l.messages = append(l.messages, fmt.Sprintf("DEBUG: %s", msg))
}

func (l *ExampleLogger) Info(msg string, fields ...reporter.Field) {
	l.messages = append(l.messages, fmt.Sprintf("INFO: %s", msg))
}

func (l *ExampleLogger) Warn(msg string, fields ...reporter.Field) {
	l.messages = append(l.messages, fmt.Sprintf("WARN: %s", msg))
}

func (l *ExampleLogger) Error(msg string, fields ...reporter.Field) {
	l.messages = append(l.messages, fmt.Sprintf("ERROR: %s", msg))
}

func (l *ExampleLogger) Fatal(msg string, fields ...reporter.Field) {
	l.messages = append(l.messages, fmt.Sprintf("FATAL: %s", msg))
}

// Example: Enterprise service with comprehensive error reporting
func ExampleStructuredReporter_enterpriseService() {
	// Setup comprehensive error handling
	logger := &ExampleLogger{}
	config := &reporter.ReporterConfig{
		MinLevel:          reporter.LevelInfo,
		MetricsLevel:      reporter.LevelWarning,
		AlertLevel:        reporter.LevelError,
		IncludeStackTrace: true,
		MaxTagsPerReport:  20,
	}

	// Create reporter with metrics and alerting
	errorReporter := reporter.NewStructuredReporter(logger, config).
		WithMetrics(&ExampleMetricsReporter{}).
		WithAlerter(&ExampleAlerter{})

	ctx := context.Background()

	// Simulate various error scenarios in enterprise service
	service := &EnterpriseService{reporter: errorReporter}

	err := service.ProcessRequest(ctx, "request123")
	if err != nil {
		fmt.Printf("Request failed: %v\n", err)
	} else {
		fmt.Println("Request processed successfully")
	}

	// Output:
	// Request processed successfully
}

// Example enterprise service
type EnterpriseService struct {
	reporter reporter.ErrorReporter
}

func (s *EnterpriseService) ProcessRequest(ctx context.Context, requestID string) error {
	// Step 1: Try primary service
	err := s.callPrimaryService(ctx, requestID)
	if err != nil {
		// Report as info - this is expected fallback behavior
		infoErr := irr.Wraps(err, "primary service unavailable, using fallback").
			Tag("request_id", requestID).
			Tag("service", "primary").
			Tag("fallback", "secondary").
			BuildWithContext(ctx)
		s.reporter.ReportInfo(infoErr)

		// Try fallback service
		err = s.callFallbackService(ctx, requestID)
		if err != nil {
			// This is a real problem - both services failed
			criticalErr := irr.Wraps(err, "both primary and fallback services failed").
				Tag("request_id", requestID).
				Trace().
				BuildWithContext(ctx)
			s.reporter.ReportCritical(criticalErr)
			return criticalErr
		}
	}

	// Step 2: Cache the result (optional)
	if err := s.cacheResult(ctx, requestID, "result"); err != nil {
		// Report as warning - caching failure shouldn't fail the request
		warningErr := irr.Wraps(err, "failed to cache result").
			Tag("request_id", requestID).
			Tag("cache_operation", "set").
			BuildWithContext(ctx)
		s.reporter.ReportWarning(warningErr)
		// Continue processing despite cache failure
	}

	return nil
}

func (s *EnterpriseService) callPrimaryService(ctx context.Context, requestID string) error {
	// Simulate service failure
	return fmt.Errorf("primary service timeout")
}

func (s *EnterpriseService) callFallbackService(ctx context.Context, requestID string) error {
	// Simulate successful fallback
	return nil
}

func (s *EnterpriseService) cacheResult(ctx context.Context, requestID, result string) error {
	// Simulate cache failure
	return fmt.Errorf("redis connection failed")
}

// Example metrics reporter
type ExampleMetricsReporter struct {
	reports int
}

func (m *ExampleMetricsReporter) ReportErrorMetrics(level reporter.ReportLevel, err error) error {
	m.reports++
	// Don't print to stdout to avoid interfering with example output
	return nil
}

// Example alerter
type ExampleAlerter struct {
	alerts int
}

func (a *ExampleAlerter) SendAlert(level reporter.ReportLevel, err error, fields []reporter.Field) error {
	a.alerts++
	fmt.Printf("Alert: %s level alert sent\n", level.String())
	return nil
}

// Example: Usage patterns comparison
func Example_usagePatterns() {
	fmt.Println("=== Before v2 (problematic) ===")
	fmt.Println("// Manual logging, no structure, no metrics")
	fmt.Println("if err := cache.Get(); err != nil {")
	fmt.Printf("    log.Printf(\"cache miss: %%v\", err) // ❌ No structure, no metrics\n")
	fmt.Println("    // fallback to database...")
	fmt.Println("}")
	fmt.Println()

	fmt.Println("=== After v2 (structured) ===")
	fmt.Println("// Structured reporting with automatic metrics")
	fmt.Println("if err := cache.Get(); err != nil {")
	fmt.Println("    warningErr := irr.Wraps(err, \"cache miss\").")
	fmt.Println("        Tag(\"cache_type\", \"redis\").")
	fmt.Println("        BuildWithContext(ctx)")
	fmt.Println("    reporter.ReportWarning(warningErr) // ✅ Structured + metrics")
	fmt.Println("    // fallback to database...")
	fmt.Println("}")

	// Output:
	// === Before v2 (problematic) ===
	// // Manual logging, no structure, no metrics
	// if err := cache.Get(); err != nil {
	//     log.Printf("cache miss: %v", err) // ❌ No structure, no metrics
	//     // fallback to database...
	// }
	//
	// === After v2 (structured) ===
	// // Structured reporting with automatic metrics
	// if err := cache.Get(); err != nil {
	//     warningErr := irr.Wraps(err, "cache miss").
	//         Tag("cache_type", "redis").
	//         BuildWithContext(ctx)
	//     reporter.ReportWarning(warningErr) // ✅ Structured + metrics
	//     // fallback to database...
	// }
}
