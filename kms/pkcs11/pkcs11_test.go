// +build cgo

package pkcs11

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"math/big"
	"reflect"
	"testing"

	"github.com/smallstep/certificates/kms/apiv1"
	"golang.org/x/crypto/cryptobyte"
	"golang.org/x/crypto/cryptobyte/asn1"
)

func TestNew(t *testing.T) {
	type args struct {
		ctx  context.Context
		opts apiv1.Options
	}
	tests := []struct {
		name    string
		args    args
		want    *PKCS11
		wantErr bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.ctx, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPKCS11_GetPublicKey(t *testing.T) {
	k := setupPKCS11(t)
	type args struct {
		req *apiv1.GetPublicKeyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    crypto.PublicKey
		wantErr bool
	}{
		{"RSA", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:id=7371;object=rsa-key",
		}}, &rsa.PublicKey{}, false},
		{"RSA by id", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:id=7371",
		}}, &rsa.PublicKey{}, false},
		{"RSA by label", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:object=rsa-key",
		}}, &rsa.PublicKey{}, false},
		{"ECDSA", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:id=7373;object=ecdsa-p256-key",
		}}, &ecdsa.PublicKey{}, false},
		{"ECDSA by id", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:id=7373",
		}}, &ecdsa.PublicKey{}, false},
		{"ECDSA by label", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:object=ecdsa-p256-key",
		}}, &ecdsa.PublicKey{}, false},
		{"fail name", args{&apiv1.GetPublicKeyRequest{
			Name: "",
		}}, nil, true},
		{"fail uri", args{&apiv1.GetPublicKeyRequest{
			Name: "https:id=9999;object=https",
		}}, nil, true},
		{"fail missing", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:id=9999;object=rsa-key",
		}}, nil, true},
		{"fail FindKeyPair", args{&apiv1.GetPublicKeyRequest{
			Name: "pkcs11:foo=bar",
		}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k.GetPublicKey(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PKCS11.GetPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("PKCS11.GetPublicKey() = %T, want %T", got, tt.want)
			}
		})
	}
}

