package providerv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	structpb "google.golang.org/protobuf/types/known/structpb"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)

	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SearchPartsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *structpb.Struct       `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	Resolved      *structpb.Struct       `protobuf:"bytes,2,opt,name=resolved,proto3" json:"resolved,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchPartsRequest) Reset() {
	*x = SearchPartsRequest{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchPartsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchPartsRequest) ProtoMessage() {}

func (x *SearchPartsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchPartsRequest) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{0}
}

func (x *SearchPartsRequest) GetConfig() *structpb.Struct {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *SearchPartsRequest) GetResolved() *structpb.Struct {
	if x != nil {
		return x.Resolved
	}
	return nil
}

type ProviderConfig struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Provider      string                 `protobuf:"bytes,1,opt,name=provider,proto3" json:"provider,omitempty"`
	Endpoint      string                 `protobuf:"bytes,2,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	Params        *structpb.Struct       `protobuf:"bytes,3,opt,name=params,proto3" json:"params,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProviderConfig) Reset() {
	*x = ProviderConfig{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProviderConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProviderConfig) ProtoMessage() {}

func (x *ProviderConfig) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ProviderConfig) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{1}
}

func (x *ProviderConfig) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *ProviderConfig) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *ProviderConfig) GetParams() *structpb.Struct {
	if x != nil {
		return x.Params
	}
	return nil
}

type SearchPartsResponse struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	Resolved       *structpb.Struct       `protobuf:"bytes,5,opt,name=resolved,proto3" json:"resolved,omitempty"`
	TransportParts []*TransportPart       `protobuf:"bytes,1,rep,name=transport_parts,json=transportParts,proto3" json:"transport_parts,omitempty"`
	HotelParts     []*HotelPart           `protobuf:"bytes,2,rep,name=hotel_parts,json=hotelParts,proto3" json:"hotel_parts,omitempty"`
	Errors         []*ProviderError       `protobuf:"bytes,3,rep,name=errors,proto3" json:"errors,omitempty"`
	Limitations    []string               `protobuf:"bytes,4,rep,name=limitations,proto3" json:"limitations,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SearchPartsResponse) Reset() {
	*x = SearchPartsResponse{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchPartsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchPartsResponse) ProtoMessage() {}

func (x *SearchPartsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchPartsResponse) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{2}
}

func (x *SearchPartsResponse) GetResolved() *structpb.Struct {
	if x != nil {
		return x.Resolved
	}
	return nil
}

func (x *SearchPartsResponse) GetTransportParts() []*TransportPart {
	if x != nil {
		return x.TransportParts
	}
	return nil
}

func (x *SearchPartsResponse) GetHotelParts() []*HotelPart {
	if x != nil {
		return x.HotelParts
	}
	return nil
}

func (x *SearchPartsResponse) GetErrors() []*ProviderError {
	if x != nil {
		return x.Errors
	}
	return nil
}

func (x *SearchPartsResponse) GetLimitations() []string {
	if x != nil {
		return x.Limitations
	}
	return nil
}

type Money struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	AmountMinor   int64                  `protobuf:"varint,1,opt,name=amount_minor,json=amountMinor,proto3" json:"amount_minor,omitempty"`
	Currency      string                 `protobuf:"bytes,2,opt,name=currency,proto3" json:"currency,omitempty"`
	Scope         string                 `protobuf:"bytes,3,opt,name=scope,proto3" json:"scope,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Money) Reset() {
	*x = Money{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Money) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Money) ProtoMessage() {}

func (x *Money) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Money) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{3}
}

func (x *Money) GetAmountMinor() int64 {
	if x != nil {
		return x.AmountMinor
	}
	return 0
}

func (x *Money) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *Money) GetScope() string {
	if x != nil {
		return x.Scope
	}
	return ""
}

