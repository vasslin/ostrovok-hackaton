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

type SearchInput struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	TripRequest   *structpb.Struct       `protobuf:"bytes,1,opt,name=trip_request,json=tripRequest,proto3" json:"trip_request,omitempty"`
	Resolved      *structpb.Struct       `protobuf:"bytes,2,opt,name=resolved,proto3" json:"resolved,omitempty"`
	InputFormat   string                 `protobuf:"bytes,3,opt,name=input_format,json=inputFormat,proto3" json:"input_format,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchInput) Reset() {
	*x = SearchInput{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchInput) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchInput) ProtoMessage() {}

func (x *SearchInput) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchInput) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{0}
}

func (x *SearchInput) GetTripRequest() *structpb.Struct {
	if x != nil {
		return x.TripRequest
	}
	return nil
}

func (x *SearchInput) GetResolved() *structpb.Struct {
	if x != nil {
		return x.Resolved
	}
	return nil
}

func (x *SearchInput) GetInputFormat() string {
	if x != nil {
		return x.InputFormat
	}
	return ""
}

type Offer struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Provider      string                 `protobuf:"bytes,2,opt,name=provider,proto3" json:"provider,omitempty"`
	Kind          string                 `protobuf:"bytes,3,opt,name=kind,proto3" json:"kind,omitempty"`
	Title         string                 `protobuf:"bytes,4,opt,name=title,proto3" json:"title,omitempty"`
	Origin        string                 `protobuf:"bytes,5,opt,name=origin,proto3" json:"origin,omitempty"`
	Destination   string                 `protobuf:"bytes,6,opt,name=destination,proto3" json:"destination,omitempty"`
	Departure     string                 `protobuf:"bytes,7,opt,name=departure,proto3" json:"departure,omitempty"`
	Arrival       string                 `protobuf:"bytes,8,opt,name=arrival,proto3" json:"arrival,omitempty"`
	PriceMinor    *int64                 `protobuf:"varint,9,opt,name=price_minor,json=priceMinor,proto3,oneof" json:"price_minor,omitempty"`
	Currency      string                 `protobuf:"bytes,10,opt,name=currency,proto3" json:"currency,omitempty"`
	PriceScope    string                 `protobuf:"bytes,11,opt,name=price_scope,json=priceScope,proto3" json:"price_scope,omitempty"`
	SourceStatus  string                 `protobuf:"bytes,12,opt,name=source_status,json=sourceStatus,proto3" json:"source_status,omitempty"`
	SourceUrl     string                 `protobuf:"bytes,13,opt,name=source_url,json=sourceUrl,proto3" json:"source_url,omitempty"`
	Attribution   string                 `protobuf:"bytes,14,opt,name=attribution,proto3" json:"attribution,omitempty"`
	Latitude      *float64               `protobuf:"fixed64,15,opt,name=latitude,proto3,oneof" json:"latitude,omitempty"`
	Longitude     *float64               `protobuf:"fixed64,16,opt,name=longitude,proto3,oneof" json:"longitude,omitempty"`
	ObservedAt    string                 `protobuf:"bytes,17,opt,name=observed_at,json=observedAt,proto3" json:"observed_at,omitempty"`
	Description   string                 `protobuf:"bytes,18,opt,name=description,proto3" json:"description,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Offer) Reset() {
	*x = Offer{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Offer) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Offer) ProtoMessage() {}

func (x *Offer) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Offer) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{1}
}

func (x *Offer) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Offer) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *Offer) GetKind() string {
	if x != nil {
		return x.Kind
	}
	return ""
}

func (x *Offer) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *Offer) GetOrigin() string {
	if x != nil {
		return x.Origin
	}
	return ""
}

func (x *Offer) GetDestination() string {
	if x != nil {
		return x.Destination
	}
	return ""
}

func (x *Offer) GetDeparture() string {
	if x != nil {
		return x.Departure
	}
	return ""
}

func (x *Offer) GetArrival() string {
	if x != nil {
		return x.Arrival
	}
	return ""
}

func (x *Offer) GetPriceMinor() int64 {
	if x != nil && x.PriceMinor != nil {
		return *x.PriceMinor
	}
	return 0
}

