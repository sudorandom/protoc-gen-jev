from google.protobuf import descriptor_pb2 as _descriptor_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor
FIELD_FIELD_NUMBER: int
field: _descriptor.FieldDescriptor
SERVICE_FIELD_NUMBER: int
service: _descriptor.FieldDescriptor
METHOD_FIELD_NUMBER: int
method: _descriptor.FieldDescriptor

class ServiceOptions(_message.Message):
    __slots__ = ("enabled",)
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    def __init__(self, enabled: bool = ...) -> None: ...

class MethodOptions(_message.Message):
    __slots__ = ("enabled",)
    ENABLED_FIELD_NUMBER: _ClassVar[int]
    enabled: bool
    def __init__(self, enabled: bool = ...) -> None: ...

class FieldOptions(_message.Message):
    __slots__ = ("instructions", "skip", "choice", "score", "noul")
    INSTRUCTIONS_FIELD_NUMBER: _ClassVar[int]
    SKIP_FIELD_NUMBER: _ClassVar[int]
    CHOICE_FIELD_NUMBER: _ClassVar[int]
    SCORE_FIELD_NUMBER: _ClassVar[int]
    NOUL_FIELD_NUMBER: _ClassVar[int]
    instructions: str
    skip: bool
    choice: ChoiceRules
    score: ScoreRules
    noul: NoulRules
    def __init__(self, instructions: _Optional[str] = ..., skip: bool = ..., choice: _Optional[_Union[ChoiceRules, _Mapping]] = ..., score: _Optional[_Union[ScoreRules, _Mapping]] = ..., noul: _Optional[_Union[NoulRules, _Mapping]] = ...) -> None: ...

class ChoiceRules(_message.Message):
    __slots__ = ("choices", "not_in", "criteria")
    class CriteriaEntry(_message.Message):
        __slots__ = ("key", "value")
        KEY_FIELD_NUMBER: _ClassVar[int]
        VALUE_FIELD_NUMBER: _ClassVar[int]
        key: str
        value: str
        def __init__(self, key: _Optional[str] = ..., value: _Optional[str] = ...) -> None: ...
    CHOICES_FIELD_NUMBER: _ClassVar[int]
    NOT_IN_FIELD_NUMBER: _ClassVar[int]
    CRITERIA_FIELD_NUMBER: _ClassVar[int]
    choices: _containers.RepeatedScalarFieldContainer[str]
    not_in: _containers.RepeatedScalarFieldContainer[str]
    criteria: _containers.ScalarMap[str, str]
    def __init__(self, choices: _Optional[_Iterable[str]] = ..., not_in: _Optional[_Iterable[str]] = ..., criteria: _Optional[_Mapping[str, str]] = ...) -> None: ...

class ScoreLevel(_message.Message):
    __slots__ = ("value", "description")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    value: float
    description: str
    def __init__(self, value: _Optional[float] = ..., description: _Optional[str] = ...) -> None: ...

class ScoreRules(_message.Message):
    __slots__ = ("levels",)
    LEVELS_FIELD_NUMBER: _ClassVar[int]
    levels: _containers.RepeatedCompositeFieldContainer[ScoreLevel]
    def __init__(self, levels: _Optional[_Iterable[_Union[ScoreLevel, _Mapping]]] = ...) -> None: ...

class NoulRules(_message.Message):
    __slots__ = ("threshold",)
    THRESHOLD_FIELD_NUMBER: _ClassVar[int]
    threshold: float
    def __init__(self, threshold: _Optional[float] = ...) -> None: ...
