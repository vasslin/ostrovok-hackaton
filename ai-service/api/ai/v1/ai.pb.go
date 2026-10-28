package aiv1

import (
	v1 "ai-service/api/common/v1"
	v11 "ai-service/api/provider/v1"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)

	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Message struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Role          string                 `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Content       string                 `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Message) Reset() {
	*x = Message{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Message) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

func (x *Message) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type ConversationContext struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Summary       string                 `protobuf:"bytes,1,opt,name=summary,proto3" json:"summary,omitempty"`
	Messages      []*Message             `protobuf:"bytes,2,rep,name=messages,proto3" json:"messages,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ConversationContext) Reset() {
	*x = ConversationContext{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ConversationContext) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ConversationContext) ProtoMessage() {}

func (x *ConversationContext) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*ConversationContext) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{1}
}

func (x *ConversationContext) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *ConversationContext) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

type Assumption struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Field         string                 `protobuf:"bytes,1,opt,name=field,proto3" json:"field,omitempty"`
	Value         string                 `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Reason        string                 `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Assumption) Reset() {
	*x = Assumption{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Assumption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Assumption) ProtoMessage() {}

func (x *Assumption) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*Assumption) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{2}
}

func (x *Assumption) GetField() string {
	if x != nil {
		return x.Field
	}
	return ""
}

func (x *Assumption) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

func (x *Assumption) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

type TravelIdea struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Title         string                 `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
	Description   string                 `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,4,opt,name=config,proto3" json:"config,omitempty"`
	Reasons       []string               `protobuf:"bytes,5,rep,name=reasons,proto3" json:"reasons,omitempty"`
	Assumptions   []*Assumption          `protobuf:"bytes,6,rep,name=assumptions,proto3" json:"assumptions,omitempty"`
	Source        string                 `protobuf:"bytes,7,opt,name=source,proto3" json:"source,omitempty"`
	City          string                 `protobuf:"bytes,8,opt,name=city,proto3" json:"city,omitempty"`
	Reason        string                 `protobuf:"bytes,9,opt,name=reason,proto3" json:"reason,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *TravelIdea) Reset() {
	*x = TravelIdea{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *TravelIdea) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TravelIdea) ProtoMessage() {}

func (x *TravelIdea) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*TravelIdea) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{3}
}

func (x *TravelIdea) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *TravelIdea) GetTitle() string {
	if x != nil {
		return x.Title
	}
	return ""
}

func (x *TravelIdea) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *TravelIdea) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *TravelIdea) GetReasons() []string {
	if x != nil {
		return x.Reasons
	}
	return nil
}

func (x *TravelIdea) GetAssumptions() []*Assumption {
	if x != nil {
		return x.Assumptions
	}
	return nil
}

func (x *TravelIdea) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *TravelIdea) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *TravelIdea) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

type PreviousSession struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Config        *v1.TripConfig         `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	Context       *ConversationContext   `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *PreviousSession) Reset() {
	*x = PreviousSession{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PreviousSession) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PreviousSession) ProtoMessage() {}

func (x *PreviousSession) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*PreviousSession) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{4}
}

func (x *PreviousSession) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *PreviousSession) GetContext() *ConversationContext {
	if x != nil {
		return x.Context
	}
	return nil
}

type GenerateOrUpdateConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Text          string                 `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	Context       *ConversationContext   `protobuf:"bytes,4,opt,name=context,proto3" json:"context,omitempty"`
	Questions     []*v1.Question         `protobuf:"bytes,5,rep,name=questions,proto3" json:"questions,omitempty"`
	Answers       []*v1.Answer           `protobuf:"bytes,6,rep,name=answers,proto3" json:"answers,omitempty"`
	SelectedIdea  *TravelIdea            `protobuf:"bytes,7,opt,name=selected_idea,json=selectedIdea,proto3" json:"selected_idea,omitempty"`
	CurrentDate   string                 `protobuf:"bytes,8,opt,name=current_date,json=currentDate,proto3" json:"current_date,omitempty"`
	ForceGenerate bool                   `protobuf:"varint,9,opt,name=force_generate,json=forceGenerate,proto3" json:"force_generate,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateOrUpdateConfigRequest) Reset() {
	*x = GenerateOrUpdateConfigRequest{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateOrUpdateConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateOrUpdateConfigRequest) ProtoMessage() {}

func (x *GenerateOrUpdateConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GenerateOrUpdateConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{5}
}

func (x *GenerateOrUpdateConfigRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *GenerateOrUpdateConfigRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *GenerateOrUpdateConfigRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *GenerateOrUpdateConfigRequest) GetContext() *ConversationContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *GenerateOrUpdateConfigRequest) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *GenerateOrUpdateConfigRequest) GetAnswers() []*v1.Answer {
	if x != nil {
		return x.Answers
	}
	return nil
}

func (x *GenerateOrUpdateConfigRequest) GetSelectedIdea() *TravelIdea {
	if x != nil {
		return x.SelectedIdea
	}
	return nil
}

func (x *GenerateOrUpdateConfigRequest) GetCurrentDate() string {
	if x != nil {
		return x.CurrentDate
	}
	return ""
}

func (x *GenerateOrUpdateConfigRequest) GetForceGenerate() bool {
	if x != nil {
		return x.ForceGenerate
	}
	return false
}

type GenerateOrUpdateConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Status        string                 `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	Advice        []*TravelIdea          `protobuf:"bytes,3,rep,name=advice,proto3" json:"advice,omitempty"`
	Questions     []*v1.Question         `protobuf:"bytes,4,rep,name=questions,proto3" json:"questions,omitempty"`
	MissingFields []string               `protobuf:"bytes,5,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Assumptions   []*Assumption          `protobuf:"bytes,6,rep,name=assumptions,proto3" json:"assumptions,omitempty"`
	Summary       string                 `protobuf:"bytes,7,opt,name=summary,proto3" json:"summary,omitempty"`
	Warnings      []string               `protobuf:"bytes,8,rep,name=warnings,proto3" json:"warnings,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateOrUpdateConfigResponse) Reset() {
	*x = GenerateOrUpdateConfigResponse{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateOrUpdateConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateOrUpdateConfigResponse) ProtoMessage() {}

func (x *GenerateOrUpdateConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GenerateOrUpdateConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{6}
}

func (x *GenerateOrUpdateConfigResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *GenerateOrUpdateConfigResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *GenerateOrUpdateConfigResponse) GetAdvice() []*TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

func (x *GenerateOrUpdateConfigResponse) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *GenerateOrUpdateConfigResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *GenerateOrUpdateConfigResponse) GetAssumptions() []*Assumption {
	if x != nil {
		return x.Assumptions
	}
	return nil
}

