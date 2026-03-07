import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:dio/dio.dart';
import '../../../core/api/api_client.dart';

class PhoneVerifyScreen extends ConsumerStatefulWidget {
  const PhoneVerifyScreen({super.key});

  @override
  ConsumerState<PhoneVerifyScreen> createState() => _PhoneVerifyScreenState();
}

class _PhoneVerifyScreenState extends ConsumerState<PhoneVerifyScreen> {
  final _phoneController = TextEditingController();
  final _codeController = TextEditingController();
  bool _loading = false;
  String? _error;
  bool _codeSent = false;
  bool _verified = false;

  Future<void> _sendCode() async {
    final phone = _phoneController.text.trim();
    if (phone.length < 10) {
      setState(() => _error = '올바른 전화번호를 입력하세요');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post('/phone/send', data: {'phone_number': phone});
      setState(() {
        _codeSent = true;
        _loading = false;
      });
    } on DioException catch (e) {
      setState(() {
        _error = e.response?.data?['message'] as String? ?? '인증번호 발송에 실패했습니다';
        _loading = false;
      });
    }
  }

  Future<void> _verify() async {
    final phone = _phoneController.text.trim();
    final code = _codeController.text.trim();
    if (code.length != 6) {
      setState(() => _error = '6자리 인증번호를 입력하세요');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final dio = ref.read(dioProvider('ko'));
      await dio.post('/phone/verify', data: {
        'phone_number': phone,
        'code': code,
      });
      setState(() {
        _verified = true;
        _loading = false;
      });
    } on DioException catch (e) {
      setState(() {
        _error = e.response?.data?['message'] as String? ?? '인증에 실패했습니다';
        _loading = false;
      });
    }
  }

  @override
  void dispose() {
    _phoneController.dispose();
    _codeController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(title: const Text('전화번호 인증')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: _verified ? _buildDone(colorScheme) : _buildForm(colorScheme),
      ),
    );
  }

  Widget _buildDone(ColorScheme colorScheme) {
    return Column(
      children: [
        const SizedBox(height: 48),
        Icon(Icons.check_circle, size: 80, color: colorScheme.primary),
        const SizedBox(height: 24),
        Text('인증 완료!', style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text(
          '이제 제한 없이 거래하실 수 있습니다',
          style: TextStyle(color: colorScheme.outline),
        ),
        const SizedBox(height: 48),
        FilledButton(
          onPressed: () => context.go('/'),
          style:
              FilledButton.styleFrom(minimumSize: const Size.fromHeight(52)),
          child: const Text('홈으로 이동'),
        ),
      ],
    );
  }

  Widget _buildForm(ColorScheme colorScheme) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('전화번호 인증',
            style: Theme.of(context).textTheme.headlineSmall),
        const SizedBox(height: 8),
        Text(
          '안전한 거래를 위해 전화번호 인증이 필요합니다',
          style: TextStyle(color: colorScheme.outline),
        ),
        const SizedBox(height: 32),
        TextFormField(
          controller: _phoneController,
          keyboardType: TextInputType.phone,
          inputFormatters: [FilteringTextInputFormatter.digitsOnly],
          decoration: const InputDecoration(
            labelText: '전화번호',
            hintText: '01012345678',
            prefixText: '+82 ',
          ),
          enabled: !_codeSent,
        ),
        if (!_codeSent) ...[
          const SizedBox(height: 16),
          FilledButton(
            onPressed: _loading ? null : _sendCode,
            style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(52)),
            child: _loading
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: Colors.white),
                  )
                : const Text('인증번호 받기'),
          ),
        ],
        if (_codeSent) ...[
          const SizedBox(height: 24),
          TextFormField(
            controller: _codeController,
            keyboardType: TextInputType.number,
            inputFormatters: [
              FilteringTextInputFormatter.digitsOnly,
              LengthLimitingTextInputFormatter(6),
            ],
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 24, letterSpacing: 8),
            decoration:
                const InputDecoration(labelText: '인증번호 6자리'),
          ),
          const SizedBox(height: 16),
          FilledButton(
            onPressed: _loading ? null : _verify,
            style: FilledButton.styleFrom(
                minimumSize: const Size.fromHeight(52)),
            child: _loading
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(
                        strokeWidth: 2, color: Colors.white),
                  )
                : const Text('인증 확인'),
          ),
          const SizedBox(height: 8),
          TextButton(
            onPressed: () => setState(() {
              _codeSent = false;
              _codeController.clear();
              _error = null;
            }),
            child: const Text('번호 다시 입력'),
          ),
        ],
        if (_error != null) ...[
          const SizedBox(height: 12),
          Text(
            _error!,
            style: TextStyle(color: colorScheme.error, fontSize: 13),
          ),
        ],
      ],
    );
  }
}