func TestPKCS11_CreateKey(t *testing.T) {
	k := setupPKCS11(t)

	// Make sure to delete the created key
	keyName := "pkcs11:id=7771;object=create-key"
	k.DeleteKey(keyName)

	type args struct {
		req *apiv1.CreateKeyRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *apiv1.CreateKeyResponse
		wantErr bool
	}{
		// SoftHSM2
		{"default", args{&apiv1.CreateKeyRequest{
			Name: keyName,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &ecdsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA SHA256WithRSA", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA256WithRSA,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA SHA384WithRSA", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA384WithRSA,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA SHA512WithRSA", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA512WithRSA,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA SHA256WithRSAPSS", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA256WithRSAPSS,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA SHA384WithRSAPSS", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA384WithRSAPSS,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA SHA512WithRSAPSS", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA512WithRSAPSS,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA 2048", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA256WithRSA,
			Bits:               2048,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"RSA 4096", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.SHA256WithRSA,
			Bits:               4096,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &rsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"ECDSA P256", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.ECDSAWithSHA256,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &ecdsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"ECDSA P384", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.ECDSAWithSHA384,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &ecdsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"ECDSA P521", args{&apiv1.CreateKeyRequest{
			Name:               keyName,
			SignatureAlgorithm: apiv1.ECDSAWithSHA512,
		}}, &apiv1.CreateKeyResponse{
			Name:      keyName,
			PublicKey: &ecdsa.PublicKey{},
			CreateSignerRequest: apiv1.CreateSignerRequest{
				SigningKey: keyName,
			},
		}, false},
		{"fail name", args{&apiv1.CreateKeyRequest{
			Name: "",
		}}, nil, true},
		{"fail bits", args{&apiv1.CreateKeyRequest{
			Name:               "pkcs11:id=9999;object=rsa-create-key",
			Bits:               -1,
			SignatureAlgorithm: apiv1.SHA256WithRSAPSS,
		}}, nil, true},
		{"fail ed25519", args{&apiv1.CreateKeyRequest{
			Name:               "pkcs11:id=9999;object=rsa-create-key",
			SignatureAlgorithm: apiv1.PureEd25519,
		}}, nil, true},
		{"fail unknown", args{&apiv1.CreateKeyRequest{
			Name:               "pkcs11:id=9999;object=rsa-create-key",
			SignatureAlgorithm: apiv1.SignatureAlgorithm(100),
		}}, nil, true},
		{"fail uri", args{&apiv1.CreateKeyRequest{
			Name:               "pkcs11:id=xxxx;object=https",
			SignatureAlgorithm: apiv1.SHA256WithRSAPSS,
		}}, nil, true},
		{"fail FindKeyPair", args{&apiv1.CreateKeyRequest{
			Name:               "pkcs11:foo=bar",
			SignatureAlgorithm: apiv1.SHA256WithRSAPSS,
		}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k.CreateKey(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PKCS11.CreateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				got.PublicKey = tt.want.PublicKey
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PKCS11.CreateKey() = %v, want %v", got, tt.want)
			}
			if got != nil {
				if err := k.DeleteKey(got.Name); err != nil {
					t.Errorf("PKCS11.DeleteKey() error = %v", err)
				}
			}
		})
	}
}

func TestPKCS11_CreateSigner(t *testing.T) {
	k := setupPKCS11(t)
	data := []byte("buggy-coheir-RUBRIC-rabbet-liberal-eaglet-khartoum-stagger")

	// VerifyASN1 verifies the ASN.1 encoded signature, sig, of hash using the
	// public key, pub. Its return value records whether the signature is valid.
	verifyASN1 := func(pub *ecdsa.PublicKey, hash, sig []byte) bool {
		var (
			r, s  = &big.Int{}, &big.Int{}
			inner cryptobyte.String
		)
		input := cryptobyte.String(sig)
		if !input.ReadASN1(&inner, asn1.SEQUENCE) ||
			!input.Empty() ||
			!inner.ReadASN1Integer(r) ||
			!inner.ReadASN1Integer(s) ||
			!inner.Empty() {
			return false
		}
		return ecdsa.Verify(pub, hash, r, s)
	}

	type args struct {
		req *apiv1.CreateSignerRequest
	}
	tests := []struct {
		name       string
		args       args
		algorithm  apiv1.SignatureAlgorithm
		signerOpts crypto.SignerOpts
		wantErr    bool
	}{
		// SoftHSM2
		{"RSA", args{&apiv1.CreateSignerRequest{
			SigningKey: "pkcs11:id=7371;object=rsa-key",
		}}, apiv1.SHA256WithRSA, crypto.SHA256, false},
		{"RSA PSS", args{&apiv1.CreateSignerRequest{
			SigningKey: "pkcs11:id=7371;object=rsa-key",
		}}, apiv1.SHA256WithRSA, crypto.SHA256, false},
		{"ECDSA P256", args{&apiv1.CreateSignerRequest{
			SigningKey: "pkcs11:id=7373;object=ecdsa-p256-key",
		}}, apiv1.ECDSAWithSHA256, crypto.SHA256, false},
		{"ECDSA P384", args{&apiv1.CreateSignerRequest{
			SigningKey: "pkcs11:id=7374;object=ecdsa-p384-key",
		}}, apiv1.ECDSAWithSHA384, crypto.SHA384, false},
		{"ECDSA P521", args{&apiv1.CreateSignerRequest{
			SigningKey: "pkcs11:id=7375;object=ecdsa-p521-key",
		}}, apiv1.ECDSAWithSHA512, crypto.SHA512, false},
		{"fail SigningKey", args{&apiv1.CreateSignerRequest{
			SigningKey: "",
		}}, 0, nil, true},
		{"fail uri", args{&apiv1.CreateSignerRequest{
			SigningKey: "https:id=7375;object=ecdsa-p521-key",
		}}, 0, nil, true},
		{"fail FindKeyPair", args{&apiv1.CreateSignerRequest{
			SigningKey: "pkcs11:foo=bar",
		}}, 0, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k.CreateSigner(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PKCS11.CreateSigner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil {
				hash := tt.signerOpts.HashFunc()
				h := hash.New()
				h.Write(data)
				digest := h.Sum(nil)
				sig, err := got.Sign(rand.Reader, digest, tt.signerOpts)
				if err != nil {
					t.Errorf("cyrpto.Signer.Sign() error = %v", err)
				}

				switch tt.algorithm {
				case apiv1.SHA256WithRSA, apiv1.SHA384WithRSA, apiv1.SHA512WithRSA:
					pub := got.Public().(*rsa.PublicKey)
					if err := rsa.VerifyPKCS1v15(pub, hash, digest, sig); err != nil {
						t.Errorf("rsa.VerifyPKCS1v15() error = %v", err)
					}
				case apiv1.UnspecifiedSignAlgorithm, apiv1.SHA256WithRSAPSS, apiv1.SHA384WithRSAPSS, apiv1.SHA512WithRSAPSS:
					pub := got.Public().(*rsa.PublicKey)
					if err := rsa.VerifyPSS(pub, hash, digest, sig, tt.signerOpts.(*rsa.PSSOptions)); err != nil {
						t.Errorf("rsa.VerifyPSS() error = %v", err)
					}
				case apiv1.ECDSAWithSHA256, apiv1.ECDSAWithSHA384, apiv1.ECDSAWithSHA512:
					pub := got.Public().(*ecdsa.PublicKey)
					if !verifyASN1(pub, digest, sig) {
						t.Error("ecdsa.VerifyASN1() failed")
					}
				default:
					t.Errorf("signature algorithm %s is not supported", tt.algorithm)
				}

			}

		})
	}
}

func TestPKCS11_LoadCertificate(t *testing.T) {
	k := setupPKCS11(t)

	getCertFn := func(i, j int) func() *x509.Certificate {
		return func() *x509.Certificate {
			return testCerts[i].Certificates[j]
		}
	}

	type args struct {
		req *apiv1.LoadCertificateRequest
	}
	tests := []struct {
		name    string
		args    args
		wantFn  func() *x509.Certificate
		wantErr bool
	}{
		{"load", args{&apiv1.LoadCertificateRequest{
			Name: "pkcs11:id=7370;object=root",
		}}, getCertFn(0, 0), false},
		{"load by id", args{&apiv1.LoadCertificateRequest{
			Name: "pkcs11:id=7370",
		}}, getCertFn(0, 0), false},
		{"load by label", args{&apiv1.LoadCertificateRequest{
			Name: "pkcs11:object=root",
		}}, getCertFn(0, 0), false},
		{"fail missing", args{&apiv1.LoadCertificateRequest{
			Name: "pkcs11:id=9999;object=root",
		}}, nil, true},
		{"fail name", args{&apiv1.LoadCertificateRequest{
			Name: "",
		}}, nil, true},
		{"fail uri", args{&apiv1.LoadCertificateRequest{
			Name: "pkcs11:id=xxxx;object=root",
		}}, nil, true},
		{"fail FindCertificate", args{&apiv1.LoadCertificateRequest{
			Name: "pkcs11:foo=bar",
		}}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := k.LoadCertificate(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("PKCS11.LoadCertificate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var want *x509.Certificate
			if tt.wantFn != nil {
				want = tt.wantFn()
				got.Raw, got.RawIssuer, got.RawSubject, got.RawTBSCertificate, got.RawSubjectPublicKeyInfo = nil, nil, nil, nil, nil
				want.Raw, want.RawIssuer, want.RawSubject, want.RawTBSCertificate, want.RawSubjectPublicKeyInfo = nil, nil, nil, nil, nil
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("PKCS11.LoadCertificate() = %v, want %v", got, want)
			}
		})
	}
}

func TestPKCS11_StoreCertificate(t *testing.T) {
	k := setupPKCS11(t)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("ed25519.GenerateKey() error = %v", err)
	}

	cert, err := generateCertificate(pub, priv)
	if err != nil {
		t.Fatalf("x509.CreateCertificate() error = %v", err)
	}

	type args struct {
		req *apiv1.StoreCertificateRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ok", args{&apiv1.StoreCertificateRequest{
			Name:        "pkcs11:id=7771;object=root",
			Certificate: cert,
		}}, false},
		{"fail name", args{&apiv1.StoreCertificateRequest{
			Name:        "",
			Certificate: cert,
		}}, true},
		{"fail certificate", args{&apiv1.StoreCertificateRequest{
			Name:        "pkcs11:id=7771;object=root",
			Certificate: nil,
		}}, true},
		{"fail uri", args{&apiv1.StoreCertificateRequest{
			Name:        "http:id=7771;object=root",
			Certificate: cert,
		}}, true},
		{"fail ImportCertificateWithLabel", args{&apiv1.StoreCertificateRequest{
			Name:        "pkcs11:foo=bar",
			Certificate: cert,
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := k.StoreCertificate(tt.args.req); (err != nil) != tt.wantErr {
				t.Errorf("PKCS11.StoreCertificate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				got, err := k.LoadCertificate(&apiv1.LoadCertificateRequest{
					Name: tt.args.req.Name,
				})
				if err != nil {
					t.Errorf("PKCS11.LoadCertificate() error = %v", err)
				}
				if !reflect.DeepEqual(got, cert) {
					t.Errorf("PKCS11.LoadCertificate() = %v, want %v", got, cert)
				}
				if err := k.DeleteCertificate(tt.args.req.Name); err != nil {
					t.Errorf("PKCS11.DeleteCertificate() error = %v", err)
				}
			}
		})
	}
}
