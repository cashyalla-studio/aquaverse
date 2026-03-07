import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

const _baseUrl = String.fromEnvironment('API_URL', defaultValue: 'http://localhost:8080');
const _storage = FlutterSecureStorage();

Dio createDio(String locale) {
  final dio = Dio(BaseOptions(
    baseUrl: '$_baseUrl/api/v1',
    connectTimeout: const Duration(seconds: 15),
    receiveTimeout: const Duration(seconds: 15),
    headers: {
      'Content-Type': 'application/json',
      'X-Locale': locale, // 게시판 로케일 분리 핵심
    },
  ));

  // 인증 인터셉터
  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) async {
      final token = await _storage.read(key: 'av_access_token');
      if (token != null) {
        options.headers['Authorization'] = 'Bearer $token';
      }
      return handler.next(options);
    },
    onError: (error, handler) async {
      if (error.response?.statusCode == 401) {
        // 토큰 갱신 시도
        final refreshToken = await _storage.read(key: 'av_refresh_token');
        if (refreshToken != null) {
          try {
            final response = await Dio().post(
              '$_baseUrl/api/v1/auth/refresh',
              data: {'refresh_token': refreshToken},
            );
            final newToken = response.data['access_token'];
            await _storage.write(key: 'av_access_token', value: newToken);
            // 원래 요청 재시도
            error.requestOptions.headers['Authorization'] = 'Bearer $newToken';
            final retryResponse = await dio.fetch(error.requestOptions);
            return handler.resolve(retryResponse);
          } catch (_) {
            await _storage.deleteAll();
          }
        }
      }
      return handler.next(error);
    },
  ));

  // 로깅 (개발 환경)
  // dio.interceptors.add(LogInterceptor(responseBody: true));

  return dio;
}

final dioProvider = Provider.family<Dio, String>((ref, locale) {
  return createDio(locale);
});
