package main

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Profile struct {
	Name        string `json:"name"`
	CommandLine string `json:"command_line"`
	User        string `json:"user"`
}

type Config struct {
	Port                string    `json:"port"`
	Profiles            []Profile `json:"profiles"`
	TimeToleranceSecond int       `json:"time_tolerance_second"`
}

type verify struct {
	Time      string `json:"date"`
	Profile   string `json:"profile"`
	Device    string `json:"device"`
	PublicKey string `json:"public_key"`
}

func loadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	profileNames := make(map[string]struct{})
	for _, profile := range config.Profiles {
		if _, exists := profileNames[profile.Name]; exists {
			return nil, fmt.Errorf("duplicate profile name found: %s", profile.Name)
		}
		profileNames[profile.Name] = struct{}{}
	}

	return &config, nil
}

func generateSelfSignedCert(hostname string) error {
	cmd := exec.Command("openssl", "req", "-x509", "-nodes", "-days", "365", "-newkey", "rsa:2048",
		"-keyout", "certkey.pem", "-out", "cert.pem", "-subj", fmt.Sprintf("/CN=%s", hostname))
	return cmd.Run()
}

func startServer(port string, cert string, key string) {
	app := fiber.New()
	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Post("/api/v0/device/add", addDevice)
	app.Post("/api/v0/verify", verifyRequest)

	log.Fatal(app.ListenTLS(":"+port, cert, key))
}

func addDevice(c *fiber.Ctx) error {
	type DeviceRequest struct {
		Name      string   `json:"name"`
		PublicKey string   `json:"public_key"`
		Profiles  []string `json:"profiles"`
	}

	var req DeviceRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	ip := c.IP()
	if ip != "127.0.0.1" && ip != "::1" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied: This endpoint can only be accessed from the host computer."})
	}

	// Check if profiles exist in the config
	profilesMap := make(map[string]struct{})
	for _, profile := range config.Profiles {
		profilesMap[profile.Name] = struct{}{}
	}

	for _, profile := range req.Profiles {
		if _, exists := profilesMap[profile]; !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Profile %s does not exist", profile)})
		}
	}

	// Add the device to the database
	if err := db.AddDevice(req.Name, req.PublicKey, req.Profiles); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to add device to database"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Device registered successfully", "device": req.Name})
}

func executeProfileCommand(profileName string) error {
	for _, profile := range config.Profiles {
		if profile.Name == profileName {
			// Use a shell to execute the command
			cmd := exec.Command("sh", "-c", profile.CommandLine)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			return cmd.Run()
		}
	}
	return fmt.Errorf("profile %s not found", profileName)
}

func verifyRequest(c *fiber.Ctx) error {
	var req verify
	t := time.Now()

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate the signature
	signature := c.Get("X-sign")
	if !verifySignature(c.Body(), signature, req.PublicKey) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Invalid signature"})
	}

	// Check if the time is within the tolerance
	verifyTime, err := time.Parse(time.RFC3339, req.Time)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid time format, use RFC3339 format"})
	}

	if t.Sub(verifyTime).Seconds() > float64(config.TimeToleranceSecond) || verifyTime.Sub(t).Seconds() > float64(config.TimeToleranceSecond) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Verification time is out of tolerance"})
	}

	// Check if the device exists
	if !db.DeviceExists(req.Device) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Device not found"})
	}

	// Check if the public key is part of the device
	if !db.KeyExists(req.Device, req.PublicKey) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Public key does not match the device"})
	}

	// Check if the device can use the profile
	if !db.CanDeviceUseProfile(req.Device, req.Profile) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Device is not authorized to use the profile"})
	}

	// Execute the command associated with the profile
	if err := executeProfileCommand(req.Profile); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to execute profile command: %v", err)})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Verification successful"})
}

func verifySignature(req []byte, signature string, publicKey string) bool {
	publicKey = strings.ReplaceAll(publicKey, "\\n", "\n")
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil || block.Type != "PUBLIC KEY" {
		fmt.Println("Failed to decode PEM block containing public key", block)
		return false
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Println("Failed to parse public key:", err)
		return false
	}

	ecdsaPub, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Public key is not of type ECDSA")
		return false
	}

	hash := sha256.Sum256(req)

	signatureBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		fmt.Println("Failed to decode signature:", err)
		return false
	}

	return ecdsa.VerifyASN1(ecdsaPub, hash[:], signatureBytes)
}

var db *DB
var config *Config

func main() {
	var err error
	config, err = loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	if _, err := os.Stat("certkey.pem"); os.IsNotExist(err) {
		if _, err := os.Stat("cert.pem"); os.IsNotExist(err) {
			hostname, err := os.Hostname()
			if err != nil {
				log.Fatalf("Error getting hostname: %v", err)
			}
			if err := generateSelfSignedCert(hostname + ".local"); err != nil {
				log.Fatalf("Error generating self-signed certificate: %v", err)
			}
		}
	}
	certPEM, err := os.ReadFile("cert.pem")
	if err != nil {
		log.Fatalf("Error reading certificate file: %v", err)
	}

	fmt.Printf("Server Cert:\n%s\n", certPEM)

	db, err = NewDB("")
	startServer(config.Port, "cert.pem", "certkey.pem")
}