func (x *Offer) GetCurrency() string {
	if x != nil {
		return x.Currency
	}
	return ""
}

func (x *Offer) GetPriceScope() string {
	if x != nil {
		return x.PriceScope
	}
	return ""
}

func (x *Offer) GetSourceStatus() string {
	if x != nil {
		return x.SourceStatus
	}
	return ""
}

func (x *Offer) GetSourceUrl() string {
	if x != nil {
		return x.SourceUrl
	}
	return ""
}

func (x *Offer) GetAttribution() string {
	if x != nil {
		return x.Attribution
	}
	return ""
}

func (x *Offer) GetLatitude() float64 {
	if x != nil && x.Latitude != nil {
		return *x.Latitude
	}
	return 0
}

func (x *Offer) GetLongitude() float64 {
	if x != nil && x.Longitude != nil {
		return *x.Longitude
	}
	return 0
}

func (x *Offer) GetObservedAt() string {
	if x != nil {
		return x.ObservedAt
	}
	return ""
}

func (x *Offer) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

type ProviderIssue struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Provider      string                 `protobuf:"bytes,1,opt,name=provider,proto3" json:"provider,omitempty"`
	Code          string                 `protobuf:"bytes,2,opt,name=code,proto3" json:"code,omitempty"`
	Message       string                 `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ProviderIssue) Reset() {
	*x = ProviderIssue{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ProviderIssue) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ProviderIssue) ProtoMessage() {}

func (x *ProviderIssue) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ProviderIssue) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{2}
}

func (x *ProviderIssue) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *ProviderIssue) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

func (x *ProviderIssue) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type SearchTransportRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Input         *SearchInput           `protobuf:"bytes,1,opt,name=input,proto3" json:"input,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchTransportRequest) Reset() {
	*x = SearchTransportRequest{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchTransportRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchTransportRequest) ProtoMessage() {}

func (x *SearchTransportRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchTransportRequest) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{3}
}

func (x *SearchTransportRequest) GetInput() *SearchInput {
	if x != nil {
		return x.Input
	}
	return nil
}

type SearchTransportResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Offers        []*Offer               `protobuf:"bytes,1,rep,name=offers,proto3" json:"offers,omitempty"`
	Issues        []*ProviderIssue       `protobuf:"bytes,2,rep,name=issues,proto3" json:"issues,omitempty"`
	Missing       []string               `protobuf:"bytes,3,rep,name=missing,proto3" json:"missing,omitempty"`
	Limitations   []string               `protobuf:"bytes,4,rep,name=limitations,proto3" json:"limitations,omitempty"`
	Discovery     *structpb.Struct       `protobuf:"bytes,5,opt,name=discovery,proto3" json:"discovery,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchTransportResponse) Reset() {
	*x = SearchTransportResponse{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchTransportResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchTransportResponse) ProtoMessage() {}

func (x *SearchTransportResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchTransportResponse) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{4}
}

func (x *SearchTransportResponse) GetOffers() []*Offer {
	if x != nil {
		return x.Offers
	}
	return nil
}

func (x *SearchTransportResponse) GetIssues() []*ProviderIssue {
	if x != nil {
		return x.Issues
	}
	return nil
}

func (x *SearchTransportResponse) GetMissing() []string {
	if x != nil {
		return x.Missing
	}
	return nil
}

func (x *SearchTransportResponse) GetLimitations() []string {
	if x != nil {
		return x.Limitations
	}
	return nil
}

func (x *SearchTransportResponse) GetDiscovery() *structpb.Struct {
	if x != nil {
		return x.Discovery
	}
	return nil
}

type SearchHotelsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Input         *SearchInput           `protobuf:"bytes,1,opt,name=input,proto3" json:"input,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchHotelsRequest) Reset() {
	*x = SearchHotelsRequest{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchHotelsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchHotelsRequest) ProtoMessage() {}

func (x *SearchHotelsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchHotelsRequest) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{5}
}

func (x *SearchHotelsRequest) GetInput() *SearchInput {
	if x != nil {
		return x.Input
	}
	return nil
}

type SearchHotelsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Offers        []*Offer               `protobuf:"bytes,1,rep,name=offers,proto3" json:"offers,omitempty"`
	Issues        []*ProviderIssue       `protobuf:"bytes,2,rep,name=issues,proto3" json:"issues,omitempty"`
	Missing       []string               `protobuf:"bytes,3,rep,name=missing,proto3" json:"missing,omitempty"`
	Limitations   []string               `protobuf:"bytes,4,rep,name=limitations,proto3" json:"limitations,omitempty"`
	Discovery     *structpb.Struct       `protobuf:"bytes,5,opt,name=discovery,proto3" json:"discovery,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchHotelsResponse) Reset() {
	*x = SearchHotelsResponse{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchHotelsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchHotelsResponse) ProtoMessage() {}

func (x *SearchHotelsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchHotelsResponse) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{6}
}

func (x *SearchHotelsResponse) GetOffers() []*Offer {
	if x != nil {
		return x.Offers
	}
	return nil
}

func (x *SearchHotelsResponse) GetIssues() []*ProviderIssue {
	if x != nil {
		return x.Issues
	}
	return nil
}

func (x *SearchHotelsResponse) GetMissing() []string {
	if x != nil {
		return x.Missing
	}
	return nil
}

func (x *SearchHotelsResponse) GetLimitations() []string {
	if x != nil {
		return x.Limitations
	}
	return nil
}

func (x *SearchHotelsResponse) GetDiscovery() *structpb.Struct {
	if x != nil {
		return x.Discovery
	}
	return nil
}

type SearchActivitiesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Input         *SearchInput           `protobuf:"bytes,1,opt,name=input,proto3" json:"input,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchActivitiesRequest) Reset() {
	*x = SearchActivitiesRequest{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchActivitiesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchActivitiesRequest) ProtoMessage() {}

func (x *SearchActivitiesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchActivitiesRequest) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{7}
}

func (x *SearchActivitiesRequest) GetInput() *SearchInput {
	if x != nil {
		return x.Input
	}
	return nil
}

type SearchActivitiesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Offers        []*Offer               `protobuf:"bytes,1,rep,name=offers,proto3" json:"offers,omitempty"`
	Issues        []*ProviderIssue       `protobuf:"bytes,2,rep,name=issues,proto3" json:"issues,omitempty"`
	Missing       []string               `protobuf:"bytes,3,rep,name=missing,proto3" json:"missing,omitempty"`
	Limitations   []string               `protobuf:"bytes,4,rep,name=limitations,proto3" json:"limitations,omitempty"`
	Discovery     *structpb.Struct       `protobuf:"bytes,5,opt,name=discovery,proto3" json:"discovery,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchActivitiesResponse) Reset() {
	*x = SearchActivitiesResponse{}
	mi := &file_api_provider_v1_provider_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchActivitiesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchActivitiesResponse) ProtoMessage() {}

func (x *SearchActivitiesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_provider_v1_provider_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchActivitiesResponse) Descriptor() ([]byte, []int) {
	return file_api_provider_v1_provider_proto_rawDescGZIP(), []int{8}
}

func (x *SearchActivitiesResponse) GetOffers() []*Offer {
	if x != nil {
		return x.Offers
	}
	return nil
}

func (x *SearchActivitiesResponse) GetIssues() []*ProviderIssue {
	if x != nil {
		return x.Issues
	}
	return nil
}

func (x *SearchActivitiesResponse) GetMissing() []string {
	if x != nil {
		return x.Missing
	}
	return nil
}

func (x *SearchActivitiesResponse) GetLimitations() []string {
	if x != nil {
		return x.Limitations
	}
	return nil
}

func (x *SearchActivitiesResponse) GetDiscovery() *structpb.Struct {
	if x != nil {
		return x.Discovery
	}
	return nil
}

var File_api_provider_v1_provider_proto protoreflect.FileDescriptor

const file_api_provider_v1_provider_proto_rawDesc = "" +
	"\n" +
	"\x1eapi/provider/v1/provider.proto\x12\x13godnota.provider.v1\x1a\x1cgoogle/protobuf/struct.proto\x1a\x1bapi/provider/v1/parts.proto\"\xa1\x01\n" +
	"\vSearchInput\x12:\n" +
	"\ftrip_request\x18\x01 \x01(\v2\x17.google.protobuf.StructR\vtripRequest\x123\n" +
	"\bresolved\x18\x02 \x01(\v2\x17.google.protobuf.StructR\bresolved\x12!\n" +
	"\finput_format\x18\x03 \x01(\tR\vinputFormat\"\xca\x04\n" +
	"\x05Offer\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x1a\n" +
	"\bprovider\x18\x02 \x01(\tR\bprovider\x12\x12\n" +
	"\x04kind\x18\x03 \x01(\tR\x04kind\x12\x14\n" +
	"\x05title\x18\x04 \x01(\tR\x05title\x12\x16\n" +
	"\x06origin\x18\x05 \x01(\tR\x06origin\x12 \n" +
	"\vdestination\x18\x06 \x01(\tR\vdestination\x12\x1c\n" +
	"\tdeparture\x18\a \x01(\tR\tdeparture\x12\x18\n" +
	"\aarrival\x18\b \x01(\tR\aarrival\x12$\n" +
	"\vprice_minor\x18\t \x01(\x03H\x00R\n" +
	"priceMinor\x88\x01\x01\x12\x1a\n" +
	"\bcurrency\x18\n" +
	" \x01(\tR\bcurrency\x12\x1f\n" +
	"\vprice_scope\x18\v \x01(\tR\n" +
	"priceScope\x12#\n" +
	"\rsource_status\x18\f \x01(\tR\fsourceStatus\x12\x1d\n" +
	"\n" +
	"source_url\x18\r \x01(\tR\tsourceUrl\x12 \n" +
	"\vattribution\x18\x0e \x01(\tR\vattribution\x12\x1f\n" +
	"\blatitude\x18\x0f \x01(\x01H\x01R\blatitude\x88\x01\x01\x12!\n" +
	"\tlongitude\x18\x10 \x01(\x01H\x02R\tlongitude\x88\x01\x01\x12\x1f\n" +
	"\vobserved_at\x18\x11 \x01(\tR\n" +
	"observedAt\x12 \n" +
	"\vdescription\x18\x12 \x01(\tR\vdescriptionB\x0e\n" +
	"\f_price_minorB\v\n" +
	"\t_latitudeB\f\n" +
	"\n" +
	"_longitude\"Y\n" +
	"\rProviderIssue\x12\x1a\n" +
	"\bprovider\x18\x01 \x01(\tR\bprovider\x12\x12\n" +
	"\x04code\x18\x02 \x01(\tR\x04code\x12\x18\n" +
	"\amessage\x18\x03 \x01(\tR\amessage\"P\n" +
	"\x16SearchTransportRequest\x126\n" +
	"\x05input\x18\x01 \x01(\v2 .godnota.provider.v1.SearchInputR\x05input\"\xfc\x01\n" +
	"\x17SearchTransportResponse\x122\n" +
	"\x06offers\x18\x01 \x03(\v2\x1a.godnota.provider.v1.OfferR\x06offers\x12:\n" +
	"\x06issues\x18\x02 \x03(\v2\".godnota.provider.v1.ProviderIssueR\x06issues\x12\x18\n" +
	"\amissing\x18\x03 \x03(\tR\amissing\x12 \n" +
	"\vlimitations\x18\x04 \x03(\tR\vlimitations\x125\n" +
	"\tdiscovery\x18\x05 \x01(\v2\x17.google.protobuf.StructR\tdiscovery\"M\n" +
	"\x13SearchHotelsRequest\x126\n" +
	"\x05input\x18\x01 \x01(\v2 .godnota.provider.v1.SearchInputR\x05input\"\xf9\x01\n" +
	"\x14SearchHotelsResponse\x122\n" +
	"\x06offers\x18\x01 \x03(\v2\x1a.godnota.provider.v1.OfferR\x06offers\x12:\n" +
	"\x06issues\x18\x02 \x03(\v2\".godnota.provider.v1.ProviderIssueR\x06issues\x12\x18\n" +
	"\amissing\x18\x03 \x03(\tR\amissing\x12 \n" +
	"\vlimitations\x18\x04 \x03(\tR\vlimitations\x125\n" +
	"\tdiscovery\x18\x05 \x01(\v2\x17.google.protobuf.StructR\tdiscovery\"Q\n" +
	"\x17SearchActivitiesRequest\x126\n" +
	"\x05input\x18\x01 \x01(\v2 .godnota.provider.v1.SearchInputR\x05input\"\xfd\x01\n" +
	"\x18SearchActivitiesResponse\x122\n" +
	"\x06offers\x18\x01 \x03(\v2\x1a.godnota.provider.v1.OfferR\x06offers\x12:\n" +
	"\x06issues\x18\x02 \x03(\v2\".godnota.provider.v1.ProviderIssueR\x06issues\x12\x18\n" +
	"\amissing\x18\x03 \x03(\tR\amissing\x12 \n" +
	"\vlimitations\x18\x04 \x03(\tR\vlimitations\x125\n" +
	"\tdiscovery\x18\x05 \x01(\v2\x17.google.protobuf.StructR\tdiscovery2\xfc\x03\n" +
	"\x0fProviderService\x12C\n" +
	"\x0fGetCapabilities\x12\x17.google.protobuf.Struct\x1a\x17.google.protobuf.Struct\x12`\n" +
	"\vSearchParts\x12'.godnota.provider.v1.SearchPartsRequest\x1a(.godnota.provider.v1.SearchPartsResponse\x12l\n" +
	"\x0fSearchTransport\x12+.godnota.provider.v1.SearchTransportRequest\x1a,.godnota.provider.v1.SearchTransportResponse\x12c\n" +
	"\fSearchHotels\x12(.godnota.provider.v1.SearchHotelsRequest\x1a).godnota.provider.v1.SearchHotelsResponse\x12o\n" +
	"\x10SearchActivities\x12,.godnota.provider.v1.SearchActivitiesRequest\x1a-.godnota.provider.v1.SearchActivitiesResponseB9Z7github.com/godnota/pkg/generated/provider/v1;providerv1b\x06proto3"

