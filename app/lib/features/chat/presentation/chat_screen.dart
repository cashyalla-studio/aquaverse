import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../domain/chat_model.dart';
import '../../auth/presentation/auth_provider.dart';

class ChatScreen extends ConsumerStatefulWidget {
  final String tradeId;
  const ChatScreen({super.key, required this.tradeId});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  static const _storage = FlutterSecureStorage();

  WebSocketChannel? _channel;
  StreamSubscription? _sub;

  final List<ChatMessage> _messages = [];
  final _controller = TextEditingController();
  final _scrollController = ScrollController();
  bool _isLoading = true;
  String _status = 'connecting'; // connecting, open, closed, error
  int _reconnectAttempts = 0;
  static const _maxReconnect = 5;
  Timer? _reconnectTimer;
  Timer? _pingTimer;

  String? _apiBase;
  String? _token;
  String? _myUserId;

  @override
  void initState() {
    super.initState();
    _init();
  }

  Future<void> _init() async {
    _token = await _storage.read(key: 'av_access_token');
    final authState = ref.read(authProvider);
    _myUserId = authState.user?.id;

    // API URL에서 WS URL 생성
    const apiUrl =
        String.fromEnvironment('API_URL', defaultValue: 'http://localhost:8080');
    _apiBase = apiUrl
        .replaceFirst('https://', 'wss://')
        .replaceFirst('http://', 'ws://');

    _connect();
  }

  void _connect() {
    if (_reconnectAttempts >= _maxReconnect) return;
    if (!mounted) return;

    setState(() => _status = 'connecting');

    try {
      final uri = Uri.parse(
        '$_apiBase/api/v1/trades/${widget.tradeId}/chat?token=$_token',
      );
      _channel = WebSocketChannel.connect(uri);

      setState(() => _status = 'open');
      _reconnectAttempts = 0;

      _sub = _channel!.stream.listen(
        _onData,
        onError: _onError,
        onDone: _onDone,
      );

      // Ping timer (30초)
      _pingTimer?.cancel();
      _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
        _channel?.sink.add(jsonEncode({'type': 'ping'}));
      });
    } catch (e) {
      _scheduleReconnect();
    }
  }

  void _onData(dynamic raw) {
    if (!mounted) return;
    try {
      final json = jsonDecode(raw as String) as Map<String, dynamic>;
      final payload = WsPayload.fromJson(json);

      setState(() {
        if (payload.type == 'history' && payload.messages != null) {
          _messages.clear();
          _messages.addAll(payload.messages!);
          _isLoading = false;
        } else if (payload.type == 'message' && payload.content != null) {
          _messages.add(ChatMessage(
            messageId: payload.messageId ?? 0,
            roomId: payload.roomId,
            senderId: payload.senderId ?? '',
            content: payload.content!,
            msgType: payload.msgType ?? 'TEXT',
            createdAt: payload.createdAt ?? DateTime.now(),
          ));
          WidgetsBinding.instance
              .addPostFrameCallback((_) => _scrollToBottom());
        }
      });
    } catch (_) {}
  }

  void _onError(Object error) {
    if (!mounted) return;
    setState(() => _status = 'error');
    _scheduleReconnect();
  }

  void _onDone() {
    if (!mounted) return;
    setState(() => _status = 'closed');
    _scheduleReconnect();
  }

  void _scheduleReconnect() {
    _pingTimer?.cancel();
    _sub?.cancel();
    _channel = null;

    if (_reconnectAttempts >= _maxReconnect) return;

    final delay = Duration(
      milliseconds:
          (3000 * (1.5 * _reconnectAttempts).clamp(1, 8)).round(),
    );
    _reconnectAttempts++;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(delay, _connect);
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  void _sendMessage() {
    final content = _controller.text.trim();
    if (content.isEmpty || _status != 'open') return;

    _channel?.sink.add(jsonEncode({'content': content}));
    _controller.clear();
  }

  @override
  void dispose() {
    _reconnectTimer?.cancel();
    _pingTimer?.cancel();
    _sub?.cancel();
    _channel?.sink.close();
    _controller.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('거래 채팅'),
        actions: [
          Padding(
            padding: const EdgeInsets.only(right: 16),
            child: Row(
              children: [
                Container(
                  width: 8,
                  height: 8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: _status == 'open'
                        ? Colors.green
                        : _status == 'connecting'
                            ? Colors.amber
                            : Colors.red,
                  ),
                ),
                const SizedBox(width: 4),
                Text(
                  _status == 'open'
                      ? '연결됨'
                      : _status == 'connecting'
                          ? '연결 중'
                          : '연결 끊김',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _messages.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.chat_bubble_outline,
                                size: 64, color: colorScheme.outlineVariant),
                            const SizedBox(height: 16),
                            Text('대화를 시작해보세요',
                                style:
                                    TextStyle(color: colorScheme.outline)),
                          ],
                        ),
                      )
                    : ListView.builder(
                        controller: _scrollController,
                        padding: const EdgeInsets.symmetric(
                            horizontal: 16, vertical: 8),
                        itemCount: _messages.length,
                        itemBuilder: (context, index) {
                          final msg = _messages[index];
                          final isMine = msg.senderId == _myUserId;
                          return _MessageBubble(msg: msg, isMine: isMine);
                        },
                      ),
          ),
          _buildInput(colorScheme),
        ],
      ),
    );
  }

  Widget _buildInput(ColorScheme colorScheme) {
    return SafeArea(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Expanded(
              child: TextField(
                controller: _controller,
                maxLines: 4,
                minLines: 1,
                textInputAction: TextInputAction.newline,
                decoration: InputDecoration(
                  hintText: _status == 'open' ? '메시지 입력...' : '연결 중...',
                  filled: true,
                  fillColor: colorScheme.surfaceContainerHighest,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(24),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16, vertical: 12),
                ),
                enabled: _status == 'open',
              ),
            ),
            const SizedBox(width: 8),
            FilledButton(
              onPressed: _status == 'open' ? _sendMessage : null,
              style: FilledButton.styleFrom(
                shape: const CircleBorder(),
                padding: const EdgeInsets.all(14),
              ),
              child: const Icon(Icons.send_rounded, size: 20),
            ),
          ],
        ),
      ),
    );
  }
}

class _MessageBubble extends StatelessWidget {
  final ChatMessage msg;
  final bool isMine;

  const _MessageBubble({required this.msg, required this.isMine});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final time = TimeOfDay.fromDateTime(msg.createdAt).format(context);

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Align(
        alignment: isMine ? Alignment.centerRight : Alignment.centerLeft,
        child: ConstrainedBox(
          constraints: BoxConstraints(
            maxWidth: MediaQuery.of(context).size.width * 0.72,
          ),
          child: Container(
            padding:
                const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: isMine
                  ? colorScheme.primary
                  : colorScheme.surfaceContainerHigh,
              borderRadius: BorderRadius.only(
                topLeft: const Radius.circular(18),
                topRight: const Radius.circular(18),
                bottomLeft: Radius.circular(isMine ? 18 : 4),
                bottomRight: Radius.circular(isMine ? 4 : 18),
              ),
            ),
            child: Column(
              crossAxisAlignment: isMine
                  ? CrossAxisAlignment.end
                  : CrossAxisAlignment.start,
              children: [
                Text(
                  msg.content,
                  style: TextStyle(
                    color: isMine
                        ? colorScheme.onPrimary
                        : colorScheme.onSurface,
                    fontSize: 15,
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  time,
                  style: TextStyle(
                    fontSize: 11,
                    color: isMine
                        ? colorScheme.onPrimary.withOpacity(0.7)
                        : colorScheme.outline,
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
