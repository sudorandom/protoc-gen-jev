from jev.v1 import options_pb2 as _options_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Mode(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    MODE_UNSPECIFIED: _ClassVar[Mode]
    MODE_FAST: _ClassVar[Mode]
    MODE_BALANCED: _ClassVar[Mode]
    MODE_ACCURATE: _ClassVar[Mode]
    MODE_DEBUG: _ClassVar[Mode]
MODE_UNSPECIFIED: Mode
MODE_FAST: Mode
MODE_BALANCED: Mode
MODE_ACCURATE: Mode
MODE_DEBUG: Mode

class RuleTestRecord(_message.Message):
    __slots__ = ("email", "sms", "push_notification", "optional_note", "execution_mode", "filtered_mode", "rating_small", "rating_strict", "discrete_code", "large_scale", "temperature", "discrete_ratio", "security_clearance", "decision_flag", "custom_bounded_score", "secret_token", "freeform_description")
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    SMS_FIELD_NUMBER: _ClassVar[int]
    PUSH_NOTIFICATION_FIELD_NUMBER: _ClassVar[int]
    OPTIONAL_NOTE_FIELD_NUMBER: _ClassVar[int]
    EXECUTION_MODE_FIELD_NUMBER: _ClassVar[int]
    FILTERED_MODE_FIELD_NUMBER: _ClassVar[int]
    RATING_SMALL_FIELD_NUMBER: _ClassVar[int]
    RATING_STRICT_FIELD_NUMBER: _ClassVar[int]
    DISCRETE_CODE_FIELD_NUMBER: _ClassVar[int]
    LARGE_SCALE_FIELD_NUMBER: _ClassVar[int]
    TEMPERATURE_FIELD_NUMBER: _ClassVar[int]
    DISCRETE_RATIO_FIELD_NUMBER: _ClassVar[int]
    SECURITY_CLEARANCE_FIELD_NUMBER: _ClassVar[int]
    DECISION_FLAG_FIELD_NUMBER: _ClassVar[int]
    CUSTOM_BOUNDED_SCORE_FIELD_NUMBER: _ClassVar[int]
    SECRET_TOKEN_FIELD_NUMBER: _ClassVar[int]
    FREEFORM_DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    email: str
    sms: str
    push_notification: str
    optional_note: str
    execution_mode: Mode
    filtered_mode: Mode
    rating_small: int
    rating_strict: int
    discrete_code: int
    large_scale: int
    temperature: float
    discrete_ratio: float
    security_clearance: str
    decision_flag: str
    custom_bounded_score: float
    secret_token: str
    freeform_description: str
    def __init__(self, email: _Optional[str] = ..., sms: _Optional[str] = ..., push_notification: _Optional[str] = ..., optional_note: _Optional[str] = ..., execution_mode: _Optional[_Union[Mode, str]] = ..., filtered_mode: _Optional[_Union[Mode, str]] = ..., rating_small: _Optional[int] = ..., rating_strict: _Optional[int] = ..., discrete_code: _Optional[int] = ..., large_scale: _Optional[int] = ..., temperature: _Optional[float] = ..., discrete_ratio: _Optional[float] = ..., security_clearance: _Optional[str] = ..., decision_flag: _Optional[str] = ..., custom_bounded_score: _Optional[float] = ..., secret_token: _Optional[str] = ..., freeform_description: _Optional[str] = ...) -> None: ...
