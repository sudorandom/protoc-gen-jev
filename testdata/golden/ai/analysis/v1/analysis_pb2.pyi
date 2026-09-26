from jev.v1 import options_pb2 as _options_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Iterable as _Iterable, Mapping as _Mapping, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Sentiment(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SENTIMENT_UNSPECIFIED: _ClassVar[Sentiment]
    SENTIMENT_POSITIVE: _ClassVar[Sentiment]
    SENTIMENT_NEUTRAL: _ClassVar[Sentiment]
    SENTIMENT_NEGATIVE: _ClassVar[Sentiment]

class Priority(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PRIORITY_UNSPECIFIED: _ClassVar[Priority]
    PRIORITY_LOW: _ClassVar[Priority]
    PRIORITY_MEDIUM: _ClassVar[Priority]
    PRIORITY_HIGH: _ClassVar[Priority]
    PRIORITY_URGENT: _ClassVar[Priority]
SENTIMENT_UNSPECIFIED: Sentiment
SENTIMENT_POSITIVE: Sentiment
SENTIMENT_NEUTRAL: Sentiment
SENTIMENT_NEGATIVE: Sentiment
PRIORITY_UNSPECIFIED: Priority
PRIORITY_LOW: Priority
PRIORITY_MEDIUM: Priority
PRIORITY_HIGH: Priority
PRIORITY_URGENT: Priority

class Entity(_message.Message):
    __slots__ = ("name", "category", "relevance")
    NAME_FIELD_NUMBER: _ClassVar[int]
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    RELEVANCE_FIELD_NUMBER: _ClassVar[int]
    name: str
    category: str
    relevance: float
    def __init__(self, name: _Optional[str] = ..., category: _Optional[str] = ..., relevance: _Optional[float] = ...) -> None: ...

class Notification(_message.Message):
    __slots__ = ("recipient", "email", "slack_channel", "webhook_url")
    RECIPIENT_FIELD_NUMBER: _ClassVar[int]
    EMAIL_FIELD_NUMBER: _ClassVar[int]
    SLACK_CHANNEL_FIELD_NUMBER: _ClassVar[int]
    WEBHOOK_URL_FIELD_NUMBER: _ClassVar[int]
    recipient: str
    email: str
    slack_channel: str
    webhook_url: str
    def __init__(self, recipient: _Optional[str] = ..., email: _Optional[str] = ..., slack_channel: _Optional[str] = ..., webhook_url: _Optional[str] = ...) -> None: ...

class ActionItem(_message.Message):
    __slots__ = ("task", "priority", "assignee", "notification", "requires_immediate_action")
    TASK_FIELD_NUMBER: _ClassVar[int]
    PRIORITY_FIELD_NUMBER: _ClassVar[int]
    ASSIGNEE_FIELD_NUMBER: _ClassVar[int]
    NOTIFICATION_FIELD_NUMBER: _ClassVar[int]
    REQUIRES_IMMEDIATE_ACTION_FIELD_NUMBER: _ClassVar[int]
    task: str
    priority: Priority
    assignee: str
    notification: Notification
    requires_immediate_action: bool
    def __init__(self, task: _Optional[str] = ..., priority: _Optional[_Union[Priority, str]] = ..., assignee: _Optional[str] = ..., notification: _Optional[_Union[Notification, _Mapping]] = ..., requires_immediate_action: bool = ...) -> None: ...

class AnalysisReport(_message.Message):
    __slots__ = ("title", "summary", "sentiment", "confidence_score", "key_findings", "entities", "action_items")
    TITLE_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    SENTIMENT_FIELD_NUMBER: _ClassVar[int]
    CONFIDENCE_SCORE_FIELD_NUMBER: _ClassVar[int]
    KEY_FINDINGS_FIELD_NUMBER: _ClassVar[int]
    ENTITIES_FIELD_NUMBER: _ClassVar[int]
    ACTION_ITEMS_FIELD_NUMBER: _ClassVar[int]
    title: str
    summary: str
    sentiment: Sentiment
    confidence_score: float
    key_findings: _containers.RepeatedScalarFieldContainer[str]
    entities: _containers.RepeatedCompositeFieldContainer[Entity]
    action_items: _containers.RepeatedCompositeFieldContainer[ActionItem]
    def __init__(self, title: _Optional[str] = ..., summary: _Optional[str] = ..., sentiment: _Optional[_Union[Sentiment, str]] = ..., confidence_score: _Optional[float] = ..., key_findings: _Optional[_Iterable[str]] = ..., entities: _Optional[_Iterable[_Union[Entity, _Mapping]]] = ..., action_items: _Optional[_Iterable[_Union[ActionItem, _Mapping]]] = ...) -> None: ...
