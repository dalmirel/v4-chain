// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dydxprotocol/daemons/pricefeed/price_feed.proto

package api

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "github.com/cosmos/gogoproto/types"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// UpdateMarketPriceRequest is a request message updating market prices.
type UpdateMarketPricesRequest struct {
	MarketPriceUpdates []*MarketPriceUpdate `protobuf:"bytes,1,rep,name=market_price_updates,json=marketPriceUpdates,proto3" json:"market_price_updates,omitempty"`
}

func (m *UpdateMarketPricesRequest) Reset()         { *m = UpdateMarketPricesRequest{} }
func (m *UpdateMarketPricesRequest) String() string { return proto.CompactTextString(m) }
func (*UpdateMarketPricesRequest) ProtoMessage()    {}
func (*UpdateMarketPricesRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_3d8cd2726a0e97cb, []int{0}
}
func (m *UpdateMarketPricesRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UpdateMarketPricesRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UpdateMarketPricesRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UpdateMarketPricesRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateMarketPricesRequest.Merge(m, src)
}
func (m *UpdateMarketPricesRequest) XXX_Size() int {
	return m.Size()
}
func (m *UpdateMarketPricesRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateMarketPricesRequest.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateMarketPricesRequest proto.InternalMessageInfo

func (m *UpdateMarketPricesRequest) GetMarketPriceUpdates() []*MarketPriceUpdate {
	if m != nil {
		return m.MarketPriceUpdates
	}
	return nil
}

// UpdateMarketPricesResponse is a response message for updating market prices.
type UpdateMarketPricesResponse struct {
}

func (m *UpdateMarketPricesResponse) Reset()         { *m = UpdateMarketPricesResponse{} }
func (m *UpdateMarketPricesResponse) String() string { return proto.CompactTextString(m) }
func (*UpdateMarketPricesResponse) ProtoMessage()    {}
func (*UpdateMarketPricesResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_3d8cd2726a0e97cb, []int{1}
}
func (m *UpdateMarketPricesResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *UpdateMarketPricesResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_UpdateMarketPricesResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *UpdateMarketPricesResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateMarketPricesResponse.Merge(m, src)
}
func (m *UpdateMarketPricesResponse) XXX_Size() int {
	return m.Size()
}
func (m *UpdateMarketPricesResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateMarketPricesResponse.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateMarketPricesResponse proto.InternalMessageInfo

// ExchangePrice represents a specific exchange's market price
type ExchangePrice struct {
	ExchangeFeedId uint32     `protobuf:"varint,1,opt,name=exchange_feed_id,json=exchangeFeedId,proto3" json:"exchange_feed_id,omitempty"`
	Price          uint64     `protobuf:"varint,2,opt,name=price,proto3" json:"price,omitempty"`
	LastUpdateTime *time.Time `protobuf:"bytes,3,opt,name=last_update_time,json=lastUpdateTime,proto3,stdtime" json:"last_update_time,omitempty"`
}

func (m *ExchangePrice) Reset()         { *m = ExchangePrice{} }
func (m *ExchangePrice) String() string { return proto.CompactTextString(m) }
func (*ExchangePrice) ProtoMessage()    {}
func (*ExchangePrice) Descriptor() ([]byte, []int) {
	return fileDescriptor_3d8cd2726a0e97cb, []int{2}
}
func (m *ExchangePrice) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ExchangePrice) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ExchangePrice.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ExchangePrice) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ExchangePrice.Merge(m, src)
}
func (m *ExchangePrice) XXX_Size() int {
	return m.Size()
}
func (m *ExchangePrice) XXX_DiscardUnknown() {
	xxx_messageInfo_ExchangePrice.DiscardUnknown(m)
}

var xxx_messageInfo_ExchangePrice proto.InternalMessageInfo

func (m *ExchangePrice) GetExchangeFeedId() uint32 {
	if m != nil {
		return m.ExchangeFeedId
	}
	return 0
}

