package travelv1

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

type Config struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Origin        *string                `protobuf:"bytes,1,opt,name=origin,proto3,oneof" json:"origin,omitempty"`
	Destination   *string                `protobuf:"bytes,2,opt,name=destination,proto3,oneof" json:"destination,omitempty"`
	DepartureDate *string                `protobuf:"bytes,3,opt,name=departure_date,json=departureDate,proto3,oneof" json:"departure_date,omitempty"`
	ReturnDate    *string                `protobuf:"bytes,4,opt,name=return_date,json=returnDate,proto3,oneof" json:"return_date,omitempty"`
	Adults        *int32                 `protobuf:"varint,5,opt,name=adults,proto3,oneof" json:"adults,omitempty"`
	ChildrenAges  []int32                `protobuf:"varint,6,rep,packed,name=children_ages,json=childrenAges,proto3" json:"children_ages,omitempty"`
	Budget        *float64               `protobuf:"fixed64,7,opt,name=budget,proto3,oneof" json:"budget,omitempty"`
	Currency      *string                `protobuf:"bytes,8,opt,name=currency,proto3,oneof" json:"currency,omitempty"`
	Preferences   []string               `protobuf:"bytes,9,rep,name=preferences,proto3" json:"preferences,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Config) Reset() {
	*x = Config{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Config) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Config) ProtoMessage() {}

func (x *Config) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Config) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{0}
}

func (x *Config) GetOrigin() string {
	if x != nil && x.Origin != nil {
		return *x.Origin
	}
	return ""
}

func (x *Config) GetDestination() string {
	if x != nil && x.Destination != nil {
		return *x.Destination
	}
	return ""
}

func (x *Config) GetDepartureDate() string {
	if x != nil && x.DepartureDate != nil {
		return *x.DepartureDate
	}
	return ""
}

func (x *Config) GetReturnDate() string {
	if x != nil && x.ReturnDate != nil {
		return *x.ReturnDate
	}
	return ""
}

func (x *Config) GetAdults() int32 {
	if x != nil && x.Adults != nil {
		return *x.Adults
	}
	return 0
}

func (x *Config) GetChildrenAges() []int32 {
	if x != nil {
		return x.ChildrenAges
	}
	return nil
}

func (x *Config) GetBudget() float64 {
	if x != nil && x.Budget != nil {
		return *x.Budget
	}
	return 0
}

func (x *Config) GetCurrency() string {
	if x != nil && x.Currency != nil {
		return *x.Currency
	}
	return ""
}

func (x *Config) GetPreferences() []string {
	if x != nil {
		return x.Preferences
	}
	return nil
}

type ExtractConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	Config        *Config                `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExtractConfigRequest) Reset() {
	*x = ExtractConfigRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExtractConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExtractConfigRequest) ProtoMessage() {}

func (x *ExtractConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ExtractConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{1}
}

func (x *ExtractConfigRequest) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *ExtractConfigRequest) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

type ExtractConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *Config                `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	MissingFields []string               `protobuf:"bytes,2,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Questions     []string               `protobuf:"bytes,3,rep,name=questions,proto3" json:"questions,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ExtractConfigResponse) Reset() {
	*x = ExtractConfigResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ExtractConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExtractConfigResponse) ProtoMessage() {}

func (x *ExtractConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ExtractConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{2}
}

func (x *ExtractConfigResponse) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ExtractConfigResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *ExtractConfigResponse) GetQuestions() []string {
	if x != nil {
		return x.Questions
	}
	return nil
}

type Hotel struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Details       *structpb.Struct       `protobuf:"bytes,3,opt,name=details,proto3" json:"details,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Hotel) Reset() {
	*x = Hotel{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Hotel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Hotel) ProtoMessage() {}

func (x *Hotel) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Hotel) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{3}
}

func (x *Hotel) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Hotel) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Hotel) GetDetails() *structpb.Struct {
	if x != nil {
		return x.Details
	}
	return nil
}

type RankHotelsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *Config                `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	Hotels        []*Hotel               `protobuf:"bytes,2,rep,name=hotels,proto3" json:"hotels,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RankHotelsRequest) Reset() {
	*x = RankHotelsRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RankHotelsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RankHotelsRequest) ProtoMessage() {}

func (x *RankHotelsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RankHotelsRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{4}
}

func (x *RankHotelsRequest) GetConfig() *Config {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *RankHotelsRequest) GetHotels() []*Hotel {
	if x != nil {
		return x.Hotels
	}
	return nil
}

type RankedHotel struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Score         float64                `protobuf:"fixed64,2,opt,name=score,proto3" json:"score,omitempty"`
	Reason        string                 `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Risks         []string               `protobuf:"bytes,4,rep,name=risks,proto3" json:"risks,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RankedHotel) Reset() {
	*x = RankedHotel{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RankedHotel) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RankedHotel) ProtoMessage() {}

func (x *RankedHotel) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RankedHotel) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{5}
}

func (x *RankedHotel) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RankedHotel) GetScore() float64 {
	if x != nil {
		return x.Score
	}
	return 0
}

func (x *RankedHotel) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *RankedHotel) GetRisks() []string {
	if x != nil {
		return x.Risks
	}
	return nil
}

type RankHotelsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Results       []*RankedHotel         `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RankHotelsResponse) Reset() {
	*x = RankHotelsResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RankHotelsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RankHotelsResponse) ProtoMessage() {}

func (x *RankHotelsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RankHotelsResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{6}
}

func (x *RankHotelsResponse) GetResults() []*RankedHotel {
	if x != nil {
		return x.Results
	}
	return nil
}

var File_api_travel_v1_travel_proto protoreflect.FileDescriptor