func (x *GenerateOrUpdateConfigResponse) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *GenerateOrUpdateConfigResponse) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

type UpdateConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Text          string                 `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	Context       *ConversationContext   `protobuf:"bytes,4,opt,name=context,proto3" json:"context,omitempty"`
	Questions     []*v1.Question         `protobuf:"bytes,5,rep,name=questions,proto3" json:"questions,omitempty"`
	Answers       []*v1.Answer           `protobuf:"bytes,6,rep,name=answers,proto3" json:"answers,omitempty"`
	SelectedIdea  *TravelIdea            `protobuf:"bytes,7,opt,name=selected_idea,json=selectedIdea,proto3" json:"selected_idea,omitempty"`
	CurrentDate   string                 `protobuf:"bytes,8,opt,name=current_date,json=currentDate,proto3" json:"current_date,omitempty"`
	ForceGenerate bool                   `protobuf:"varint,9,opt,name=force_generate,json=forceGenerate,proto3" json:"force_generate,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateConfigRequest) Reset() {
	*x = UpdateConfigRequest{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateConfigRequest) ProtoMessage() {}

func (x *UpdateConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[7]
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
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateConfigRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *UpdateConfigRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *UpdateConfigRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *UpdateConfigRequest) GetContext() *ConversationContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *UpdateConfigRequest) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *UpdateConfigRequest) GetAnswers() []*v1.Answer {
	if x != nil {
		return x.Answers
	}
	return nil
}

func (x *UpdateConfigRequest) GetSelectedIdea() *TravelIdea {
	if x != nil {
		return x.SelectedIdea
	}
	return nil
}

func (x *UpdateConfigRequest) GetCurrentDate() string {
	if x != nil {
		return x.CurrentDate
	}
	return ""
}

func (x *UpdateConfigRequest) GetForceGenerate() bool {
	if x != nil {
		return x.ForceGenerate
	}
	return false
}

type UpdateConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Status        string                 `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	Advice        []*TravelIdea          `protobuf:"bytes,3,rep,name=advice,proto3" json:"advice,omitempty"`
	Questions     []*v1.Question         `protobuf:"bytes,4,rep,name=questions,proto3" json:"questions,omitempty"`
	MissingFields []string               `protobuf:"bytes,5,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Assumptions   []*Assumption          `protobuf:"bytes,6,rep,name=assumptions,proto3" json:"assumptions,omitempty"`
	Summary       string                 `protobuf:"bytes,7,opt,name=summary,proto3" json:"summary,omitempty"`
	Warnings      []string               `protobuf:"bytes,8,rep,name=warnings,proto3" json:"warnings,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateConfigResponse) Reset() {
	*x = UpdateConfigResponse{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateConfigResponse) ProtoMessage() {}

func (x *UpdateConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[8]
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
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{8}
}

func (x *UpdateConfigResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *UpdateConfigResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *UpdateConfigResponse) GetAdvice() []*TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

func (x *UpdateConfigResponse) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *UpdateConfigResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *UpdateConfigResponse) GetAssumptions() []*Assumption {
	if x != nil {
		return x.Assumptions
	}
	return nil
}

func (x *UpdateConfigResponse) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *UpdateConfigResponse) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

type RefineConfigRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	SessionId     string                 `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Text          string                 `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,3,opt,name=config,proto3" json:"config,omitempty"`
	Context       *ConversationContext   `protobuf:"bytes,4,opt,name=context,proto3" json:"context,omitempty"`
	Questions     []*v1.Question         `protobuf:"bytes,5,rep,name=questions,proto3" json:"questions,omitempty"`
	Answers       []*v1.Answer           `protobuf:"bytes,6,rep,name=answers,proto3" json:"answers,omitempty"`
	SelectedIdea  *TravelIdea            `protobuf:"bytes,7,opt,name=selected_idea,json=selectedIdea,proto3" json:"selected_idea,omitempty"`
	CurrentDate   string                 `protobuf:"bytes,8,opt,name=current_date,json=currentDate,proto3" json:"current_date,omitempty"`
	ForceGenerate bool                   `protobuf:"varint,9,opt,name=force_generate,json=forceGenerate,proto3" json:"force_generate,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RefineConfigRequest) Reset() {
	*x = RefineConfigRequest{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RefineConfigRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RefineConfigRequest) ProtoMessage() {}

func (x *RefineConfigRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RefineConfigRequest) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{9}
}

func (x *RefineConfigRequest) GetSessionId() string {
	if x != nil {
		return x.SessionId
	}
	return ""
}

func (x *RefineConfigRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *RefineConfigRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *RefineConfigRequest) GetContext() *ConversationContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *RefineConfigRequest) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *RefineConfigRequest) GetAnswers() []*v1.Answer {
	if x != nil {
		return x.Answers
	}
	return nil
}

func (x *RefineConfigRequest) GetSelectedIdea() *TravelIdea {
	if x != nil {
		return x.SelectedIdea
	}
	return nil
}

func (x *RefineConfigRequest) GetCurrentDate() string {
	if x != nil {
		return x.CurrentDate
	}
	return ""
}

func (x *RefineConfigRequest) GetForceGenerate() bool {
	if x != nil {
		return x.ForceGenerate
	}
	return false
}

type RefineConfigResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Status        string                 `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	Advice        []*TravelIdea          `protobuf:"bytes,3,rep,name=advice,proto3" json:"advice,omitempty"`
	Questions     []*v1.Question         `protobuf:"bytes,4,rep,name=questions,proto3" json:"questions,omitempty"`
	MissingFields []string               `protobuf:"bytes,5,rep,name=missing_fields,json=missingFields,proto3" json:"missing_fields,omitempty"`
	Assumptions   []*Assumption          `protobuf:"bytes,6,rep,name=assumptions,proto3" json:"assumptions,omitempty"`
	Summary       string                 `protobuf:"bytes,7,opt,name=summary,proto3" json:"summary,omitempty"`
	Warnings      []string               `protobuf:"bytes,8,rep,name=warnings,proto3" json:"warnings,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RefineConfigResponse) Reset() {
	*x = RefineConfigResponse{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RefineConfigResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RefineConfigResponse) ProtoMessage() {}

func (x *RefineConfigResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RefineConfigResponse) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{10}
}

func (x *RefineConfigResponse) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *RefineConfigResponse) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *RefineConfigResponse) GetAdvice() []*TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

func (x *RefineConfigResponse) GetQuestions() []*v1.Question {
	if x != nil {
		return x.Questions
	}
	return nil
}

func (x *RefineConfigResponse) GetMissingFields() []string {
	if x != nil {
		return x.MissingFields
	}
	return nil
}

func (x *RefineConfigResponse) GetAssumptions() []*Assumption {
	if x != nil {
		return x.Assumptions
	}
	return nil
}

func (x *RefineConfigResponse) GetSummary() string {
	if x != nil {
		return x.Summary
	}
	return ""
}

func (x *RefineConfigResponse) GetWarnings() []string {
	if x != nil {
		return x.Warnings
	}
	return nil
}

type GenerateRecommendationsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Text          string                 `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	Config        *v1.TripConfig         `protobuf:"bytes,2,opt,name=config,proto3" json:"config,omitempty"`
	Context       *ConversationContext   `protobuf:"bytes,3,opt,name=context,proto3" json:"context,omitempty"`
	LastSessions  []*PreviousSession     `protobuf:"bytes,4,rep,name=last_sessions,json=lastSessions,proto3" json:"last_sessions,omitempty"`
	CurrentDate   string                 `protobuf:"bytes,5,opt,name=current_date,json=currentDate,proto3" json:"current_date,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateRecommendationsRequest) Reset() {
	*x = GenerateRecommendationsRequest{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateRecommendationsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateRecommendationsRequest) ProtoMessage() {}

func (x *GenerateRecommendationsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GenerateRecommendationsRequest) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{11}
}

func (x *GenerateRecommendationsRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *GenerateRecommendationsRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *GenerateRecommendationsRequest) GetContext() *ConversationContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *GenerateRecommendationsRequest) GetLastSessions() []*PreviousSession {
	if x != nil {
		return x.LastSessions
	}
	return nil
}

func (x *GenerateRecommendationsRequest) GetCurrentDate() string {
	if x != nil {
		return x.CurrentDate
	}
	return ""
}

type GenerateRecommendationsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Advice        []*TravelIdea          `protobuf:"bytes,1,rep,name=advice,proto3" json:"advice,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GenerateRecommendationsResponse) Reset() {
	*x = GenerateRecommendationsResponse{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenerateRecommendationsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenerateRecommendationsResponse) ProtoMessage() {}

