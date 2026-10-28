package travelv1

import (
	v11 "github.com/godnota/pkg/generated/ai/v1"
	v1 "github.com/godnota/pkg/generated/common/v1"
	v12 "github.com/godnota/pkg/generated/provider/v1"
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

type PlanTripRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Text          string                 `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	UserId        string                 `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PlanTripRequest) Reset() {
	*x = PlanTripRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlanTripRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlanTripRequest) ProtoMessage() {}

func (x *PlanTripRequest) ProtoReflect() protoreflect.Message {
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

func (*PlanTripRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{0}
}

func (x *PlanTripRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *PlanTripRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *PlanTripRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type PlanTripResponse struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	SessionId      string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Intent         string                 `protobuf:"bytes,2,opt,name=intent,proto3" json:"intent,omitempty"`
	Config         *v1.TripConfig         `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	Questions      []*v1.Question         `protobuf:"bytes,4,rep,name=questions,proto3" json:"questions,omitempty"`
	ReadyForSearch bool                   `protobuf:"varint,5,opt,name=ready_for_search,json=readyForSearch,proto3" json:"ready_for_search,omitempty"`
	Advice         []*v11.TravelIdea      `protobuf:"bytes,6,rep,name=advice,proto3" json:"advice,omitempty"`
	Summary        string                 `protobuf:"bytes,7,opt,name=summary,proto3" json:"summary,omitempty"`
	MissingFields  []string               `protobuf:"bytes,8,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Warnings       []string               `protobuf:"bytes,9,rep,name=warnings,proto3" json:"warnings,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *PlanTripResponse) Reset() {
	*x = PlanTripResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PlanTripResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PlanTripResponse) ProtoMessage() {}

func (x *PlanTripResponse) ProtoReflect() protoreflect.Message {
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

func (*PlanTripResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{1}
}

func (x *PlanTripResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *PlanTripResponse) GetIntent() string {
	if x != nil {
		return x.Intent
	}
	return ""
}

func (x *PlanTripResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *PlanTripResponse) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *PlanTripResponse) GetReadyForSearch() bool {
	if x != nil {
		return x.ReadyForSearch
	}
	return false
}

func (x *PlanTripResponse) GetAdvice() []*v11.TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

func (x *PlanTripResponse) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *PlanTripResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *PlanTripResponse) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

type SubmitAnswersRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Answers       []*v1.Answer           `protobuf:"bytes,2,rep,name=answers,proto3" json:"answers,omitempty"`
	UserId        string                 `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SubmitAnswersRequest) Reset() {
	*x = SubmitAnswersRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitAnswersRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitAnswersRequest) ProtoMessage() {}

func (x *SubmitAnswersRequest) ProtoReflect() protoreflect.Message {
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

func (*SubmitAnswersRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{2}
}

func (x *SubmitAnswersRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SubmitAnswersRequest) GetAnswers() []*v1.Answer {
	if x != nil {
		return x.Answers
	}
	return nil
}

func (x *SubmitAnswersRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type SubmitAnswersResponse struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	SessionId      string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Config         *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	Questions      []*v1.Question         `protobuf:"bytes,3,rep,name=questions,proto3" json:"questions,omitempty"`
	ReadyForSearch bool                   `protobuf:"varint,4,opt,name=ready_for_search,json=readyForSearch,proto3" json:"ready_for_search,omitempty"`
	Advice         []*v11.TravelIdea      `protobuf:"bytes,5,rep,name=advice,proto3" json:"advice,omitempty"`
	Summary        string                 `protobuf:"bytes,6,opt,name=summary,proto3" json:"summary,omitempty"`
	MissingFields  []string               `protobuf:"bytes,7,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Warnings       []string               `protobuf:"bytes,8,rep,name=warnings,proto3" json:"warnings,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SubmitAnswersResponse) Reset() {
	*x = SubmitAnswersResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SubmitAnswersResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SubmitAnswersResponse) ProtoMessage() {}

