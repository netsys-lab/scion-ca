// Copyright 2021 Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/scionproto/scion/pkg/log"
	"github.com/scionproto/scion/pkg/private/serrors"
	cppb "github.com/scionproto/scion/pkg/proto/control_plane"
	"github.com/scionproto/scion/pkg/scrypto/cppki"
	"github.com/scionproto/scion/private/ca/api"
)

// DelegatingHandler delegates requests to the CA service.
type DelegatingHandler struct {
}

// HandleCMSRequest handles a certificate renewal request that was signed with
// CMS by delegating it to the CA Service.
func (h *DelegatingHandler) HandleCMSRequest(
	ctx context.Context,
	req *cppb.ChainRenewalRequest,
) ([]*x509.Certificate, error) {

	logger := log.FromCtx(ctx)

	chain, err := extractChain(req.CmsSignedRequest)
	if err != nil {
		logger.Info("Failed to extract client certificate", "err", err)

		return nil, status.Error(
			codes.InvalidArgument,
			"malformed request: cannot extract client certificate chain",
		)
	}
	_, err = cppki.ExtractIA(chain[0].Subject)
	if err != nil {
		logger.Info("Failed to extract IA from AS certificate",
			"err", err,
			"subject", chain[0].Subject,
		)

		return nil, status.Error(
			codes.InvalidArgument,
			"malformed request: cannot extract ISD-AS from subject",
		)
	}

	return nil, nil
}

func (h *DelegatingHandler) parseChain(rep api.RenewalResponse) ([]*x509.Certificate, error) {
	switch content := rep.CertificateChain.(type) {
	case string:
		raw, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			return nil, serrors.WrapStr("malformed certificate_chain", err)
		}
		return extractChain(raw)
	case map[string]interface{}:
		decode := func(key string) ([]byte, error) {
			b64, ok := content[key]
			if !ok {
				return nil, serrors.New("certificate missing")
			}
			s, ok := b64.(string)
			if !ok {
				return nil, serrors.New("wrong type", "type", fmt.Sprintf("%T", s))
			}
			return base64.StdEncoding.DecodeString(s)
		}
		as, err := decode("as_certificate")
		if err != nil {
			return nil, serrors.WrapStr("parsing AS certificate", err, "key", "as_certificate")
		}
		ca, err := decode("ca_certificate")
		if err != nil {
			return nil, serrors.WrapStr("parsing AS certificate", err, "key", "ca_certificate")
		}
		return h.parseChainJSON(api.CertificateChain{
			AsCertificate: as,
			CaCertificate: ca,
		})
	default:
		return nil, serrors.New("certificate_chain unset", "type", fmt.Sprintf("%T", content))
	}
}

func (h *DelegatingHandler) parseChainJSON(rep api.CertificateChain) ([]*x509.Certificate, error) {
	as, err := x509.ParseCertificate(rep.AsCertificate)
	if err != nil {
		return nil, serrors.WrapStr("parsing AS certificate", err)
	}
	ca, err := x509.ParseCertificate(rep.CaCertificate)
	if err != nil {
		return nil, serrors.WrapStr("parsing CA certificate", err)
	}
	chain := []*x509.Certificate{as, ca}
	if err := cppki.ValidateChain(chain); err != nil {
		return nil, serrors.WrapStr("validating certificate chain", err)
	}
	return chain, nil
}
