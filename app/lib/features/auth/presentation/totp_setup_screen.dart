import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/api_client.dart';

class TotpSetupScreen extends ConsumerStatefulWidget {
  const TotpSetupScreen({super.key});

  @override
  ConsumerState<TotpSetupScreen> createState() => _TotpSetupScreenState();
}

class _TotpSetupScreenState extends ConsumerState<TotpSetupScreen> {
  String _step = 'idle'; // idle, qr, verify, done, disable
  String _secret = '';
  String _qrUrl = '';
  final _codeCtrl = TextEditingController();
  List<String> _backupCodes = [];
  String _error = '';
  bool _loading = false;

  @override
  void dispose() {
    _codeCtrl.dispose();
    super.dispose();
  }

  Future<void> _enable() async {
    setState(() { _loading = true; _error = ''; });
    try {
      final dio = ref.read(dioProvider('ko'));
      final r = await dio.post('/auth/totp/enable');
      setState(() {
        _secret = r.data['secret'] ?? '';
        _qrUrl = r.data['qr_url'] ?? '';
        _step = 'qr';
      });
    } catch (e) {
      setState(() => _error = '오류가 발생했습니다');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _verify() async {
    setState(() { _loading = true; _error = ''; });
    try {
      final dio = ref.read(dioProvider('ko'));
      final r = await dio.post('/auth/totp/verify', data: {'code': _codeCtrl.text});
      setState(() {
        _backupCodes = List<String>.from(r.data['backup_codes'] ?? []);
        _step = 'done';
      });
    } catch (e) {
      setState(() => _error = '코드가 올바르지 않습니다');
    } finally {
      setState(() => _loading = false);
    }
  }

  Future<void> _disable() async {
    setState(() { _loading = true; _error = ''; });
    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.delete('/auth/totp', data: {'code': _codeCtrl.text});
      setState(() { _step = 'idle'; _codeCtrl.clear(); });
    } catch (e) {
      setState(() => _error = '코드가 올바르지 않습니다');
    } finally {
      setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('2단계 인증')),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: _buildStep(),
      ),
    );
  }

  Widget _buildStep() {
    switch (_step) {
      case 'idle':
        return Column(
          children: [
            const Text('Google Authenticator, Authy 등의 앱으로 계정을 보호하세요.'),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _loading ? null : _enable,
              child: _loading ? const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2)) : const Text('2단계 인증 활성화'),
            ),
            TextButton(
              onPressed: () => setState(() => _step = 'disable'),
              child: const Text('비활성화', style: TextStyle(color: Colors.red)),
            ),
          ],
        );

      case 'qr':
        return SingleChildScrollView(
          child: Column(
            children: [
              const Text('인증 앱으로 QR 코드를 스캔하거나 비밀키를 수동 입력하세요.'),
              const SizedBox(height: 12),
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(border: Border.all(color: Colors.grey)),
                child: SelectableText(_secret, style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
              ),
              const SizedBox(height: 8),
              GestureDetector(
                onTap: () {
                  Clipboard.setData(ClipboardData(text: _secret));
                  ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('복사됨')));
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(4)),
                  child: const Text('비밀키 복사', style: TextStyle(color: Colors.blue)),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _codeCtrl,
                keyboardType: TextInputType.number,
                maxLength: 6,
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 24, letterSpacing: 8),
                decoration: const InputDecoration(labelText: '6자리 코드', hintText: '000000'),
              ),
              if (_error.isNotEmpty) Text(_error, style: const TextStyle(color: Colors.red)),
              const SizedBox(height: 8),
              ElevatedButton(
                onPressed: _codeCtrl.text.length == 6 && !_loading ? _verify : null,
                child: const Text('코드 확인 및 활성화'),
              ),
            ],
          ),
        );

      case 'done':
        return Column(
          children: [
            const Icon(Icons.check_circle, color: Colors.green, size: 64),
            const SizedBox(height: 8),
            const Text('2단계 인증 활성화 완료!', style: TextStyle(fontWeight: FontWeight.bold)),
            const SizedBox(height: 12),
            const Text('아래 백업 코드를 안전한 곳에 저장하세요. 각 코드는 1회만 사용 가능합니다.'),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8, runSpacing: 8,
              children: _backupCodes.map((c) => Chip(label: Text(c, style: const TextStyle(fontFamily: 'monospace')))).toList(),
            ),
          ],
        );

      case 'disable':
        return Column(
          children: [
            const Text('비활성화하려면 현재 인증 앱의 코드를 입력하세요.'),
            const SizedBox(height: 12),
            TextField(
              controller: _codeCtrl,
              keyboardType: TextInputType.number,
              maxLength: 6,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 24, letterSpacing: 8),
              decoration: const InputDecoration(labelText: '6자리 코드'),
            ),
            if (_error.isNotEmpty) Text(_error, style: const TextStyle(color: Colors.red)),
            ElevatedButton(
              onPressed: _codeCtrl.text.length == 6 && !_loading ? _disable : null,
              style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
              child: const Text('비활성화', style: TextStyle(color: Colors.white)),
            ),
            TextButton(onPressed: () => setState(() => _step = 'idle'), child: const Text('취소')),
          ],
        );

      default:
        return const SizedBox.shrink();
    }
  }
}