func (x *SubmitAnswersResponse) ProtoReflect() protoreflect.Message {
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

func (*SubmitAnswersResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{3}
}

func (x *SubmitAnswersResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SubmitAnswersResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *SubmitAnswersResponse) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *SubmitAnswersResponse) GetReadyForSearch() bool {
	if x != nil {
		return x.ReadyForSearch
	}
	return false
}

func (x *SubmitAnswersResponse) GetAdvice() []*v11.TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

func (x *SubmitAnswersResponse) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *SubmitAnswersResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *SubmitAnswersResponse) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

type SelectIdeaRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	IdeaId        string                 `protobuf:"bytes,2,opt,name=idea_id,json=ideaId,proto3" json:"idea_id,omitempty"`
	Comment       string                 `protobuf:"bytes,3,opt,name=comment,proto3" json:"comment,omitempty"`
	UserId        string                 `protobuf:"bytes,4,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SelectIdeaRequest) Reset() {
	*x = SelectIdeaRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SelectIdeaRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectIdeaRequest) ProtoMessage() {}

func (x *SelectIdeaRequest) ProtoReflect() protoreflect.Message {
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

func (*SelectIdeaRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{4}
}

func (x *SelectIdeaRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SelectIdeaRequest) GetIdeaId() string {
	if x != nil {
		return x.IdeaId
	}
	return ""
}

func (x *SelectIdeaRequest) GetComment() string {
	if x != nil {
		return x.Comment
	}
	return ""
}

func (x *SelectIdeaRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type SelectIdeaResponse struct {
	state          protoimpl.MessageState `protogen:"open.v1"`
	SessionId      string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Config         *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	Questions      []*v1.Question         `protobuf:"bytes,3,rep,name=questions,proto3" json:"questions,omitempty"`
	Advice         []*v11.TravelIdea      `protobuf:"bytes,4,rep,name=advice,proto3" json:"advice,omitempty"`
	Summary        string                 `protobuf:"bytes,5,opt,name=summary,proto3" json:"summary,omitempty"`
	MissingFields  []string               `protobuf:"bytes,6,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Warnings       []string               `protobuf:"bytes,7,rep,name=warnings,proto3" json:"warnings,omitempty"`
	ReadyForSearch bool                   `protobuf:"varint,8,opt,name=ready_for_search,json=readyForSearch,proto3" json:"ready_for_search,omitempty"`
	unknownFields  protoimpl.UnknownFields
	sizeCache      protoimpl.SizeCache
}

func (x *SelectIdeaResponse) Reset() {
	*x = SelectIdeaResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SelectIdeaResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SelectIdeaResponse) ProtoMessage() {}

func (x *SelectIdeaResponse) ProtoReflect() protoreflect.Message {
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

func (*SelectIdeaResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{5}
}

func (x *SelectIdeaResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SelectIdeaResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *SelectIdeaResponse) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *SelectIdeaResponse) GetAdvice() []*v11.TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

func (x *SelectIdeaResponse) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *SelectIdeaResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *SelectIdeaResponse) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

func (x *SelectIdeaResponse) GetReadyForSearch() bool {
	if x != nil {
		return x.ReadyForSearch
	}
	return false
}

type GetConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetConfigRequest) Reset() {
	*x = GetConfigRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfigRequest) ProtoMessage() {}

