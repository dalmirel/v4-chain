// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dydxprotocol/prices/market.proto

package types

import (
	fmt "fmt"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Market defines the price configuration for a single market relative to
// quoteCurrency.
type Market struct {
	// Unique, sequentially-generated value.
	Id uint32 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// The human readable name of the market pair (e.g. `BTC-USD` or `BTC-ETH`).
	Pair string `protobuf:"bytes,2,opt,name=pair,proto3" json:"pair,omitempty"`
	// Static value. The exponent of the price.
	// For example if `Exponent == -5` then a `Value` of `1,000,000,000`
	// represents “$10,000`. Therefore `10 ^ Exponent` represents the smallest
	// price step (in dollars) that can be recorded.
	Exponent int32 `protobuf:"zigzag32,3,opt,name=exponent,proto3" json:"exponent,omitempty"`
	// The list of exchanges to query to determine the price.
	Exchanges []uint32 `protobuf:"varint,4,rep,packed,name=exchanges,proto3" json:"exchanges,omitempty"`
	// The minimum number of exchanges that should be reporting a live price for
	// a price update to be considered valid.
	MinExchanges uint32 `protobuf:"varint,5,opt,name=min_exchanges,json=minExchanges,proto3" json:"min_exchanges,omitempty"`
	// The minimum allowable change in the `Price` value for a given update.
	// Measured as `1e-6`.
	MinPriceChangePpm uint32 `protobuf:"varint,6,opt,name=min_price_change_ppm,json=minPriceChangePpm,proto3" json:"min_price_change_ppm,omitempty"`
	// The variable value that is updated by oracle price updates. `0` if it has
	// never been updated, `>0` otherwise.
	Price uint64 `protobuf:"varint,7,opt,name=price,proto3" json:"price,omitempty"`
}

func (m *Market) Reset()         { *m = Market{} }
func (m *Market) String() string { return proto.CompactTextString(m) }
func (*Market) ProtoMessage()    {}
func (*Market) Descriptor() ([]byte, []int) {
	return fileDescriptor_d5fa1fbd42754dc9, []int{0}
}
func (m *Market) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Market) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Market.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Market) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Market.Merge(m, src)
}
func (m *Market) XXX_Size() int {
	return m.Size()
}
func (m *Market) XXX_DiscardUnknown() {
	xxx_messageInfo_Market.DiscardUnknown(m)
}

var xxx_messageInfo_Market proto.InternalMessageInfo

func (m *Market) GetId() uint32 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *Market) GetPair() string {
	if m != nil {
		return m.Pair
	}
	return ""
}

func (m *Market) GetExponent() int32 {
	if m != nil {
		return m.Exponent
	}
	return 0
}

func (m *Market) GetExchanges() []uint32 {
	if m != nil {
		return m.Exchanges
	}
	return nil
}

func (m *Market) GetMinExchanges() uint32 {
	if m != nil {
		return m.MinExchanges
	}
	return 0
}

func (m *Market) GetMinPriceChangePpm() uint32 {
	if m != nil {
		return m.MinPriceChangePpm
	}
	return 0
}

func (m *Market) GetPrice() uint64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func init() {
	proto.RegisterType((*Market)(nil), "dydxprotocol.prices.Market")
}

func init() { proto.RegisterFile("dydxprotocol/prices/market.proto", fileDescriptor_d5fa1fbd42754dc9) }

var fileDescriptor_d5fa1fbd42754dc9 = []byte{
	// 268 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0xd0, 0x4f, 0x4b, 0xc3, 0x30,
	0x18, 0x06, 0xf0, 0xa6, 0xeb, 0xaa, 0x0b, 0x56, 0x58, 0xdc, 0x21, 0x88, 0x84, 0xa0, 0x97, 0x7a,
	0x69, 0x0e, 0xfa, 0x09, 0x1c, 0x1e, 0x85, 0x91, 0xa3, 0x97, 0xd2, 0xb5, 0x61, 0x0b, 0x9a, 0x3f,
	0xb4, 0x51, 0xba, 0x6f, 0xe1, 0xc7, 0xf2, 0x38, 0x3c, 0x79, 0x94, 0xf6, 0x8b, 0xc8, 0x52, 0xdc,
	0xdc, 0x2d, 0xef, 0xf3, 0xfc, 0x20, 0x2f, 0x2f, 0xa4, 0xd5, 0xa6, 0x6a, 0x6d, 0x6d, 0x9c, 0x29,
	0xcd, 0x2b, 0xb3, 0xb5, 0x2c, 0x45, 0xc3, 0x54, 0x51, 0xbf, 0x08, 0x97, 0xf9, 0x18, 0x5d, 0xfc,
	0x17, 0xd9, 0x20, 0xae, 0xbf, 0x00, 0x8c, 0x9f, 0xbc, 0x42, 0xe7, 0x30, 0x94, 0x15, 0x06, 0x14,
	0xa4, 0x09, 0x0f, 0x65, 0x85, 0x10, 0x8c, 0x6c, 0x21, 0x6b, 0x1c, 0x52, 0x90, 0x4e, 0xb8, 0x7f,
	0xa3, 0x4b, 0x78, 0x2a, 0x5a, 0x6b, 0xb4, 0xd0, 0x0e, 0x8f, 0x28, 0x48, 0xa7, 0x7c, 0x3f, 0xa3,
	0x2b, 0x38, 0x11, 0x6d, 0xb9, 0x2e, 0xf4, 0x4a, 0x34, 0x38, 0xa2, 0xa3, 0x34, 0xe1, 0x87, 0x00,
	0xdd, 0xc0, 0x44, 0x49, 0x9d, 0x1f, 0xc4, 0xd8, 0x7f, 0x74, 0xa6, 0xa4, 0x7e, 0xdc, 0x23, 0x06,
	0x67, 0x3b, 0xe4, 0x77, 0xcb, 0x87, 0x30, 0xb7, 0x56, 0xe1, 0xd8, 0xdb, 0xa9, 0x92, 0x7a, 0xb1,
	0xab, 0xe6, 0xbe, 0x59, 0x58, 0x85, 0x66, 0x70, 0xec, 0x31, 0x3e, 0xa1, 0x20, 0x8d, 0xf8, 0x30,
	0x3c, 0xcc, 0x3f, 0x3b, 0x02, 0xb6, 0x1d, 0x01, 0x3f, 0x1d, 0x01, 0x1f, 0x3d, 0x09, 0xb6, 0x3d,
	0x09, 0xbe, 0x7b, 0x12, 0x3c, 0xdf, 0xae, 0xa4, 0x5b, 0xbf, 0x2d, 0xb3, 0xd2, 0x28, 0x76, 0x74,
	0xb0, 0xf7, 0x7b, 0xd6, 0xfe, 0x5d, 0xcd, 0x6d, 0xac, 0x68, 0x96, 0xb1, 0xef, 0xee, 0x7e, 0x03,
	0x00, 0x00, 0xff, 0xff, 0x51, 0x60, 0xde, 0x3f, 0x59, 0x01, 0x00, 0x00,
}