func (x *GenerateRecommendationsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*GenerateRecommendationsResponse) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{12}
}

func (x *GenerateRecommendationsResponse) GetAdvice() []*TravelIdea {
	if x != nil {
		return x.Advice
	}
	return nil
}

type RouteCandidate struct {
	state               protoimpl.MessageState `protogen:"open.v1"`
	Id                  string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	TransportPartIds    []string               `protobuf:"bytes,2,rep,name=transport_part_ids,json=transportPartIds,proto3" json:"transport_part_ids,omitempty"`
	HotelPartIds        []string               `protobuf:"bytes,3,rep,name=hotel_part_ids,json=hotelPartIds,proto3" json:"hotel_part_ids,omitempty"`
	ConnectionsVerified bool                   `protobuf:"varint,4,opt,name=connections_verified,json=connectionsVerified,proto3" json:"connections_verified,omitempty"`
	unknownFields       protoimpl.UnknownFields
	sizeCache           protoimpl.SizeCache
}

func (x *RouteCandidate) Reset() {
	*x = RouteCandidate{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RouteCandidate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RouteCandidate) ProtoMessage() {}

func (x *RouteCandidate) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RouteCandidate) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{13}
}

func (x *RouteCandidate) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RouteCandidate) GetTransportPartIds() []string {
	if x != nil {
		return x.TransportPartIds
	}
	return nil
}

func (x *RouteCandidate) GetHotelPartIds() []string {
	if x != nil {
		return x.HotelPartIds
	}
	return nil
}

func (x *RouteCandidate) GetConnectionsVerified() bool {
	if x != nil {
		return x.ConnectionsVerified
	}
	return false
}

type RankRoutesRequest struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	Config        *v1.TripConfig           `protobuf:"bytes,1,opt,name=config,proto3" json:"config,omitempty"`
	Context       *ConversationContext     `protobuf:"bytes,2,opt,name=context,proto3" json:"context,omitempty"`
	Parts         *v11.SearchPartsResponse `protobuf:"bytes,3,opt,name=parts,proto3" json:"parts,omitempty"`
	Candidates    []*RouteCandidate        `protobuf:"bytes,4,rep,name=candidates,proto3" json:"candidates,omitempty"`
	Limit         int32                    `protobuf:"varint,5,opt,name=limit,proto3" json:"limit,omitempty"`
	CurrentDate   string                   `protobuf:"bytes,6,opt,name=current_date,json=currentDate,proto3" json:"current_date,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RankRoutesRequest) Reset() {
	*x = RankRoutesRequest{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RankRoutesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RankRoutesRequest) ProtoMessage() {}

func (x *RankRoutesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RankRoutesRequest) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{14}
}

func (x *RankRoutesRequest) GetConfig() *v1.TripConfig {
	if x != nil {
		return x.Config
	}
	return nil
}

func (x *RankRoutesRequest) GetContext() *ConversationContext {
	if x != nil {
		return x.Context
	}
	return nil
}

func (x *RankRoutesRequest) GetParts() *v11.SearchPartsResponse {
	if x != nil {
		return x.Parts
	}
	return nil
}

func (x *RankRoutesRequest) GetCandidates() []*RouteCandidate {
	if x != nil {
		return x.Candidates
	}
	return nil
}

func (x *RankRoutesRequest) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

func (x *RankRoutesRequest) GetCurrentDate() string {
	if x != nil {
		return x.CurrentDate
	}
	return ""
}

type RankedRoute struct {
	state                protoimpl.MessageState `protogen:"open.v1"`
	Id                   string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Score                float64                `protobuf:"fixed64,2,opt,name=score,proto3" json:"score,omitempty"`
	Reason               string                 `protobuf:"bytes,3,opt,name=reason,proto3" json:"reason,omitempty"`
	Risks                []string               `protobuf:"bytes,4,rep,name=risks,proto3" json:"risks,omitempty"`
	TotalPrice           *v11.Money             `protobuf:"bytes,5,opt,name=total_price,json=totalPrice,proto3" json:"total_price,omitempty"`
	PriceComplete        bool                   `protobuf:"varint,6,opt,name=price_complete,json=priceComplete,proto3" json:"price_complete,omitempty"`
	RequiresVerification bool                   `protobuf:"varint,7,opt,name=requires_verification,json=requiresVerification,proto3" json:"requires_verification,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *RankedRoute) Reset() {
	*x = RankedRoute{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RankedRoute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RankedRoute) ProtoMessage() {}

