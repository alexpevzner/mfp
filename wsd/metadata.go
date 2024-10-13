// MFP - Miulti-Function Printers and scanners toolkit
// WSD core protocol
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Metadata exchange definitions
//
// Specification: Devices Profile for Web Services
// https://specs.xmlsoap.org/ws/2006/02/devprof/devicesprofile.pdf

package wsd

import "github.com/alexpevzner/mfp/xmldoc"

// Dialect attribute values for ThisDevice, ThisModel and Relationship
// sections.
const (
	ThisDeviceDialect   = "http://schemas.xmlsoap.org/ws/2006/02/devprof/ThisDevice"
	ThisModelDialect    = "http://schemas.xmlsoap.org/ws/2006/02/devprof/ThisModel"
	RelationshipDialect = "http://schemas.xmlsoap.org/ws/2006/02/devprof/host"
)

// Relationship types for the needs of Metadata exchange, implemented here.
const (
	RelationshipHost = "http://schemas.xmlsoap.org/ws/2006/02/devprof/host"
)

// Metadata is the device description, returned as response to
// the [Get] request.
type Metadata struct {
	ThisDevice   ThisDeviceMetadata // Device description
	ThisModel    ThisModelMetadata  // Model description
	Relationship Relationship       // Host and hosted services
}

// ThisDeviceMetadata contains information about the particular device.
type ThisDeviceMetadata struct {
	FriendlyName    LocalizedStringList // Device user-friendly name
	FirmwareVersion string              // Firmware version
	SerialNumber    string              // Serial number
}

// ThisModelMetadata contains information about the model.
type ThisModelMetadata struct {
	Manufacturer    LocalizedStringList // Manufacturer name
	ManufacturerURL string              // Manufacturer URL
	ModelName       LocalizedStringList // Model name
	ModelNumber     string              // Model number
	ModelURL        string              // Model URL
	PresentationURL string              // HTML page for this model
}

// Relationship defines relationship between host (i.e., the device)
// and hosted services (i.e., print/scan serviced, implemented by
// the device).
type Relationship struct {
	Host   *ServiceMetadata  // The host inself (very optional)
	Hosted []ServiceMetadata // Hosted services
}

// ServiceMetadata contains information about the host or hosted service.
type ServiceMetadata struct {
	EndpointReference []EndpointReference // Service endpoints
	Types             Types               // Service types
	ServiceID         AnyURI              // Service identifier
}

// ToXML generates XML tree for Metadata.
func (meta Metadata) ToXML() xmldoc.Element {
	// Generate sections
	thisDevice := meta.ThisDevice.ToXML()
	thisModel := meta.ThisModel.ToXML()
	relationship := meta.Relationship.ToXML()

	// Build metadata
	metadata := xmldoc.Element{
		Name: NsMex + ":Metadata",
		Children: []xmldoc.Element{
			thisDevice,
			thisModel,
			relationship,
		},
	}

	return metadata
}

// DecodeThisDeviceMetadata decodes ThisDeviceMetadata from the
// XML tree.
func DecodeThisDeviceMetadata(root xmldoc.Element) (
	thisdev ThisDeviceMetadata, err error) {

	defer func() { err = xmlErrWrap(root, err) }()

	// Find MetadataSection element
	data, ok := root.ChildByName(NsMex + ":MetadataSection")
	if !ok {
		err = xmlErrMissed(NsDevprof + ":MetadataSection")
		return
	}

	defer func() { err = xmlErrWrap(data, err) }()

	// Decode FriendlyName
	for _, chld := range data.Children {
		if chld.Name == NsDevprof+":FriendlyName" {
			ls := decodeLocalizedString(chld)
			thisdev.FriendlyName = append(thisdev.FriendlyName, ls)
		}
	}

	if len(thisdev.FriendlyName) == 0 {
		err = xmlErrMissed(NsDevprof + ":FriendlyName")
		return
	}

	// Decode other elements
	firmwareVersion := xmldoc.Lookup{Name: NsDevprof + ":FirmwareVersion",
		Required: true}
	serialNumber := xmldoc.Lookup{Name: NsDevprof + ":SerialNumber",
		Required: true}

	missed := data.Lookup(&firmwareVersion, &serialNumber)
	if missed != nil {
		err = xmlErrMissed(missed.Name)
		return
	}

	thisdev.FirmwareVersion = firmwareVersion.Elem.Text
	thisdev.SerialNumber = serialNumber.Elem.Text

	return
}

