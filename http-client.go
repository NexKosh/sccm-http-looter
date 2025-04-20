package main

import (
    "crypto/tls"
    "log/slog"
    "net/http"
    "os"
    "time"

    "github.com/Azure/go-ntlmssp"
)

// customTransport: Process User-Agent & NTLM 
type customTransport struct {
    Transport http.RoundTripper
    UserAgent string
    Username  string
    Password  string
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Setting User-Agent
    if t.UserAgent != "" {
        req.Header.Set("User-Agent", t.UserAgent)
    }

    // Add Username & Password
    if t.Username != "" && t.Password != "" {
        req.SetBasicAuth(t.Username, t.Password)
    }

    // Execute
    return t.Transport.RoundTrip(req)
}

func createCustomHTTPClient(userAgent string, insecureSkipVerify bool, httpTimeout string, useNTLM bool, username, password string) http.Client {
    // Base transport
    baseTransport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: insecureSkipVerify,
        },
    }

    // Default transport
    transport := http.RoundTripper(baseTransport)

    // If NTLM is needed, wrap the transport with NTLM negotiator
    if useNTLM {
        transport = ntlmssp.Negotiator{
            RoundTripper: baseTransport,
        }
    }

    // Parse timeout
    timeout, err := time.ParseDuration(httpTimeout)
    if err != nil {
        slog.Error("Invalid HTTP timeout value", "timeout", httpTimeout, "error", err)
        os.Exit(1)
    }

    // Create the HTTP client with custom transport
    client := http.Client{
        Timeout: timeout,
        Transport: &customTransport{
            Transport: transport,
            UserAgent: userAgent,
            Username:  username,
            Password:  password,
        },
    }

    return client
}
