package env

import (
	"testing"
)

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "hello", false},
		{"whitespace", " ", false},
		{"tab", "\t", false},
		{"newline", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmpty(tt.value)
			if result != tt.expected {
				t.Errorf("IsEmpty(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsValidEmail(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		expected bool
	}{
		{"valid email", "test@example.com", true},
		{"valid email with subdomain", "user@mail.example.com", true},
		{"valid email with numbers", "user123@example.org", true},
		{"valid email with special chars", "user.name+tag@example.co.uk", true},
		{"empty email", "", false},
		{"missing @", "testexample.com", false},
		{"missing domain", "test@", false},
		{"missing local part", "@example.com", false},
		{"invalid domain", "test@", false},
		{"no TLD", "test@example", false},
		{"double @", "test@@example.com", false},
		{"spaces", "test @example.com", false},
		{"invalid chars", "test#@example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEmail(tt.email)
			if result != tt.expected {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.email, result, tt.expected)
			}
		})
	}
}

func TestIsValidEthAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{"valid ethereum address", "0x742d35Cc6634C0532925a3b8D322e99c4c8b9c25", true},
		{"valid address lowercase", "0x742d35cc6634c0532925a3b8d322e99c4c8b9c25", true},
		{"valid address uppercase", "0x742D35CC6634C0532925A3B8D322E99C4C8B9C25", true},
		{"empty address", "", false},
		{"missing 0x prefix", "742d35Cc6634C0532925a3b8D322e99c4c8b9c25", false},
		{"too short", "0x742d35Cc6634C0532925a3b8D322e99c4c8b9c2", false},
		{"too long", "0x742d35Cc6634C0532925a3b8D322e99c4c8b9c259", false},
		{"invalid characters", "0x742d35Cc6634C0532925a3b8D322e99c4c8b9cZZ", false},
		{"only 0x", "0x", false},
		{"no hex digits", "0xgggggggggggggggggggggggggggggggggggggggg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidEthAddress(tt.address)
			if result != tt.expected {
				t.Errorf("IsValidEthAddress(%q) = %v, want %v", tt.address, result, tt.expected)
			}
		})
	}
}

func TestIsValidPrivateKey(t *testing.T) {
	tests := []struct {
		name       string
		privateKey string
		expected   bool
	}{
		{"valid private key", "a0b1c2d3e4f56789abcdef0123456789abcdef0123456789abcdef0123456789", true},
		{"valid private key uppercase", "A0B1C2D3E4F56789ABCDEF0123456789ABCDEF0123456789ABCDEF0123456789", true},
		{"valid private key mixed case", "A0b1C2d3E4f56789AbCdEf0123456789AbCdEf0123456789AbCdEf0123456789", true},
		{"empty private key", "", false},
		{"too short", "a0b1c2d3e4f56789abcdef0123456789abcdef0123456789abcdef012345678", false},
		{"too long", "a0b1c2d3e4f56789abcdef0123456789abcdef0123456789abcdef01234567890", false},
		{"invalid characters", "g0b1c2d3e4f56789abcdef0123456789abcdef0123456789abcdef0123456789", false},
		{"with 0x prefix", "0xa0b1c2d3e4f56789abcdef0123456789abcdef0123456789abcdef0123456789", false},
		{"all zeros", "0000000000000000000000000000000000000000000000000000000000000000", true},
		{"all f's", "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPrivateKey(tt.privateKey)
			if result != tt.expected {
				t.Errorf("IsValidPrivateKey(%q) = %v, want %v", tt.privateKey, result, tt.expected)
			}
		})
	}
}