const file_api_travel_v1_travel_proto_rawDesc = "" +
	"\n" +
	"\x1aapi/travel/v1/travel.proto\x12\ttravel.v1\x1a\x1cgoogle/protobuf/struct.proto\"\xa1\x03\n" +
	"\x06Config\x12\x1b\n" +
	"\x06origin\x18\x01 \x01(\tH\x00R\x06origin\x88\x01\x01\x12%\n" +
	"\vdestination\x18\x02 \x01(\tH\x01R\vdestination\x88\x01\x01\x12*\n" +
	"\x0edeparture_date\x18\x03 \x01(\tH\x02R\rdepartureDate\x88\x01\x01\x12$\n" +
	"\vreturn_date\x18\x04 \x01(\tH\x03R\n" +
	"returnDate\x88\x01\x01\x12\x1b\n" +
	"\x06adults\x18\x05 \x01(\x05H\x04R\x06adults\x88\x01\x01\x12#\n" +
	"\rchildren_ages\x18\x06 \x03(\x05R\fchildrenAges\x12\x1b\n" +
	"\x06budget\x18\a \x01(\x01H\x05R\x06budget\x88\x01\x01\x12\x1f\n" +
	"\bcurrency\x18\b \x01(\tH\x06R\bcurrency\x88\x01\x01\x12 \n" +
	"\vpreferences\x18\t \x03(\tR\vpreferencesB\t\n" +
	"\a_originB\x0e\n" +
	"\f_destinationB\x11\n" +
	"\x0f_departure_dateB\x0e\n" +
	"\f_return_dateB\t\n" +
	"\a_adultsB\t\n" +
	"\a_budgetB\v\n" +
	"\t_currency\"[\n" +
	"\x14ExtractConfigRequest\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\x12)\n" +
	"\x06config\x18\x02 \x01(\v2\x11.travel.v1.ConfigR\x06config\"\x87\x01\n" +
	"\x15ExtractConfigResponse\x12)\n" +
	"\x06config\x18\x01 \x01(\v2\x11.travel.v1.ConfigR\x06config\x12%\n" +
	"\x0emissing_fields\x18\x02 \x03(\tR\rmissingFields\x12\x1c\n" +
	"\tquestions\x18\x03 \x03(\tR\tquestions\"^\n" +
	"\x05Hotel\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x121\n" +
	"\adetails\x18\x03 \x01(\v2\x17.google.protobuf.StructR\adetails\"h\n" +
	"\x11RankHotelsRequest\x12)\n" +
	"\x06config\x18\x01 \x01(\v2\x11.travel.v1.ConfigR\x06config\x12(\n" +
	"\x06hotels\x18\x02 \x03(\v2\x10.travel.v1.HotelR\x06hotels\"a\n" +
	"\vRankedHotel\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x14\n" +
	"\x05score\x18\x02 \x01(\x01R\x05score\x12\x16\n" +
	"\x06reason\x18\x03 \x01(\tR\x06reason\x12\x14\n" +
	"\x05risks\x18\x04 \x03(\tR\x05risks\"F\n" +
	"\x12RankHotelsResponse\x120\n" +
	"\aresults\x18\x01 \x03(\v2\x16.travel.v1.RankedHotelR\aresults2\xa9\x01\n" +
	"\bTravelAI\x12R\n" +
	"\rExtractConfig\x12\x1f.travel.v1.ExtractConfigRequest\x1a .travel.v1.ExtractConfigResponse\x12I\n" +
	"\n" +
	"RankHotels\x12\x1c.travel.v1.RankHotelsRequest\x1a\x1d.travel.v1.RankHotelsResponseB#Z!ai-service/api/travel/v1;travelv1b\x06proto3"

var (
	file_api_travel_v1_travel_proto_rawDescOnce sync.Once
	file_api_travel_v1_travel_proto_rawDescData []byte
)

func file_api_travel_v1_travel_proto_rawDescGZIP() []byte {
	file_api_travel_v1_travel_proto_rawDescOnce.Do(func() {
		file_api_travel_v1_travel_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_travel_v1_travel_proto_rawDesc), len(file_api_travel_v1_travel_proto_rawDesc)))
	})
	return file_api_travel_v1_travel_proto_rawDescData
}

var file_api_travel_v1_travel_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_api_travel_v1_travel_proto_goTypes = []any{
	(*Config)(nil),
	(*ExtractConfigRequest)(nil),
	(*ExtractConfigResponse)(nil),
	(*Hotel)(nil),
	(*RankHotelsRequest)(nil),
	(*RankedHotel)(nil),
	(*RankHotelsResponse)(nil),
	(*structpb.Struct)(nil),
}
var file_api_travel_v1_travel_proto_depIdxs = []int32{
	0,
	0,
	7,
	0,
	3,
	5,
	1,
	4,
	2,
	6,
	8,
	6,
	6,
	6,
	0,
}

func init() { file_api_travel_v1_travel_proto_init() }
func file_api_travel_v1_travel_proto_init() {
	if File_api_travel_v1_travel_proto != nil {
		return
	}
	file_api_travel_v1_travel_proto_msgTypes[0].OneofWrappers = []any{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_travel_v1_travel_proto_rawDesc), len(file_api_travel_v1_travel_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_travel_v1_travel_proto_goTypes,
		DependencyIndexes: file_api_travel_v1_travel_proto_depIdxs,
		MessageInfos:      file_api_travel_v1_travel_proto_msgTypes,
	}.Build()
	File_api_travel_v1_travel_proto = out.File
	file_api_travel_v1_travel_proto_goTypes = nil
	file_api_travel_v1_travel_proto_depIdxs = nil
}
