package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// OpenAPISpec represents a simplified structure of an OpenAPI specification.
type OpenAPISpec struct {
	Paths      map[string]map[string]interface{} `json:"paths"`
	Components struct {
		Schemas map[string]interface{} `json:"schemas"`
	} `json:"components"`
}

// CoverageReport provides a comprehensive summary of API and schema coverage.
type CoverageReport struct {
	TotalEndpoints     int                         `json:"total_endpoints"`
	CoveredEndpoints   int                         `json:"covered_endpoints"`
	CoveragePercent    float64                     `json:"coverage_percent"`
	EndpointDetails    map[string]Coverage         `json:"endpoint_details"`
	SchemaAnalysis     SchemaAnalysis              `json:"schema_analysis"`
	ResourceAnalysis   map[string]ResourceCoverage `json:"resource_analysis"`
	DataSourceAnalysis map[string]ResourceCoverage `json:"data_source_analysis"`
	DebugInfo          DebugInfo                   `json:"debug_info"`
}

// Coverage details the status of a single API endpoint.
type Coverage struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
	Covered []string `json:"covered"`
	Missing []string `json:"missing"`
}

// SchemaAnalysis provides analysis of OpenAPI schemas vs Terraform schemas.
type SchemaAnalysis struct {
	TotalSchemas          int                       `json:"total_schemas"`
	CoveredSchemas        int                       `json:"covered_schemas"`
	SchemaCoveragePercent float64                   `json:"schema_coverage_percent"`
	SchemaDetails         map[string]SchemaCoverage `json:"schema_details"`
}

// SchemaCoverage details the coverage of a single schema.
type SchemaCoverage struct {
	SchemaName        string              `json:"schema_name"`
	TotalProperties   int                 `json:"total_properties"`
	CoveredProperties int                 `json:"covered_properties"`
	Properties        map[string]Property `json:"properties"`
	MissingProperties []string            `json:"missing_properties"`
	ExtraProperties   []string            `json:"extra_properties"`
}

// Property represents a schema property.
type Property struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Required      bool   `json:"required"`
	Covered       bool   `json:"covered"`
	TerraformType string `json:"terraform_type,omitempty"`
}

// ResourceCoverage details the coverage of a Terraform resource or data source.
type ResourceCoverage struct {
	ResourceName        string              `json:"resource_name"`
	SchemaProperties    map[string]Property `json:"schema_properties"`
	SupportedOperations []string            `json:"supported_operations"`
	RelatedEndpoints    []string            `json:"related_endpoints"`
	CoverageScore       float64             `json:"coverage_score"`
}

// DebugInfo provides debugging information about the analysis.
type DebugInfo struct {
	FilesAnalyzed      int                 `json:"files_analyzed"`
	FilesWithErrors    int                 `json:"files_with_errors"`
	TotalAPICallsFound int                 `json:"total_api_calls_found"`
	AnalyzedFiles      []string            `json:"analyzed_files"`
	Errors             []string            `json:"errors"`
	FoundEndpoints     map[string][]string `json:"found_endpoints"`
	FoundResources     []string            `json:"found_resources"`
	FoundDataSources   []string            `json:"found_data_sources"`
}

// Constants to avoid repetition.
const (
	trueString   = "true"
	postMethod   = "POST"
	getMethod    = "GET"
	putMethod    = "PUT"
	deleteMethod = "DELETE"
	patchMethod  = "PATCH"
)

