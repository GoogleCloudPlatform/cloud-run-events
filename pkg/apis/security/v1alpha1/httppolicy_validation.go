/*
Copyright 2020 Google LLC.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"context"
	"net/http"

	"knative.dev/pkg/apis"
)

var validHTTPMethods = map[string]bool{
	http.MethodConnect: true,
	http.MethodDelete:  true,
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodPatch:   true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodTrace:   true,
}

// Validate validates a HTTPPolicy.
func (p *HTTPPolicy) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError
	if p.Spec.JWT != nil {
		if err := p.Spec.JWT.Validate(ctx); err != nil {
			errs = errs.Also(err.ViaField("jwt").ViaField("spec"))
		}
	}
	for i, r := range p.Spec.Rules {
		if err := r.Validate(ctx); err != nil {
			errs = errs.Also(err.ViaFieldIndex("rules", i).ViaField("spec"))
		}
	}
	return errs
}

// Validate validates a HTTPPolicyRuleSpec.
func (r *HTTPPolicyRuleSpec) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError
	if err := r.JWTRule.Validate(ctx); err != nil {
		errs = errs.Also(err)
	}
	for i, h := range r.Headers {
		if err := h.Validate(ctx); err != nil {
			errs = errs.Also(err.ViaFieldIndex("headers", i))
		}
	}
	for i, op := range r.Operations {
		if err := ValidateStringMatches(ctx, op.Hosts, "hosts"); err != nil {
			errs = errs.Also(err.ViaFieldIndex("operations", i))
		}
		if err := ValidateStringMatches(ctx, op.Paths, "paths"); err != nil {
			errs = errs.Also(err.ViaFieldIndex("operations", i))
		}
		for j, m := range op.Methods {
			if !validHTTPMethods[m] {
				errs = errs.Also(apis.ErrInvalidArrayValue(m, "methods", j))
			}
		}
	}
	return errs
}
