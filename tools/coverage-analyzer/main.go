package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// OpenAPISpec represents a simplified structure of an OpenAPI specification
type OpenAPISpec struct {
	Paths      map[string]map[string]interface{} `json:"paths"`
	Components struct {
		Schemas map[string]interface{} `json:"schemas"`
	} `json:"components"`
}

// CoverageReport provides a comprehensive summary of API and schema coverage
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

// Coverage details the status of a single API endpoint
type Coverage struct {
	Path    string   `json:"path"`
	Methods []string `json:"methods"`
	Covered []string `json:"covered"`
	Missing []string `json:"missing"`
}

// SchemaAnalysis provides analysis of OpenAPI schemas vs Terraform schemas
type SchemaAnalysis struct {
	TotalSchemas          int                       `json:"total_schemas"`
	CoveredSchemas        int                       `json:"covered_schemas"`
	SchemaCoveragePercent float64                   `json:"schema_coverage_percent"`
	SchemaDetails         map[string]SchemaCoverage `json:"schema_details"`
}

// SchemaCoverage details the coverage of a single schema
type SchemaCoverage struct {
	SchemaName        string              `json:"schema_name"`
	TotalProperties   int                 `json:"total_properties"`
	CoveredProperties int                 `json:"covered_properties"`
	Properties        map[string]Property `json:"properties"`
	MissingProperties []string            `json:"missing_properties"`
	ExtraProperties   []string            `json:"extra_properties"`
}

// Property represents a schema property
type Property struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Required      bool   `json:"required"`
	Covered       bool   `json:"covered"`
	TerraformType string `json:"terraform_type,omitempty"`
}

// ResourceCoverage details the coverage of a Terraform resource or data source
type ResourceCoverage struct {
	ResourceName        string              `json:"resource_name"`
	SchemaProperties    map[string]Property `json:"schema_properties"`
	SupportedOperations []string            `json:"supported_operations"`
	RelatedEndpoints    []string            `json:"related_endpoints"`
	CoverageScore       float64             `json:"coverage_score"`
}

// DebugInfo provides debugging information about the analysis
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

func main() {
	openAPIFile := flag.String("openapi", "api/openapi.json", "Path to the OpenAPI JSON file")
	providerDir := flag.String("provider", "bastion/", "Path to the provider's source code directory")
	outputFile := flag.String("output", "coverage-report.json", "Path to save the JSON coverage report")
	verbose := flag.Bool("verbose", false, "Enable verbose output for debugging")
	analyzeSchemas := flag.Bool("schemas", true, "Analyze OpenAPI schemas vs Terraform schemas")

	flag.Parse()

	if err := run(*openAPIFile, *providerDir, *outputFile, *verbose, *analyzeSchemas); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(openAPIFile, providerDir, outputFile string, verbose, analyzeSchemas bool) error {
	fmt.Printf("Analyzing OpenAPI spec from %s...\n", openAPIFile)
	spec, err := parseOpenAPI(openAPIFile)
	if err != nil {
		return fmt.Errorf("error parsing OpenAPI spec: %w", err)
	}

	fmt.Printf("Analyzing provider code in %s...\n", providerDir)

	if _, err := os.Stat(providerDir); os.IsNotExist(err) {
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
		fmt.Println("Analyzing schemas...")
		schemaAnalysis = analyzeSchemas_(spec, resourceAnalysis, dataSourceAnalysis, verbose)
	}

	if verbose {
		fmt.Printf("Debug: Analyzed %d files\n", debugInfo.FilesAnalyzed)
		fmt.Printf("Debug: Found %d API calls\n", debugInfo.TotalAPICallsFound)
		fmt.Printf("Debug: Found %d resources\n", len(debugInfo.FoundResources))
		fmt.Printf("Debug: Found %d data sources\n", len(debugInfo.FoundDataSources))
	}

	fmt.Println("Generating comprehensive coverage report...")
	report := generateComprehensiveCoverageReport(spec, providerEndpoints, schemaAnalysis, resourceAnalysis, dataSourceAnalysis, debugInfo)

	fmt.Println("\n=== Comprehensive Coverage Report ===")
	printComprehensiveReport(report)

	fmt.Printf("Saving report to %s...\n", outputFile)
	return saveReport(report, outputFile)
}

func parseOpenAPI(filename string) (*OpenAPISpec, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var spec OpenAPISpec
	if err = json.Unmarshal(data, &spec); err != nil {
		return nil, err
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

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			debugInfo.Errors = append(debugInfo.Errors, fmt.Sprintf("Walk error in %s: %v", path, err))
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		debugInfo.FilesAnalyzed++
		debugInfo.AnalyzedFiles = append(debugInfo.AnalyzedFiles, path)

		if verbose {
			fmt.Printf("Analyzing file: %s\n", path)
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			debugInfo.FilesWithErrors++
			debugInfo.Errors = append(debugInfo.Errors, fmt.Sprintf("Parse error in %s: %v", path, err))
			return nil
		}

		// Analyze file content using multiple strategies
		analyzeFileContent(node, path, endpoints, debugInfo, verbose)

		return nil
	})

	// Normalize endpoints
	normalizedEndpoints := normalizeEndpoints(endpoints, debugInfo, verbose)

	return normalizedEndpoints, debugInfo, err
}