func main() {
	openAPIFile := flag.String("openapi", "api/openapi.json", "Path to the OpenAPI JSON file")
	providerDir := flag.String("provider", "bastion/", "Path to the provider's source code directory")
	outputFile := flag.String("output", "coverage-report.json", "Path to save the JSON coverage report")
	verbose := flag.Bool("verbose", false, "Enable verbose output for debugging")
	analyzeSchemas := flag.Bool("schemas", true, "Analyze OpenAPI schemas vs Terraform schemas")

	flag.Parse()

	err := run(*openAPIFile, *providerDir, *outputFile, *verbose, *analyzeSchemas)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run(openAPIFile, providerDir, outputFile string, verbose, analyzeSchemas bool) error {
	log.Printf("Analyzing OpenAPI spec from %s...", openAPIFile)

	spec, err := parseOpenAPI(openAPIFile)
	if err != nil {
		return fmt.Errorf("error parsing OpenAPI spec: %w", err)
	}

	log.Printf("Analyzing provider code in %s...", providerDir)

	_, statErr := os.Stat(providerDir)
	if os.IsNotExist(statErr) {
		return fmt.Errorf("provider directory '%s' does not exist", providerDir)
	}

	providerEndpoints, debugInfo, err := analyzeProviderWithDebug(providerDir, verbose)
	if err != nil {
		return fmt.Errorf("error analyzing provider code: %w", err)
	}

	// Analyze Terraform resources and data sources
	resourceAnalysis, dataSourceAnalysis, err := analyzeTerraformResources(providerDir, verbose, debugInfo)
	if err != nil {
		return fmt.Errorf("error analyzing Terraform resources: %w", err)
	}

	var schemaAnalysis SchemaAnalysis

	if analyzeSchemas {
		log.Println("Analyzing schemas...")

		schemaAnalysis = analyzeOpenAPISchemas(spec, resourceAnalysis, dataSourceAnalysis, verbose)
	}

	if verbose {
		log.Printf("Debug: Analyzed %d files", debugInfo.FilesAnalyzed)
		log.Printf("Debug: Found %d API calls", debugInfo.TotalAPICallsFound)
		log.Printf("Debug: Found %d resources", len(debugInfo.FoundResources))
		log.Printf("Debug: Found %d data sources", len(debugInfo.FoundDataSources))
	}

	log.Println("Generating comprehensive coverage report...")

	report := generateComprehensiveCoverageReport(spec, providerEndpoints, schemaAnalysis,
		resourceAnalysis, dataSourceAnalysis, debugInfo)

	log.Println("\n=== Comprehensive Coverage Report ===")
	printComprehensiveReport(report)

	log.Printf("Saving report to %s...", outputFile)

	return saveReport(report, outputFile)
}

func parseOpenAPI(filename string) (*OpenAPISpec, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var spec OpenAPISpec

	err = json.Unmarshal(data, &spec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &spec, nil
}

func analyzeProviderWithDebug(dir string, verbose bool) (map[string][]string, *DebugInfo, error) {
	endpoints := make(map[string][]string)
	debugInfo := &DebugInfo{
		AnalyzedFiles:    []string{},
		Errors:           []string{},
		FoundEndpoints:   make(map[string][]string),
		FoundResources:   []string{},
		FoundDataSources: []string{},
	}

	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errMsg := fmt.Sprintf("Walk error in %s: %v", path, err)
			debugInfo.Errors = append(debugInfo.Errors, errMsg)

			return nil
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		debugInfo.FilesAnalyzed++
		debugInfo.AnalyzedFiles = append(debugInfo.AnalyzedFiles, path)

		if verbose {
			log.Printf("Analyzing file: %s", path)
		}

		fset := token.NewFileSet()

		node, parseErr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if parseErr != nil {
			debugInfo.FilesWithErrors++
			errMsg := fmt.Sprintf("Parse error in %s: %v", path, parseErr)
			debugInfo.Errors = append(debugInfo.Errors, errMsg)

			return nil
		}

		// Analyze file content using multiple strategies
		analyzeFileContent(node, path, endpoints, debugInfo, verbose)

		return nil
	})

	// Normalize endpoints
	normalizedEndpoints := normalizeEndpoints(endpoints, debugInfo, verbose)

	if walkErr != nil {
		return normalizedEndpoints, debugInfo, fmt.Errorf("error walking provider directory: %w", walkErr)
	}

	return normalizedEndpoints, debugInfo, nil
}

func analyzeTerraformResources(dir string, verbose bool,
	debugInfo *DebugInfo) (map[string]ResourceCoverage, map[string]ResourceCoverage, error) {
	resources := make(map[string]ResourceCoverage)
	dataSources := make(map[string]ResourceCoverage)

	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		filename := filepath.Base(path)

		// Analyze resource files
		if strings.HasPrefix(filename, "resource_") {
			resourceName := extractResourceName(filename)
			if resourceName != "" {
				coverage := analyzeResourceFile(path, resourceName, verbose)
				resources[resourceName] = coverage
				debugInfo.FoundResources = append(debugInfo.FoundResources, resourceName)
			}
		}

		// Analyze data source files
		if strings.HasPrefix(filename, "data_source_") {
			dataSourceName := extractDataSourceName(filename)
			if dataSourceName != "" {
				coverage := analyzeDataSourceFile(path, dataSourceName, verbose)
				dataSources[dataSourceName] = coverage
				debugInfo.FoundDataSources = append(debugInfo.FoundDataSources, dataSourceName)
			}
		}

		return nil
	})
	if walkErr != nil {
		return resources, dataSources, fmt.Errorf("error walking provider directory: %w", walkErr)
	}

	return resources, dataSources, nil
}

