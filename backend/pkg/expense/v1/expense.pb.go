// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: expense/v1/expense.proto

package expensev1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Expense struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Amount        float64                `protobuf:"fixed64,3,opt,name=amount,proto3" json:"amount,omitempty"`
	DayOfMonthDue int32                  `protobuf:"varint,4,opt,name=day_of_month_due,json=dayOfMonthDue,proto3" json:"day_of_month_due,omitempty"`
	IsAutopay     bool                   `protobuf:"varint,5,opt,name=is_autopay,json=isAutopay,proto3" json:"is_autopay,omitempty"`
	CreatedAt     int64                  `protobuf:"varint,6,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt     int64                  `protobuf:"varint,7,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *Expense) Reset() {
	*x = Expense{}
	mi := &file_expense_v1_expense_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Expense) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Expense) ProtoMessage() {}

func (x *Expense) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Expense.ProtoReflect.Descriptor instead.
func (*Expense) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{0}
}

func (x *Expense) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *Expense) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Expense) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Expense) GetDayOfMonthDue() int32 {
	if x != nil {
		return x.DayOfMonthDue
	}
	return 0
}

func (x *Expense) GetIsAutopay() bool {
	if x != nil {
		return x.IsAutopay
	}
	return false
}

func (x *Expense) GetCreatedAt() int64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Expense) GetUpdatedAt() int64 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

type SortedExpense struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Day           int32                  `protobuf:"varint,1,opt,name=day,proto3" json:"day,omitempty"`
	Expenses      []*Expense             `protobuf:"bytes,2,rep,name=expenses,proto3" json:"expenses,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SortedExpense) Reset() {
	*x = SortedExpense{}
	mi := &file_expense_v1_expense_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SortedExpense) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SortedExpense) ProtoMessage() {}

func (x *SortedExpense) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SortedExpense.ProtoReflect.Descriptor instead.
func (*SortedExpense) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{1}
}

func (x *SortedExpense) GetDay() int32 {
	if x != nil {
		return x.Day
	}
	return 0
}

func (x *SortedExpense) GetExpenses() []*Expense {
	if x != nil {
		return x.Expenses
	}
	return nil
}

type CreateExpenseRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Amount        float64                `protobuf:"fixed64,2,opt,name=amount,proto3" json:"amount,omitempty"`
	DayOfMonthDue int32                  `protobuf:"varint,3,opt,name=day_of_month_due,json=dayOfMonthDue,proto3" json:"day_of_month_due,omitempty"`
	IsAutopay     bool                   `protobuf:"varint,4,opt,name=is_autopay,json=isAutopay,proto3" json:"is_autopay,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateExpenseRequest) Reset() {
	*x = CreateExpenseRequest{}
	mi := &file_expense_v1_expense_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateExpenseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateExpenseRequest) ProtoMessage() {}

func (x *CreateExpenseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateExpenseRequest.ProtoReflect.Descriptor instead.
func (*CreateExpenseRequest) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{2}
}

func (x *CreateExpenseRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreateExpenseRequest) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *CreateExpenseRequest) GetDayOfMonthDue() int32 {
	if x != nil {
		return x.DayOfMonthDue
	}
	return 0
}

func (x *CreateExpenseRequest) GetIsAutopay() bool {
	if x != nil {
		return x.IsAutopay
	}
	return false
}

type CreateExpenseResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Expense       *Expense               `protobuf:"bytes,1,opt,name=expense,proto3" json:"expense,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateExpenseResponse) Reset() {
	*x = CreateExpenseResponse{}
	mi := &file_expense_v1_expense_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateExpenseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateExpenseResponse) ProtoMessage() {}

func (x *CreateExpenseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateExpenseResponse.ProtoReflect.Descriptor instead.
func (*CreateExpenseResponse) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{3}
}

func (x *CreateExpenseResponse) GetExpense() *Expense {
	if x != nil {
		return x.Expense
	}
	return nil
}

type GetExpenseRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetExpenseRequest) Reset() {
	*x = GetExpenseRequest{}
	mi := &file_expense_v1_expense_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetExpenseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetExpenseRequest) ProtoMessage() {}

func (x *GetExpenseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetExpenseRequest.ProtoReflect.Descriptor instead.
func (*GetExpenseRequest) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{4}
}

func (x *GetExpenseRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type GetExpenseResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Expense       *Expense               `protobuf:"bytes,1,opt,name=expense,proto3" json:"expense,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetExpenseResponse) Reset() {
	*x = GetExpenseResponse{}
	mi := &file_expense_v1_expense_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetExpenseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetExpenseResponse) ProtoMessage() {}