var (
	file_api_provider_v1_provider_proto_rawDescOnce sync.Once
	file_api_provider_v1_provider_proto_rawDescData []byte
)

func file_api_provider_v1_provider_proto_rawDescGZIP() []byte {
	file_api_provider_v1_provider_proto_rawDescOnce.Do(func() {
		file_api_provider_v1_provider_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_provider_v1_provider_proto_rawDesc), len(file_api_provider_v1_provider_proto_rawDesc)))
	})
	return file_api_provider_v1_provider_proto_rawDescData
}

var file_api_provider_v1_provider_proto_msgTypes = make([]protoimpl.MessageInfo, 9)
var file_api_provider_v1_provider_proto_goTypes = []any{
	(*SearchInput)(nil),
	(*Offer)(nil),
	(*ProviderIssue)(nil),
	(*SearchTransportRequest)(nil),
	(*SearchTransportResponse)(nil),
	(*SearchHotelsRequest)(nil),
	(*SearchHotelsResponse)(nil),
	(*SearchActivitiesRequest)(nil),
	(*SearchActivitiesResponse)(nil),
	(*structpb.Struct)(nil),
	(*SearchPartsRequest)(nil),
	(*SearchPartsResponse)(nil),
}
var file_api_provider_v1_provider_proto_depIdxs = []int32{
	9,
	9,
	0,
	1,
	2,
	9,
	0,
	1,
	2,
	9,
	0,
	1,
	2,
	9,
	9,
	10,
	3,
	5,
	7,
	9,
	11,
	4,
	6,
	8,
	19,
	14,
	14,
	14,
	0,
}

func init() { file_api_provider_v1_provider_proto_init() }
func file_api_provider_v1_provider_proto_init() {
	if File_api_provider_v1_provider_proto != nil {
		return
	}
	file_api_provider_v1_parts_proto_init()
	file_api_provider_v1_provider_proto_msgTypes[1].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_provider_v1_provider_proto_rawDesc), len(file_api_provider_v1_provider_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   9,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_provider_v1_provider_proto_goTypes,
		DependencyIndexes: file_api_provider_v1_provider_proto_depIdxs,
		MessageInfos:      file_api_provider_v1_provider_proto_msgTypes,
	}.Build()
	File_api_provider_v1_provider_proto = out.File
	file_api_provider_v1_provider_proto_goTypes = nil
	file_api_provider_v1_provider_proto_depIdxs = nil
}