func analyzeResourceFile(path, resourceName string, verbose bool) ResourceCoverage {
	coverage := ResourceCoverage{
		ResourceName:        resourceName,
		SchemaProperties:    make(map[string]Property),
		SupportedOperations: []string{},
		RelatedEndpoints:    []string{},
	}

	file, err := os.Open(path)
	if err != nil {
		return coverage
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return coverage
	}

	fileContent := string(content)

	if verbose {
		log.Printf("Analyzing resource file: %s", path)
	}

	// Extract schema properties from the file
	coverage.SchemaProperties = extractSchemaProperties(fileContent, verbose)

	// Determine supported operations based on function presence
	operations := []string{"Create", "Read", "Update", "Delete"}
	for _, op := range operations {
		patterns := []string{
			"resource" + op,
			op + ":",
			"func.*" + op,
		}

		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, fileContent); matched {
				if !contains(coverage.SupportedOperations, strings.ToUpper(op)) {
					coverage.SupportedOperations = append(coverage.SupportedOperations, strings.ToUpper(op))
				}

				break
			}
		}
	}

	// Extract related endpoints
	coverage.RelatedEndpoints = extractRelatedEndpoints(fileContent)

	if verbose {
		log.Printf("  Found %d properties, %d operations, %d endpoints",
			len(coverage.SchemaProperties), len(coverage.SupportedOperations), len(coverage.RelatedEndpoints))
	}

	return coverage
}

func analyzeDataSourceFile(path, dataSourceName string, verbose bool) ResourceCoverage {
	coverage := ResourceCoverage{
		ResourceName:        dataSourceName,
		SchemaProperties:    make(map[string]Property),
		SupportedOperations: []string{"READ"}, // Data sources only support read
		RelatedEndpoints:    []string{},
	}

	file, err := os.Open(path)
	if err != nil {
		return coverage
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return coverage
	}

	fileContent := string(content)

	// Extract schema properties from the file
	coverage.SchemaProperties = extractSchemaProperties(fileContent, verbose)

	// Extract related endpoints
	coverage.RelatedEndpoints = extractRelatedEndpoints(fileContent)

	return coverage
}

func checkPropertyCoverage(coverage SchemaCoverage, resources,
	dataSources map[string]ResourceCoverage, verbose bool) SchemaCoverage {
	allCoverage := make(map[string]ResourceCoverage)
	for k, v := range resources {
		allCoverage[k] = v
	}

	for k, v := range dataSources {
		allCoverage[k] = v
	}

	if verbose {
		log.Printf("  Checking coverage for schema '%s' with %d properties", coverage.SchemaName, len(coverage.Properties))
		log.Printf("  Available resources/datasources: %d", len(allCoverage))
	}

	for propName, property := range coverage.Properties {
		found, updatedProperty := findPropertyCoverage(propName, property, coverage.SchemaName, allCoverage, verbose)

		coverage.Properties[propName] = updatedProperty
		if found {
			coverage.CoveredProperties++
		} else {
			coverage.MissingProperties = append(coverage.MissingProperties, propName)
			if verbose {
				log.Printf("    ✗ Property %s.%s not found in any resource",
					coverage.SchemaName, propName)
			}
		}
	}

	return coverage
}

// findPropertyCoverage checks if a property is covered in resources/dataSources.
func findPropertyCoverage(
	propName string,
	property Property,
	schemaName string,
	allCoverage map[string]ResourceCoverage,
	verbose bool,
) (bool, Property) {
	// Direct search by property name
	for resourceName, resourceCoverage := range allCoverage {
		if _, exists := resourceCoverage.SchemaProperties[propName]; exists {
			property.Covered = true

			if verbose {
				log.Printf("    ✓ Property %s.%s found in resource %s",
					schemaName, propName, resourceName)
			}

			return true, property
		}
	}
	// Approximate matches
	for resourceName, resourceCoverage := range allCoverage {
		for terraformProp := range resourceCoverage.SchemaProperties {
			if isPropertyMatch(propName, terraformProp) {
				property.Covered = true

				if verbose {
					log.Printf("    ≈ Property %s.%s matches %s in resource %s",
						schemaName, propName, terraformProp, resourceName)
				}

				return true, property
			}
		}
	}
	// Search based on schema name
	schemaBaseName := extractSchemaBaseName(schemaName)
	for resourceName, resourceCoverage := range allCoverage {
		if strings.Contains(resourceName, schemaBaseName) || strings.Contains(schemaBaseName, resourceName) {
			if _, exists := resourceCoverage.SchemaProperties[propName]; exists {
				property.Covered = true

				if verbose {
					log.Printf("    ~ Property %s.%s found via schema match in resource %s",
						schemaName, propName, resourceName)
				}

				return true, property
			}
		}
	}

	return false, property
}

