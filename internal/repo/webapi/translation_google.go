package webapi

import (
	"fmt"

	translator "github.com/Conight/go-googletrans"

	"go-boilerplate/internal/entity"
	"go-boilerplate/pkg/resilience"
)

// TranslationWebAPI implements translation using Google Translate.
type TranslationWebAPI struct {
	conf translator.Config
	cb   *resilience.CircuitBreaker
}

// New creates a new TranslationWebAPI with circuit breaker protection.
func New() *TranslationWebAPI {
	conf := translator.Config{
		UserAgent:   []string{"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:15.0) Gecko/20100101 Firefox/15.0.1"},
		ServiceUrls: []string{"translate.google.com"},
	}

	return &TranslationWebAPI{
		conf: conf,
		cb:   resilience.New(resilience.DefaultConfig("google-translate")),
	}
}

// NewWithCircuitBreaker creates a new TranslationWebAPI with custom circuit breaker config.
func NewWithCircuitBreaker(cbConfig resilience.Config) *TranslationWebAPI {
	conf := translator.Config{
		UserAgent:   []string{"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:15.0) Gecko/20100101 Firefox/15.0.1"},
		ServiceUrls: []string{"translate.google.com"},
	}

	return &TranslationWebAPI{
		conf: conf,
		cb:   resilience.New(cbConfig),
	}
}

// Translate translates text using Google Translate with circuit breaker protection.
func (t *TranslationWebAPI) Translate(translation *entity.Translation) (*entity.Translation, error) {
	result, err := t.cb.Execute(func() (any, error) {
		trans := translator.New(t.conf)
		return trans.Translate(translation.Original, translation.Source, translation.Destination)
	})
	if err != nil {
		return nil, fmt.Errorf("TranslationWebAPI - Translate: %w (circuit: %s)", err, t.cb.State())
	}

	translated := result.(*translator.Translated)
	translation.Translation = translated.Text

	return translation, nil
}

// CircuitState returns the current state of the circuit breaker.
func (t *TranslationWebAPI) CircuitState() resilience.State {
	return t.cb.State()
}
