package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// ValidationError represents a validation error with field information
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("field '%s': %s", e.Field, e.Message)
}

// validateRequiredFields checks for required fields in the configuration
func validateRequiredFields(config *Config) error {
	var errs []error

	// Validate LLM providers
	for i, provider := range config.LLMProviders {
		if strings.TrimSpace(provider.Name) == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("llm_providers[%d].name", i), Message: "name is required"})
		}
		if strings.TrimSpace(provider.Model) == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("llm_providers[%d].model", i), Message: "model is required"})
		}
	}

	// Validate embedding providers
	for i, provider := range config.EmbeddingProviders {
		if strings.TrimSpace(provider.Name) == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("embedding_providers[%d].name", i), Message: "name is required"})
		}
		if strings.TrimSpace(provider.Model) == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("embedding_providers[%d].model", i), Message: "model is required"})
		}
	}

	// Validate retrievers
	for i, retriever := range config.Retrievers {
		if strings.TrimSpace(retriever.Name) == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("retrievers[%d].name", i), Message: "name is required"})
		}
		// Specific validations based on retriever type
		switch retriever.Name {
		case "sqlite":
			if strings.TrimSpace(retriever.DBPath) == "" {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("retrievers[%d].db_path", i), Message: "db_path is required for sqlite retriever"})
			}
		case "qdrant":
			if strings.TrimSpace(retriever.Collection) == "" {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("retrievers[%d].collection", i), Message: "collection is required for qdrant retriever"})
			}
		}
	}

	// Validate default fields
	if strings.TrimSpace(config.DefaultLLM) == "" {
		errs = append(errs, ValidationError{Field: "default_llm", Message: "default_llm is required"})
	}
	if strings.TrimSpace(config.DefaultEmbedding) == "" {
		errs = append(errs, ValidationError{Field: "default_embedding", Message: "default_embedding is required"})
	}
	if strings.TrimSpace(config.DefaultRetriever) == "" {
		errs = append(errs, ValidationError{Field: "default_retriever", Message: "default_retriever is required"})
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateCrossReferences ensures default providers exist in their respective lists
func validateCrossReferences(config *Config) error {
	var errs []error

	// Check default LLM
	llmNames := make(map[string]bool)
	for _, provider := range config.LLMProviders {
		llmNames[provider.Name] = true
	}
	if !llmNames[config.DefaultLLM] {
		errs = append(errs, ValidationError{Field: "default_llm", Message: fmt.Sprintf("default LLM provider '%s' not found in llm_providers list", config.DefaultLLM)})
	}

	// Check default embedding
	embeddingNames := make(map[string]bool)
	for _, provider := range config.EmbeddingProviders {
		embeddingNames[provider.Name] = true
	}
	if !embeddingNames[config.DefaultEmbedding] {
		errs = append(errs, ValidationError{Field: "default_embedding", Message: fmt.Sprintf("default embedding provider '%s' not found in embedding_providers list", config.DefaultEmbedding)})
	}

	// Check default retriever
	retrieverNames := make(map[string]bool)
	for _, retriever := range config.Retrievers {
		retrieverNames[retriever.Name] = true
	}
	if !retrieverNames[config.DefaultRetriever] {
		errs = append(errs, ValidationError{Field: "default_retriever", Message: fmt.Sprintf("default retriever '%s' not found in retrievers list", config.DefaultRetriever)})
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// validateEnvironmentVariables checks for required environment variables based on configured providers
func validateEnvironmentVariables(config *Config) error {
	var errs []error

	// Check for OPENAI_API_KEY if openai is used in LLM or embedding providers
	openaiUsed := false
	for _, provider := range config.LLMProviders {
		if provider.Name == "openai" {
			openaiUsed = true
			break
		}
	}
	if !openaiUsed {
		for _, provider := range config.EmbeddingProviders {
			if provider.Name == "openai" {
				openaiUsed = true
				break
			}
		}
	}
	if openaiUsed {
		if os.Getenv("OPENAI_API_KEY") == "" {
			errs = append(errs, ValidationError{Field: "environment.OPENAI_API_KEY", Message: "OPENAI_API_KEY environment variable is required when using openai provider"})
		}
	}

	// Check for QDRANT_URL if qdrant retriever is configured
	qdrantUsed := false
	for _, retriever := range config.Retrievers {
		if retriever.Name == "qdrant" {
			qdrantUsed = true
			break
		}
	}
	if qdrantUsed {
		if os.Getenv("QDRANT_URL") == "" {
			errs = append(errs, ValidationError{Field: "environment.QDRANT_URL", Message: "QDRANT_URL environment variable is required when using qdrant retriever"})
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// ValidateConfig performs comprehensive validation of the configuration
func ValidateConfig(config *Config) error {
	var allErrs []error

	if err := validateRequiredFields(config); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validateCrossReferences(config); err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validateEnvironmentVariables(config); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) > 0 {
		return errors.Join(allErrs...)
	}
	return nil
}