func (x *RankedRoute) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RankedRoute) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{15}
}

func (x *RankedRoute) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RankedRoute) GetScore() float64 {
	if x != nil {
		return x.Score
	}
	return 0
}

func (x *RankedRoute) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

func (x *RankedRoute) GetRisks() []string {
	if x != nil {
		return x.Risks
	}
	return nil
}

func (x *RankedRoute) GetTotalPrice() *v11.Money {
	if x != nil {
		return x.TotalPrice
	}
	return nil
}

func (x *RankedRoute) GetPriceComplete() bool {
	if x != nil {
		return x.PriceComplete
	}
	return false
}

func (x *RankedRoute) GetRequiresVerification() bool {
	if x != nil {
		return x.RequiresVerification
	}
	return false
}

type RejectedRoute struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Reason        string                 `protobuf:"bytes,2,opt,name=reason,proto3" json:"reason,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RejectedRoute) Reset() {
	*x = RejectedRoute{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[16]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RejectedRoute) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RejectedRoute) ProtoMessage() {}

func (x *RejectedRoute) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[16]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RejectedRoute) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{16}
}

func (x *RejectedRoute) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *RejectedRoute) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

type RankRoutesResponse struct {
	state         protoimpl.MessageState   `protogen:"open.v1"`
	Results       []*RankedRoute           `protobuf:"bytes,1,rep,name=results,proto3" json:"results,omitempty"`
	Rejected      []*RejectedRoute         `protobuf:"bytes,2,rep,name=rejected,proto3" json:"rejected,omitempty"`
	Parts         *v11.SearchPartsResponse `protobuf:"bytes,3,opt,name=parts,proto3" json:"parts,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RankRoutesResponse) Reset() {
	*x = RankRoutesResponse{}
	mi := &file_api_ai_v1_ai_proto_msgTypes[17]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RankRoutesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RankRoutesResponse) ProtoMessage() {}