func (m *Market) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Market) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Market) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Price != 0 {
		i = encodeVarintMarket(dAtA, i, uint64(m.Price))
		i--
		dAtA[i] = 0x38
	}
	if m.MinPriceChangePpm != 0 {
		i = encodeVarintMarket(dAtA, i, uint64(m.MinPriceChangePpm))
		i--
		dAtA[i] = 0x30
	}
	if m.MinExchanges != 0 {
		i = encodeVarintMarket(dAtA, i, uint64(m.MinExchanges))
		i--
		dAtA[i] = 0x28
	}
	if len(m.Exchanges) > 0 {
		dAtA2 := make([]byte, len(m.Exchanges)*10)
		var j1 int
		for _, num := range m.Exchanges {
			for num >= 1<<7 {
				dAtA2[j1] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j1++
			}
			dAtA2[j1] = uint8(num)
			j1++
		}
		i -= j1
		copy(dAtA[i:], dAtA2[:j1])
		i = encodeVarintMarket(dAtA, i, uint64(j1))
		i--
		dAtA[i] = 0x22
	}
	if m.Exponent != 0 {
		i = encodeVarintMarket(dAtA, i, uint64((uint32(m.Exponent)<<1)^uint32((m.Exponent>>31))))
		i--
		dAtA[i] = 0x18
	}
	if len(m.Pair) > 0 {
		i -= len(m.Pair)
		copy(dAtA[i:], m.Pair)
		i = encodeVarintMarket(dAtA, i, uint64(len(m.Pair)))
		i--
		dAtA[i] = 0x12
	}
	if m.Id != 0 {
		i = encodeVarintMarket(dAtA, i, uint64(m.Id))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func encodeVarintMarket(dAtA []byte, offset int, v uint64) int {
	offset -= sovMarket(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Market) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Id != 0 {
		n += 1 + sovMarket(uint64(m.Id))
	}
	l = len(m.Pair)
	if l > 0 {
		n += 1 + l + sovMarket(uint64(l))
	}
	if m.Exponent != 0 {
		n += 1 + sozMarket(uint64(m.Exponent))
	}
	if len(m.Exchanges) > 0 {
		l = 0
		for _, e := range m.Exchanges {
			l += sovMarket(uint64(e))
		}
		n += 1 + sovMarket(uint64(l)) + l
	}
	if m.MinExchanges != 0 {
		n += 1 + sovMarket(uint64(m.MinExchanges))
	}
	if m.MinPriceChangePpm != 0 {
		n += 1 + sovMarket(uint64(m.MinPriceChangePpm))
	}
	if m.Price != 0 {
		n += 1 + sovMarket(uint64(m.Price))
	}
	return n
}

func sovMarket(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMarket(x uint64) (n int) {
	return sovMarket(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Market) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMarket
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
			return fmt.Errorf("proto: Market: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Market: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			m.Id = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMarket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Id |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Pair", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMarket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMarket
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMarket
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Pair = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Exponent", wireType)
			}
			var v int32
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMarket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			v = int32((uint32(v) >> 1) ^ uint32(((v&1)<<31)>>31))
			m.Exponent = v
		case 4:
			if wireType == 0 {
				var v uint32
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowMarket
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= uint32(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Exchanges = append(m.Exchanges, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowMarket
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthMarket
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthMarket
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				var count int
				for _, integer := range dAtA[iNdEx:postIndex] {
					if integer < 128 {
						count++
					}
				}
				elementCount = count
				if elementCount != 0 && len(m.Exchanges) == 0 {
					m.Exchanges = make([]uint32, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v uint32
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowMarket
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= uint32(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Exchanges = append(m.Exchanges, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Exchanges", wireType)
			}
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinExchanges", wireType)
			}
			m.MinExchanges = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMarket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MinExchanges |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinPriceChangePpm", wireType)
			}
			m.MinPriceChangePpm = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMarket
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MinPriceChangePpm |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Price", wireType)
			}
			m.Price = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMarket
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
		default:
			iNdEx = preIndex
			skippy, err := skipMarket(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMarket
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
func skipMarket(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMarket
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
					return 0, ErrIntOverflowMarket
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
					return 0, ErrIntOverflowMarket
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
				return 0, ErrInvalidLengthMarket
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMarket
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMarket
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMarket        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMarket          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMarket = fmt.Errorf("proto: unexpected end of group")
)