// isPropertyMatch checks if two property names match.
func isPropertyMatch(openAPIProp, terraformProp string) bool {
	// Normalize names
	openAPILower := strings.ToLower(openAPIProp)
	terraformLower := strings.ToLower(terraformProp)

	// Exact match
	if openAPILower == terraformLower {
		return true
	}

	// Convert snake_case <-> camelCase
	openAPISnake := camelToSnake(openAPIProp)
	terraformSnake := camelToSnake(terraformProp)

	if strings.EqualFold(openAPISnake, terraformSnake) {
		return true
	}

	// Common mappings
	commonMappings := map[string][]string{
		"id":          {"identifier", "uuid"},
		"name":        {"display_name", "title", "label"},
		"description": {"desc", "comment"},
		"enabled":     {"is_enabled", "active", "is_active"},
		"type":        {"kind", "category"},
		"url":         {"uri", "link", "href"},
		"host":        {"hostname", "server", "address"},
		"port":        {"port_number"},
		"login":       {"username", "user_name", "account_login"},
		"password":    {"secret", "passphrase"},
	}

	for canonical, variants := range commonMappings {
		if (openAPILower == canonical && contains(variants, terraformLower)) ||
			(terraformLower == canonical && contains(variants, openAPILower)) ||
			(contains(variants, openAPILower) && contains(variants, terraformLower)) {
			return true
		}
	}

	return false
}

// camelToSnake converts camelCase to snake_case.
func camelToSnake(str string) string {
	snake := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	snake = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}

// extractSchemaBaseName extracts the base name of a schema.
func extractSchemaBaseName(schemaName string) string {
	// Remove common suffixes
	suffixes := []string{"_get", "_post", "_put", "_delete", "_patch"}
	base := schemaName

	for _, suffix := range suffixes {
		if strings.HasSuffix(base, suffix) {
			base = strings.TrimSuffix(base, suffix)

			break
		}
	}

	return base
}

// extractSchemaProperties enhanced for better extraction.
func extractSchemaProperties(content string, verbose bool) map[string]Property {
	properties := extractSchemaBlockProperties(content, verbose)
	if len(properties) == 0 {
		properties = extractSimpleProperties(content, verbose)
	}

	return properties
}

func extractSchemaBlockProperties(content string, verbose bool) map[string]Property {
	properties := make(map[string]Property)
	schemaPattern := regexp.MustCompile(`"([a-zA-Z_][a-zA-Z0-9_]*)"\s*:\s*&?schema\.Schema\s*{([^}]+)}`)
	typePattern := regexp.MustCompile(`Type:\s*schema\.Type([A-Za-z]+)`)
	requiredPattern := regexp.MustCompile(`Required:\s*(` + trueString + `|false)`)
	optionalPattern := regexp.MustCompile(`Optional:\s*(` + trueString + `|false)`)
	computedPattern := regexp.MustCompile(`Computed:\s*(` + trueString + `|false)`)

	matches := schemaPattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) > 2 {
			propertyName := match[1]
			schemaBlock := match[2]

			property := Property{
				Name:     propertyName,
				Type:     "unknown",
				Required: false,
				Covered:  true,
			}
			if typeMatch := typePattern.FindStringSubmatch(schemaBlock); len(typeMatch) > 1 {
				property.Type = strings.ToLower(typeMatch[1])
				property.TerraformType = "Type" + typeMatch[1]
			}

			if requiredMatch := requiredPattern.FindStringSubmatch(schemaBlock); len(requiredMatch) > 1 {
				property.Required = requiredMatch[1] == trueString
			} else if optionalMatch := optionalPattern.FindStringSubmatch(schemaBlock); len(optionalMatch) > 1 {
				property.Required = optionalMatch[1] != trueString
			}

			computedMatch := computedPattern.FindStringSubmatch(schemaBlock)
			if len(computedMatch) > 1 && computedMatch[1] == trueString {
				property.Required = false
			}

			properties[propertyName] = property
			if verbose {
				log.Printf("    Found property: %s (type: %s, required: %t)",
					propertyName, property.Type, property.Required)
			}
		}
	}

	return properties
}

