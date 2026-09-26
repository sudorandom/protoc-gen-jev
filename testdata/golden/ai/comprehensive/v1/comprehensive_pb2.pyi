from jev.v1 import options_pb2 as _options_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    STATUS_UNSPECIFIED: _ClassVar[Status]
    STATUS_ACTIVE: _ClassVar[Status]
    STATUS_PAUSED: _ClassVar[Status]
    STATUS_ARCHIVED: _ClassVar[Status]
STATUS_UNSPECIFIED: Status
STATUS_ACTIVE: Status
STATUS_PAUSED: Status
STATUS_ARCHIVED: Status

class Metadata(_message.Message):
    __slots__ = ("source", "version")
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    source: str
    version: int
    def __init__(self, source: _Optional[str] = ..., version: _Optional[int] = ...) -> None: ...

class ComprehensiveRecord(_message.Message):
    __slots__ = ("text_bounded", "email_address", "website_url", "custom_pattern", "prefixed_code", "score_int", "big_count", "ratio", "latitude", "is_enabled", "status", "tags", "metadata", "notes", "attributes", "text_payload", "binary_payload", "churn_risk", "internal_debug_info", "account_tier")
    class AttributesEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    TEXT_BOUNDED_FIELD_NUMBER: _ClassVar[int]
    EMAIL_ADDRESS_FIELD_NUMBER: _ClassVar[int]
    WEBSITE_URL_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_PATTERN_FIELD_NUMBER: _ClassVar[int]
    PREFIXED_CODE_FIELD_NUMBER: _ClassVar[int]
    SCORE_INT_FIELD_NUMBER: _ClassVar[int]
    BIG_COUNT_FIELD_NUMBER: _ClassVar[int]
    RATIO_FIELD_NUMBER: _ClassVar[int]
    LATITUDE_FIELD_NUMBER: _ClassVar[int]
    IS_ENABLED_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    TAGS_FIELD_NUMBER: _ClassVar[int]
    METADATA_FIELD_NUMBER: _ClassVar[int]
    NOTES_FIELD_NUMBER: _ClassVar[int]
    ATTRIBUTES_FIELD_NUMBER: _ClassVar[int]
    TEXT_PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    BINARY_PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    CHURN_RISK_FIELD_NUMBER: _ClassVar[int]
    INTERNAL_DEBUG_INFO_FIELD_NUMBER: _ClassVar[int]
    ACCOUNT_TIER_FIELD_NUMBER: _ClassVar[int]
    text_bounded: str
    email_address: str
    website_url: str
    custom_pattern: str
    prefixed_code: str
    score_int: int
    big_count: int
    ratio: float
    latitude: float
    is_enabled: bool
    status: Status
    tags: _containers.RepeatedScalarFieldContainer[str]
    metadata: Metadata
    notes: str
    attributes: _containers.ScalarMap[str, str]
    text_payload: str
    binary_payload: bytes
    churn_risk: bool
    internal_debug_info: str
    account_tier: str
    def __init__(self, text_bounded: _Optional[str] = ..., email_address: _Optional[str] = ..., website_url: _Optional[str] = ..., custom_pattern: _Optional[str] = ..., prefixed_code: _Optional[str] = ..., score_int: _Optional[int] = ..., big_count: _Optional[int] = ..., ratio: _Optional[float] = ..., latitude: _Optional[float] = ..., is_enabled: bool = ..., status: _Optional[_Union[Status, str]] = ..., tags: _Optional[_Iterable[str]] = ..., metadata: _Optional[_Union[Metadata, _Mapping]] = ..., notes: _Optional[str] = ..., attributes: _Optional[_Mapping[str, str]] = ..., text_payload: _Optional[str] = ..., binary_payload: _Optional[bytes] = ..., churn_risk: bool = ..., internal_debug_info: _Optional[str] = ..., account_tier: _Optional[str] = ...) -> None: ...