func analyzeTerraformResources(dir string, verbose bool, debugInfo *DebugInfo) (map[string]ResourceCoverage, map[string]ResourceCoverage, error) {
	resources := make(map[string]ResourceCoverage)
	dataSources := make(map[string]ResourceCoverage)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
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

	return resources, dataSources, err
}

func analyzeResourceFile(path, resourceName string, verbose bool) ResourceCoverage {
	coverage := ResourceCoverage{
		ResourceName:        resourceName,
		SchemaProperties:    make(map[string]Property),
		SupportedOperations: []string{},
		RelatedEndpoints:    []string{},
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return coverage
	}

	fileContent := string(content)

	if verbose {
		fmt.Printf("Analyzing resource file: %s\n", path)
	}

	// Extract schema properties from the file
	coverage.SchemaProperties = extractSchemaProperties(fileContent, verbose)

	// Determine supported operations based on function presence
	operations := []string{"Create", "Read", "Update", "Delete"}
	for _, op := range operations {
		patterns := []string{
			fmt.Sprintf("resource%s", op),
			fmt.Sprintf("%s:", op),
			fmt.Sprintf("func.*%s", op),
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
		fmt.Printf("  Found %d properties, %d operations, %d endpoints\n",
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

	content, err := ioutil.ReadFile(path)
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

func checkPropertyCoverage(coverage SchemaCoverage, resources, dataSources map[string]ResourceCoverage, verbose bool) SchemaCoverage {
	// Combiner toutes les couvertures de ressources et data sources
	allCoverage := make(map[string]ResourceCoverage)
	for k, v := range resources {
		allCoverage[k] = v
	}
	for k, v := range dataSources {
		allCoverage[k] = v
	}

	if verbose {
		fmt.Printf("  Checking coverage for schema '%s' with %d properties\n", coverage.SchemaName, len(coverage.Properties))
		fmt.Printf("  Available resources/datasources: %d\n", len(allCoverage))
	}

	for propName, property := range coverage.Properties {
		found := false

		// Recherche directe par nom de propriété
		for resourceName, resourceCoverage := range allCoverage {
			if _, exists := resourceCoverage.SchemaProperties[propName]; exists {
				property.Covered = true
				found = true
				if verbose {
					fmt.Printf("    ✓ Property %s.%s found in resource %s\n",
						coverage.SchemaName, propName, resourceName)
				}
				break
			}
		}

		// Si pas trouvé, essayer des correspondances approximatives
		if !found {
			for resourceName, resourceCoverage := range allCoverage {
				for terraformProp := range resourceCoverage.SchemaProperties {
					// Correspondances possibles (snake_case vs camelCase, etc.)
					if isPropertyMatch(propName, terraformProp) {
						property.Covered = true
						found = true
						if verbose {
							fmt.Printf("    ≈ Property %s.%s matches %s in resource %s\n",
								coverage.SchemaName, propName, terraformProp, resourceName)
						}
						break
					}
				}
				if found {
					break
				}
			}
		}

		// Recherche basée sur le nom du schéma
		if !found {
			schemaBaseName := extractSchemaBaseName(coverage.SchemaName)
			for resourceName, resourceCoverage := range allCoverage {
				if strings.Contains(resourceName, schemaBaseName) || strings.Contains(schemaBaseName, resourceName) {
					if _, exists := resourceCoverage.SchemaProperties[propName]; exists {
						property.Covered = true
						found = true
						if verbose {
							fmt.Printf("    ~ Property %s.%s found via schema match in resource %s\n",
								coverage.SchemaName, propName, resourceName)
						}
						break
					}
				}
			}
		}

		coverage.Properties[propName] = property

		if found {
			coverage.CoveredProperties++
		} else {
			coverage.MissingProperties = append(coverage.MissingProperties, propName)
			if verbose {
				fmt.Printf("    ✗ Property %s.%s not found in any resource\n",
					coverage.SchemaName, propName)
			}
		}
	}

	return coverage
}

// isPropertyMatch vérifie si deux noms de propriétés correspondent
func isPropertyMatch(openAPIProp, terraformProp string) bool {
	// Normalisation des noms
	openAPILower := strings.ToLower(openAPIProp)
	terraformLower := strings.ToLower(terraformProp)

	// Correspondance exacte
	if openAPILower == terraformLower {
		return true
	}

	// Conversion snake_case <-> camelCase
	openAPISnake := camelToSnake(openAPIProp)
	terraformSnake := camelToSnake(terraformProp)

	if strings.ToLower(openAPISnake) == strings.ToLower(terraformSnake) {
		return true
	}

	// Correspondances communes
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

// camelToSnake convertit camelCase en snake_case
func camelToSnake(str string) string {
	snake := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(str, "${1}_${2}")
	snake = regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// extractSchemaBaseName extrait le nom de base d'un schéma
func extractSchemaBaseName(schemaName string) string {
	// Supprimer les suffixes communs
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

// extractSchemaProperties amélioré pour une meilleure extraction
func extractSchemaProperties(content string, verbose bool) map[string]Property {
	properties := make(map[string]Property)

	// Patterns regex améliorés
	schemaPattern := regexp.MustCompile(`"([a-zA-Z_][a-zA-Z0-9_]*)"\s*:\s*&?schema\.Schema\s*{([^}]+)}`)
	typePattern := regexp.MustCompile(`Type:\s*schema\.Type([A-Za-z]+)`)
	requiredPattern := regexp.MustCompile(`Required:\s*(true|false)`)
	optionalPattern := regexp.MustCompile(`Optional:\s*(true|false)`)
	computedPattern := regexp.MustCompile(`Computed:\s*(true|false)`)

	// Recherche des définitions de schéma avec contenu
	matches := schemaPattern.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		if len(match) > 2 {
			propertyName := match[1]
			schemaBlock := match[2]

			property := Property{
				Name:     propertyName,
				Type:     "unknown",
				Required: false,
				Covered:  true, // La propriété existe dans Terraform
			}

			// Extraction du type
			if typeMatch := typePattern.FindStringSubmatch(schemaBlock); len(typeMatch) > 1 {
				property.Type = strings.ToLower(typeMatch[1])
				property.TerraformType = "Type" + typeMatch[1]
			}

			// Extraction required/optional
			if requiredMatch := requiredPattern.FindStringSubmatch(schemaBlock); len(requiredMatch) > 1 {
				property.Required = requiredMatch[1] == "true"
			} else if optionalMatch := optionalPattern.FindStringSubmatch(schemaBlock); len(optionalMatch) > 1 {
				property.Required = optionalMatch[1] != "true"
			}

			// Si c'est computed, généralement pas required
			if computedMatch := computedPattern.FindStringSubmatch(schemaBlock); len(computedMatch) > 1 && computedMatch[1] == "true" {
				property.Required = false
			}

			properties[propertyName] = property

			if verbose {
				fmt.Printf("    Found property: %s (type: %s, required: %t)\n",
					propertyName, property.Type, property.Required)
			}
		}
	}

	// Recherche alternative avec pattern plus simple
	if len(properties) == 0 {
		simplePattern := regexp.MustCompile(`"([a-zA-Z_][a-zA-Z0-9_]*)"\s*:`)
		simpleMatches := simplePattern.FindAllStringSubmatch(content, -1)

		for _, match := range simpleMatches {
			if len(match) > 1 {
				propertyName := match[1]
				// Éviter les mots-clés Go et autres patterns non-propriétés
				if !isGoKeyword(propertyName) && !strings.HasPrefix(propertyName, "Test") {
					properties[propertyName] = Property{
						Name:     propertyName,
						Type:     "unknown",
						Required: false,
						Covered:  true,
					}

					if verbose {
						fmt.Printf("    Found simple property: %s\n", propertyName)
					}
				}
			}
		}
	}

	return properties
}

// isGoKeyword vérifie si c'est un mot-clé Go à éviter
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

// Existing functions (analyzeFileContent, extractAPICall, normalizeEndpoints, etc.) remain the same...
// [Rest of the previous functions would go here - keeping them unchanged]

func analyzeFileContent(node *ast.File, filepath string, endpoints map[string][]string, debugInfo *DebugInfo, verbose bool) {
	// Strategy 1: AST analysis for function calls
	ast.Inspect(node, func(n ast.Node) bool {
		if callExpr, ok := n.(*ast.CallExpr); ok {
			method, url := extractAPICall(callExpr, verbose)
			if method != "" && url != "" {
				debugInfo.TotalAPICallsFound++
				if verbose {
					fmt.Printf("  AST Found: %s %s\n", method, url)
				}
				endpoints[url] = append(endpoints[url], method)
				debugInfo.FoundEndpoints[url] = append(debugInfo.FoundEndpoints[url], method)
			}
		}
		return true
	})

	// Strategy 2: String analysis for URL patterns
	analyzeStringLiterals(node, filepath, endpoints, debugInfo, verbose)
}

func analyzeStringLiterals(node *ast.File, filepath string, endpoints map[string][]string, debugInfo *DebugInfo, verbose bool) {
	// Regex patterns pour identifier les endpoints d'API
	apiPatterns := []*regexp.Regexp{
		regexp.MustCompile(`"(/[a-zA-Z0-9_/-]*[a-zA-Z0-9_])"`),               // "/path/to/endpoint"
		regexp.MustCompile(`"/api/v\d+\.\d+(/[a-zA-Z0-9_/-]*[a-zA-Z0-9_])"`), // "/api/v3.12/endpoint"
		regexp.MustCompile(`"(/config/[a-zA-Z0-9_/-]*)"`),                    // "/config/something"
		regexp.MustCompile(`"(/auth[a-zA-Z0-9_/-]*)"`),                       // "/auth..." patterns
		regexp.MustCompile(`"(/users?[a-zA-Z0-9_/-]*)"`),                     // "/users" patterns
		regexp.MustCompile(`"(/devices?[a-zA-Z0-9_/-]*)"`),                   // "/devices" patterns
		regexp.MustCompile(`"(/domains?[a-zA-Z0-9_/-]*)"`),                   // "/domains" patterns
		regexp.MustCompile(`"(/applications?[a-zA-Z0-9_/-]*)"`),              // "/applications" patterns
	}

	methodPatterns := map[*regexp.Regexp]string{
		regexp.MustCompile(`http\.MethodGet`):    "GET",
		regexp.MustCompile(`http\.MethodPost`):   "POST",
		regexp.MustCompile(`http\.MethodPut`):    "PUT",
		regexp.MustCompile(`http\.MethodDelete`): "DELETE",
		regexp.MustCompile(`http\.MethodPatch`):  "PATCH",
		regexp.MustCompile(`"GET"`):              "GET",
		regexp.MustCompile(`"POST"`):             "POST",
		regexp.MustCompile(`"PUT"`):              "PUT",
		regexp.MustCompile(`"DELETE"`):           "DELETE",
		regexp.MustCompile(`"PATCH"`):            "PATCH",
	}

	// Lire le contenu du fichier pour l'analyse regex
	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		return
	}

	fileContent := string(content)

	// Chercher les URLs
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

	// Chercher les méthodes HTTP dans le contexte
	var foundMethods []string
	for pattern, method := range methodPatterns {
		if pattern.MatchString(fileContent) {
			foundMethods = append(foundMethods, method)
		}
	}

	// Associer les URLs trouvées avec des méthodes
	if len(foundURLs) > 0 {
		for _, url := range foundURLs {
			if len(foundMethods) > 0 {
				for _, method := range foundMethods {
					debugInfo.TotalAPICallsFound++
					if verbose {
						fmt.Printf("  String Found: %s %s\n", method, url)
					}
					endpoints[url] = append(endpoints[url], method)
					debugInfo.FoundEndpoints[url] = append(debugInfo.FoundEndpoints[url], method)
				}
			} else {
				// Si aucune méthode trouvée, essayer de deviner basé sur le contexte
				method := guessHTTPMethod(fileContent, url)
				debugInfo.TotalAPICallsFound++
				if verbose {
					fmt.Printf("  String Found (guessed): %s %s\n", method, url)
				}
				endpoints[url] = append(endpoints[url], method)
				debugInfo.FoundEndpoints[url] = append(debugInfo.FoundEndpoints[url], method)
			}
		}
	}
}

func guessHTTPMethod(content string, url string) string {
	lowerContent := strings.ToLower(content)

	// Patterns pour deviner la méthode HTTP basée sur le contexte
	if strings.Contains(lowerContent, "create") || strings.Contains(lowerContent, "add") {
		return "POST"
	}
	if strings.Contains(lowerContent, "update") || strings.Contains(lowerContent, "modify") {
		return "PUT"
	}
	if strings.Contains(lowerContent, "delete") || strings.Contains(lowerContent, "remove") {
		return "DELETE"
	}
	if strings.Contains(lowerContent, "read") || strings.Contains(lowerContent, "get") || strings.Contains(lowerContent, "fetch") {
		return "GET"
	}

	// Par défaut, supposer GET pour les data sources et POST pour les resources
	if strings.Contains(lowerContent, "data_source") {
		return "GET"
	}
	if strings.Contains(lowerContent, "resource") {
		return "POST"
	}

	return "GET" // Par défaut
}

func extractAPICall(callExpr *ast.CallExpr, verbose bool) (method, url string) {
	// Pattern 1: client.newRequest ou client.doRequest
	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
		if selExpr.Sel.Name == "newRequest" || selExpr.Sel.Name == "doRequest" {
			if len(callExpr.Args) >= 3 {
				// URL argument
				if urlLit, ok := callExpr.Args[1].(*ast.BasicLit); ok {
					url = strings.Trim(urlLit.Value, `"`)
				}

				// Method argument
				if methodExpr, ok := callExpr.Args[2].(*ast.SelectorExpr); ok {
					if methodExpr.Sel != nil {
						methodName := methodExpr.Sel.Name
						switch methodName {
						case "MethodGet":
							method = "GET"
						case "MethodPost":
							method = "POST"
						case "MethodPut":
							method = "PUT"
						case "MethodDelete":
							method = "DELETE"
						case "MethodPatch":
							method = "PATCH"
						default:
							method = methodName
						}
					}
				} else if methodLit, ok := callExpr.Args[2].(*ast.BasicLit); ok {
					method = strings.Trim(methodLit.Value, `"`)
				}
			}
		}
	}

	// Pattern 2: http.NewRequest
	if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok && selExpr.Sel.Name == "NewRequest" {
		if len(callExpr.Args) >= 2 {
			// Method is first argument
			if methodLit, ok := callExpr.Args[0].(*ast.BasicLit); ok {
				method = strings.Trim(methodLit.Value, `"`)
			}
			// URL is second argument
			if urlLit, ok := callExpr.Args[1].(*ast.BasicLit); ok {
				url = strings.Trim(urlLit.Value, `"`)
			}
		}
	}

	return method, url
}

func normalizeEndpoints(endpoints map[string][]string, debugInfo *DebugInfo, verbose bool) map[string][]string {
	normalized := make(map[string][]string)

	for rawURL, methods := range endpoints {
		// Nettoyer l'URL
		cleanURL := strings.TrimSpace(rawURL)
		cleanURL = strings.TrimSuffix(cleanURL, "/")

		// Ignorer les URLs trop courtes ou invalides
		if len(cleanURL) < 2 || !strings.HasPrefix(cleanURL, "/") {
			continue
		}

		// Supprimer les préfixes d'API si présents
		cleanURL = regexp.MustCompile(`^/api/v\d+\.\d+`).ReplaceAllString(cleanURL, "")
		if cleanURL == "" {
			cleanURL = "/"
		}

		// Dédupliquer les méthodes
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
				fmt.Printf("  Normalized: %s -> %s with methods %v\n", rawURL, cleanURL, result)
			}
		}
	}

	return normalized
}

func generateComprehensiveCoverageReport(spec *OpenAPISpec, providerEndpoints map[string][]string, schemaAnalysis SchemaAnalysis, resourceAnalysis, dataSourceAnalysis map[string]ResourceCoverage, debugInfo *DebugInfo) *CoverageReport {
	report := &CoverageReport{
		EndpointDetails:    make(map[string]Coverage),
		SchemaAnalysis:     schemaAnalysis,
		ResourceAnalysis:   resourceAnalysis,
		DataSourceAnalysis: dataSourceAnalysis,
		DebugInfo:          *debugInfo,
	}

	for path, methods := range spec.Paths {
		var availableMethods []string
		for method := range methods {
			if method != "parameters" && method != "x-amazon-apigateway-integration" {
				availableMethods = append(availableMethods, strings.ToUpper(method))
			}
		}

		if len(availableMethods) == 0 {
			continue
		}

		coverage := Coverage{
			Path:    path,
			Methods: availableMethods,
			Covered: []string{},
			Missing: []string{},
		}

		// Vérification exacte ET approximative
		if providerMethods, exists := providerEndpoints[path]; exists {
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
		} else {
			// Vérification approximative (sans paramètres)
			basePath := regexp.MustCompile(`\{[^}]+\}`).ReplaceAllString(path, "")
			basePath = strings.TrimSuffix(basePath, "/")

			for endpoint, providerMethods := range providerEndpoints {
				if strings.HasPrefix(endpoint, basePath) || strings.HasPrefix(basePath, endpoint) {
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
					break
				}
			}

			// Si toujours rien trouvé, tout est manquant
			if len(coverage.Covered) == 0 {
				coverage.Missing = availableMethods
			}
		}

		report.EndpointDetails[path] = coverage
		report.TotalEndpoints++
		if len(coverage.Covered) > 0 {
			report.CoveredEndpoints++
		}
	}

	if report.TotalEndpoints > 0 {
		report.CoveragePercent = float64(report.CoveredEndpoints) / float64(report.TotalEndpoints) * 100
	}

	return report
}

func printComprehensiveReport(report *CoverageReport) {
	fmt.Printf("=== ENDPOINT COVERAGE ===\n")
	fmt.Printf("Total Endpoints: %d\n", report.TotalEndpoints)
	fmt.Printf("Covered Endpoints: %d\n", report.CoveredEndpoints)
	fmt.Printf("Coverage Percentage: %.2f%%\n", report.CoveragePercent)

	fmt.Printf("\n=== SCHEMA COVERAGE ===\n")
	fmt.Printf("Total Schemas: %d\n", report.SchemaAnalysis.TotalSchemas)
	fmt.Printf("Covered Schemas: %d\n", report.SchemaAnalysis.CoveredSchemas)
	fmt.Printf("Schema Coverage Percentage: %.2f%%\n", report.SchemaAnalysis.SchemaCoveragePercent)

	fmt.Printf("\n=== RESOURCE ANALYSIS ===\n")
	fmt.Printf("Total Resources: %d\n", len(report.ResourceAnalysis))
	resourceNames := make([]string, 0, len(report.ResourceAnalysis))
	for name := range report.ResourceAnalysis {
		resourceNames = append(resourceNames, name)
	}
	sort.Strings(resourceNames)

	for _, name := range resourceNames {
		resource := report.ResourceAnalysis[name]
		fmt.Printf("- %s: %d properties, %v operations\n",
			name, len(resource.SchemaProperties), resource.SupportedOperations)
	}

	fmt.Printf("\n=== DATA SOURCE ANALYSIS ===\n")
	fmt.Printf("Total Data Sources: %d\n", len(report.DataSourceAnalysis))
	dataSourceNames := make([]string, 0, len(report.DataSourceAnalysis))
	for name := range report.DataSourceAnalysis {
		dataSourceNames = append(dataSourceNames, name)
	}
	sort.Strings(dataSourceNames)

	for _, name := range dataSourceNames {
		dataSource := report.DataSourceAnalysis[name]
		fmt.Printf("- %s: %d properties\n", name, len(dataSource.SchemaProperties))
	}

	fmt.Printf("\n=== TOP MISSING ENDPOINTS ===\n")
	missingCount := 0
	for path, coverage := range report.EndpointDetails {
		if len(coverage.Missing) > 0 && missingCount < 10 {
			fmt.Printf("- %s (%s)\n", path, strings.Join(coverage.Missing, ", "))
			missingCount++
		}
	}

	if len(report.SchemaAnalysis.SchemaDetails) > 0 {
		fmt.Printf("\n=== SCHEMA DETAILS (Top 5) ===\n")
		count := 0
		for schemaName, schema := range report.SchemaAnalysis.SchemaDetails {
			if count >= 5 {
				break
			}
			fmt.Printf("Schema: %s\n", schemaName)
			fmt.Printf("  Total Properties: %d\n", schema.TotalProperties)
			fmt.Printf("  Covered Properties: %d\n", schema.CoveredProperties)
			fmt.Printf("  Missing Properties: %v\n", schema.MissingProperties)
			fmt.Println()
			count++
		}
	}
}

func saveReport(report *CoverageReport, filename string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Ajouter ces fonctions manquantes à la fin du fichier

// analyzeSchemas_ analyse les schémas OpenAPI par rapport aux ressources Terraform
func analyzeSchemas_(spec *OpenAPISpec, resources, dataSources map[string]ResourceCoverage, verbose bool) SchemaAnalysis {
	analysis := SchemaAnalysis{
		SchemaDetails: make(map[string]SchemaCoverage),
	}

	// Analyser les schémas OpenAPI
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

// analyzeSingleSchema analyse un seul schéma OpenAPI
func analyzeSingleSchema(schemaName string, schemaData interface{}, resources, dataSources map[string]ResourceCoverage, verbose bool) SchemaCoverage {
	coverage := SchemaCoverage{
		SchemaName: schemaName,
		Properties: make(map[string]Property),
	}

	// Convertir les données du schéma en map
	schemaMap, ok := schemaData.(map[string]interface{})
	if !ok {
		return coverage
	}

	// Extraire les propriétés du schéma OpenAPI
	if properties, exists := schemaMap["properties"]; exists {
		if propMap, ok := properties.(map[string]interface{}); ok {
			for propName, propData := range propMap {
				property := extractPropertyFromOpenAPI(propName, propData)
				coverage.Properties[propName] = property
				coverage.TotalProperties++
			}
		}
	}

	// Vérifier la couverture par rapport aux ressources Terraform
	coverage = checkPropertyCoverage(coverage, resources, dataSources, verbose)

	return coverage
}

// extractPropertyFromOpenAPI extrait une propriété depuis les données OpenAPI
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

// extractResourceName extrait le nom d'une ressource à partir du nom de fichier
func extractResourceName(filename string) string {
	// Extraire le nom de ressource depuis un nom de fichier comme "resource_application.go"
	name := strings.TrimPrefix(filename, "resource_")
	name = strings.TrimSuffix(name, ".go")
	return name
}

// extractDataSourceName extrait le nom d'une source de données à partir du nom de fichier
func extractDataSourceName(filename string) string {
	// Extraire le nom de source de données depuis un nom de fichier comme "data_source_version.go"
	name := strings.TrimPrefix(filename, "data_source_")
	name = strings.TrimSuffix(name, ".go")
	return name
}

// extractRelatedEndpoints extrait les endpoints liés depuis le contenu d'un fichier
func extractRelatedEndpoints(content string) []string {
	var endpoints []string

	// Patterns regex pour trouver les endpoints d'API dans le code
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
					// Nettoyer et normaliser l'endpoint
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