func extractSimpleProperties(content string, verbose bool) map[string]Property {
	properties := make(map[string]Property)
	simplePattern := regexp.MustCompile(`"([a-zA-Z_][a-zA-Z0-9_]*)"\s*:`)

	simpleMatches := simplePattern.FindAllStringSubmatch(content, -1)
	for _, match := range simpleMatches {
		if len(match) > 1 {
			propertyName := match[1]
			if !isGoKeyword(propertyName) && !strings.HasPrefix(propertyName, "Test") {
				properties[propertyName] = Property{
					Name:     propertyName,
					Type:     "unknown",
					Required: false,
					Covered:  true,
				}
				if verbose {
					log.Printf("    Found simple property: %s", propertyName)
				}
			}
		}
	}

	return properties
}

// isGoKeyword checks if it's a Go keyword to avoid.
func isGoKeyword(word string) bool {
	keywords := []string{
		"package", "import", "func", "var", "const", "type", "struct",
		"interface", "map", "chan", "select", "case", "default", "if",
		"else", "switch", "for", "range", "go", "defer", "return",
		"break", "continue", "fallthrough", "goto",
	}

	for _, keyword := range keywords {
		if word == keyword {
			return true
		}
	}

	return false
}

func analyzeFileContent(
	_ *ast.File,
	filePath string,
	endpoints map[string][]string,
	debugInfo *DebugInfo,
	verbose bool,
) {
	// Strategy 1: String analysis for URL patterns
	analyzeStringLiterals(filePath, endpoints, debugInfo, verbose)
}

func analyzeStringLiterals(filePath string, endpoints map[string][]string, debugInfo *DebugInfo, verbose bool) {
	// Regex patterns to identify API endpoints
	apiPatterns := []*regexp.Regexp{
		regexp.MustCompile(`"(/[a-zA-Z0-9_/-]*[a-zA-Z0-9_])"`),
		regexp.MustCompile(`"/api/v\d+\.\d+(/[a-zA-Z0-9_/-]*[a-zA-Z0-9_])"`),
		regexp.MustCompile(`"(/config/[a-zA-Z0-9_/-]*)"`),
		regexp.MustCompile(`"(/auth[a-zA-Z0-9_/-]*)"`),
		regexp.MustCompile(`"(/users?[a-zA-Z0-9_/-]*)"`),
		regexp.MustCompile(`"(/devices?[a-zA-Z0-9_/-]*)"`),
		regexp.MustCompile(`"(/domains?[a-zA-Z0-9_/-]*)"`),
		regexp.MustCompile(`"(/applications?[a-zA-Z0-9_/-]*)"`),
	}

	methodPatterns := map[*regexp.Regexp]string{
		regexp.MustCompile(`http\.MethodGet`):        getMethod,
		regexp.MustCompile(`http\.MethodPost`):       postMethod,
		regexp.MustCompile(`http\.MethodPut`):        putMethod,
		regexp.MustCompile(`http\.MethodDelete`):     deleteMethod,
		regexp.MustCompile(`http\.MethodPatch`):      patchMethod,
		regexp.MustCompile(`"` + getMethod + `"`):    getMethod,
		regexp.MustCompile(`"` + postMethod + `"`):   postMethod,
		regexp.MustCompile(`"` + putMethod + `"`):    putMethod,
		regexp.MustCompile(`"` + deleteMethod + `"`): deleteMethod,
		regexp.MustCompile(`"` + patchMethod + `"`):  patchMethod,
	}

	// Read file content for regex analysis
	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return
	}

	fileContent := string(content)

	// Search for URLs
	var foundURLs []string

	for _, pattern := range apiPatterns {
		matches := pattern.FindAllStringSubmatch(fileContent, -1)
		for _, match := range matches {
			if len(match) > 1 {
				url := match[1]
				foundURLs = append(foundURLs, url)
			}
		}
	}

	// Search for HTTP methods in context
	var foundMethods []string

	for pattern, method := range methodPatterns {
		if pattern.MatchString(fileContent) {
			foundMethods = append(foundMethods, method)
		}
	}

	// Associate found URLs with methods
	if len(foundURLs) > 0 {
		for _, url := range foundURLs {
			if len(foundMethods) > 0 {
				for _, method := range foundMethods {
					debugInfo.TotalAPICallsFound++

					if verbose {
						log.Printf("  String Found: %s %s", method, url)
					}

					endpoints[url] = append(endpoints[url], method)
					debugInfo.FoundEndpoints[url] = append(debugInfo.FoundEndpoints[url], method)
				}
			} else {
				// If no method found, try to guess based on context
				method := guessHTTPMethod(fileContent, url)
				debugInfo.TotalAPICallsFound++

				if verbose {
					log.Printf("  String Found (guessed): %s %s", method, url)
				}

				endpoints[url] = append(endpoints[url], method)
				debugInfo.FoundEndpoints[url] = append(debugInfo.FoundEndpoints[url], method)
			}
		}
	}
}