func (x *RankRoutesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_api_ai_v1_ai_proto_msgTypes[17]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

func (*RankRoutesResponse) Descriptor() ([]byte, []int) {
	return file_api_ai_v1_ai_proto_rawDescGZIP(), []int{17}
}

func (x *RankRoutesResponse) GetResults() []*RankedRoute {
	if x != nil {
		return x.Results
	}
	return nil
}

func (x *RankRoutesResponse) GetRejected() []*RejectedRoute {
	if x != nil {
		return x.Rejected
	}
	return nil
}

func (x *RankRoutesResponse) GetParts() *v11.SearchPartsResponse {
	if x != nil {
		return x.Parts
	}
	return nil
}

var File_api_ai_v1_ai_proto protoreflect.FileDescriptor

const file_api_ai_v1_ai_proto_rawDesc = "" +
	"\n" +
	"\x12api/ai/v1/ai.proto\x12\rgodnota.ai.v1\x1a\x18api/common/v1/trip.proto\x1a\x1bapi/provider/v1/parts.proto\"7\n" +
	"\aMessage\x12\x12\n" +
	"\x04role\x18\x01 \x01(\tR\x04role\x12\x18\n" +
	"\acontent\x18\x02 \x01(\tR\acontent\"c\n" +
	"\x13ConversationContext\x12\x18\n" +
	"\asummary\x18\x01 \x01(\tR\asummary\x122\n" +
	"\bmessages\x18\x02 \x03(\v2\x16.godnota.ai.v1.MessageR\bmessages\"P\n" +
	"\n" +
	"Assumption\x12\x14\n" +
	"\x05field\x18\x01 \x01(\tR\x05field\x12\x14\n" +
	"\x05value\x18\x02 \x01(\tR\x05value\x12\x16\n" +
	"\x06reason\x18\x03 \x01(\tR\x06reason\"\xa6\x02\n" +
	"\n" +
	"TravelIdea\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x14\n" +
	"\x05title\x18\x02 \x01(\tR\x05title\x12 \n" +
	"\vdescription\x18\x03 \x01(\tR\vdescription\x125\n" +
	"\x06config\x18\x04 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12\x18\n" +
	"\areasons\x18\x05 \x03(\tR\areasons\x12;\n" +
	"\vassumptions\x18\x06 \x03(\v2\x19.godnota.ai.v1.AssumptionR\vassumptions\x12\x16\n" +
	"\x06source\x18\a \x01(\tR\x06source\x12\x12\n" +
	"\x04city\x18\b \x01(\tR\x04city\x12\x16\n" +
	"\x06reason\x18\t \x01(\tR\x06reason\"\x86\x01\n" +
	"\x0fPreviousSession\x125\n" +
	"\x06config\x18\x01 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12<\n" +
	"\acontext\x18\x02 \x01(\v2\".godnota.ai.v1.ConversationContextR\acontext\"\xc1\x03\n" +
	"\x1dGenerateOrUpdateConfigRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x12\n" +
	"\x04text\x18\x02 \x01(\tR\x04text\x125\n" +
	"\x06config\x18\x03 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12<\n" +
	"\acontext\x18\x04 \x01(\v2\".godnota.ai.v1.ConversationContextR\acontext\x129\n" +
	"\tquestions\x18\x05 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x123\n" +
	"\aanswers\x18\x06 \x03(\v2\x19.godnota.common.v1.AnswerR\aanswers\x12>\n" +
	"\rselected_idea\x18\a \x01(\v2\x19.godnota.ai.v1.TravelIdeaR\fselectedIdea\x12!\n" +
	"\fcurrent_date\x18\b \x01(\tR\vcurrentDate\x12%\n" +
	"\x0eforce_generate\x18\t \x01(\bR\rforceGenerate\"\xf7\x02\n" +
	"\x1eGenerateOrUpdateConfigResponse\x12\x16\n" +
	"\x06status\x18\x01 \x01(\tR\x06status\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x121\n" +
	"\x06advice\x18\x03 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\x129\n" +
	"\tquestions\x18\x04 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x12%\n" +
	"\x0emissing_fields\x18\x05 \x03(\tR\rmissingFields\x12;\n" +
	"\vassumptions\x18\x06 \x03(\v2\x19.godnota.ai.v1.AssumptionR\vassumptions\x12\x18\n" +
	"\asummary\x18\a \x01(\tR\asummary\x12\x1a\n" +
	"\bwarnings\x18\b \x03(\tR\bwarnings\"\xb7\x03\n" +
	"\x13UpdateConfigRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x12\n" +
	"\x04text\x18\x02 \x01(\tR\x04text\x125\n" +
	"\x06config\x18\x03 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12<\n" +
	"\acontext\x18\x04 \x01(\v2\".godnota.ai.v1.ConversationContextR\acontext\x129\n" +
	"\tquestions\x18\x05 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x123\n" +
	"\aanswers\x18\x06 \x03(\v2\x19.godnota.common.v1.AnswerR\aanswers\x12>\n" +
	"\rselected_idea\x18\a \x01(\v2\x19.godnota.ai.v1.TravelIdeaR\fselectedIdea\x12!\n" +
	"\fcurrent_date\x18\b \x01(\tR\vcurrentDate\x12%\n" +
	"\x0eforce_generate\x18\t \x01(\bR\rforceGenerate\"\xed\x02\n" +
	"\x14UpdateConfigResponse\x12\x16\n" +
	"\x06status\x18\x01 \x01(\tR\x06status\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x121\n" +
	"\x06advice\x18\x03 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\x129\n" +
	"\tquestions\x18\x04 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x12%\n" +
	"\x0emissing_fields\x18\x05 \x03(\tR\rmissingFields\x12;\n" +
	"\vassumptions\x18\x06 \x03(\v2\x19.godnota.ai.v1.AssumptionR\vassumptions\x12\x18\n" +
	"\asummary\x18\a \x01(\tR\asummary\x12\x1a\n" +
	"\bwarnings\x18\b \x03(\tR\bwarnings\"\xb7\x03\n" +
	"\x13RefineConfigRequest\x12\x1d\n" +
	"\n" +
	"session_id\x18\x01 \x01(\tR\tsessionId\x12\x12\n" +
	"\x04text\x18\x02 \x01(\tR\x04text\x125\n" +
	"\x06config\x18\x03 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12<\n" +
	"\acontext\x18\x04 \x01(\v2\".godnota.ai.v1.ConversationContextR\acontext\x129\n" +
	"\tquestions\x18\x05 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x123\n" +
	"\aanswers\x18\x06 \x03(\v2\x19.godnota.common.v1.AnswerR\aanswers\x12>\n" +
	"\rselected_idea\x18\a \x01(\v2\x19.godnota.ai.v1.TravelIdeaR\fselectedIdea\x12!\n" +
	"\fcurrent_date\x18\b \x01(\tR\vcurrentDate\x12%\n" +
	"\x0eforce_generate\x18\t \x01(\bR\rforceGenerate\"\xed\x02\n" +
	"\x14RefineConfigResponse\x12\x16\n" +
	"\x06status\x18\x01 \x01(\tR\x06status\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x121\n" +
	"\x06advice\x18\x03 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\x129\n" +
	"\tquestions\x18\x04 \x03(\v2\x1b.godnota.common.v1.QuestionR\tquestions\x12%\n" +
	"\x0emissing_fields\x18\x05 \x03(\tR\rmissingFields\x12;\n" +
	"\vassumptions\x18\x06 \x03(\v2\x19.godnota.ai.v1.AssumptionR\vassumptions\x12\x18\n" +
	"\asummary\x18\a \x01(\tR\asummary\x12\x1a\n" +
	"\bwarnings\x18\b \x03(\tR\bwarnings\"\x91\x02\n" +
	"\x1eGenerateRecommendationsRequest\x12\x12\n" +
	"\x04text\x18\x01 \x01(\tR\x04text\x125\n" +
	"\x06config\x18\x02 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12<\n" +
	"\acontext\x18\x03 \x01(\v2\".godnota.ai.v1.ConversationContextR\acontext\x12C\n" +
	"\rlast_sessions\x18\x04 \x03(\v2\x1e.godnota.ai.v1.PreviousSessionR\flastSessions\x12!\n" +
	"\fcurrent_date\x18\x05 \x01(\tR\vcurrentDate\"T\n" +
	"\x1fGenerateRecommendationsResponse\x121\n" +
	"\x06advice\x18\x01 \x03(\v2\x19.godnota.ai.v1.TravelIdeaR\x06advice\"\xa7\x01\n" +
	"\x0eRouteCandidate\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12,\n" +
	"\x12transport_part_ids\x18\x02 \x03(\tR\x10transportPartIds\x12$\n" +
	"\x0ehotel_part_ids\x18\x03 \x03(\tR\fhotelPartIds\x121\n" +
	"\x14connections_verified\x18\x04 \x01(\bR\x13connectionsVerified\"\xc0\x02\n" +
	"\x11RankRoutesRequest\x125\n" +
	"\x06config\x18\x01 \x01(\v2\x1d.godnota.common.v1.TripConfigR\x06config\x12<\n" +
	"\acontext\x18\x02 \x01(\v2\".godnota.ai.v1.ConversationContextR\acontext\x12>\n" +
	"\x05parts\x18\x03 \x01(\v2(.godnota.provider.v1.SearchPartsResponseR\x05parts\x12=\n" +
	"\n" +
	"candidates\x18\x04 \x03(\v2\x1d.godnota.ai.v1.RouteCandidateR\n" +
	"candidates\x12\x14\n" +
	"\x05limit\x18\x05 \x01(\x05R\x05limit\x12!\n" +
	"\fcurrent_date\x18\x06 \x01(\tR\vcurrentDate\"\xfa\x01\n" +
	"\vRankedRoute\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x14\n" +
	"\x05score\x18\x02 \x01(\x01R\x05score\x12\x16\n" +
	"\x06reason\x18\x03 \x01(\tR\x06reason\x12\x14\n" +
	"\x05risks\x18\x04 \x03(\tR\x05risks\x12;\n" +
	"\vtotal_price\x18\x05 \x01(\v2\x1a.godnota.provider.v1.MoneyR\n" +
	"totalPrice\x12%\n" +
	"\x0eprice_complete\x18\x06 \x01(\bR\rpriceComplete\x123\n" +
	"\x15requires_verification\x18\a \x01(\bR\x14requiresVerification\"7\n" +
	"\rRejectedRoute\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x16\n" +
	"\x06reason\x18\x02 \x01(\tR\x06reason\"\xc4\x01\n" +
	"\x12RankRoutesResponse\x124\n" +
	"\aresults\x18\x01 \x03(\v2\x1a.godnota.ai.v1.RankedRouteR\aresults\x128\n" +
	"\brejected\x18\x02 \x03(\v2\x1c.godnota.ai.v1.RejectedRouteR\brejected\x12>\n" +
	"\x05parts\x18\x03 \x01(\v2(.godnota.provider.v1.SearchPartsResponseR\x05parts2\xb1\x05\n" +
	"\tAIService\x12u\n" +
	"\x16GenerateOrUpdateConfig\x12,.godnota.ai.v1.GenerateOrUpdateConfigRequest\x1a-.godnota.ai.v1.GenerateOrUpdateConfigResponse\x12W\n" +
	"\fUpdateConfig\x12\".godnota.ai.v1.UpdateConfigRequest\x1a#.godnota.ai.v1.UpdateConfigResponse\x12W\n" +
	"\fRefineConfig\x12\".godnota.ai.v1.RefineConfigRequest\x1a#.godnota.ai.v1.RefineConfigResponse\x12x\n" +
	"\x17GenerateRecommendations\x12-.godnota.ai.v1.GenerateRecommendationsRequest\x1a..godnota.ai.v1.GenerateRecommendationsResponse\x12Q\n" +
	"\n" +
	"RankRoutes\x12 .godnota.ai.v1.RankRoutesRequest\x1a!.godnota.ai.v1.RankRoutesResponse\x12V\n" +
	"\x0fRankJoinOptions\x12 .godnota.ai.v1.RankRoutesRequest\x1a!.godnota.ai.v1.RankRoutesResponse\x12V\n" +
	"\x0fRankFinalRoutes\x12 .godnota.ai.v1.RankRoutesRequest\x1a!.godnota.ai.v1.RankRoutesResponseB-Z+github.com/godnota/pkg/generated/ai/v1;aiv1b\x06proto3"

var (
	file_api_ai_v1_ai_proto_rawDescOnce sync.Once
	file_api_ai_v1_ai_proto_rawDescData []byte
)

func file_api_ai_v1_ai_proto_rawDescGZIP() []byte {
	file_api_ai_v1_ai_proto_rawDescOnce.Do(func() {
		file_api_ai_v1_ai_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_api_ai_v1_ai_proto_rawDesc), len(file_api_ai_v1_ai_proto_rawDesc)))
	})
	return file_api_ai_v1_ai_proto_rawDescData
}

