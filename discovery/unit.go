// MFP - Miulti-Function Printers and scanners toolkit
// Device discovery
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Device units

package discovery

import (
	"fmt"
	"strings"

	"github.com/alexpevzner/mfp/uuid"
)

// PrintUnit represents a print unit.
type PrintUnit struct {
	ID        UnitID            // Unit identity
	Meta      Metadata          // Unit metadata
	Params    PrinterParameters // Printer parameters
	Endpoints []string          // URLs of printer endpoints
}

// ScanUnit represents a scan unit.
type ScanUnit struct {
	ID        UnitID            // Unit identity
	Meta      Metadata          // Unit metadata
	Params    ScannerParameters // Scanner parameters
	Endpoints []string          // URLs of printer endpoints
}

// FaxoutUnit represents a fax unit.
type FaxoutUnit struct {
	ID        UnitID            // Unit identity
	Meta      Metadata          // Unit metadata
	Params    PrinterParameters // Printer parameters
	Endpoints []string          // URLs of printer endpoints
}

// unit is the internal representation of the PrintUnit, ScanUnit
// or FaxoutUnit
type unit struct {
	id        UnitID   // Unit identity
	meta      Metadata // Unit metadata
	params    any      // PrinterParameters or ScannerParameters
	endpoints []string // Unit endpoints
}

// Export exports unit ad PrintUnit, ScanUnit or FaxoutUnit
func (un unit) Export() any {
	switch params := un.params.(type) {
	case PrinterParameters:
		// PrinterParameters can be used either with PrintUnit
		// or FaxoutUnit
		switch un.id.SvcType {
		case ServicePrinter:
			return PrintUnit{
				ID:        un.id,
				Meta:      un.meta,
				Params:    params,
				Endpoints: un.endpoints,
			}
		case ServiceFaxout:
			return FaxoutUnit{
				ID:        un.id,
				Meta:      un.meta,
				Params:    params,
				Endpoints: un.endpoints,
			}
		}

	case ScannerParameters:
		return ScanUnit{
			ID:        un.id,
			Meta:      un.meta,
			Params:    params,
			Endpoints: un.endpoints,
		}
	}

	return nil
}

// UnitID contains combination of parameters that identifies a device.
//
// Please note, depending on a discovery protocol being used, not
// all the fields of the following structure will have any sense.
//
// Note also, that device UUID is not necessary the same between
// protocols. Some Canon devices known to use different UUID for
// DNS-SD and WS-Discovery.
//
// The intended fields usage is the following:
//
//	DeviceName - realm-unique device name, in the DNS-SD sense.
//	             E.g., "Kyocera ECOSYS M2040dn",
//	UUID       - device UUID
//	Queue      - Job queue name for units with logical sub-units,
//	             like LPD server with multiple queues
//	Realm      - search realm. Different realms are treated as
//	             independent namespaces.
//	Zone       - allows backend to further divide its namespace
//	             (for example, to split it between network interfaces)
//	Variant    - used to distinguish between logically equivalent
//	             variants of discovered units, that backend sees as
//	             independent instances (for example IP4/IP6, HTTP/HTTPS)
//	SvcType    - service type, printer/scanner/faxout
//	SvcProto   - service protocol, i.e., IPP, LPD, eSCL etc
//	Serial     - device serial number, if appropriate (i.e., for USB)
type UnitID struct {
	DeviceName string       // Realm-unique device name
	UUID       uuid.UUID    // uuid.NilUUID if not available
	Queue      string       // Logical unit within a device
	Realm      SearchRealm  // Search realm
	Zone       string       // Namespace zone within the Realm
	Variant    string       // Finding variant of the same unit
	SvcType    ServiceType  // Service type
	SvcProto   ServiceProto // Service protocol
	Serial     string       // "" if not avaliable
}

// SameDevice reports if two [UnitID]s belong to the same device.
func (id UnitID) SameDevice(id2 UnitID) bool {
	if id.UUID == id2.UUID {
		return true
	}

	if id.DeviceName == id2.DeviceName &&
		id.Realm == id2.Realm &&
		id.Zone == id2.Zone {
		return true
	}

	return false
}

// SameService reports if two [UnitID]s belong to the same service of
// the same device.
func (id UnitID) SameService(id2 UnitID) bool {
	return id.SvcType == id2.SvcType && id.SameDevice(id2)
}

// SameUnit reports if two [UnitID]s belong to the same unit of
// the same device.
func (id UnitID) SameUnit(id2 UnitID) bool {
	return id.Queue == id2.Queue && id.SameService(id2)
}

// MarshalText dumps [UnitID] as text, for [log.Object].
// It implements [encoding.TextMarshaler] interface.
func (id UnitID) MarshalText() ([]byte, error) {
	var line string
	lines := make([]string, 0, 6)

	if id.DeviceName != "" {
		line = fmt.Sprintf("DeviceName: %q", id.DeviceName)
		lines = append(lines, line)
	}
	if id.UUID != uuid.NilUUID {
		line = fmt.Sprintf("UUID:       %s", id.UUID)
		lines = append(lines, line)
	}
	if id.Queue != "" {
		line = fmt.Sprintf("Queue:      %q", id.Queue)
		lines = append(lines, line)
	}

	line = fmt.Sprintf("Realm:      %s", id.Realm)
	lines = append(lines, line)

	if id.Zone != "" {
		line = fmt.Sprintf("Zone:       %s", id.Zone)
		lines = append(lines, line)
	}

	if id.Variant != "" {
		line = fmt.Sprintf("Variant:    %s", id.Variant)
		lines = append(lines, line)
	}

	line = fmt.Sprintf("Service:    %s %s", id.SvcProto, id.SvcType)
	lines = append(lines, line)

	if id.Serial != "" {
		line = fmt.Sprintf("Serial:     %s", id.Serial)
		lines = append(lines, line)
	}

	return []byte(strings.Join(lines, "\n")), nil
}