func (x *GetExpenseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetExpenseResponse.ProtoReflect.Descriptor instead.
func (*GetExpenseResponse) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{5}
}

func (x *GetExpenseResponse) GetExpense() *Expense {
	if x != nil {
		return x.Expense
	}
	return nil
}

type UpdateExpenseRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Amount        float64                `protobuf:"fixed64,3,opt,name=amount,proto3" json:"amount,omitempty"`
	DayOfMonthDue int32                  `protobuf:"varint,4,opt,name=day_of_month_due,json=dayOfMonthDue,proto3" json:"day_of_month_due,omitempty"`
	IsAutopay     bool                   `protobuf:"varint,5,opt,name=is_autopay,json=isAutopay,proto3" json:"is_autopay,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateExpenseRequest) Reset() {
	*x = UpdateExpenseRequest{}
	mi := &file_expense_v1_expense_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateExpenseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateExpenseRequest) ProtoMessage() {}

func (x *UpdateExpenseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateExpenseRequest.ProtoReflect.Descriptor instead.
func (*UpdateExpenseRequest) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{6}
}

func (x *UpdateExpenseRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *UpdateExpenseRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *UpdateExpenseRequest) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *UpdateExpenseRequest) GetDayOfMonthDue() int32 {
	if x != nil {
		return x.DayOfMonthDue
	}
	return 0
}

func (x *UpdateExpenseRequest) GetIsAutopay() bool {
	if x != nil {
		return x.IsAutopay
	}
	return false
}

type UpdateExpenseResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Expense       *Expense               `protobuf:"bytes,1,opt,name=expense,proto3" json:"expense,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *UpdateExpenseResponse) Reset() {
	*x = UpdateExpenseResponse{}
	mi := &file_expense_v1_expense_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *UpdateExpenseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateExpenseResponse) ProtoMessage() {}

func (x *UpdateExpenseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateExpenseResponse.ProtoReflect.Descriptor instead.
func (*UpdateExpenseResponse) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{7}
}

func (x *UpdateExpenseResponse) GetExpense() *Expense {
	if x != nil {
		return x.Expense
	}
	return nil
}

type DeleteExpenseRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Id            int64                  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteExpenseRequest) Reset() {
	*x = DeleteExpenseRequest{}
	mi := &file_expense_v1_expense_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteExpenseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteExpenseRequest) ProtoMessage() {}

func (x *DeleteExpenseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteExpenseRequest.ProtoReflect.Descriptor instead.
func (*DeleteExpenseRequest) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{8}
}

func (x *DeleteExpenseRequest) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

type DeleteExpenseResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Success       bool                   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DeleteExpenseResponse) Reset() {
	*x = DeleteExpenseResponse{}
	mi := &file_expense_v1_expense_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DeleteExpenseResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteExpenseResponse) ProtoMessage() {}

func (x *DeleteExpenseResponse) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteExpenseResponse.ProtoReflect.Descriptor instead.
func (*DeleteExpenseResponse) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{9}
}

func (x *DeleteExpenseResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type ListExpensesRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	PageSize      int32                  `protobuf:"varint,1,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
	PageToken     string                 `protobuf:"bytes,2,opt,name=page_token,json=pageToken,proto3" json:"page_token,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListExpensesRequest) Reset() {
	*x = ListExpensesRequest{}
	mi := &file_expense_v1_expense_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListExpensesRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListExpensesRequest) ProtoMessage() {}

func (x *ListExpensesRequest) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListExpensesRequest.ProtoReflect.Descriptor instead.
func (*ListExpensesRequest) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{10}
}

func (x *ListExpensesRequest) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

func (x *ListExpensesRequest) GetPageToken() string {
	if x != nil {
		return x.PageToken
	}
	return ""
}

type ListExpensesResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Expenses      []*SortedExpense       `protobuf:"bytes,1,rep,name=expenses,proto3" json:"expenses,omitempty"`
	NextPageToken string                 `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListExpensesResponse) Reset() {
	*x = ListExpensesResponse{}
	mi := &file_expense_v1_expense_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListExpensesResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListExpensesResponse) ProtoMessage() {}

func (x *ListExpensesResponse) ProtoReflect() protoreflect.Message {
	mi := &file_expense_v1_expense_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListExpensesResponse.ProtoReflect.Descriptor instead.
func (*ListExpensesResponse) Descriptor() ([]byte, []int) {
	return file_expense_v1_expense_proto_rawDescGZIP(), []int{11}
}

func (x *ListExpensesResponse) GetExpenses() []*SortedExpense {
	if x != nil {
		return x.Expenses
	}
	return nil
}

func (x *ListExpensesResponse) GetNextPageToken() string {
	if x != nil {
		return x.NextPageToken
	}
	return ""
}

var File_expense_v1_expense_proto protoreflect.FileDescriptor

const file_expense_v1_expense_proto_rawDesc = "" +
	"\n" +
	"\x18expense/v1/expense.proto\x12\n" +
	"expense.v1\"\xcb\x01\n" +
	"\aExpense\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x16\n" +
	"\x06amount\x18\x03 \x01(\x01R\x06amount\x12'\n" +
	"\x10day_of_month_due\x18\x04 \x01(\x05R\rdayOfMonthDue\x12\x1d\n" +
	"\n" +
	"is_autopay\x18\x05 \x01(\bR\tisAutopay\x12\x1d\n" +
	"\n" +
	"created_at\x18\x06 \x01(\x03R\tcreatedAt\x12\x1d\n" +
	"\n" +
	"updated_at\x18\a \x01(\x03R\tupdatedAt\"R\n" +
	"\rSortedExpense\x12\x10\n" +
	"\x03day\x18\x01 \x01(\x05R\x03day\x12/\n" +
	"\bexpenses\x18\x02 \x03(\v2\x13.expense.v1.ExpenseR\bexpenses\"\x8a\x01\n" +
	"\x14CreateExpenseRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\x12\x16\n" +
	"\x06amount\x18\x02 \x01(\x01R\x06amount\x12'\n" +
	"\x10day_of_month_due\x18\x03 \x01(\x05R\rdayOfMonthDue\x12\x1d\n" +
	"\n" +
	"is_autopay\x18\x04 \x01(\bR\tisAutopay\"F\n" +
	"\x15CreateExpenseResponse\x12-\n" +
	"\aexpense\x18\x01 \x01(\v2\x13.expense.v1.ExpenseR\aexpense\"#\n" +
	"\x11GetExpenseRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"C\n" +
	"\x12GetExpenseResponse\x12-\n" +
	"\aexpense\x18\x01 \x01(\v2\x13.expense.v1.ExpenseR\aexpense\"\x9a\x01\n" +
	"\x14UpdateExpenseRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\x12\x16\n" +
	"\x06amount\x18\x03 \x01(\x01R\x06amount\x12'\n" +
	"\x10day_of_month_due\x18\x04 \x01(\x05R\rdayOfMonthDue\x12\x1d\n" +
	"\n" +
	"is_autopay\x18\x05 \x01(\bR\tisAutopay\"F\n" +
	"\x15UpdateExpenseResponse\x12-\n" +
	"\aexpense\x18\x01 \x01(\v2\x13.expense.v1.ExpenseR\aexpense\"&\n" +
	"\x14DeleteExpenseRequest\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\x03R\x02id\"1\n" +
	"\x15DeleteExpenseResponse\x12\x18\n" +
	"\asuccess\x18\x01 \x01(\bR\asuccess\"Q\n" +
	"\x13ListExpensesRequest\x12\x1b\n" +
	"\tpage_size\x18\x01 \x01(\x05R\bpageSize\x12\x1d\n" +
	"\n" +
	"page_token\x18\x02 \x01(\tR\tpageToken\"u\n" +
	"\x14ListExpensesResponse\x125\n" +
	"\bexpenses\x18\x01 \x03(\v2\x19.expense.v1.SortedExpenseR\bexpenses\x12&\n" +
	"\x0fnext_page_token\x18\x02 \x01(\tR\rnextPageToken2\xb2\x03\n" +
	"\x0eExpenseService\x12T\n" +
	"\rCreateExpense\x12 .expense.v1.CreateExpenseRequest\x1a!.expense.v1.CreateExpenseResponse\x12K\n" +
	"\n" +
	"GetExpense\x12\x1d.expense.v1.GetExpenseRequest\x1a\x1e.expense.v1.GetExpenseResponse\x12T\n" +
	"\rUpdateExpense\x12 .expense.v1.UpdateExpenseRequest\x1a!.expense.v1.UpdateExpenseResponse\x12T\n" +
	"\rDeleteExpense\x12 .expense.v1.DeleteExpenseRequest\x1a!.expense.v1.DeleteExpenseResponse\x12Q\n" +
	"\fListExpenses\x12\x1f.expense.v1.ListExpensesRequest\x1a .expense.v1.ListExpensesResponseB+Z)expenses-backend/pkg/expense/v1;expensev1b\x06proto3"

var (
	file_expense_v1_expense_proto_rawDescOnce sync.Once
	file_expense_v1_expense_proto_rawDescData []byte
)

func file_expense_v1_expense_proto_rawDescGZIP() []byte {
	file_expense_v1_expense_proto_rawDescOnce.Do(func() {
		file_expense_v1_expense_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_expense_v1_expense_proto_rawDesc), len(file_expense_v1_expense_proto_rawDesc)))
	})
	return file_expense_v1_expense_proto_rawDescData
}

var file_expense_v1_expense_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_expense_v1_expense_proto_goTypes = []any{
	(*Expense)(nil),               // 0: expense.v1.Expense
	(*SortedExpense)(nil),         // 1: expense.v1.SortedExpense
	(*CreateExpenseRequest)(nil),  // 2: expense.v1.CreateExpenseRequest
	(*CreateExpenseResponse)(nil), // 3: expense.v1.CreateExpenseResponse
	(*GetExpenseRequest)(nil),     // 4: expense.v1.GetExpenseRequest
	(*GetExpenseResponse)(nil),    // 5: expense.v1.GetExpenseResponse
	(*UpdateExpenseRequest)(nil),  // 6: expense.v1.UpdateExpenseRequest
	(*UpdateExpenseResponse)(nil), // 7: expense.v1.UpdateExpenseResponse
	(*DeleteExpenseRequest)(nil),  // 8: expense.v1.DeleteExpenseRequest
	(*DeleteExpenseResponse)(nil), // 9: expense.v1.DeleteExpenseResponse
	(*ListExpensesRequest)(nil),   // 10: expense.v1.ListExpensesRequest
	(*ListExpensesResponse)(nil),  // 11: expense.v1.ListExpensesResponse
}
var file_expense_v1_expense_proto_depIdxs = []int32{
	0,  // 0: expense.v1.SortedExpense.expenses:type_name -> expense.v1.Expense
	0,  // 1: expense.v1.CreateExpenseResponse.expense:type_name -> expense.v1.Expense
	0,  // 2: expense.v1.GetExpenseResponse.expense:type_name -> expense.v1.Expense
	0,  // 3: expense.v1.UpdateExpenseResponse.expense:type_name -> expense.v1.Expense
	1,  // 4: expense.v1.ListExpensesResponse.expenses:type_name -> expense.v1.SortedExpense
	2,  // 5: expense.v1.ExpenseService.CreateExpense:input_type -> expense.v1.CreateExpenseRequest
	4,  // 6: expense.v1.ExpenseService.GetExpense:input_type -> expense.v1.GetExpenseRequest
	6,  // 7: expense.v1.ExpenseService.UpdateExpense:input_type -> expense.v1.UpdateExpenseRequest
	8,  // 8: expense.v1.ExpenseService.DeleteExpense:input_type -> expense.v1.DeleteExpenseRequest
	10, // 9: expense.v1.ExpenseService.ListExpenses:input_type -> expense.v1.ListExpensesRequest
	3,  // 10: expense.v1.ExpenseService.CreateExpense:output_type -> expense.v1.CreateExpenseResponse
	5,  // 11: expense.v1.ExpenseService.GetExpense:output_type -> expense.v1.GetExpenseResponse
	7,  // 12: expense.v1.ExpenseService.UpdateExpense:output_type -> expense.v1.UpdateExpenseResponse
	9,  // 13: expense.v1.ExpenseService.DeleteExpense:output_type -> expense.v1.DeleteExpenseResponse
	11, // 14: expense.v1.ExpenseService.ListExpenses:output_type -> expense.v1.ListExpensesResponse
	10, // [10:15] is the sub-list for method output_type
	5,  // [5:10] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_expense_v1_expense_proto_init() }
func file_expense_v1_expense_proto_init() {
	if File_expense_v1_expense_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_expense_v1_expense_proto_rawDesc), len(file_expense_v1_expense_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_expense_v1_expense_proto_goTypes,
		DependencyIndexes: file_expense_v1_expense_proto_depIdxs,
		MessageInfos:      file_expense_v1_expense_proto_msgTypes,
	}.Build()
	File_expense_v1_expense_proto = out.File
	file_expense_v1_expense_proto_goTypes = nil
	file_expense_v1_expense_proto_depIdxs = nil
}
