// FCM 토큰 등록 서비스 (firebase_messaging 설치 후 사용)
// 현재는 스텁으로 구현 — P3 후반에 firebase_messaging 패키지 추가 시 구현

import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'api/api_client.dart';

class FCMService {
  // ref: Ref (StateNotifier) 또는 WidgetRef (UI) 모두 사용 가능
  static Future<void> registerToken(Ref ref, String locale) async {
    // TODO: firebase_messaging 패키지 추가 후 구현
    // final token = await FirebaseMessaging.instance.getToken();
    // if (token == null) return;
    // final dio = ref.read(dioProvider(locale));
    // await dio.post('/notifications/fcm/register', data: {
    //   'token': token,
    //   'platform': defaultTargetPlatform == TargetPlatform.iOS ? 'ios' : 'android',
    // });
    debugPrint('FCM: firebase_messaging 패키지 추가 후 활성화');
  }
}