func guessHTTPMethod(content string, _ string) string {
	lowerContent := strings.ToLower(content)

	// Patterns to guess HTTP method based on context
	if strings.Contains(lowerContent, "create") || strings.Contains(lowerContent, "add") {
		return postMethod
	}

	if strings.Contains(lowerContent, "update") || strings.Contains(lowerContent, "modify") {
		return putMethod
	}

	if strings.Contains(lowerContent, "delete") || strings.Contains(lowerContent, "remove") {
		return deleteMethod
	}

	if strings.Contains(lowerContent, "read") ||
		strings.Contains(lowerContent, "get") ||
		strings.Contains(lowerContent, "fetch") {
		return getMethod
	}

	// By default, assume GET for data sources and POST for resources
	if strings.Contains(lowerContent, "data_source") {
		return getMethod
	}

	if strings.Contains(lowerContent, "resource") {
		return postMethod
	}

	return getMethod // Default
}

func normalizeEndpoints(endpoints map[string][]string, _ *DebugInfo, verbose bool) map[string][]string {
	normalized := make(map[string][]string)

	for rawURL, methods := range endpoints {
		// Clean URL
		cleanURL := strings.TrimSpace(rawURL)
		cleanURL = strings.TrimSuffix(cleanURL, "/")

		// Ignore URLs that are too short or invalid
		if len(cleanURL) < 2 || !strings.HasPrefix(cleanURL, "/") {
			continue
		}

		// Remove API prefixes if present
		cleanURL = regexp.MustCompile(`^/api/v\d+\.\d+`).ReplaceAllString(cleanURL, "")
		if cleanURL == "" {
			cleanURL = "/"
		}

		// Deduplicate methods
		uniqueMethods := make(map[string]bool)

		var result []string

		for _, method := range methods {
			methodUpper := strings.ToUpper(strings.TrimSpace(method))
			if methodUpper != "" && methodUpper != "UNKNOWN" {
				if !uniqueMethods[methodUpper] {
					uniqueMethods[methodUpper] = true
					result = append(result, methodUpper)
				}
			}
		}

		if len(result) > 0 {
			normalized[cleanURL] = result
			if verbose {
				log.Printf("  Normalized: %s -> %s with methods %v", rawURL, cleanURL, result)
			}
		}
	}

	return normalized
}

func generateComprehensiveCoverageReport(
	spec *OpenAPISpec,
	providerEndpoints map[string][]string,
	schemaAnalysis SchemaAnalysis,
	resourceAnalysis, dataSourceAnalysis map[string]ResourceCoverage,
	debugInfo *DebugInfo,
) *CoverageReport {
	report := &CoverageReport{
		EndpointDetails:    make(map[string]Coverage),
		SchemaAnalysis:     schemaAnalysis,
		ResourceAnalysis:   resourceAnalysis,
		DataSourceAnalysis: dataSourceAnalysis,
		DebugInfo:          *debugInfo,
	}

	for path, methods := range spec.Paths {
		coverage, covered := calculateEndpointCoverage(path, methods, providerEndpoints)
		report.EndpointDetails[path] = coverage

		report.TotalEndpoints++
		if covered {
			report.CoveredEndpoints++
		}
	}

	if report.TotalEndpoints > 0 {
		report.CoveragePercent = float64(report.CoveredEndpoints) / float64(report.TotalEndpoints) * 100
	}

	return report
}

// calculateEndpointCoverage handles the endpoint coverage logic for a single path.
func calculateEndpointCoverage(
	path string,
	methods map[string]interface{},
	providerEndpoints map[string][]string,
) (Coverage, bool) {
	availableMethods := extractAvailableMethods(methods)
	if len(availableMethods) == 0 {
		return Coverage{
			Path:    path,
			Methods: availableMethods,
			Covered: []string{},
			Missing: []string{},
		}, false
	}

	coverage := Coverage{
		Path:    path,
		Methods: availableMethods,
		Covered: []string{},
		Missing: []string{},
	}

	providerMethods, exists := providerEndpoints[path]
	if exists {
		coverage = matchMethods(coverage, availableMethods, providerMethods)
	} else {
		coverage = matchMethodsWithBasePath(coverage, path, availableMethods, providerEndpoints)
	}

	return coverage, len(coverage.Covered) > 0
}