func (x *GetConfigRequest) ProtoReflect() protoreflect.Message {
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

func (*GetConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{6}
}

func (x *GetConfigRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *GetConfigRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GetConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetConfigResponse) Reset() {
	*x = GetConfigResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetConfigResponse) ProtoMessage() {}

func (x *GetConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{7}
}

func (x *GetConfigResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *GetConfigResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type UpdateConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	UserId        string                 `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateConfigRequest) Reset() {
	*x = UpdateConfigRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateConfigRequest) ProtoMessage() {}

func (x *UpdateConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*UpdateConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{8}
}

func (x *UpdateConfigRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *UpdateConfigRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *UpdateConfigRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type UpdateConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateConfigResponse) Reset() {
	*x = UpdateConfigResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateConfigResponse) ProtoMessage() {}

func (x *UpdateConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*UpdateConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{9}
}

func (x *UpdateConfigResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *UpdateConfigResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

type ConfirmConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	UserId        string                 `protobuf:"bytes,3,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConfirmConfigRequest) Reset() {
	*x = ConfirmConfigRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConfirmConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfirmConfigRequest) ProtoMessage() {}

func (x *ConfirmConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ConfirmConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{10}
}

func (x *ConfirmConfigRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *ConfirmConfigRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *ConfirmConfigRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type ConfirmConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Confirmed     bool                   `protobuf:"varint,2,opt,name=confirmed,proto3" json:"confirmed,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConfirmConfigResponse) Reset() {
	*x = ConfirmConfigResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConfirmConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConfirmConfigResponse) ProtoMessage() {}

func (x *ConfirmConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ConfirmConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{11}
}

func (x *ConfirmConfigResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *ConfirmConfigResponse) GetConfirmed() bool {
	if x != nil {
		return x.Confirmed
	}
	return false
}

type SearchRoutesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchRoutesRequest) Reset() {
	*x = SearchRoutesRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchRoutesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchRoutesRequest) ProtoMessage() {}

func (x *SearchRoutesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchRoutesRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{12}
}

func (x *SearchRoutesRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SearchRoutesRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type SearchRoutesResponse struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	SessionId     string                   `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Parts         *v12.SearchPartsResponse `protobuf:"bytes,2,opt,name=parts,proto3" json:"parts,omitempty"`
	Candidates    []*structpb.Struct       `protobuf:"bytes,3,rep,name=candidates,proto3" json:"candidates,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchRoutesResponse) Reset() {
	*x = SearchRoutesResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchRoutesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchRoutesResponse) ProtoMessage() {}

func (x *SearchRoutesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*SearchRoutesResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{13}
}

func (x *SearchRoutesResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *SearchRoutesResponse) GetParts() *v12.SearchPartsResponse {
	if x != nil {
		return x.Parts
	}
	return nil
}

func (x *SearchRoutesResponse) GetCandidates() []*structpb.Struct {
	if x != nil {
		return x.Candidates
	}
	return nil
}

type GetRoutesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	UserId        string                 `protobuf:"bytes,2,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetRoutesRequest) Reset() {
	*x = GetRoutesRequest{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRoutesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRoutesRequest) ProtoMessage() {}

func (x *GetRoutesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetRoutesRequest) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{14}
}

func (x *GetRoutesRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *GetRoutesRequest) GetUserId() string {
	if x != nil {
		return x.UserId
	}
	return ""
}

type GetRoutesResponse struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	SessionId     string                   `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Parts         *v12.SearchPartsResponse `protobuf:"bytes,2,opt,name=parts,proto3" json:"parts,omitempty"`
	Candidates    []*structpb.Struct       `protobuf:"bytes,3,rep,name=candidates,proto3" json:"candidates,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetRoutesResponse) Reset() {
	*x = GetRoutesResponse{}
	mi := &file_api_travel_v1_travel_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetRoutesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetRoutesResponse) ProtoMessage() {}

func (x *GetRoutesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_travel_v1_travel_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GetRoutesResponse) Descriptor() ([]byte, []int) {
	return file_api_travel_v1_travel_proto_rawDescGZIP(), []int{15}
}

func (x *GetRoutesResponse) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *GetRoutesResponse) GetParts() *v12.SearchPartsResponse {
	if x != nil {
		return x.Parts
	}
	return nil
}

func (x *GetRoutesResponse) GetCandidates() []*structpb.Struct {
	if x != nil {
		return x.Candidates
	}
	return nil
}

var File_api_travel_v1_travel_proto protoreflect.FileDescriptor