func (m *ExchangePrice) GetPrice() uint64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *ExchangePrice) GetLastUpdateTime() *time.Time {
	if m != nil {
		return m.LastUpdateTime
	}
	return nil
}

// MarketPriceUpdate represents an update to a single market
type MarketPriceUpdate struct {
	MarketId       uint32           `protobuf:"varint,1,opt,name=market_id,json=marketId,proto3" json:"market_id,omitempty"`
	ExchangePrices []*ExchangePrice `protobuf:"bytes,2,rep,name=exchange_prices,json=exchangePrices,proto3" json:"exchange_prices,omitempty"`
}

func (m *MarketPriceUpdate) Reset()         { *m = MarketPriceUpdate{} }
func (m *MarketPriceUpdate) String() string { return proto.CompactTextString(m) }
func (*MarketPriceUpdate) ProtoMessage()    {}
func (*MarketPriceUpdate) Descriptor() ([]byte, []int) {
	return fileDescriptor_3d8cd2726a0e97cb, []int{3}
}
func (m *MarketPriceUpdate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MarketPriceUpdate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MarketPriceUpdate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MarketPriceUpdate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MarketPriceUpdate.Merge(m, src)
}
func (m *MarketPriceUpdate) XXX_Size() int {
	return m.Size()
}
func (m *MarketPriceUpdate) XXX_DiscardUnknown() {
	xxx_messageInfo_MarketPriceUpdate.DiscardUnknown(m)
}

var xxx_messageInfo_MarketPriceUpdate proto.InternalMessageInfo

func (m *MarketPriceUpdate) GetMarketId() uint32 {
	if m != nil {
		return m.MarketId
	}
	return 0
}

func (m *MarketPriceUpdate) GetExchangePrices() []*ExchangePrice {
	if m != nil {
		return m.ExchangePrices
	}
	return nil
}

func init() {
	proto.RegisterType((*UpdateMarketPricesRequest)(nil), "dydxprotocol.daemons.pricefeed.UpdateMarketPricesRequest")
	proto.RegisterType((*UpdateMarketPricesResponse)(nil), "dydxprotocol.daemons.pricefeed.UpdateMarketPricesResponse")
	proto.RegisterType((*ExchangePrice)(nil), "dydxprotocol.daemons.pricefeed.ExchangePrice")
	proto.RegisterType((*MarketPriceUpdate)(nil), "dydxprotocol.daemons.pricefeed.MarketPriceUpdate")
}

func init() {
	proto.RegisterFile("dydxprotocol/daemons/pricefeed/price_feed.proto", fileDescriptor_3d8cd2726a0e97cb)
}

