from jev.v1 import options_pb2 as _options_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PriorityLevel(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PRIORITY_LEVEL_UNSPECIFIED: _ClassVar[PriorityLevel]
    PRIORITY_LEVEL_LOW: _ClassVar[PriorityLevel]
    PRIORITY_LEVEL_MEDIUM: _ClassVar[PriorityLevel]
    PRIORITY_LEVEL_HIGH: _ClassVar[PriorityLevel]
    PRIORITY_LEVEL_CRITICAL: _ClassVar[PriorityLevel]
PRIORITY_LEVEL_UNSPECIFIED: PriorityLevel
PRIORITY_LEVEL_LOW: PriorityLevel
PRIORITY_LEVEL_MEDIUM: PriorityLevel
PRIORITY_LEVEL_HIGH: PriorityLevel
PRIORITY_LEVEL_CRITICAL: PriorityLevel

class TriageResponse(_message.Message):
    __slots__ = ("automated_runbook", "oncall_engineer", "incident_commander", "requires_immediate_paging", "priority", "urgency_rating", "blast_radius_percentage", "compliance_classification")
    AUTOMATED_RUNBOOK_FIELD_NUMBER: _ClassVar[int]
    ONCALL_ENGINEER_FIELD_NUMBER: _ClassVar[int]
    INCIDENT_COMMANDER_FIELD_NUMBER: _ClassVar[int]
    REQUIRES_IMMEDIATE_PAGING_FIELD_NUMBER: _ClassVar[int]
    PRIORITY_FIELD_NUMBER: _ClassVar[int]
    URGENCY_RATING_FIELD_NUMBER: _ClassVar[int]
    BLAST_RADIUS_PERCENTAGE_FIELD_NUMBER: _ClassVar[int]
    COMPLIANCE_CLASSIFICATION_FIELD_NUMBER: _ClassVar[int]
    automated_runbook: str
    oncall_engineer: str
    incident_commander: str
    requires_immediate_paging: bool
    priority: PriorityLevel
    urgency_rating: int
    blast_radius_percentage: float
    compliance_classification: str
    def __init__(self, automated_runbook: _Optional[str] = ..., oncall_engineer: _Optional[str] = ..., incident_commander: _Optional[str] = ..., requires_immediate_paging: bool = ..., priority: _Optional[_Union[PriorityLevel, str]] = ..., urgency_rating: _Optional[int] = ..., blast_radius_percentage: _Optional[float] = ..., compliance_classification: _Optional[str] = ...) -> None: ...

class TriageRequest(_message.Message):
    __slots__ = ("incident_id", "title", "description", "raw_logs")
    INCIDENT_ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    RAW_LOGS_FIELD_NUMBER: _ClassVar[int]
    incident_id: str
    title: str
    description: str
    raw_logs: str
    def __init__(self, incident_id: _Optional[str] = ..., title: _Optional[str] = ..., description: _Optional[str] = ..., raw_logs: _Optional[str] = ...) -> None: ...