// ToXML generates XML tree for ThisDeviceMetadata
func (thisdev ThisDeviceMetadata) ToXML() xmldoc.Element {
	data := xmldoc.Element{
		Name: NsDevprof + ":ThisDevice",
	}

	for _, fn := range thisdev.FriendlyName {
		data.Children = append(data.Children,
			fn.ToXML(NsDevprof+":FriendlyName"))
	}

	data.Children = append(data.Children,
		xmldoc.Element{
			Name: NsDevprof + ":FirmwareVersion",
			Text: thisdev.FirmwareVersion,
		})

	data.Children = append(data.Children,
		xmldoc.Element{
			Name: NsDevprof + ":SerialNumber",
			Text: thisdev.SerialNumber,
		})

	thisDevice := xmldoc.Element{
		Name: NsMex + ":MetadataSection",
		Attrs: []xmldoc.Attr{{
			Name:  "Dialect",
			Value: ThisDeviceDialect,
		}},
		Children: []xmldoc.Element{data},
	}

	return thisDevice
}

// DecodeThisModelMetadata decodes ThisModelMetadata from the XML tree.
func DecodeThisModelMetadata(root xmldoc.Element) (
	thismdl ThisModelMetadata, err error) {

	defer func() { err = xmlErrWrap(root, err) }()

	// Find MetadataSection element
	data, ok := root.ChildByName(NsMex + ":MetadataSection")
	if !ok {
		err = xmlErrMissed(NsDevprof + ":MetadataSection")
		return
	}

	defer func() { err = xmlErrWrap(data, err) }()

	// Decode repeated elements, i.e. Manufacturer and ModelName
	for _, chld := range data.Children {
		switch chld.Name {
		case NsDevprof + ":Manufacturer":
			ls := decodeLocalizedString(chld)
			thismdl.Manufacturer = append(thismdl.Manufacturer, ls)
		case NsDevprof + ":ModelName":
			ls := decodeLocalizedString(chld)
			thismdl.ModelName = append(thismdl.ModelName, ls)
		}
	}

	if len(thismdl.Manufacturer) == 0 {
		err = xmlErrMissed(NsDevprof + ":Manufacturer")
		return
	}

	if len(thismdl.ModelName) == 0 {
		err = xmlErrMissed(NsDevprof + ":ModelName")
		return
	}

	// Decode other elements
	manufacturerURL := xmldoc.Lookup{Name: NsDevprof + ":ManufacturerUrl"}
	modelNumber := xmldoc.Lookup{Name: NsDevprof + ":ModelNumber",
		Required: true}
	modelURL := xmldoc.Lookup{Name: NsDevprof + ":ModelUrl"}
	presentationURL := xmldoc.Lookup{Name: NsDevprof + ":PresentationUrl"}

	missed := data.Lookup(&manufacturerURL, &modelNumber,
		&modelURL, &presentationURL)

	if missed != nil {
		err = xmlErrMissed(missed.Name)
		return
	}

	if manufacturerURL.Found {
		thismdl.ManufacturerURL = manufacturerURL.Elem.Text
	}

	thismdl.ModelNumber = modelNumber.Elem.Text

	if modelURL.Found {
		thismdl.ModelURL = modelURL.Elem.Text
	}

	if presentationURL.Found {
		thismdl.PresentationURL = presentationURL.Elem.Text
	}

	return
}