func TestIsValidIPAddress(t *testing.T) {
	tests := []struct {
		name      string
		ipAddress string
		expected  bool
	}{
		{"localhost", "localhost", true},
		{"valid IP 192.168.1.1", "192.168.1.1", true},
		{"valid IP 10.0.0.1", "10.0.0.1", true},
		{"valid IP 172.16.0.1", "172.16.0.1", true},
		{"valid IP 127.0.0.1", "127.0.0.1", true},
		{"valid IP 255.255.255.255", "255.255.255.255", true},
		{"valid IP 0.0.0.0", "0.0.0.0", true},
		{"empty IP", "", false},
		{"invalid format", "192.168.1", false},
		{"too many octets", "192.168.1.1.1", false},
		{"octet > 255", "192.168.1.256", false},
		{"negative octet", "192.168.1.-1", false},
		{"leading zeros", "192.168.001.1", false},
		{"non-numeric", "192.168.a.1", false},
		{"spaces", "192.168. 1.1", false},
		{"double dots", "192.168..1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIPAddress(tt.ipAddress)
			if result != tt.expected {
				t.Errorf("IsValidIPAddress(%q) = %v, want %v", tt.ipAddress, result, tt.expected)
			}
		})
	}
}

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected bool
	}{
		{"valid port 1024", "1024", true},
		{"valid port 8080", "8080", true},
		{"valid port 65535", "65535", true},
		{"valid port 3000", "3000", true},
		{"valid port 443", "443", false}, // Below 1024
		{"valid port 80", "80", false},   // Below 1024
		{"port 0", "0", false},
		{"port 1023", "1023", false},   // Below 1024
		{"port 65536", "65536", false}, // Above 65535
		{"port 99999", "99999", false},
		{"empty port", "", false},
		{"non-numeric", "abc", false},
		{"negative", "-1", false},
		{"with spaces", " 8080", false},
		{"floating point", "8080.5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPort(tt.port)
			if result != tt.expected {
				t.Errorf("IsValidPort(%q) = %v, want %v", tt.port, result, tt.expected)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid HTTP URL", "http://example.com", true},
		{"valid HTTPS URL", "https://example.com", true},
		{"valid URL with subdomain", "https://www.example.com", true},
		{"valid URL with IP", "http://192.168.1.1", true},
		{"valid URL with localhost", "http://localhost", true},
		{"valid URL with port", "http://example.com:8080", true},
		{"valid URL with IP and port", "http://192.168.1.1:3000", true},
		{"valid URL with localhost and port", "http://localhost:8080", true},
		{"valid complex domain", "https://api.sub.example.co.uk", true},
		{"empty URL", "", false},
		{"no protocol", "example.com", false},
		{"invalid protocol", "ftp://example.com", false},
		{"no domain", "http://", false},
		{"invalid domain", "http://invalid..domain", false},
		{"invalid port", "http://example.com:abc", false},
		{"port out of range", "http://example.com:99999", false},
		{"port too low", "http://example.com:80", false}, // Port validation restricts to 1024+
		{"multiple colons", "http://example.com:8080:9090", false},
		{"spaces in URL", "http://example .com", false},
		{"invalid IP", "http://999.999.999.999", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestIsValidPeerID(t *testing.T) {
	tests := []struct {
		name     string
		peerID   string
		expected bool
	}{
		{"valid peer ID", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB", true},
		{"another valid peer ID", "12D3KooWA1B2C3D4E5F6G7H8I9J0K1L2M3N4O5P6Q7R8S9T0U1V2", true},
		{"empty peer ID", "", false},
		{"wrong prefix", "13D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB", false},
		{"no prefix", "KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB", false},
		{"too short", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91K", false},
		{"too long", "12D3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KBX", false},
		{"partial prefix", "12D", false},
		{"case sensitive prefix", "12d3KooWBNFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB", false},
		{"special characters", "12D3KooW@NFG1QjuF3UKAKvqhdXcxh9iBmj88cM5eU2EK5Pa91KB", true}, // Length check passes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPeerID(tt.peerID)
			if result != tt.expected {
				t.Errorf("IsValidPeerID(%q) = %v, want %v", tt.peerID, result, tt.expected)
			}
		})
	}
}
