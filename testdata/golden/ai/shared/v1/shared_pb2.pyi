from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional

DESCRIPTOR: _descriptor.FileDescriptor

class Status(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    STATUS_UNSPECIFIED: _ClassVar[Status]
    STATUS_READY: _ClassVar[Status]
    STATUS_BLOCKED: _ClassVar[Status]
STATUS_UNSPECIFIED: Status
STATUS_READY: Status
STATUS_BLOCKED: Status

class ExternalRequest(_message.Message):
    __slots__ = ("input_text",)
    INPUT_TEXT_FIELD_NUMBER: _ClassVar[int]
    input_text: str
    def __init__(self, input_text: _Optional[str] = ...) -> None: ...

class ExternalResponse(_message.Message):
    __slots__ = ("accepted",)
    ACCEPTED_FIELD_NUMBER: _ClassVar[int]
    accepted: bool
    def __init__(self, accepted: bool = ...) -> None: ...