// extractAvailableMethods extracts valid HTTP methods from the OpenAPI methods map.
func extractAvailableMethods(methods map[string]interface{}) []string {
	var availableMethods []string

	for method := range methods {
		if method != "parameters" && method != "x-amazon-apigateway-integration" {
			availableMethods = append(availableMethods, strings.ToUpper(method))
		}
	}

	return availableMethods
}

// matchMethods matches available methods with provider methods for a given path.
func matchMethods(coverage Coverage, availableMethods, providerMethods []string) Coverage {
	for _, method := range availableMethods {
		found := false

		for _, providerMethod := range providerMethods {
			if strings.ToUpper(providerMethod) == method {
				coverage.Covered = append(coverage.Covered, method)
				found = true

				break
			}
		}

		if !found {
			coverage.Missing = append(coverage.Missing, method)
		}
	}

	return coverage
}

// matchMethodsWithBasePath tries to match methods using a base path if an exact match is not found.
func matchMethodsWithBasePath(
	coverage Coverage,
	path string,
	availableMethods []string,
	providerEndpoints map[string][]string,
) Coverage {
	basePath := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, "")
	basePath = strings.TrimSuffix(basePath, "/")
	foundAny := false

	for endpoint, providerMethods := range providerEndpoints {
		if strings.HasPrefix(endpoint, basePath) || strings.HasPrefix(basePath, endpoint) {
			coverage = matchMethods(coverage, availableMethods, providerMethods)
			if len(coverage.Covered) > 0 {
				foundAny = true
			}

			break
		}
	}

	if !foundAny && len(coverage.Covered) == 0 {
		coverage.Missing = availableMethods
	}

	return coverage
}

func printComprehensiveReport(report *CoverageReport) {
	log.Printf("=== ENDPOINT COVERAGE ===")
	log.Printf("Total Endpoints: %d", report.TotalEndpoints)
	log.Printf("Covered Endpoints: %d", report.CoveredEndpoints)
	log.Printf("Coverage Percentage: %.2f%%", report.CoveragePercent)

	log.Printf("\n=== SCHEMA COVERAGE ===")
	log.Printf("Total Schemas: %d", report.SchemaAnalysis.TotalSchemas)
	log.Printf("Covered Schemas: %d", report.SchemaAnalysis.CoveredSchemas)
	log.Printf("Schema Coverage Percentage: %.2f%%", report.SchemaAnalysis.SchemaCoveragePercent)

	log.Printf("\n=== RESOURCE ANALYSIS ===")
	log.Printf("Total Resources: %d", len(report.ResourceAnalysis))

	resourceNames := make([]string, 0, len(report.ResourceAnalysis))
	for name := range report.ResourceAnalysis {
		resourceNames = append(resourceNames, name)
	}

	sort.Strings(resourceNames)

	for _, name := range resourceNames {
		resource := report.ResourceAnalysis[name]
		log.Printf("- %s: %d properties, %v operations",
			name, len(resource.SchemaProperties), resource.SupportedOperations)
	}

	log.Printf("\n=== DATA SOURCE ANALYSIS ===")
	log.Printf("Total Data Sources: %d", len(report.DataSourceAnalysis))

	dataSourceNames := make([]string, 0, len(report.DataSourceAnalysis))
	for name := range report.DataSourceAnalysis {
		dataSourceNames = append(dataSourceNames, name)
	}

	sort.Strings(dataSourceNames)

	for _, name := range dataSourceNames {
		dataSource := report.DataSourceAnalysis[name]
		log.Printf("- %s: %d properties", name, len(dataSource.SchemaProperties))
	}

	log.Printf("\n=== TOP MISSING ENDPOINTS ===")

	missingCount := 0

	for path, coverage := range report.EndpointDetails {
		if len(coverage.Missing) > 0 && missingCount < 10 {
			log.Printf("- %s (%s)", path, strings.Join(coverage.Missing, ", "))

			missingCount++
		}
	}

	if len(report.SchemaAnalysis.SchemaDetails) > 0 {
		log.Printf("\n=== SCHEMA DETAILS (Top 5) ===")

		count := 0

		for schemaName, schema := range report.SchemaAnalysis.SchemaDetails {
			if count >= 5 {
				break
			}

			log.Printf("Schema: %s", schemaName)
			log.Printf("  Total Properties: %d", schema.TotalProperties)
			log.Printf("  Covered Properties: %d", schema.CoveredProperties)
			log.Printf("  Missing Properties: %v", schema.MissingProperties)
			log.Println()

			count++
		}
	}
}

