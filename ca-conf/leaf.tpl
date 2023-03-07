{
    "subject": {
    "country": {{ toJson .Subject.Country }},
    "organization": {{ toJson .Subject.Organization }},
    "commonName": {{toJson .Subject.CommonName }},
    "extraNames": [{"type": "1.3.6.1.4.1.55324.1.2.1", "value": {{ toJson .Insecure.User.isdAS }} }]
  },
    "sans": {{ toJson .SANs }},
{{- if typeIs "*rsa.PublicKey" .Insecure.CR.PublicKey }}
    "keyUsage": ["keyEncipherment", "digitalSignature"],
{{- else }}
    "keyUsage": ["digitalSignature"],
{{- end }}
    "extKeyUsage": ["serverAuth", "clientAuth", "timestamping"]
}