// ToXML generates XML tree for ThisModelMetadata
func (thismdl ThisModelMetadata) ToXML() xmldoc.Element {
	data := xmldoc.Element{
		Name: NsDevprof + ":ThisModel",
	}

	for _, mfg := range thismdl.Manufacturer {
		data.Children = append(data.Children,
			mfg.ToXML(NsDevprof+":Manufacturer"))
	}

	if thismdl.ManufacturerURL != "" {
		data.Children = append(data.Children,
			xmldoc.Element{
				Name: NsDevprof + ":ManufacturerUrl",
				Text: thismdl.ManufacturerURL,
			})
	}

	for _, mdl := range thismdl.ModelName {
		data.Children = append(data.Children,
			mdl.ToXML(NsDevprof+":ModelName"))
	}

	data.Children = append(data.Children,
		xmldoc.Element{
			Name: NsDevprof + ":ModelNumber",
			Text: thismdl.ModelNumber,
		})

	if thismdl.ModelURL != "" {
		data.Children = append(data.Children,
			xmldoc.Element{
				Name: NsDevprof + ":ModelUrl",
				Text: thismdl.ModelURL,
			})
	}

	if thismdl.PresentationURL != "" {
		data.Children = append(data.Children,
			xmldoc.Element{
				Name: NsDevprof + ":PresentationUrl",
				Text: thismdl.PresentationURL,
			})
	}

	thisModel := xmldoc.Element{
		Name: NsMex + ":MetadataSection",
		Attrs: []xmldoc.Attr{{
			Name:  "Dialect",
			Value: ThisModelDialect,
		}},
		Children: []xmldoc.Element{data},
	}

	return thisModel
}

// DecodeRelationship decodes Relationship from the XML tree
func DecodeRelationship(root xmldoc.Element) (rel Relationship, err error) {
	defer func() { err = xmlErrWrap(root, err) }()

	// Find MetadataSection element
	data, ok := root.ChildByName(NsMex + ":MetadataSection")
	if !ok {
		err = xmlErrMissed(NsDevprof + ":MetadataSection")
		return
	}

	defer func() { err = xmlErrWrap(data, err) }()

	// Decode Host/Hosted
	for _, chld := range data.Children {
		switch chld.Name {
		case NsDevprof + ":Host":
			if rel.Host == nil {
				var host ServiceMetadata
				host, err = DecodeServiceMetadata(chld)
				if err != nil {
					return
				}
				rel.Host = &host
			}

		case NsDevprof + ":Hosted":
			var hosted ServiceMetadata
			hosted, err = DecodeServiceMetadata(chld)
			if err != nil {
				return
			}
			rel.Hosted = append(rel.Hosted, hosted)
		}
	}

	return
}

// ToXML generates XML tree for Relationship
func (rel Relationship) ToXML() xmldoc.Element {
	root := xmldoc.Element{
		Name: NsDevprof + ":Relationship",
		Attrs: []xmldoc.Attr{{
			Name:  "Type",
			Value: RelationshipHost,
		}},
	}

	if rel.Host != nil {
		root.Children = append(root.Children,
			rel.Host.ToXML(NsDevprof+":Host"))
	}

	for _, hosted := range rel.Hosted {
		root.Children = append(root.Children,
			hosted.ToXML(NsDevprof+":Hosted"))
	}

	return root
}

// DecodeServiceMetadata decodes ServiceMetadata from the XML tree.
func DecodeServiceMetadata(root xmldoc.Element) (
	svcmeta ServiceMetadata, err error) {

	defer func() { err = xmlErrWrap(root, err) }()

	// Decode EndpointReference
	for _, chld := range root.Children {
		if chld.Name == NsAddressing+":EndpointReference" {
			var ep EndpointReference
			ep, err = DecodeEndpointReference(chld)
			if err != nil {
				return
			}

			svcmeta.EndpointReference = append(
				svcmeta.EndpointReference, ep)
		}
	}

	// Decode other elements
	types := xmldoc.Lookup{Name: NsDiscovery + ":" + "Types"}
	serviceID := xmldoc.Lookup{Name: NsDevprof + ":ServiceId"}

	root.Lookup(&types, &serviceID)

	if types.Found {
		svcmeta.Types, err = DecodeTypes(types.Elem)
		if err != nil {
			return
		}
	}

	if serviceID.Found {
		svcmeta.ServiceID = AnyURI(types.Elem.Text)
	}

	return
}

// ToXML generates XML tree for the ServiceMetadata
func (svcmeta ServiceMetadata) ToXML(name string) xmldoc.Element {
	elm := xmldoc.Element{
		Name: name,
	}

	for _, ep := range svcmeta.EndpointReference {
		elm.Children = append(elm.Children,
			ep.ToXML(NsAddressing+":EndpointReference"))
	}

	elm.Children = append(elm.Children, svcmeta.Types.ToXML())

	if svcmeta.ServiceID != "" {
		elm.Children = append(elm.Children,
			xmldoc.Element{
				Name: NsDevprof + ":ServiceId",
				Text: string(svcmeta.ServiceID),
			})
	}

	return elm
}
