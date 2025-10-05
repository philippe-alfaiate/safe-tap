package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

type Device struct {
	Name         string
	ProfileNames []string
	CreationDate time.Time
	Keys         []KeyInfo
}

type KeyInfo struct {
	PublicKey    string
	CreationDate time.Time
	IsRevoked    bool
}

type DB struct {
	Devices  []Device
	filename string
}

func NewDB(filename string) (*DB, error) {
	if filename == "" {
		filename = "db.json"
	}
	devices, err := loadDevicesFromFile(filename)
	if err != nil {
		return nil, err
	}
	return &DB{Devices: devices, filename: filename}, nil
}

func loadDevicesFromFile(filename string) ([]Device, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var devices []Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return nil, err
	}

	return devices, nil
}

func (db *DB) save() error {

	data, err := json.Marshal(db.Devices)
	if err != nil {
		return err
	}

	return os.WriteFile(db.filename, data, 0644)
}

func (db *DB) ListRevokedKeys(deviceName string) ([]KeyInfo, error) {
	for _, device := range db.Devices {
		if device.Name == deviceName {
			var revokedKeys []KeyInfo
			for _, key := range device.Keys {
				if key.IsRevoked {
					revokedKeys = append(revokedKeys, key)
				}
			}
			return revokedKeys, nil // Return the list of revoked keys, empty if no revoked keys found
		}
	}
	return nil, nil // Return nil if device not found
}

func (db *DB) RevokeKey(deviceName, publicKey string) error {
	for i, device := range db.Devices {
		if device.Name == deviceName {
			for j, key := range device.Keys {
				if key.PublicKey == publicKey {
					if db.Devices[i].Keys[j].IsRevoked {
						return nil // Key is already revoked, no action needed
					}
					db.Devices[i].Keys[j].IsRevoked = true
					return db.save()
				}
			}
			return nil // Key not found, no action needed
		}
	}
	return nil // Device not found, no action needed
}

func (db *DB) AddDevice(deviceName, publicKey string, Profiles []string) error {
	t := time.Now()
	if db.DeviceExists(deviceName) {
		return fmt.Errorf("device with name %s already exists", deviceName)
	}
	key := KeyInfo{
		PublicKey:    publicKey,
		CreationDate: t,
		IsRevoked:    false,
	}
	device := Device{
		Name:         deviceName,
		ProfileNames: Profiles,
		CreationDate: t,
		Keys:         []KeyInfo{key},
	}
	db.Devices = append(db.Devices, device)
	return db.save()
}

func (db *DB) AddKeyToDevice(deviceName, publicKey string) error {
	for i, device := range db.Devices {
		if device.Name == deviceName {
			for _, key := range device.Keys {
				if key.PublicKey == publicKey {
					return nil // Key already exists, no action needed
				}
			}
			newKey := KeyInfo{
				PublicKey:    publicKey,
				CreationDate: time.Now(),
				IsRevoked:    false,
			}
			db.Devices[i].Keys = append(db.Devices[i].Keys, newKey)
			return db.save()
		}
	}
	return nil // Device not found, no action needed
}

func (db *DB) DeviceExists(deviceName string) bool {
	for _, device := range db.Devices {
		if device.Name == deviceName {
			return true
		}
	}
	return false
}

func (db *DB) KeyExists(deviceName, publicKey string) bool {
	for _, device := range db.Devices {
		if device.Name == deviceName {
			for _, key := range device.Keys {
				if key.PublicKey == publicKey {
					return true
				}
			}
		}
	}
	return false
}

func (db *DB) CanDeviceUseProfile(deviceName, profileName string) bool {
	for _, device := range db.Devices {
		if device.Name == deviceName {
			for _, profile := range device.ProfileNames {
				if profile == profileName {
					return true // Device can use the profile
				}
			}
			return false // Profile not found for the device
		}
	}
	return false // Device not found
}