const file_api_travel_v1_travel_proto_rawDesc = "" +
	"\n" +
	"\x1aapi/travel/v1/travel.proto\x12\x11godnota.travel.v1\x1a\x18api/common/v1/trip.proto\x1a\x1bapi/provider/v1/parts.proto\x1a\x12api/ai/v1/ai.proto\x1a\x1cgoogle/protobuf/struct.proto\"]\n" +
	"\x0fPlanTripRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x12\n" +
	"\x04text\x18\x02 \x01(\tR\x04text\x12\x17\n" +
	"\auser_id\x18\x03 \x01(\tR\x06userId\"\xf5\x02\n" +
	"\x10PlanTripResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x16\n" +
	"\x06intent\x18\x02 \x01(\tR\x06intent\x125\n" +
	"\x06config\x18\x03 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x129\n" +
	"\tquestions\x18\x04 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x12(\n" +
	"\x10ready_for_search\x18\x05 \x01(\bR\x0ereadyForSearch\x121\n" +
	"\x06advice\x18\x06 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\x12\x18\n" +
	"\asummary\x18\a \x01(\tR\asummary\x12%\n" +
	"\x0emissing_fields\x18\b \x03(\tR\rmissingFields\x12\x1a\n" +
	"\bwarnings\x18\t \x03(\tR\bwarnings\"\x83\x01\n" +
	"\x14SubmitAnswersRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x123\n" +
	"\aanswers\x18\x02 \x03(\v2\x19.godnota.common.v1.AnswerR\aanswers\x12\x17\n" +
	"\auser_id\x18\x03 \x01(\tR\x06userId\"\xe2\x02\n" +
	"\x15SubmitAnswersResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x129\n" +
	"\tquestions\x18\x03 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x12(\n" +
	"\x10ready_for_search\x18\x04 \x01(\bR\x0ereadyForSearch\x121\n" +
	"\x06advice\x18\x05 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\x12\x18\n" +
	"\asummary\x18\x06 \x01(\tR\asummary\x12%\n" +
	"\x0emissing_fields\x18\a \x03(\tR\rmissingFields\x12\x1a\n" +
	"\bwarnings\x18\b \x03(\tR\bwarnings\"~\n" +
	"\x11SelectIdeaRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x17\n" +
	"\aidea_id\x18\x02 \x01(\tR\x06ideaId\x12\x18\n" +
	"\acomment\x18\x03 \x01(\tR\acomment\x12\x17\n" +
	"\auser_id\x18\x04 \x01(\tR\x06userId\"\xdf\x02\n" +
	"\x12SelectIdeaResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x129\n" +
	"\tquestions\x18\x03 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x121\n" +
	"\x06advice\x18\x04 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\x12\x18\n" +
	"\asummary\x18\x05 \x01(\tR\asummary\x12%\n" +
	"\x0emissing_fields\x18\x06 \x03(\tR\rmissingFields\x12\x1a\n" +
	"\bwarnings\x18\a \x03(\tR\bwarnings\x12(\n" +
	"\x10ready_for_search\x18\b \x01(\bR\x0ereadyForSearch\"J\n" +
	"\x10GetConfigRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\"i\n" +
	"\x11GetConfigResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\"\x84\x01\n" +
	"\x13UpdateConfigRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12\x17\n" +
	"\auser_id\x18\x03 \x01(\tR\x06userId\"l\n" +
	"\x14UpdateConfigResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\"\x85\x01\n" +
	"\x14ConfirmConfigRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12\x17\n" +
	"\auser_id\x18\x03 \x01(\tR\x06userId\"T\n" +
	"\x15ConfirmConfigResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x1c\n" +
	"\tconfirmed\x18\x02 \x01(\bR\tconfirmed\"M\n" +
	"\x13SearchRoutesRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\"\xae\x01\n" +
	"\x14SearchRoutesResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12>\n" +
	"\x05parts\x18\x02 \x01(\v2(.godnota.provider.v1.SearchPartsResponseR\x05parts\x127\n" +
	"\n" +
	"candidates\x18\x03 \x03(\v2\x17.google.protobuf.StructR\n" +
	"candidates\"J\n" +
	"\x10GetRoutesRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x17\n" +
	"\auser_id\x18\x02 \x01(\tR\x06userId\"\xab\x01\n" +
	"\x11GetRoutesResponse\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12>\n" +
	"\x05parts\x18\x02 \x01(\v2(.godnota.provider.v1.SearchPartsResponseR\x05parts\x127\n" +
	"\n" +
	"candidates\x18\x03 \x03(\v2\x17.google.protobuf.StructR\n" +
	"candidates2\xca\x06\n" +
	"\rTravelService\x12O\n" +
	"\x0fGetCapabilities\x12#.godnota.travel.v1.GetConfigRequest\x1a\x17.google.protobuf.Struct\x12S\n" +
	"\bPlanTrip\x12\".godnota.travel.v1.PlanTripRequest\x1a#.godnota.travel.v1.PlanTripResponse\x12b\n" +
	"\rSubmitAnswers\x12'.godnota.travel.v1.SubmitAnswersRequest\x1a(.godnota.travel.v1.SubmitAnswersResponse\x12Y\n" +
	"\n" +
	"SelectIdea\x12$.godnota.travel.v1.SelectIdeaRequest\x1a%.godnota.travel.v1.SelectIdeaResponse\x12V\n" +
	"\tGetConfig\x12#.godnota.travel.v1.GetConfigRequest\x1a$.godnota.travel.v1.GetConfigResponse\x12_\n" +
	"\fUpdateConfig\x12&.godnota.travel.v1.UpdateConfigRequest\x1a'.godnota.travel.v1.UpdateConfigResponse\x12b\n" +
	"\rConfirmConfig\x12'.godnota.travel.v1.ConfirmConfigRequest\x1a(.godnota.travel.v1.ConfirmConfigResponse\x12_\n" +
	"\fSearchRoutes\x12&.godnota.travel.v1.SearchRoutesRequest\x1a'.godnota.travel.v1.SearchRoutesResponse\x12V\n" +
	"\tGetRoutes\x12#.godnota.travel.v1.GetRoutesRequest\x1a$.godnota.travel.v1.GetRoutesResponseB5Z3github.com/godnota/pkg/generated/travel/v1;travelv1b\x06proto3"

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

