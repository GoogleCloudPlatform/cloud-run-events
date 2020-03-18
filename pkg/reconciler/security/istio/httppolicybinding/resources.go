/*
Copyright 2020 Google LLC

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

package httppolicybinding

import (
	"fmt"

	istiosecurity "istio.io/api/security/v1beta1"
	istiotype "istio.io/api/type/v1beta1"
	istioclient "istio.io/client-go/pkg/apis/security/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"

	"github.com/google/knative-gcp/pkg/apis/security/v1alpha1"
)

// See: https://istio.io/docs/reference/config/security/conditions/
const (
	istioClaimKeyPattern  = "request.auth.claims[%s]"
	istioHeaderKeyPattern = "request.headers[%s]"
)

// MakeRequestAuthentication makes an Istio RequestAuthentication.
// Reference: https://istio.io/docs/reference/config/security/request_authentication/
func MakeRequestAuthentication(
	b *v1alpha1.HTTPPolicyBinding,
	subjectSelector *metav1.LabelSelector,
	jwt v1alpha1.JWTSpec) istioclient.RequestAuthentication {

	var rhs []*istiosecurity.JWTHeader
	for _, rh := range jwt.FromHeaders {
		rhs = append(rhs, &istiosecurity.JWTHeader{Name: rh.Name, Prefix: rh.Prefix})
	}

	return istioclient.RequestAuthentication{
		ObjectMeta: metav1.ObjectMeta{
			Name:            b.Name,
			Namespace:       b.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(b)},
		},
		Spec: istiosecurity.RequestAuthentication{
			Selector: &istiotype.WorkloadSelector{
				MatchLabels: subjectSelector.MatchLabels,
			},
			JwtRules: []*istiosecurity.JWTRule{{
				Jwks:                 jwt.Jwks,
				JwksUri:              jwt.JwksURI,
				ForwardOriginalToken: true,
				FromHeaders:          rhs,
			}},
		},
	}
}

// MakeAuthorizationPolicy makes an Istio AuthorizationPolicy.
// Reference: https://istio.io/docs/reference/config/security/authorization-policy/
func MakeAuthorizationPolicy(
	b *v1alpha1.HTTPPolicyBinding,
	subjectSelector *metav1.LabelSelector,
	rules []v1alpha1.HTTPPolicyRuleSpec) istioclient.AuthorizationPolicy {

	var rs []*istiosecurity.Rule
	for _, r := range rules {
		var rfs []*istiosecurity.Rule_From
		if len(r.Principals) > 0 {
			rfs = []*istiosecurity.Rule_From{{
				Source: &istiosecurity.Source{
					RequestPrincipals: r.Principals,
				},
			}}
		}

		var rts []*istiosecurity.Rule_To
		for _, op := range r.Operations {
			rt := &istiosecurity.Rule_To{
				Operation: &istiosecurity.Operation{},
			}
			for _, h := range op.Hosts {
				rt.Operation.Hosts = append(rt.Operation.Hosts, h.ToExpression())
			}
			for _, p := range op.Paths {
				rt.Operation.Paths = append(rt.Operation.Paths, p.ToExpression())
			}
			for _, m := range op.Methods {
				rt.Operation.Methods = append(rt.Operation.Methods, m)
			}
			rts = append(rts, rt)
		}

		var rcs []*istiosecurity.Condition
		for _, cl := range r.Claims {
			rc := &istiosecurity.Condition{
				Key: istioClaimKey(cl.Key),
			}
			for _, v := range cl.Values {
				rc.Values = append(rc.Values, v.ToExpression())
			}
			rcs = append(rcs, rc)
		}
		for _, h := range r.Headers {
			rc := &istiosecurity.Condition{
				Key: istioHeaderKey(h.Key),
			}
			for _, v := range h.Values {
				rc.Values = append(rc.Values, v.ToExpression())
			}
			rcs = append(rcs, rc)
		}

		rs = append(rs, &istiosecurity.Rule{
			From: rfs,
			To:   rts,
			When: rcs,
		})
	}

	return istioclient.AuthorizationPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:            b.Name,
			Namespace:       b.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(b)},
		},
		Spec: istiosecurity.AuthorizationPolicy{
			Selector: &istiotype.WorkloadSelector{
				MatchLabels: subjectSelector.MatchLabels,
			},
			Action: istiosecurity.AuthorizationPolicy_ALLOW,
			Rules:  rs,
		},
	}
}

func istioClaimKey(cl string) string {
	return fmt.Sprintf(istioClaimKeyPattern, cl)
}

func istioHeaderKey(h string) string {
	return fmt.Sprintf(istioHeaderKeyPattern, h)
}
