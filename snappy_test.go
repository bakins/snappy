package snappy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnappy(t *testing.T) {
	tests := []struct {
		name   string
		server bool
		client bool
	}{
		{
			name: "disabled",
		},
		{
			name:   "client only",
			client: true,
		},
		{
			name:   "server only",
			server: true,
		},
		{
			name:   "enabled",
			server: true,
			client: true,
		},
	}

	bodyLen := 65536

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			handler := http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body := make([]byte, bodyLen)
				_, err := w.Write(body)
				require.NoError(t, err)
			}))

			if test.server {
				handler = Handler(handler)
			}
			s := httptest.NewServer(handler)
			defer s.Close()

			req, err := http.NewRequest(http.MethodGet, s.URL, nil)
			require.NoError(t, err)

			client := &http.Client{}
			if test.client {
				client.Transport = Transport(nil)
			}

			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)

			if test.server {
				require.Contains(t, resp.Header.Get("Vary"), "Accept-Encoding")

				if test.client {
					require.True(t, hasSnappyEncoding(resp.Header["Content-Encoding"]...))
					require.IsType(t, &snappyReader{}, resp.Body)
					body, err := ioutil.ReadAll(resp.Body)
					require.NoError(t, err)
					require.Len(t, body, bodyLen)
					return
				}
			}

			require.False(t, hasSnappyEncoding(resp.Header["Content-Encoding"]...))
		})
	}
}
