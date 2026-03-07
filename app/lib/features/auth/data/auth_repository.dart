import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../../../core/api/api_client.dart';
import '../domain/auth_model.dart';

const _keyAccessToken = 'av_access_token';
const _keyRefreshToken = 'av_refresh_token';
const _keyUser = 'av_user';

class AuthRepository {
  final Dio _dio;
  final FlutterSecureStorage _storage;

  AuthRepository(this._dio, this._storage);

  Future<AuthTokens> login(String email, String password) async {
    final resp = await _dio.post(
      '/auth/login',
      data: {'email': email, 'password': password},
    );
    final tokens = AuthTokens.fromJson(resp.data as Map<String, dynamic>);
    await saveTokens(tokens);

    // Fetch user profile after login if provided inline
    if (resp.data['user'] != null) {
      final user = AuthUser.fromJson(resp.data['user'] as Map<String, dynamic>);
      await _storage.write(key: _keyUser, value: jsonEncode(user.toJson()));
    }

    return tokens;
  }

  Future<void> register(
    String email,
    String password,
    String nickname,
    String locale,
  ) async {
    await _dio.post(
      '/auth/register',
      data: {
        'email': email,
        'password': password,
        'nickname': nickname,
        'locale': locale,
      },
    );
  }

  Future<void> logout() async {
    try {
      await _dio.post('/auth/logout');
    } catch (_) {
      // Ignore network errors on logout – still clear local tokens
    } finally {
      await clearTokens();
    }
  }

  Future<AuthUser?> getStoredUser() async {
    final raw = await _storage.read(key: _keyUser);
    if (raw == null) return null;
    try {
      final json = jsonDecode(raw) as Map<String, dynamic>;
      return AuthUser.fromJson(json);
    } catch (_) {
      return null;
    }
  }

  Future<bool> hasValidToken() async {
    final token = await _storage.read(key: _keyAccessToken);
    return token != null && token.isNotEmpty;
  }

  Future<void> saveTokens(AuthTokens tokens) async {
    await Future.wait([
      _storage.write(key: _keyAccessToken, value: tokens.accessToken),
      _storage.write(key: _keyRefreshToken, value: tokens.refreshToken),
    ]);
  }

  Future<void> clearTokens() async {
    await Future.wait([
      _storage.delete(key: _keyAccessToken),
      _storage.delete(key: _keyRefreshToken),
      _storage.delete(key: _keyUser),
    ]);
  }
}

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  // Auth calls use a fixed locale; the interceptor in api_client handles tokens
  final dio = ref.watch(dioProvider('en-US'));
  const storage = FlutterSecureStorage();
  return AuthRepository(dio, storage);
});
