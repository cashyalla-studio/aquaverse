class ChatMessage {
  final int messageId;
  final int roomId;
  final String senderId;
  final String content;
  final String msgType; // TEXT, IMAGE, SYSTEM
  final DateTime createdAt;

  const ChatMessage({
    required this.messageId,
    required this.roomId,
    required this.senderId,
    required this.content,
    required this.msgType,
    required this.createdAt,
  });

  factory ChatMessage.fromJson(Map<String, dynamic> json) => ChatMessage(
        messageId: json['message_id'] as int? ?? 0,
        roomId: json['room_id'] as int? ?? 0,
        senderId: json['sender_id'] as String? ?? '',
        content: json['content'] as String? ?? '',
        msgType: json['msg_type'] as String? ?? 'TEXT',
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'] as String)
            : DateTime.now(),
      );
}

class WsPayload {
  final String type; // "message" | "history" | "error" | "system"
  final int roomId;
  final List<ChatMessage>? messages;
  final String? senderId;
  final String? content;
  final String? msgType;
  final int? messageId;
  final DateTime? createdAt;
  final String? error;

  const WsPayload({
    required this.type,
    required this.roomId,
    this.messages,
    this.senderId,
    this.content,
    this.msgType,
    this.messageId,
    this.createdAt,
    this.error,
  });

  factory WsPayload.fromJson(Map<String, dynamic> json) => WsPayload(
        type: json['type'] as String? ?? '',
        roomId: json['room_id'] as int? ?? 0,
        messages: (json['messages'] as List<dynamic>?)
            ?.map((e) => ChatMessage.fromJson(e as Map<String, dynamic>))
            .toList(),
        senderId: json['sender_id'] as String?,
        content: json['content'] as String?,
        msgType: json['msg_type'] as String?,
        messageId: json['message_id'] as int?,
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'] as String)
            : null,
        error: json['error'] as String?,
      );
}