func saveReport(report *CoverageReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	return os.WriteFile(filename, data, 0600)
}

// Helper function.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

// analyzeOpenAPISchemas analyzes OpenAPI schemas against Terraform resources.
func analyzeOpenAPISchemas(spec *OpenAPISpec, resources,
	dataSources map[string]ResourceCoverage, verbose bool) SchemaAnalysis {
	analysis := SchemaAnalysis{
		SchemaDetails: make(map[string]SchemaCoverage),
	}

	// Analyze OpenAPI schemas
	for schemaName, schemaData := range spec.Components.Schemas {
		coverage := analyzeSingleSchema(schemaName, schemaData, resources, dataSources, verbose)
		analysis.SchemaDetails[schemaName] = coverage
		analysis.TotalSchemas++

		if coverage.CoveredProperties > 0 {
			analysis.CoveredSchemas++
		}
	}

	if analysis.TotalSchemas > 0 {
		analysis.SchemaCoveragePercent = float64(analysis.CoveredSchemas) / float64(analysis.TotalSchemas) * 100
	}

	return analysis
}

// analyzeSingleSchema analyzes a single OpenAPI schema.
func analyzeSingleSchema(schemaName string, schemaData interface{}, resources,
	dataSources map[string]ResourceCoverage, verbose bool) SchemaCoverage {
	coverage := SchemaCoverage{
		SchemaName: schemaName,
		Properties: make(map[string]Property),
	}

	// Convert schema data to map
	schemaMap, ok := schemaData.(map[string]interface{})
	if !ok {
		return coverage
	}

	// Extract properties from OpenAPI schema
	if properties, exists := schemaMap["properties"]; exists {
		if propMap, ok := properties.(map[string]interface{}); ok {
			for propName, propData := range propMap {
				property := extractPropertyFromOpenAPI(propName, propData)
				coverage.Properties[propName] = property
				coverage.TotalProperties++
			}
		}
	}

	// Check coverage against Terraform resources
	coverage = checkPropertyCoverage(coverage, resources, dataSources, verbose)

	return coverage
}

// extractPropertyFromOpenAPI extracts a property from OpenAPI data.
func extractPropertyFromOpenAPI(propName string, propData interface{}) Property {
	property := Property{
		Name:    propName,
		Type:    "unknown",
		Covered: false,
	}

	if propMap, ok := propData.(map[string]interface{}); ok {
		if typeVal, exists := propMap["type"]; exists {
			if typeStr, ok := typeVal.(string); ok {
				property.Type = typeStr
			}
		}
	}

	return property
}

// extractResourceName extracts the resource name from a filename.
func extractResourceName(filename string) string {
	// Extract resource name from a filename like "resource_application.go"
	name := strings.TrimPrefix(filename, "resource_")
	name = strings.TrimSuffix(name, ".go")

	return name
}

// extractDataSourceName extracts the data source name from a filename.
func extractDataSourceName(filename string) string {
	// Extract data source name from a filename like "data_source_version.go"
	name := strings.TrimPrefix(filename, "data_source_")
	name = strings.TrimSuffix(name, ".go")

	return name
}

// extractRelatedEndpoints extracts related endpoints from file content.
func extractRelatedEndpoints(content string) []string {
	var endpoints []string

	// Regex patterns to find API endpoints in code
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`"(/[a-zA-Z0-9_/-]+)"`),
		regexp.MustCompile(`client\.newRequest\([^,]+,\s*"([^"]+)"`),
		regexp.MustCompile(`fmt\.Sprintf\("([^"]*%[^"]*)"[^)]*\)`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 {
				endpoint := match[1]
				if strings.HasPrefix(endpoint, "/") && len(endpoint) > 1 {
					// Clean and normalize endpoint
					endpoint = strings.TrimSuffix(endpoint, "/")
					if !contains(endpoints, endpoint) {
						endpoints = append(endpoints, endpoint)
					}
				}
			}
		}
	}

	return endpoints
}