var file_api_travel_v1_travel_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_api_travel_v1_travel_proto_goTypes = []any{
	(*PlanTripRequest)(nil),
	(*PlanTripResponse)(nil),
	(*SubmitAnswersRequest)(nil),
	(*SubmitAnswersResponse)(nil),
	(*SelectIdeaRequest)(nil),
	(*SelectIdeaResponse)(nil),
	(*GetConfigRequest)(nil),
	(*GetConfigResponse)(nil),
	(*UpdateConfigRequest)(nil),
	(*UpdateConfigResponse)(nil),
	(*ConfirmConfigRequest)(nil),
	(*ConfirmConfigResponse)(nil),
	(*SearchRoutesRequest)(nil),
	(*SearchRoutesResponse)(nil),
	(*GetRoutesRequest)(nil),
	(*GetRoutesResponse)(nil),
	(*v1.TripConfig)(nil),
	(*v1.Question)(nil),
	(*v11.TravelIdea)(nil),
	(*v1.Answer)(nil),
	(*v12.SearchPartsResponse)(nil),
	(*structpb.Struct)(nil),
}
var file_api_travel_v1_travel_proto_depIdxs = []int32{
	16,
	17,
	18,
	19,
	16,
	17,
	18,
	16,
	17,
	18,
	16,
	16,
	16,
	16,
	20,
	21,
	20,
	21,
	6,
	0,
	2,
	4,
	6,
	8,
	10,
	12,
	14,
	21,
	1,
	3,
	5,
	7,
	9,
	11,
	13,
	15,
	27,
	18,
	18,
	18,
	0,
}

func init() { file_api_travel_v1_travel_proto_init() }
func file_api_travel_v1_travel_proto_init() {
	if File_api_travel_v1_travel_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_travel_v1_travel_proto_rawDesc), len(file_api_travel_v1_travel_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   16,
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