var fileDescriptor_3d8cd2726a0e97cb = []byte{
	// 431 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0x3f, 0x8f, 0xd3, 0x30,
	0x18, 0xc6, 0xe3, 0x3b, 0x40, 0x87, 0x4f, 0x77, 0x14, 0xab, 0x43, 0x08, 0x28, 0x8d, 0x32, 0x65,
	0xc1, 0x81, 0xc2, 0x02, 0xe3, 0x49, 0x20, 0x1d, 0x12, 0x08, 0x85, 0x3f, 0x03, 0x4b, 0x94, 0xc6,
	0xef, 0xa5, 0x11, 0x4d, 0x1d, 0x62, 0xa7, 0x2a, 0x1b, 0x23, 0x0b, 0x52, 0xbf, 0x01, 0x12, 0x9f,
	0xa6, 0x63, 0x47, 0x26, 0x40, 0xed, 0x17, 0x41, 0xb6, 0xd3, 0xd2, 0xd2, 0x42, 0xa5, 0xdb, 0x5e,
	0xbf, 0x7e, 0x1f, 0xfb, 0x79, 0x7e, 0x89, 0x71, 0xc8, 0x3e, 0xb2, 0x71, 0x59, 0x71, 0xc9, 0x53,
	0x3e, 0x08, 0x59, 0x02, 0x05, 0x1f, 0x8a, 0xb0, 0xac, 0xf2, 0x14, 0x2e, 0x00, 0x98, 0xa9, 0x62,
	0x55, 0x52, 0x3d, 0x45, 0xdc, 0x75, 0x01, 0x6d, 0x04, 0x74, 0x25, 0x70, 0xda, 0x19, 0xcf, 0xb8,
	0xde, 0x0f, 0x55, 0x65, 0x54, 0x4e, 0x27, 0xe3, 0x3c, 0x1b, 0x40, 0xa8, 0x57, 0xbd, 0xfa, 0x22,
	0x94, 0x79, 0x01, 0x42, 0x26, 0x45, 0x69, 0x06, 0xfc, 0x4f, 0x08, 0xdf, 0x7a, 0x53, 0xb2, 0x44,
	0xc2, 0xf3, 0xa4, 0x7a, 0x0f, 0xf2, 0xa5, 0x3a, 0x50, 0x44, 0xf0, 0xa1, 0x06, 0x21, 0x49, 0x8a,
	0xdb, 0x85, 0x6e, 0xc7, 0xc6, 0x4f, 0xad, 0x27, 0x85, 0x8d, 0xbc, 0xc3, 0xe0, 0xb8, 0x7b, 0x9f,
	0xfe, 0xdf, 0x13, 0x5d, 0x3b, 0xd2, 0xdc, 0x11, 0x91, 0xe2, 0xef, 0x96, 0xf0, 0xef, 0x60, 0x67,
	0x97, 0x03, 0x51, 0xf2, 0xa1, 0x00, 0xff, 0x2b, 0xc2, 0x27, 0x4f, 0xc6, 0x69, 0x3f, 0x19, 0x66,
	0xa0, 0xb7, 0x48, 0x80, 0x5b, 0xd0, 0x34, 0x34, 0xa0, 0x38, 0x67, 0x36, 0xf2, 0x50, 0x70, 0x12,
	0x9d, 0x2e, 0xfb, 0x4f, 0x01, 0xd8, 0x39, 0x23, 0x6d, 0x7c, 0x55, 0x9b, 0xb1, 0x0f, 0x3c, 0x14,
	0x5c, 0x89, 0xcc, 0x82, 0xbc, 0xc0, 0xad, 0x41, 0x22, 0x64, 0x13, 0x26, 0x56, 0x44, 0xec, 0x43,
	0x0f, 0x05, 0xc7, 0x5d, 0x87, 0x1a, 0x5c, 0x74, 0x89, 0x8b, 0xbe, 0x5e, 0xe2, 0x3a, 0x3b, 0x9a,
	0xfe, 0xe8, 0xa0, 0xc9, 0xcf, 0x0e, 0x8a, 0x4e, 0x95, 0xda, 0x38, 0x56, 0xdb, 0xfe, 0x67, 0x84,
	0x6f, 0x6e, 0x25, 0x25, 0xb7, 0xf1, 0xf5, 0x06, 0xdd, 0xca, 0xde, 0x91, 0x69, 0x9c, 0x33, 0xf2,
	0x16, 0xdf, 0x58, 0x45, 0xd0, 0xa6, 0x84, 0x7d, 0xa0, 0x91, 0xde, 0xdd, 0x87, 0x74, 0x03, 0xc5,
	0x9f, 0xc0, 0x06, 0x5a, 0xf7, 0x1b, 0xc2, 0x2d, 0x5d, 0x2a, 0x00, 0xaf, 0xa0, 0x1a, 0xa9, 0xbc,
	0x5f, 0x10, 0x26, 0xdb, 0x80, 0xc9, 0xa3, 0x7d, 0x57, 0xfd, 0xf3, 0xb7, 0x70, 0x1e, 0x5f, 0x46,
	0xda, 0x7c, 0x4f, 0xeb, 0xec, 0xd9, 0x74, 0xee, 0xa2, 0xd9, 0xdc, 0x45, 0xbf, 0xe6, 0x2e, 0x9a,
	0x2c, 0x5c, 0x6b, 0xb6, 0x70, 0xad, 0xef, 0x0b, 0xd7, 0x7a, 0x77, 0x2f, 0xcb, 0x65, 0xbf, 0xee,
	0xd1, 0x94, 0x17, 0x9b, 0xef, 0x63, 0xf4, 0x70, 0xc7, 0x13, 0x49, 0xca, 0xbc, 0x77, 0x4d, 0x8f,
	0x3c, 0xf8, 0x1d, 0x00, 0x00, 0xff, 0xff, 0x7c, 0x43, 0xb3, 0x7c, 0x4f, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PriceFeedServiceClient is the client API for PriceFeedService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PriceFeedServiceClient interface {
	// Updates market prices.
	UpdateMarketPrices(ctx context.Context, in *UpdateMarketPricesRequest, opts ...grpc.CallOption) (*UpdateMarketPricesResponse, error)
}

type priceFeedServiceClient struct {
	cc grpc1.ClientConn
}

func NewPriceFeedServiceClient(cc grpc1.ClientConn) PriceFeedServiceClient {
	return &priceFeedServiceClient{cc}
}

func (c *priceFeedServiceClient) UpdateMarketPrices(ctx context.Context, in *UpdateMarketPricesRequest, opts ...grpc.CallOption) (*UpdateMarketPricesResponse, error) {
	out := new(UpdateMarketPricesResponse)
	err := c.cc.Invoke(ctx, "/dydxprotocol.daemons.pricefeed.PriceFeedService/UpdateMarketPrices", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PriceFeedServiceServer is the server API for PriceFeedService service.
type PriceFeedServiceServer interface {
	// Updates market prices.
	UpdateMarketPrices(context.Context, *UpdateMarketPricesRequest) (*UpdateMarketPricesResponse, error)
}

// UnimplementedPriceFeedServiceServer can be embedded to have forward compatible implementations.
type UnimplementedPriceFeedServiceServer struct {
}

func (*UnimplementedPriceFeedServiceServer) UpdateMarketPrices(ctx context.Context, req *UpdateMarketPricesRequest) (*UpdateMarketPricesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateMarketPrices not implemented")
}

func RegisterPriceFeedServiceServer(s grpc1.Server, srv PriceFeedServiceServer) {
	s.RegisterService(&_PriceFeedService_serviceDesc, srv)
}

func _PriceFeedService_UpdateMarketPrices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateMarketPricesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PriceFeedServiceServer).UpdateMarketPrices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dydxprotocol.daemons.pricefeed.PriceFeedService/UpdateMarketPrices",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PriceFeedServiceServer).UpdateMarketPrices(ctx, req.(*UpdateMarketPricesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _PriceFeedService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "dydxprotocol.daemons.pricefeed.PriceFeedService",
	HandlerType: (*PriceFeedServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "UpdateMarketPrices",
			Handler:    _PriceFeedService_UpdateMarketPrices_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dydxprotocol/daemons/pricefeed/price_feed.proto",
}

func (m *UpdateMarketPricesRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UpdateMarketPricesRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UpdateMarketPricesRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MarketPriceUpdates) > 0 {
		for iNdEx := len(m.MarketPriceUpdates) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.MarketPriceUpdates[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPriceFeed(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *UpdateMarketPricesResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *UpdateMarketPricesResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *UpdateMarketPricesResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *ExchangePrice) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ExchangePrice) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ExchangePrice) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.LastUpdateTime != nil {
		n1, err1 := github_com_cosmos_gogoproto_types.StdTimeMarshalTo(*m.LastUpdateTime, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdTime(*m.LastUpdateTime):])
		if err1 != nil {
			return 0, err1
		}
		i -= n1
		i = encodeVarintPriceFeed(dAtA, i, uint64(n1))
		i--
		dAtA[i] = 0x1a
	}
	if m.Price != 0 {
		i = encodeVarintPriceFeed(dAtA, i, uint64(m.Price))
		i--
		dAtA[i] = 0x10
	}
	if m.ExchangeFeedId != 0 {
		i = encodeVarintPriceFeed(dAtA, i, uint64(m.ExchangeFeedId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *MarketPriceUpdate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MarketPriceUpdate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MarketPriceUpdate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ExchangePrices) > 0 {
		for iNdEx := len(m.ExchangePrices) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ExchangePrices[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintPriceFeed(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.MarketId != 0 {
		i = encodeVarintPriceFeed(dAtA, i, uint64(m.MarketId))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintPriceFeed(dAtA []byte, offset int, v uint64) int {
	offset -= sovPriceFeed(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *UpdateMarketPricesRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.MarketPriceUpdates) > 0 {
		for _, e := range m.MarketPriceUpdates {
			l = e.Size()
			n += 1 + l + sovPriceFeed(uint64(l))
		}
	}
	return n
}

func (m *UpdateMarketPricesResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *ExchangePrice) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.ExchangeFeedId != 0 {
		n += 1 + sovPriceFeed(uint64(m.ExchangeFeedId))
	}
	if m.Price != 0 {
		n += 1 + sovPriceFeed(uint64(m.Price))
	}
	if m.LastUpdateTime != nil {
		l = github_com_cosmos_gogoproto_types.SizeOfStdTime(*m.LastUpdateTime)
		n += 1 + l + sovPriceFeed(uint64(l))
	}
	return n
}

func (m *MarketPriceUpdate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.MarketId != 0 {
		n += 1 + sovPriceFeed(uint64(m.MarketId))
	}
	if len(m.ExchangePrices) > 0 {
		for _, e := range m.ExchangePrices {
			l = e.Size()
			n += 1 + l + sovPriceFeed(uint64(l))
		}
	}
	return n
}

func sovPriceFeed(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozPriceFeed(x uint64) (n int) {
	return sovPriceFeed(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *UpdateMarketPricesRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPriceFeed
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: UpdateMarketPricesRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UpdateMarketPricesRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MarketPriceUpdates", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPriceFeed
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MarketPriceUpdates = append(m.MarketPriceUpdates, &MarketPriceUpdate{})
			if err := m.MarketPriceUpdates[len(m.MarketPriceUpdates)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPriceFeed(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *UpdateMarketPricesResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPriceFeed
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: UpdateMarketPricesResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: UpdateMarketPricesResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipPriceFeed(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ExchangePrice) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPriceFeed
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ExchangePrice: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ExchangePrice: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExchangeFeedId", wireType)
			}
			m.ExchangeFeedId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ExchangeFeedId |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Price", wireType)
			}
			m.Price = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Price |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field LastUpdateTime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPriceFeed
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.LastUpdateTime == nil {
				m.LastUpdateTime = new(time.Time)
			}
			if err := github_com_cosmos_gogoproto_types.StdTimeUnmarshal(m.LastUpdateTime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPriceFeed(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *MarketPriceUpdate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowPriceFeed
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MarketPriceUpdate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MarketPriceUpdate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MarketId", wireType)
			}
			m.MarketId = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MarketId |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ExchangePrices", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthPriceFeed
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ExchangePrices = append(m.ExchangePrices, &ExchangePrice{})
			if err := m.ExchangePrices[len(m.ExchangePrices)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipPriceFeed(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthPriceFeed
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipPriceFeed(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowPriceFeed
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowPriceFeed
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthPriceFeed
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupPriceFeed
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthPriceFeed
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthPriceFeed        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowPriceFeed          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupPriceFeed = fmt.Errorf("proto: unexpected end of group")
)