type TransportPart struct {
	state             protoimpl.MessageState `protogen:"open.v1"`
	Id                string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Provider          string                 `protobuf:"bytes,2,opt,name=provider,proto3" json:"provider,omitempty"`
	Source            string                 `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	FetchedAt         string                 `protobuf:"bytes,4,opt,name=fetched_at,json=fetchedAt,proto3" json:"fetched_at,omitempty"`
	Mode              string                 `protobuf:"bytes,5,opt,name=mode,proto3" json:"mode,omitempty"`
	Origin            string                 `protobuf:"bytes,6,opt,name=origin,proto3" json:"origin,omitempty"`
	Destination       string                 `protobuf:"bytes,7,opt,name=destination,proto3" json:"destination,omitempty"`
	DepartureAt       string                 `protobuf:"bytes,8,opt,name=departure_at,json=departureAt,proto3" json:"departure_at,omitempty"`
	ArrivalAt         string                 `protobuf:"bytes,9,opt,name=arrival_at,json=arrivalAt,proto3" json:"arrival_at,omitempty"`
	Price             *Money                 `protobuf:"bytes,10,opt,name=price,proto3" json:"price,omitempty"`
	BookingUrl        string                 `protobuf:"bytes,11,opt,name=booking_url,json=bookingUrl,proto3" json:"booking_url,omitempty"`
	Warnings          []string               `protobuf:"bytes,12,rep,name=warnings,proto3" json:"warnings,omitempty"`
	ReturnDepartureAt string                 `protobuf:"bytes,13,opt,name=return_departure_at,json=returnDepartureAt,proto3" json:"return_departure_at,omitempty"`
	ReturnArrivalAt   string                 `protobuf:"bytes,14,opt,name=return_arrival_at,json=returnArrivalAt,proto3" json:"return_arrival_at,omitempty"`
	OriginCity        string                 `protobuf:"bytes,15,opt,name=origin_city,json=originCity,proto3" json:"origin_city,omitempty"`
	DestinationCity   string                 `protobuf:"bytes,16,opt,name=destination_city,json=destinationCity,proto3" json:"destination_city,omitempty"`
	Transfers         int32                  `protobuf:"varint,17,opt,name=transfers,proto3" json:"transfers,omitempty"`
	ReturnTransfers   int32                  `protobuf:"varint,18,opt,name=return_transfers,json=returnTransfers,proto3" json:"return_transfers,omitempty"`
	Legs              []*TransferLeg         `protobuf:"bytes,19,rep,name=legs,proto3" json:"legs,omitempty"`
	unknownFields     protoimpl.UnknownFields
	sizeCache         protoimpl.SizeCache
}

func (x *TransportPart) Reset() {
	*x = TransportPart{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransportPart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransportPart) ProtoMessage() {}

func (x *TransportPart) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TransportPart) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{4}
}

func (x *TransportPart) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TransportPart) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *TransportPart) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *TransportPart) GetFetchedAt() string {
	if x != nil {
		return x.FetchedAt
	}
	return ""
}

func (x *TransportPart) GetMode() string {
	if x != nil {
		return x.Mode
	}
	return ""
}

func (x *TransportPart) GetOrigin() string {
	if x != nil {
		return x.Origin
	}
	return ""
}

func (x *TransportPart) GetDestination() string {
	if x != nil {
		return x.Destination
	}
	return ""
}

func (x *TransportPart) GetDepartureAt() string {
	if x != nil {
		return x.DepartureAt
	}
	return ""
}

func (x *TransportPart) GetArrivalAt() string {
	if x != nil {
		return x.ArrivalAt
	}
	return ""
}

func (x *TransportPart) GetPrice() *Money {
	if x != nil {
		return x.Price
	}
	return nil
}

func (x *TransportPart) GetBookingUrl() string {
	if x != nil {
		return x.BookingUrl
	}
	return ""
}

func (x *TransportPart) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

func (x *TransportPart) GetReturnDepartureAt() string {
	if x != nil {
		return x.ReturnDepartureAt
	}
	return ""
}

func (x *TransportPart) GetReturnArrivalAt() string {
	if x != nil {
		return x.ReturnArrivalAt
	}
	return ""
}

func (x *TransportPart) GetOriginCity() string {
	if x != nil {
		return x.OriginCity
	}
	return ""
}

func (x *TransportPart) GetDestinationCity() string {
	if x != nil {
		return x.DestinationCity
	}
	return ""
}

func (x *TransportPart) GetTransfers() int32 {
	if x != nil {
		return x.Transfers
	}
	return 0
}

func (x *TransportPart) GetReturnTransfers() int32 {
	if x != nil {
		return x.ReturnTransfers
	}
	return 0
}

func (x *TransportPart) GetLegs() []*TransferLeg {
	if x != nil {
		return x.Legs
	}
	return nil
}

type TransferLeg struct {
	state              protoimpl.MessageState `protogen:"open.v1"`
	OriginStation      string                 `protobuf:"bytes,1,opt,name=origin_station,json=originStation,proto3" json:"origin_station,omitempty"`
	DestinationStation string                 `protobuf:"bytes,2,opt,name=destination_station,json=destinationStation,proto3" json:"destination_station,omitempty"`
	DepartureAt        string                 `protobuf:"bytes,3,opt,name=departure_at,json=departureAt,proto3" json:"departure_at,omitempty"`
	ArrivalAt          string                 `protobuf:"bytes,4,opt,name=arrival_at,json=arrivalAt,proto3" json:"arrival_at,omitempty"`
	Mode               string                 `protobuf:"bytes,5,opt,name=mode,proto3" json:"mode,omitempty"`
	ServiceId          string                 `protobuf:"bytes,6,opt,name=service_id,json=serviceId,proto3" json:"service_id,omitempty"`
	unknownFields      protoimpl.UnknownFields
	sizeCache          protoimpl.SizeCache
}

func (x *TransferLeg) Reset() {
	*x = TransferLeg{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TransferLeg) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransferLeg) ProtoMessage() {}

func (x *TransferLeg) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TransferLeg) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{5}
}

func (x *TransferLeg) GetOriginStation() string {
	if x != nil {
		return x.OriginStation
	}
	return ""
}

func (x *TransferLeg) GetDestinationStation() string {
	if x != nil {
		return x.DestinationStation
	}
	return ""
}

func (x *TransferLeg) GetDepartureAt() string {
	if x != nil {
		return x.DepartureAt
	}
	return ""
}

func (x *TransferLeg) GetArrivalAt() string {
	if x != nil {
		return x.ArrivalAt
	}
	return ""
}

func (x *TransferLeg) GetMode() string {
	if x != nil {
		return x.Mode
	}
	return ""
}

func (x *TransferLeg) GetServiceId() string {
	if x != nil {
		return x.ServiceId
	}
	return ""
}

type HotelPart struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Id               string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Provider         string                 `protobuf:"bytes,2,opt,name=provider,proto3" json:"provider,omitempty"`
	Source           string                 `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	FetchedAt        string                 `protobuf:"bytes,4,opt,name=fetched_at,json=fetchedAt,proto3" json:"fetched_at,omitempty"`
	City             string                 `protobuf:"bytes,5,opt,name=city,proto3" json:"city,omitempty"`
	Name             string                 `protobuf:"bytes,6,opt,name=name,proto3" json:"name,omitempty"`
	CheckIn          string                 `protobuf:"bytes,7,opt,name=check_in,json=checkIn,proto3" json:"check_in,omitempty"`
	CheckOut         string                 `protobuf:"bytes,8,opt,name=check_out,json=checkOut,proto3" json:"check_out,omitempty"`
	Stars            int32                  `protobuf:"varint,9,opt,name=stars,proto3" json:"stars,omitempty"`
	Price            *Money                 `protobuf:"bytes,10,opt,name=price,proto3" json:"price,omitempty"`
	FreeCancellation bool                   `protobuf:"varint,11,opt,name=free_cancellation,json=freeCancellation,proto3" json:"free_cancellation,omitempty"`
	Warnings         []string               `protobuf:"bytes,12,rep,name=warnings,proto3" json:"warnings,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *HotelPart) Reset() {
	*x = HotelPart{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HotelPart) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HotelPart) ProtoMessage() {}

func (x *HotelPart) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*HotelPart) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{6}
}

func (x *HotelPart) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *HotelPart) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *HotelPart) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *HotelPart) GetFetchedAt() string {
	if x != nil {
		return x.FetchedAt
	}
	return ""
}

func (x *HotelPart) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *HotelPart) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *HotelPart) GetCheckIn() string {
	if x != nil {
		return x.CheckIn
	}
	return ""
}

func (x *HotelPart) GetCheckOut() string {
	if x != nil {
		return x.CheckOut
	}
	return ""
}

func (x *HotelPart) GetStars() int32 {
	if x != nil {
		return x.Stars
	}
	return 0
}

func (x *HotelPart) GetPrice() *Money {
	if x != nil {
		return x.Price
	}
	return nil
}

func (x *HotelPart) GetFreeCancellation() bool {
	if x != nil {
		return x.FreeCancellation
	}
	return false
}

func (x *HotelPart) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

type ProviderError struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Provider      string                 `protobuf:"bytes,1,opt,name=provider,proto3" json:"provider,omitempty"`
	Kind          string                 `protobuf:"bytes,2,opt,name=kind,proto3" json:"kind,omitempty"`
	Code          string                 `protobuf:"bytes,3,opt,name=code,proto3" json:"code,omitempty"`
	Message       string                 `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	Retryable     bool                   `protobuf:"varint,5,opt,name=retryable,proto3" json:"retryable,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProviderError) Reset() {
	*x = ProviderError{}
	mi := &file_api_provider_v1_parts_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProviderError) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProviderError) ProtoMessage() {}

func (x *ProviderError) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_parts_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ProviderError) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_parts_proto_rawDescGZIP(), []int{7}
}

func (x *ProviderError) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *ProviderError) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *ProviderError) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *ProviderError) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ProviderError) GetRetryable() bool {
	if x != nil {
		return x.Retryable
	}
	return false
}

var File_api_provider_v1_parts_proto protoreflect.FileDescriptor

const file_api_provider_v1_parts_proto_rawDesc = "" +
	"\n" +
	"\x1bapi/provider/v1/parts.proto\x12\x13godnota.provider.v1\x1a\x1cgoogle/protobuf/struct.proto\"z\n" +
	"\x12SearchPartsRequest\x12/\n" +
	"\x06config\x18\x01 \x01(\v2\x17.google.protobuf.StructR\x06config\x123\n" +
	"\bresolved\x18\x02 \x01(\v2\x17.google.protobuf.StructR\bresolved\"y\n" +
	"\x0eProviderConfig\x12\x1a\n" +
	"\bprovider\x18\x01 \x01(\tR\bprovider\x12\x1a\n" +
	"\bendpoint\x18\x02 \x01(\tR\bendpoint\x12/\n" +
	"\x06params\x18\x03 \x01(\v2\x17.google.protobuf.StructR\x06params\"\xb6\x02\n" +
	"\x13SearchPartsResponse\x123\n" +
	"\bresolved\x18\x05 \x01(\v2\x17.google.protobuf.StructR\bresolved\x12K\n" +
	"\x0ftransport_parts\x18\x01 \x03(\v2\".godnota.provider.v1.TransportPartR\x0etransportParts\x12?\n" +
	"\vhotel_parts\x18\x02 \x03(\v2\x1e.godnota.provider.v1.HotelPartR\n" +
	"hotelParts\x12:\n" +
	"\x06errors\x18\x03 \x03(\v2\".godnota.provider.v1.ProviderErrorR\x06errors\x12 \n" +
	"\vlimitations\x18\x04 \x03(\tR\vlimitations\"\\\n" +
	"\x05Money\x12!\n" +
	"\famount_minor\x18\x01 \x01(\x03R\vamountMinor\x12\x1a\n" +
	"\bcurrency\x18\x02 \x01(\tR\bcurrency\x12\x14\n" +
	"\x05scope\x18\x03 \x01(\tR\x05scope\"\x98\x05\n" +
	"\rTransportPart\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1a\n" +
	"\bprovider\x18\x02 \x01(\tR\bprovider\x12\x16\n" +
	"\x06source\x18\x03 \x01(\tR\x06source\x12\x1d\n" +
	"\n" +
	"fetched_at\x18\x04 \x01(\tR\tfetchedAt\x12\x12\n" +
	"\x04mode\x18\x05 \x01(\tR\x04mode\x12\x16\n" +
	"\x06origin\x18\x06 \x01(\tR\x06origin\x12 \n" +
	"\vdestination\x18\a \x01(\tR\vdestination\x12!\n" +
	"\fdeparture_at\x18\b \x01(\tR\vdepartureAt\x12\x1d\n" +
	"\n" +
	"arrival_at\x18\t \x01(\tR\tarrivalAt\x120\n" +
	"\x05price\x18\n" +
	" \x01(\v2\x1a.godnota.provider.v1.MoneyR\x05price\x12\x1f\n" +
	"\vbooking_url\x18\v \x01(\tR\n" +
	"bookingUrl\x12\x1a\n" +
	"\bwarnings\x18\f \x03(\tR\bwarnings\x12.\n" +
	"\x13return_departure_at\x18\r \x01(\tR\x11returnDepartureAt\x12*\n" +
	"\x11return_arrival_at\x18\x0e \x01(\tR\x0freturnArrivalAt\x12\x1f\n" +
	"\vorigin_city\x18\x0f \x01(\tR\n" +
	"originCity\x12)\n" +
	"\x10destination_city\x18\x10 \x01(\tR\x0fdestinationCity\x12\x1c\n" +
	"\ttransfers\x18\x11 \x01(\x05R\ttransfers\x12)\n" +
	"\x10return_transfers\x18\x12 \x01(\x05R\x0freturnTransfers\x124\n" +
	"\x04legs\x18\x13 \x03(\v2 .godnota.provider.v1.TransferLegR\x04legs\"\xda\x01\n" +
	"\vTransferLeg\x12%\n" +
	"\x0eorigin_station\x18\x01 \x01(\tR\roriginStation\x12/\n" +
	"\x13destination_station\x18\x02 \x01(\tR\x12destinationStation\x12!\n" +
	"\fdeparture_at\x18\x03 \x01(\tR\vdepartureAt\x12\x1d\n" +
	"\n" +
	"arrival_at\x18\x04 \x01(\tR\tarrivalAt\x12\x12\n" +
	"\x04mode\x18\x05 \x01(\tR\x04mode\x12\x1d\n" +
	"\n" +
	"service_id\x18\x06 \x01(\tR\tserviceId\"\xdf\x02\n" +
	"\tHotelPart\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1a\n" +
	"\bprovider\x18\x02 \x01(\tR\bprovider\x12\x16\n" +
	"\x06source\x18\x03 \x01(\tR\x06source\x12\x1d\n" +
	"\n" +
	"fetched_at\x18\x04 \x01(\tR\tfetchedAt\x12\x12\n" +
	"\x04city\x18\x05 \x01(\tR\x04city\x12\x12\n" +
	"\x04name\x18\x06 \x01(\tR\x04name\x12\x19\n" +
	"\bcheck_in\x18\a \x01(\tR\acheckIn\x12\x1b\n" +
	"\tcheck_out\x18\b \x01(\tR\bcheckOut\x12\x14\n" +
	"\x05stars\x18\t \x01(\x05R\x05stars\x120\n" +
	"\x05price\x18\n" +
	" \x01(\v2\x1a.godnota.provider.v1.MoneyR\x05price\x12+\n" +
	"\x11free_cancellation\x18\v \x01(\bR\x10freeCancellation\x12\x1a\n" +
	"\bwarnings\x18\f \x03(\tR\bwarnings\"\x8b\x01\n" +
	"\rProviderError\x12\x1a\n" +
	"\bprovider\x18\x01 \x01(\tR\bprovider\x12\x12\n" +
	"\x04kind\x18\x02 \x01(\tR\x04kind\x12\x12\n" +
	"\x04code\x18\x03 \x01(\tR\x04code\x12\x18\n" +
	"\amessage\x18\x04 \x01(\tR\amessage\x12\x1c\n" +
	"\tretryable\x18\x05 \x01(\bR\tretryableB9Z7github.com/godnota/pkg/generated/provider/v1;providerv1b\x06proto3"

var (
	file_api_provider_v1_parts_proto_rawDescOnce sync.Once
	file_api_provider_v1_parts_proto_rawDescData []byte
)

func file_api_provider_v1_parts_proto_rawDescGZIP() []byte {
	file_api_provider_v1_parts_proto_rawDescOnce.Do(func() {
		file_api_provider_v1_parts_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_provider_v1_parts_proto_rawDesc), len(file_api_provider_v1_parts_proto_rawDesc)))
	})
	return file_api_provider_v1_parts_proto_rawDescData
}

var file_api_provider_v1_parts_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_api_provider_v1_parts_proto_goTypes = []any{
	(*SearchPartsRequest)(nil),
	(*ProviderConfig)(nil),
	(*SearchPartsResponse)(nil),
	(*Money)(nil),
	(*TransportPart)(nil),
	(*TransferLeg)(nil),
	(*HotelPart)(nil),
	(*ProviderError)(nil),
	(*structpb.Struct)(nil),
}
var file_api_provider_v1_parts_proto_depIdxs = []int32{
	8,
	8,
	8,
	8,
	4,
	6,
	7,
	3,
	5,
	3,
	10,
	10,
	10,
	10,
	0,
}

func init() { file_api_provider_v1_parts_proto_init() }
func file_api_provider_v1_parts_proto_init() {
	if File_api_provider_v1_parts_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_provider_v1_parts_proto_rawDesc), len(file_api_provider_v1_parts_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_provider_v1_parts_proto_goTypes,
		DependencyIndexes: file_api_provider_v1_parts_proto_depIdxs,
		MessageInfos:      file_api_provider_v1_parts_proto_msgTypes,
	}.Build()
	File_api_provider_v1_parts_proto = out.File
	file_api_provider_v1_parts_proto_goTypes = nil
	file_api_provider_v1_parts_proto_depIdxs = nil
}