var file_api_ai_v1_ai_proto_msgTypes = make([]protoimpl.MessageInfo, 18)
var file_api_ai_v1_ai_proto_goTypes = []any{
	(*Message)(nil),
	(*ConversationContext)(nil),
	(*Assumption)(nil),
	(*TravelIdea)(nil),
	(*PreviousSession)(nil),
	(*GenerateOrUpdateConfigRequest)(nil),
	(*GenerateOrUpdateConfigResponse)(nil),
	(*UpdateConfigRequest)(nil),
	(*UpdateConfigResponse)(nil),
	(*RefineConfigRequest)(nil),
	(*RefineConfigResponse)(nil),
	(*GenerateRecommendationsRequest)(nil),
	(*GenerateRecommendationsResponse)(nil),
	(*RouteCandidate)(nil),
	(*RankRoutesRequest)(nil),
	(*RankedRoute)(nil),
	(*RejectedRoute)(nil),
	(*RankRoutesResponse)(nil),
	(*v1.TripConfig)(nil),
	(*v1.Question)(nil),
	(*v1.Answer)(nil),
	(*v11.SearchPartsResponse)(nil),
	(*v11.Money)(nil),
}
var file_api_ai_v1_ai_proto_depIdxs = []int32{
	0,
	18,
	2,
	18,
	1,
	18,
	1,
	19,
	20,
	3,
	18,
	3,
	19,
	2,
	18,
	1,
	19,
	20,
	3,
	18,
	3,
	19,
	2,
	18,
	1,
	19,
	20,
	3,
	18,
	3,
	19,
	2,
	18,
	1,
	4,
	3,
	18,
	1,
	21,
	13,
	22,
	15,
	16,
	21,
	5,
	7,
	9,
	11,
	14,
	14,
	14,
	6,
	8,
	10,
	12,
	17,
	17,
	17,
	51,
	44,
	44,
	44,
	0,
}

func init() { file_api_ai_v1_ai_proto_init() }
func file_api_ai_v1_ai_proto_init() {
	if File_api_ai_v1_ai_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_api_ai_v1_ai_proto_rawDesc), len(file_api_ai_v1_ai_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   18,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_ai_v1_ai_proto_goTypes,
		DependencyIndexes: file_api_ai_v1_ai_proto_depIdxs,
		MessageInfos:      file_api_ai_v1_ai_proto_msgTypes,
	}.Build()
	File_api_ai_v1_ai_proto = out.File
	file_api_ai_v1_ai_proto_goTypes = nil
	file_api_ai_v1_ai_proto_depIdxs = nil
}
