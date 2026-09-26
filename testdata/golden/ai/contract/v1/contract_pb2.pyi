from jev.v1 import options_pb2 as _options_pb2
from ai.shared.v1 import shared_pb2 as _shared_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class EvaluateRequest(_message.Message):
    __slots__ = ("input_text", "input_count")
    INPUT_TEXT_FIELD_NUMBER: _ClassVar[int]
    INPUT_COUNT_FIELD_NUMBER: _ClassVar[int]
    input_text: str
    input_count: int
    def __init__(self, input_text: _Optional[str] = ..., input_count: _Optional[int] = ...) -> None: ...

class EvaluateResponse(_message.Message):
    __slots__ = ("flag", "rating", "count", "big_count", "big_unsigned", "fixed_count", "signed_fixed_count", "risk", "status", "label", "boundary", "email", "binary", "excluded")
    FLAG_FIELD_NUMBER: _ClassVar[int]
    RATING_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    BIG_COUNT_FIELD_NUMBER: _ClassVar[int]
    BIG_UNSIGNED_FIELD_NUMBER: _ClassVar[int]
    FIXED_COUNT_FIELD_NUMBER: _ClassVar[int]
    SIGNED_FIXED_COUNT_FIELD_NUMBER: _ClassVar[int]
    RISK_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    BOUNDARY_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    BINARY_FIELD_NUMBER: _ClassVar[int]
    EXCLUDED_FIELD_NUMBER: _ClassVar[int]
    flag: bool
    rating: int
    count: int
    big_count: int
    big_unsigned: int
    fixed_count: int
    signed_fixed_count: int
    risk: bool
    status: _shared_pb2.Status
    label: str
    boundary: int
    email: str
    binary: bytes
    excluded: str
    def __init__(self, flag: bool = ..., rating: _Optional[int] = ..., count: _Optional[int] = ..., big_count: _Optional[int] = ..., big_unsigned: _Optional[int] = ..., fixed_count: _Optional[int] = ..., signed_fixed_count: _Optional[int] = ..., risk: bool = ..., status: _Optional[_Union[_shared_pb2.Status, str]] = ..., label: _Optional[str] = ..., boundary: _Optional[int] = ..., email: _Optional[str] = ..., binary: _Optional[bytes] = ..., excluded: _Optional[str] = ...) -> None: ...

class Container(_message.Message):
    __slots__ = ()
    class NestedRequest(_message.Message):
        __slots__ = ("input_text",)
        INPUT_TEXT_FIELD_NUMBER: _ClassVar[int]
        input_text: str
        def __init__(self, input_text: _Optional[str] = ...) -> None: ...
    class NestedResponse(_message.Message):
        __slots__ = ("accepted",)
        ACCEPTED_FIELD_NUMBER: _ClassVar[int]
        accepted: bool
        def __init__(self, accepted: bool = ...) -> None: ...
    def __init__(self) -> None: ...
