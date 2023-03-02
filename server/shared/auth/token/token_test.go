package token

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const TestPublicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAk1373RtlDiFACtMRG6TT
yse8XmXgAxVfnfyEwtuoH8rBA7q/uluNkIO18C+7GNXbMyZlDVg2xuCsjpC9rq9V
dMbfzuUCFUBGyjCPpYvH8mi8YWCjJzjcBcvtt5FPitqJM7UKnYfxi876VYSAREV3
dCZ5L731R4Y9Bg9QrYTGVEPkVFhFrBNC82mu7q4xJJkLW3l2BiMG/QvsBmcDx+Sd
C3HE3UorW0JuOz3VUPjZP6xQ9LgRgDFcR/wuGIhtnNUVwNxeyh63dEgmj/a+3y5L
uRsWkKx5OmOIy4hG6Qggej1SLkjaENuGGhuMTwvYgx1vV9w33ueQDB0O9K5Af4r8
zwIDAQAB
-----END PUBLIC KEY-----`

func TestVerify(t *testing.T) {
	pubkey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(TestPublicKey))
	if err != nil {
		t.Fatalf("connot parse private key: %v", err)
	}

	v := &JWTTokenVerifier{
		PublicKey: pubkey,
	}

	cases := []struct {
		name    string
		token   string
		now     time.Time
		want    string
		wantErr bool
	}{
		{
			name:    "valid_token",
			token:   "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNWY3YzMxNjhlMjI4M2FhNzIyZTM1MWEzIn0.FxiQjhLUlyImiIvRAerPYUZCMpzTtDHNHQ09AV5GiJ1LGxZjPfiqeunmLDJ8Tkzc8xk0i1NVwflpLeo_e6rh-N0u5kWoe8Ip75Rpg8eR-6ola6TfWXax1WEO0on6K2hIlwFLbIQtI0v8UIGkm-6UOUdlnvlbW28MJem-4q7gURgcqHe6DJyTRRDaJUL2Aqz8pxFSU4hGU3_d9AhvmW7NGY7Cy_z3O615R7hYrbm72FNKS_OHJGJWUI-tHjkktqNEeAdUvdSx7n3Q8YlYUtWY6nxlyL7NqU185oCWuhUncXxdrTWRn8hMl78NeoQn_M6zPprYjrX7eUMehcDt6e6kqQ",
			now:     time.Unix(1516239122, 0),
			want:    "5f7c3168e2283aa722e351a3",
			wantErr: false,
		},
		{
			name:    "token_expired",
			token:   "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNWY3YzMxNjhlMjI4M2FhNzIyZTM1MWEzIn0.FxiQjhLUlyImiIvRAerPYUZCMpzTtDHNHQ09AV5GiJ1LGxZjPfiqeunmLDJ8Tkzc8xk0i1NVwflpLeo_e6rh-N0u5kWoe8Ip75Rpg8eR-6ola6TfWXax1WEO0on6K2hIlwFLbIQtI0v8UIGkm-6UOUdlnvlbW28MJem-4q7gURgcqHe6DJyTRRDaJUL2Aqz8pxFSU4hGU3_d9AhvmW7NGY7Cy_z3O615R7hYrbm72FNKS_OHJGJWUI-tHjkktqNEeAdUvdSx7n3Q8YlYUtWY6nxlyL7NqU185oCWuhUncXxdrTWRn8hMl78NeoQn_M6zPprYjrX7eUMehcDt6e6kqQ",
			now:     time.Unix(1517239122, 0),
			wantErr: true,
		},
		{
			name:    "bad_token",
			token:   "bad_token",
			now:     time.Unix(1516239122, 0),
			wantErr: true,
		},
		{
			name:    "wrong_signature",
			token:   "eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyNDYyMjIsImlhdCI6MTUxNjIzOTAyMiwiaXNzIjoiY29vbGNhci9hdXRoIiwic3ViIjoiNWY3YzMxNjhlMjI4M2FhNzIyZTM1MWEzIn0.VbOv03FQOOQyh6DK5fZVyGYsp0QErGtVaKbRtSLsRLhYjItlQQvOkZmJvsA5OVl3A3GnG2CQwtKa7q9b0KiFIwD7Cx4hxs7L2TEdjLnP238jmcMEukrEfQ6XtRnUFYayCJiYZZPpDu9ZmNwFcWcXUB6eycqFz-wUbWWf0-Tr5DAW72BNh8GXtDijZW3Ada7OL7N9gUAyNlzVNTf311mV5-8I0p0tv6zWWFByKI_i-1bTLd8jxD3Z7V_yESDYpWuFFGADpnZWGKlU5QohWkF4Hy8cjcs5_ZTMuvN42hWwJsPXHE9eHkdbbsjMadq8yQkfZJxItILb8YKDbP6-AVfARw",
			now:     time.Unix(1516239122, 0),
			wantErr: true,
		},
	}

	for _, cc := range cases {
		t.Run(cc.name, func(t *testing.T) {
			jwt.TimeFunc = func() time.Time {
				return cc.now
			}
			accountID, err := v.Verify(cc.token)
			if !cc.wantErr && err != nil {
				t.Errorf("verification failed: %v", err)
			}
			if cc.wantErr && err == nil {
				t.Errorf("want error; got no error")
			}
			if cc.want != accountID {
				t.Fatalf("accountID is not right: want: %q, got: %q", cc.want, accountID)
			}
		})
	}
}
